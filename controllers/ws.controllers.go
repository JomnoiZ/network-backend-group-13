package controllers

import (
	"log"
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type websocketController struct {
    websocketService services.WebsocketService
}

type WebsocketController interface {
    HandleWebSocket(c *gin.Context)
}

func NewWebsocketController(websocketService services.WebsocketService) WebsocketController {
    return &websocketController{
        websocketService: websocketService,
    }
}

func (c *websocketController) HandleWebSocket(ctx *gin.Context) {
    username := ctx.Query("username")
    if username == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing username"})
        return
    }

    var upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin: func(r *http.Request) bool { return true },
    }

    conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error for user %s: %v", username, err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
        return
    }

    c.websocketService.HandleConnection(username, conn)
}