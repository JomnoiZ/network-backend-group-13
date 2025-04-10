package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/routes"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
)

func main() {
	websocketService := services.NewWebsocketService()

	routes.WebsocketRoute(websocketService)

	// Serve frontend from /public
	http.Handle("/", http.FileServer(http.Dir("./public")))

	fmt.Println("Server running on PORT 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
