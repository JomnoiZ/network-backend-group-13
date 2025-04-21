package controllers

import (
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

type groupController struct {
	groupService services.GroupService
}

type GroupController interface {
	GetAllGroups(c *gin.Context)
	GetGroup(c *gin.Context)
	CreateGroup(c *gin.Context)
	AddMember(c *gin.Context)
	KickMember(c *gin.Context)
	AddAdmin(c *gin.Context)
	RemoveAdmin(c *gin.Context)
	GetGroupMessages(c *gin.Context)
}

func NewGroupController(groupService services.GroupService) GroupController {
	return &groupController{
		groupService: groupService,
	}
}

func (c *groupController) GetAllGroups(ctx *gin.Context) {
	groups, err := c.groupService.GetAllGroups()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, groups)
}

func (c *groupController) GetGroup(ctx *gin.Context) {
	groupID := ctx.Param("id")
	group, err := c.groupService.GetGroup(groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if group == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}
	ctx.JSON(http.StatusOK, group)
}

func (c *groupController) CreateGroup(ctx *gin.Context) {
	var groupDTO struct {
		Name  string `json:"name" binding:"required"`
		Owner string `json:"owner" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&groupDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	group, err := c.groupService.CreateGroup(groupDTO.Name, groupDTO.Owner)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, group)
}

func (c *groupController) AddMember(ctx *gin.Context) {
	groupID := ctx.Param("id")
	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	err := c.groupService.AddMember(groupID, req.Username)
	if err != nil {
		if err.Error() == "group not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Member added"})
}

func (c *groupController) KickMember(ctx *gin.Context) {
	groupID := ctx.Param("id")
	var req struct {
		Requester string `json:"requester" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	username := ctx.Param("username")
	err := c.groupService.KickMember(groupID, username, req.Requester)
	if err != nil {
		if err.Error() == "group not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Member removed"})
}

func (c *groupController) AddAdmin(ctx *gin.Context) {
	groupID := ctx.Param("id")
	var req struct {
		Username  string `json:"username" binding:"required"`
		Requester string `json:"requester" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	err := c.groupService.AddAdmin(groupID, req.Username, req.Requester)
	if err != nil {
		if err.Error() == "group not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" || err.Error() == "user not a member" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Admin added"})
}

func (c *groupController) RemoveAdmin(ctx *gin.Context) {
	groupID := ctx.Param("id")
	var req struct {
		Requester string `json:"requester" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	username := ctx.Param("username")
	err := c.groupService.RemoveAdmin(groupID, username, req.Requester)
	if err != nil {
		if err.Error() == "group not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" || err.Error() == "user not an admin" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Admin removed"})
}

func (c *groupController) GetGroupMessages(ctx *gin.Context) {
	groupID := ctx.Param("id")
	messages, err := c.groupService.GetGroupMessages(groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, messages)
}
