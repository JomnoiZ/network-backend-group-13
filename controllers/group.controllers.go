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
        Name    string `json:"name" binding:"required"`
        OwnerID string `json:"owner_id" binding:"required"`
    }
    if err := ctx.ShouldBindJSON(&groupDTO); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    group, err := c.groupService.CreateGroup(groupDTO.Name, groupDTO.OwnerID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusCreated, group)
}

func (c *groupController) AddMember(ctx *gin.Context) {
    groupID := ctx.Param("id")
    var req struct {
        UserID      string `json:"user_id" binding:"required"`
        RequesterID string `json:"requester_id" binding:"required"`
    }
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    err := c.groupService.AddMember(groupID, req.UserID, req.RequesterID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"message": "Member added"})
}

func (c *groupController) KickMember(ctx *gin.Context) {
    groupID := ctx.Param("id")
    userID := ctx.Param("user_id")
    requesterID := ctx.Query("requester_id")
    if requesterID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Requester ID required"})
        return
    }
    err := c.groupService.KickMember(groupID, userID, requesterID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"message": "Member removed"})
}

func (c *groupController) AddAdmin(ctx *gin.Context) {
    groupID := ctx.Param("id")
    userID := ctx.Param("user_id")
    requesterID := ctx.Query("requester_id")
    if requesterID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Requester ID required"})
        return
    }
    err := c.groupService.AddAdmin(groupID, userID, requesterID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"message": "Admin added"})
}

func (c *groupController) RemoveAdmin(ctx *gin.Context) {
    groupID := ctx.Param("id")
    userID := ctx.Param("user_id")
    requesterID := ctx.Query("requester_id")
    if requesterID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Requester ID required"})
        return
    }
    err := c.groupService.RemoveAdmin(groupID, userID, requesterID)
    if err != nil {
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