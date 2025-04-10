package controllers

import (
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

type groupController struct {
	groupService services.GroupService
}

type GroupController interface {
	GetGroup(c *gin.Context)
	CreateGroup(c *gin.Context)
}

func NewGroupController(groupService services.GroupService) GroupController {
	return &groupController{
		groupService: groupService,
	}
}

func (c *groupController) GetGroup(ctx *gin.Context) {
	groupID := ctx.Param("id")

	group, err := c.groupService.GetGroup(groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, group)
}

func (c *groupController) CreateGroup(ctx *gin.Context) {
	var groupDTO *models.Group
	if err := ctx.ShouldBindJSON(&groupDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	group, err := c.groupService.CreateGroup(groupDTO)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, group)
}
