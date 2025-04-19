package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 10000
	sendTimeout    = 3 * time.Second 
)

type WebsocketService interface {
	HandleConnection(username string, conn *websocket.Conn)
	GetClients() map[string]*models.Client
	AddToGroup(client *models.Client, groupID string)
	KickFromGroup(username string, groupID string)
	NotifyGroupUpdate(groupID string, updateType string, data interface{})
}

type websocketService struct {
	clients     map[string]*models.Client
	groups      map[string]map[string]*models.Client
	mutex       sync.RWMutex
	messageRepo database.MessageRepository
}

func NewWebsocketService(messageRepo database.MessageRepository) WebsocketService {
	return &websocketService{
		clients:     make(map[string]*models.Client),
		groups:      make(map[string]map[string]*models.Client),
		messageRepo: messageRepo,
	}
}

func (s *websocketService) HandleConnection(username string, conn *websocket.Conn) {
	client := &models.Client{
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
	}

	s.mutex.Lock()
	if oldClient, exists := s.clients[username]; exists {
		// Signal old client to close
		select {
		case oldClient.Send <- []byte{}: // Send empty message to unblock writePump
		default:
		}
		close(oldClient.Send)
		oldClient.Conn.Close()
	}
	s.clients[username] = client
	s.mutex.Unlock()

	go s.writePump(client)
	go s.readPump(client)

	s.broadcastStatus(username, "online")
}

func (s *websocketService) GetClients() map[string]*models.Client {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	clients := make(map[string]*models.Client, len(s.clients))
	for k, v := range s.clients {
		clients[k] = v
	}
	return clients
}

func (s *websocketService) AddToGroup(client *models.Client, groupID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Only add if client exists in s.clients
	if actualClient, exists := s.clients[client.Username]; exists {
		if s.groups[groupID] == nil {
			s.groups[groupID] = make(map[string]*models.Client)
		}
		s.groups[groupID][client.Username] = actualClient
		actualClient.Groups[groupID] = true
	}
}

func (s *websocketService) KickFromGroup(username string, groupID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if groupClients, exists := s.groups[groupID]; exists {
		delete(groupClients, username)
	}

	if client, exists := s.clients[username]; exists {
		delete(client.Groups, groupID)
		// Notify the kicked client
		message := models.Message{
			Type:    "group_update",
			GroupID: groupID,
			Data: map[string]interface{}{
				"type": "kick",
				"data": map[string]string{"username": username},
			},
		}
		if messageJSON, err := json.Marshal(message); err == nil {
			s.sendMessage(client, messageJSON)
		}
	}
}

func (s *websocketService) NotifyGroupUpdate(groupID string, updateType string, data interface{}) {
	message := models.Message{
		Type:    "group_update",
		GroupID: groupID,
		Data: map[string]interface{}{
			"type": updateType,
			"data": data,
		},
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Failed to marshal group update:", err)
		return
	}

	s.mutex.RLock()
	groupClients, exists := s.groups[groupID]
	s.mutex.RUnlock()

	if exists {
		for _, client := range groupClients {
			s.sendMessage(client, messageJSON)
		}
	}
}

func (s *websocketService) broadcastStatus(username string, status string) {
	message := models.Message{
		Type:   "status",
		Sender: username,
		Status: status,
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Failed to marshal status message:", err)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, client := range s.clients {
		if client.Username != username {
			s.sendMessage(client, messageJSON)
		}
	}
}

func (s *websocketService) readPump(client *models.Client) {
	defer func() {
		s.mutex.Lock()
		delete(s.clients, client.Username)
		for groupID := range client.Groups {
			if groupClients, exists := s.groups[groupID]; exists {
				delete(groupClients, client.Username)
			}
		}
		s.mutex.Unlock()
		close(client.Send)
		client.Conn.Close()
		s.broadcastStatus(client.Username, "offline")
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error for user %s: %v", client.Username, err)
			}
			return
		}

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}

		msg.Sender = client.Username

		switch msg.Type {
		case "message":
			s.handleChatMessage(client, &msg)
		case "typing":
			s.handleTypingStatus(client, &msg)
		case "join_group":
			if msg.GroupID != "" {
				s.AddToGroup(client, msg.GroupID)
				log.Printf("User %s joined group %s", client.Username, msg.GroupID)
			}
		}
	}
}

func (s *websocketService) writePump(client *models.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Skip empty messages (used for signaling closure)
			if len(message) == 0 {
				continue
			}

			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error for user %s: %v", client.Username, err)
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error for user %s: %v", client.Username, err)
				return
			}
		}
	}
}

func (s *websocketService) sendMessage(client *models.Client, message []byte) {
	if client == nil || client.Send == nil {
		log.Printf("Cannot send message: nil client or send channel")
		return
	}

	s.mutex.RLock()
	_, exists := s.clients[client.Username]
	s.mutex.RUnlock()

	if !exists {
		log.Printf("Cannot send message: client %s not found", client.Username)
		return
	}

	select {
	case client.Send <- message:
	case <-time.After(sendTimeout):
		log.Printf("Timeout sending to client %s, marking for cleanup", client.Username)
		// Do not close Send here; let readPump/writePump handle cleanup
		s.mutex.Lock()
		delete(s.clients, client.Username)
		for groupID := range client.Groups {
			if groupClients, exists := s.groups[groupID]; exists {
				delete(groupClients, client.Username)
			}
		}
		s.mutex.Unlock()
	}
}

func (s *websocketService) handleChatMessage(client *models.Client, msg *models.Message) {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// Validate message
	if msg.Content == "" || (msg.GroupID == "" && msg.Receiver == "") {
		log.Printf("Invalid message from %s: empty content or no recipient", client.Username)
		return
	}

	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal message:", err)
		return
	}

	// Save to database before sending
	if msg.GroupID != "" || msg.Receiver != "" {
		dbMsg := &models.MessageDB{
			ID:        msg.ID,
			Sender:    msg.Sender,
			Receiver:  msg.Receiver,
			GroupID:   msg.GroupID,
			Content:   msg.Content,
			Timestamp: time.Now(),
		}
		if err := s.messageRepo.SaveMessage(dbMsg); err != nil {
			log.Println("Failed to save message to database:", err)
			// Continue to send message even if DB save fails
		}
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if msg.GroupID != "" {
		if groupClients, exists := s.groups[msg.GroupID]; exists {
			for _, c := range groupClients {
				s.sendMessage(c, messageJSON)
			}
		}
		return
	}

	if msg.Receiver != "" {
		if receiver, exists := s.clients[msg.Receiver]; exists && receiver.Username != msg.Sender {
			s.sendMessage(receiver, messageJSON)
		}
		// Echo back to sender
		s.sendMessage(client, messageJSON)
	}
}

func (s *websocketService) handleTypingStatus(client *models.Client, msg *models.Message) {
	// Skip typing events for groups (not supported by frontend)
	if msg.GroupID != "" {
		return
	}

	if msg.Receiver == "" || msg.Receiver == msg.Sender {
		return
	}

	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal typing status:", err)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if client, exists := s.clients[msg.Receiver]; exists {
		s.sendMessage(client, messageJSON)
	}
}