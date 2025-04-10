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

type websocketService struct {
	messageRepository database.MessageRepository
	Clients           map[string]map[*models.Client]bool
	Groups            map[string]map[string]bool
	Mutex             sync.RWMutex
}

type WebsocketService interface {
	HandleConnection(userID string, conn *websocket.Conn)
	RegisterClient(c *models.Client)
	UnregisterClient(c *models.Client)
	BroadcastStatus(userID string, status string)
	HandleMessage(sender *models.Client, msg *models.Message)
}

func NewWebsocketService() WebsocketService {
	return &websocketService{
		Clients: make(map[string]map[*models.Client]bool),
		Groups:  make(map[string]map[string]bool),
	}
}

func (s *websocketService) HandleConnection(userID string, conn *websocket.Conn) {
	client := &models.Client{
		ID:     userID,
		Conn:   conn,
		Send:   make(chan []byte, 10),
		Groups: make(map[string]bool),
	}

	s.RegisterClient(client)

	go s.readMessages(client)
	go s.writeMessages(client)

	s.BroadcastStatus(userID, "online")
}

func (s *websocketService) RegisterClient(client *models.Client) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Clients[client.ID] == nil {
		s.Clients[client.ID] = make(map[*models.Client]bool)
	}
	s.Clients[client.ID][client] = true

	log.Printf("Client %s connected (%d connections)", client.ID, len(s.Clients[client.ID]))
}

func (s *websocketService) UnregisterClient(client *models.Client) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	conns, ok := s.Clients[client.ID]
	if !ok {
		return
	}

	delete(conns, client)
	client.Conn.Close()
	close(client.Send)

	if len(conns) == 0 {
		delete(s.Clients, client.ID)
		log.Printf("Client %s disconnected (last connection)", client.ID)
	} else {
		log.Printf("Client %s disconnected (remaining %d connections)", client.ID, len(conns))
	}

	// Remove from groups
	for gid := range client.Groups {
		delete(s.Groups[gid], client.ID)
	}
}

func (s *websocketService) BroadcastStatus(userID string, status string) {
	msg := models.Message{
		Type:     "status",
		SenderID: userID,
		Status:   status,
	}
	data, _ := json.Marshal(msg)

	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	for id, conns := range s.Clients {
		if id == userID {
			continue
		}

		for client := range conns {
			select {
			case client.Send <- data:
			default:
				log.Printf("Warning: Client %s send channel full or closed", client.ID)
			}
		}
	}
}

func (s *websocketService) HandleMessage(sender *models.Client, msg *models.Message) {
	switch msg.Type {
	case "message":
		s.sendMessage(sender.ID, msg)
	case "typing":
		s.sendTyping(sender.ID, msg)
	case "join_group":
		s.addToGroup(sender, msg.GroupID)
	}
}

func (s *websocketService) sendMessage(senderID string, msg *models.Message) {
	data, _ := json.Marshal(msg)

	if msg.GroupID != "" {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()

		for memberID := range s.Groups[msg.GroupID] {
			if conns, ok := s.Clients[memberID]; ok {
				for client := range conns {
					if memberID != senderID || msg.Type == "typing" {
						select {
						case client.Send <- data:
						default:
							log.Printf("Warning: Send buffer full for group member %s", memberID)
						}
					}
				}
			}
		}
	} else if msg.ReceiverID != "" {
		s.sendToUser(msg.ReceiverID, data)
	}

	// go s.messageRepository.SaveMessage(&models.MessageDB{
	// 	SenderID:   msg.SenderID,
	// 	ReceiverID: msg.ReceiverID,
	// 	GroupID:    msg.GroupID,
	// 	Content:    msg.Content,
	// })
}

func (s *websocketService) sendTyping(senderID string, msg *models.Message) {
	data, _ := json.Marshal(msg)

	if msg.GroupID != "" {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()

		for memberID := range s.Groups[msg.GroupID] {
			if memberID == senderID {
				continue
			}
			for client := range s.Clients[memberID] {
				select {
				case client.Send <- data:
				default:
					log.Printf("Typing: buffer full for %s", memberID)
				}
			}
		}
	} else if msg.ReceiverID != "" {
		s.sendToUser(msg.ReceiverID, data)
	}
}

func (s *websocketService) sendToUser(userID string, data []byte) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	for client := range s.Clients[userID] {
		select {
		case client.Send <- data:
		default:
			log.Printf("Send buffer full or closed for user %s", userID)
		}
	}
}

func (s *websocketService) addToGroup(client *models.Client, groupID string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Groups[groupID] == nil {
		s.Groups[groupID] = make(map[string]bool)
	}

	s.Groups[groupID][client.ID] = true
	client.Groups[groupID] = true
}

func (s *websocketService) readMessages(client *models.Client) {
	defer func() {
		s.BroadcastStatus(client.ID, "offline")
		s.UnregisterClient(client)
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msg, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Read error (%s): %v", client.ID, err)
			break
		}

		var message models.Message
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Println("Invalid JSON:", err)
			continue
		}

		s.HandleMessage(client, &message)
	}
}

func (s *websocketService) writeMessages(client *models.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-client.Send:
			if !ok {
				return
			}
			err := client.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("Write error (%s): %v", client.ID, err)
				return
			}
		case <-ticker.C:
			err := client.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Printf("Ping error (%s): %v", client.ID, err)
				return
			}
		}
	}
}
