package controllers

import (
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

type userController struct {
    userService services.UserService
}

type UserController interface {
    GetUser(c *gin.Context)
    CreateUser(c *gin.Context)
    ListOnlineUsers(c *gin.Context)
    ListUserGroups(c *gin.Context)
    GetDirectMessages(c *gin.Context)
}

func NewUserController(userService services.UserService) UserController {
    return &userController{
        userService: userService,
    }
}

func (c *userController) GetUser(ctx *gin.Context) {
    userID := ctx.Param("id")
    user, err := c.userService.GetUser(userID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if user == nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    ctx.JSON(http.StatusOK, user)
}

func (c *userController) CreateUser(ctx *gin.Context) {
    var userDTO models.User
    if err := ctx.ShouldBindJSON(&userDTO); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    user, err := c.userService.CreateUser(&userDTO)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusCreated, user)
}

func (c *userController) ListOnlineUsers(ctx *gin.Context) {
    users, err := c.userService.ListOnlineUsers()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, users)
}

func (c *userController) ListUserGroups(ctx *gin.Context) {
    userID := ctx.Param("id")
    groups, err := c.userService.ListUserGroups(userID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, groups)
}

func (c *userController) GetDirectMessages(ctx *gin.Context) {
    userID := ctx.Param("id")
    targetID := ctx.Param("target_id")
    messages, err := c.userService.GetDirectMessages(userID, targetID)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, messages)
}