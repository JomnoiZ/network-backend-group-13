package routes

import (
	"github.com/JomnoiZ/network-backend-group-13.git/controllers"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

func WebsocketRoute(websocketService services.WebsocketService, r *gin.Engine) {
	websocketController := controllers.NewWebsocketController(websocketService)
	r.GET("/ws", websocketController.HandleWebSocket)
}