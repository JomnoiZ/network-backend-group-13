package routes

import (
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/controllers"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
)

func WebsocketRoute(websockerService services.WebsocketService) {
	websocketController := controllers.NewWebsocketController(websockerService)

	r := http.NewServeMux()
	r.HandleFunc("/ws", websocketController.HandleWebSocket)

	http.Handle("/ws", r)
}
