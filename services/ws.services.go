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
	// Validate inputs
	if username == "" || conn == nil {
		log.Printf("Invalid HandleConnection parameters: username=%s, conn=%v", username, conn)
		if conn != nil {
			conn.Close()
		}
		return
	}

	client := &models.Client{
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Groups:   make(map[string]bool),
	}

	s.mutex.Lock()
	if oldClient, exists := s.clients[username]; exists {
		// Signal old client to close by sending a close message
		if oldClient.Conn != nil {
			oldClient.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			oldClient.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}
	}
	s.clients[username] = client
	s.mutex.Unlock()

	go s.writePump(client)
	go s.readPump(client)

	s.broadcastStatus(username, "online")
	log.Printf("User %s connected via WebSocket", username)
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
	if client == nil || client.Username == "" {
		log.Printf("Cannot add nil or invalid client to group %s", groupID)
		return
	}

	s.mutex.Lock()
	// Only add if client exists in s.clients
	if actualClient, exists := s.clients[client.Username]; exists {
		if s.groups[groupID] == nil {
			s.groups[groupID] = make(map[string]*models.Client)
		}
		s.groups[groupID][client.Username] = actualClient
		actualClient.Groups[groupID] = true
		log.Printf("Added user %s to group %s", client.Username, groupID)
	} else {
		log.Printf("Client %s not found in clients map for group %s", client.Username, groupID)
		s.mutex.Unlock()
		return
	}
	s.mutex.Unlock()

	// Notify all group members of the new member
	s.NotifyGroupUpdate(groupID, "add", map[string]string{"username": client.Username})
}

func (s *websocketService) KickFromGroup(username string, groupID string) {
	s.mutex.Lock()
	if groupClients, exists := s.groups[groupID]; exists {
		delete(groupClients, username)
		if len(groupClients) == 0 {
			delete(s.groups, groupID) // Clean up empty groups
		}
	}

	var kickedClient *models.Client
	if client, exists := s.clients[username]; exists {
		delete(client.Groups, groupID)
		kickedClient = client
	}
	s.mutex.Unlock()

	// Notify all group members of the kick
	s.NotifyGroupUpdate(groupID, "kick", map[string]string{"username": username})

	// Notify the kicked client
	if kickedClient != nil {
		message := models.Message{
			Type:    "group_update",
			GroupID: groupID,
			Data: map[string]interface{}{
				"type": "kick",
				"data": map[string]string{"username": username},
			},
		}
		if messageJSON, err := json.Marshal(message); err == nil {
			s.sendMessage(kickedClient, messageJSON)
		} else {
			log.Printf("Failed to marshal kick message for %s: %v", username, err)
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
		log.Printf("Failed to marshal group update for group %s, type %s: %v", groupID, updateType, err)
		return
	}

	s.mutex.RLock()
	groupClients, exists := s.groups[groupID]
	s.mutex.RUnlock()

	if exists {
		for _, client := range groupClients {
			s.sendMessage(client, messageJSON)
		}
		log.Printf("Notified group %s of update type %s", groupID, updateType)
	} else {
		log.Printf("No clients found for group %s to notify update type %s", groupID, updateType)
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
		log.Printf("Failed to marshal status message for %s: %v", username, err)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, client := range s.clients {
		if client.Username != username {
			s.sendMessage(client, messageJSON)
		}
	}
	log.Printf("Broadcasted status %s for user %s", status, username)
}

func (s *websocketService) readPump(client *models.Client) {
	// Validate client
	if client == nil || client.Conn == nil || client.Username == "" || client.Send == nil || client.Groups == nil {
		log.Printf("Invalid client state in readPump: %+v", client)
		return
	}

	defer func() {
		// Recover from any panic to prevent server crash
		if r := recover(); r != nil {
			log.Printf("Recovered panic in readPump for user %s: %v", client.Username, r)
		}

		s.mutex.Lock()
		// Only delete if client is still in s.clients
		if actualClient, exists := s.clients[client.Username]; exists && actualClient == client {
			delete(s.clients, client.Username)
			for groupID := range client.Groups {
				if groupClients, exists := s.groups[groupID]; exists {
					delete(groupClients, client.Username)
					if len(groupClients) == 0 {
						delete(s.groups, groupID)
					}
				}
			}
		}
		s.mutex.Unlock()

		// Do not close client.Send here; let writePump handle it
		if client.Conn != nil {
			client.Conn.Close()
		}
		if client.Username != "" {
			s.broadcastStatus(client.Username, "offline")
		}
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
			log.Printf("Failed to unmarshal message for user %s: %v", client.Username, err)
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
	// Validate client
	if client == nil || client.Conn == nil || client.Username == "" || client.Send == nil {
		log.Printf("Invalid client state in writePump: %+v", client)
		return
	}

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if client.Conn != nil {
			client.Conn.Close()
		}
		// Close Send channel only in writePump
		if client.Send != nil {
			close(client.Send)
		}
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			// Skip empty messages (used for signaling)
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
	if client == nil || client.Send == nil || client.Conn == nil {
		log.Printf("Cannot send message: nil client or invalid state")
		return
	}

	s.mutex.RLock()
	actualClient, exists := s.clients[client.Username]
	s.mutex.RUnlock()

	if !exists || actualClient != client {
		log.Printf("Cannot send message: client %s not found or replaced", client.Username)
		return
	}

	select {
	case client.Send <- message:
	case <-time.After(sendTimeout):
		log.Printf("Timeout sending to client %s", client.Username)
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
		log.Printf("Failed to marshal message from %s: %v", client.Username, err)
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
			log.Printf("Failed to save message to database: %v", err)
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
	if msg.GroupID != "" {
		return
	}

	if msg.Receiver == "" || msg.Receiver == msg.Sender {
		return
	}

	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal typing status for %s: %v", client.Username, err)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if client, exists := s.clients[msg.Receiver]; exists {
		s.sendMessage(client, messageJSON)
	}
}