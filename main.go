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

	cors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Apply CORS to WebSocket route
	routes.WebsocketRoute(websocketService, cors)

	// Serve static files with CORS
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", cors(fs))

	fmt.Println("Server running on PORT 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
