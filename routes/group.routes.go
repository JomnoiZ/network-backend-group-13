package routes

import (
	"github.com/JomnoiZ/network-backend-group-13.git/controllers"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

func GroupRoute(r *gin.Engine, groupService services.GroupService) {
    groupController := controllers.NewGroupController(groupService)

    rgu := r.Group("/groups")
    {
        rgu.GET("/:id", groupController.GetGroup)
        rgu.POST("/", groupController.CreateGroup)
        rgu.POST("/:id/members", groupController.AddMember)
        rgu.DELETE("/:id/members/:user_id", groupController.KickMember)
        rgu.PUT("/:id/admins/:user_id", groupController.AddAdmin)
        rgu.DELETE("/:id/admins/:user_id", groupController.RemoveAdmin)
        rgu.GET("/:id/messages", groupController.GetGroupMessages)
    }
}