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
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (c *userController) CreateUser(ctx *gin.Context) {
	var userDTO *models.User
	if err := ctx.ShouldBindJSON(&userDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	user, err := c.userService.CreateUser(userDTO)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}
