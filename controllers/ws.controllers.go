package controllers

import (
	"log"
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gorilla/websocket"
)

type websocketController struct {
	websocketService services.WebsocketService
}

type WebsocketController interface {
	HandleWebSocket(w http.ResponseWriter, r *http.Request)
}

func NewWebsocketController(websocketService services.WebsocketService) WebsocketController {
	return &websocketController{
		websocketService: websocketService,
	}
}

func (c *websocketController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	c.websocketService.HandleConnection(userID, conn)
}
