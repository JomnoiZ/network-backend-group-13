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
        rgu.DELETE("/:id/members/:username", groupController.KickMember)
        rgu.PUT("/:id/admins/:username", groupController.AddAdmin)
        rgu.DELETE("/:id/admins/:username", groupController.RemoveAdmin)
        rgu.GET("/:id/messages", groupController.GetGroupMessages)
    }
}