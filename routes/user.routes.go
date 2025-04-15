package routes

import (
	"github.com/JomnoiZ/network-backend-group-13.git/controllers"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

func UserRoute(r *gin.Engine, userService services.UserService) {
    userController := controllers.NewUserController(userService)

    rgu := r.Group("/users")
    {
        rgu.GET("/:id", userController.GetUser)
        rgu.POST("/", userController.CreateUser)
        rgu.GET("/online", userController.ListOnlineUsers)
        rgu.GET("/:id/groups", userController.ListUserGroups)
        rgu.GET("/:user_id/messages/:target_id", userController.GetDirectMessages)
    }
}