package services

import (
	"errors"
	"sync"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
)

type userService struct {
    userRepository   database.UserRepository
    websocketService WebsocketService
    mutex            sync.RWMutex
}

type UserService interface {
    GetUser(userID string) (*models.User, error)
    CreateUser(user *models.User) (*models.User, error)
    ListOnlineUsers() ([]*models.User, error)
    ListUserGroups(userID string) ([]*models.Group, error)
}

func NewUserService(userRepo database.UserRepository, wsService WebsocketService) UserService {
    return &userService{
        userRepository:   userRepo,
        websocketService: wsService,
    }
}

func (s *userService) GetUser(userID string) (*models.User, error) {
    return s.userRepository.GetUser(userID)
}

func (s *userService) CreateUser(user *models.User) (*models.User, error) {
    if user.ID == "" || user.Username == "" {
        return nil, errors.New("user ID and username are required")
    }
    return s.userRepository.CreateUser(user)
}

func (s *userService) ListOnlineUsers() ([]*models.User, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    onlineUsers := make([]*models.User, 0)
    for userID := range s.websocketService.GetClients() {
        user, err := s.userRepository.GetUser(userID)
        if err != nil || user == nil {
            continue
        }
        onlineUsers = append(onlineUsers, user)
    }
    return onlineUsers, nil
}

func (s *userService) ListUserGroups(userID string) ([]*models.Group, error) {
    return s.userRepository.GetUserGroups(userID)
}