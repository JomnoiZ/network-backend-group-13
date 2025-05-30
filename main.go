package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/configs"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
	"github.com/JomnoiZ/network-backend-group-13.git/routes"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize MongoDB
	mongoClient := configs.NewMongoDBClient()
	userRepo := database.NewMongoUserRepository(mongoClient)
	groupRepo := database.NewMongoGroupRepository(mongoClient)
	messageRepo := database.NewMongoMessageRepository(mongoClient)

	// Initialize services
	websocketService := services.NewWebsocketService(messageRepo)
	userService := services.NewUserService(userRepo, messageRepo, websocketService)
	groupService := services.NewGroupService(groupRepo, userRepo, messageRepo, websocketService)

	// Set up Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// Set up routes
	routes.WebsocketRoute(websocketService, r)
	routes.UserRoute(r, userService, websocketService)
	routes.GroupRoute(r, groupService)

	// Serve static files under /static/
	r.Static("/static", "./public")

	// Optional: Serve index.html for root to support frontend
	r.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	fmt.Println("Server running on PORT 8080")
	log.Fatal(r.Run(":8080"))
}
