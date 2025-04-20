package controllers

import (
	"net/http"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/services"
	"github.com/gin-gonic/gin"
)

type userController struct {
	userService      services.UserService
	websocketService services.WebsocketService
}

type UserController interface {
	GetUser(c *gin.Context)
	GetAllUsers(c *gin.Context)
	CreateUser(c *gin.Context)
	ListOnlineUsers(c *gin.Context)
	ListUserGroups(c *gin.Context)
	GetDirectMessages(c *gin.Context)
}

func NewUserController(userService services.UserService, websocketService services.WebsocketService) UserController {
	return &userController{
		userService:      userService,
		websocketService: websocketService,
	}
}

func (c *userController) GetUser(ctx *gin.Context) {
	username := ctx.Param("username")
	user, err := c.userService.GetUser(username)
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

func (c *userController) GetAllUsers(ctx *gin.Context) {
	users, err := c.userService.GetAllUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(users) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No users found"})
		return
	}
	ctx.JSON(http.StatusOK, users)
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

	c.websocketService.BroadcastStatus(user.Username, "offline")
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
	username := ctx.Param("username")
	groups, err := c.userService.ListUserGroups(username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, groups)
}

func (c *userController) GetDirectMessages(ctx *gin.Context) {
	sender := ctx.Param("username")
	receiver := ctx.Param("receiver")
	messages, err := c.userService.GetDirectMessages(sender, receiver)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, messages)
}
