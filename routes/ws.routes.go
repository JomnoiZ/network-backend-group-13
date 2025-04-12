package routes

import (
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/controllers"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
)

func WebsocketRoute(websocketService services.WebsocketService, corsMiddleware func(http.Handler) http.Handler) {
	websocketController := controllers.NewWebsocketController(websocketService)

	http.Handle("/ws", corsMiddleware(http.HandlerFunc(websocketController.HandleWebSocket)))
}
