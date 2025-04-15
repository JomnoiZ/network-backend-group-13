package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 10000
)

type WebsocketService interface {
	HandleConnection(userID string, conn *websocket.Conn)
	GetClients() map[string]*models.Client
	AddToGroup(client *models.Client, groupID string)
	KickFromGroup(userID string, groupID string)
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

func (s *websocketService) HandleConnection(userID string, conn *websocket.Conn) {
	client := &models.Client{
		ID:     userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Groups: make(map[string]bool),
	}

	s.mutex.Lock()
	s.clients[userID] = client
	s.mutex.Unlock()

	go s.writePump(client)
	go s.readPump(client)

	s.broadcastStatus(userID, "online")
}

func (s *websocketService) GetClients() map[string]*models.Client {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.clients
}

func (s *websocketService) AddToGroup(client *models.Client, groupID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.groups[groupID] == nil {
		s.groups[groupID] = make(map[string]*models.Client)
	}

	if actualClient, exists := s.clients[client.ID]; exists {
		s.groups[groupID][client.ID] = actualClient
		actualClient.Groups[groupID] = true
	} else {
		s.groups[groupID][client.ID] = client
		client.Groups[groupID] = true
	}
}

func (s *websocketService) KickFromGroup(userID string, groupID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.groups[groupID] != nil {
		delete(s.groups[groupID], userID)
	}

	if client, exists := s.clients[userID]; exists {
		delete(client.Groups, groupID)
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
	defer s.mutex.RUnlock()

	if groupClients, exists := s.groups[groupID]; exists {
		for _, client := range groupClients {
			select {
			case client.Send <- messageJSON:
			default:
				close(client.Send)
				delete(s.clients, client.ID)
				for gid := range s.groups {
					delete(s.groups[gid], client.ID)
				}
			}
		}
	}
}

func (s *websocketService) broadcastStatus(userID string, status string) {
	message := models.Message{
		Type:     "status",
		SenderID: userID,
		Status:   status,
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Failed to marshal status message:", err)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, client := range s.clients {
		if client.ID != userID {
			select {
			case client.Send <- messageJSON:
			default:
				close(client.Send)
				delete(s.clients, client.ID)
				for groupID := range s.groups {
					delete(s.groups[groupID], client.ID)
				}
			}
		}
	}
}

func (s *websocketService) readPump(client *models.Client) {
	defer func() {
		s.mutex.Lock()
		if _, ok := s.clients[client.ID]; ok {
			delete(s.clients, client.ID)
			for groupID := range s.groups {
				delete(s.groups[groupID], client.ID)
			}
		}
		s.mutex.Unlock()
		client.Conn.Close()
		s.broadcastStatus(client.ID, "offline")
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
				log.Printf("Read error for user %s: %v", client.ID, err)
			}
			break
		}

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}

		msg.SenderID = client.ID

		switch msg.Type {
		case "message":
			s.handleChatMessage(&msg)
		case "typing":
			s.handleTypingStatus(&msg)
		case "join_group":
			if msg.GroupID != "" {
				s.AddToGroup(client, msg.GroupID)
				log.Printf("User %s joined group %s", client.ID, msg.GroupID)
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
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *websocketService) handleChatMessage(msg *models.Message) {
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal message:", err)
		return
	}

	if msg.GroupID != "" || msg.ReceiverID != "" {
		dbMsg := &models.MessageDB{
			ID:         msg.ID,
			SenderID:   msg.SenderID,
			ReceiverID: msg.ReceiverID,
			GroupID:    msg.GroupID,
			Content:    msg.Content,
		}
		if err := s.messageRepo.SaveMessage(dbMsg); err != nil {
			log.Println("Failed to save message to database:", err)
		}
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if msg.GroupID != "" {
		if groupClients, exists := s.groups[msg.GroupID]; exists {
			for _, client := range groupClients {
				select {
				case client.Send <- messageJSON:
				default:
					close(client.Send)
					delete(s.clients, client.ID)
					for gid := range s.groups {
						delete(s.groups[gid], client.ID)
					}
				}
			}
		}
		return
	}

	if msg.ReceiverID != "" {
		if client, exists := s.clients[msg.ReceiverID]; exists {
			select {
			case client.Send <- messageJSON:
			default:
				close(client.Send)
				delete(s.clients, client.ID)
				for gid := range s.groups {
					delete(s.groups[gid], client.ID)
				}
			}
		}

		if sender, exists := s.clients[msg.SenderID]; exists && sender.ID != msg.ReceiverID {
			select {
			case sender.Send <- messageJSON:
			default:
				close(sender.Send)
				delete(s.clients, sender.ID)
				for gid := range s.groups {
					delete(s.groups[gid], sender.ID)
				}
			}
		}
	}
}

func (s *websocketService) handleTypingStatus(msg *models.Message) {
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal typing status:", err)
		return
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if msg.GroupID != "" {
		if groupClients, exists := s.groups[msg.GroupID]; exists {
			for _, client := range groupClients {
				if client.ID != msg.SenderID {
					select {
					case client.Send <- messageJSON:
					default:
						close(client.Send)
						delete(s.clients, client.ID)
						for gid := range s.groups {
							delete(s.groups[gid], client.ID)
						}
					}
				}
			}
		}
		return
	}

	if msg.ReceiverID != "" && msg.ReceiverID != msg.SenderID {
		if client, exists := s.clients[msg.ReceiverID]; exists {
			select {
			case client.Send <- messageJSON:
			default:
				close(client.Send)
				delete(s.clients, client.ID)
				for gid := range s.groups {
					delete(s.groups[gid], client.ID)
				}
			}
		}
	}
}