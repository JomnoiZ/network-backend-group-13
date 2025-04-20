package routes

import (
	"github.com/JomnoiZ/network-backend-group-13.git/controllers"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

func UserRoute(r *gin.Engine, userService services.UserService, websocketService services.WebsocketService) {
	userController := controllers.NewUserController(userService, websocketService)

	rgu := r.Group("/users")
	{
		rgu.GET("/:username", userController.GetUser)
		rgu.GET("/", userController.GetAllUsers)
		rgu.POST("/", userController.CreateUser)
		rgu.GET("/online", userController.ListOnlineUsers)
		rgu.GET("/:username/groups", userController.ListUserGroups)
		rgu.GET("/:username/messages/:receiver", userController.GetDirectMessages)
	}
}
