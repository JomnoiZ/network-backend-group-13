package database

import "github.com/JomnoiZ/network-backend-group-13.git/models"

type UserRepository interface {
    GetUser(userID string) (*models.User, error)
    GetAllUser() (*models.User, error)
    CreateUser(user *models.User) (*models.User, error)
    GetUserGroups(userID string) ([]*models.Group, error)
}

type GroupRepository interface {
    GetGroup(groupID string) (*models.Group, error)
    CreateGroup(group *models.Group) (*models.Group, error)
    UpdateGroup(group *models.Group) error
}

type MessageRepository interface {
    SaveMessage(message *models.MessageDB) error
    GetGroupMessages(groupID string) ([]*models.MessageDB, error)
    GetDirectMessages(userID, targetID string) ([]*models.MessageDB, error)
}