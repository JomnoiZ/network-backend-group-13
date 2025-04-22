package database

import "github.com/JomnoiZ/network-backend-group-13.git/models"

type UserRepository interface {
	GetUser(username string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	GetUserGroups(username string) ([]*models.Group, error)
}

type GroupRepository interface {
	GetAllGroups() ([]*models.Group, error)
	GetGroup(groupID string) (*models.Group, error)
	CreateGroup(group *models.Group) (*models.Group, error)
	UpdateGroup(group *models.Group) error
}

type MessageRepository interface {
	SaveMessage(message *models.MessageDB) error
	GetGroupMessages(groupID string) ([]*models.MessageDB, error)
	GetDirectMessages(sender, receiver string) ([]*models.MessageDB, error)
}
