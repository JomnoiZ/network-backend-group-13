package services

import (
	"errors"
	"sync"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
)

type userService struct {
	userRepository    database.UserRepository
	messageRepository database.MessageRepository
	websocketService  WebsocketService
	mutex             sync.RWMutex
}

type UserService interface {
	GetUser(username string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	ListOnlineUsers() ([]*models.User, error)
	ListUserGroups(username string) ([]*models.Group, error)
	GetDirectMessages(sender, receiver string) ([]*models.MessageDB, error)
}

func NewUserService(userRepo database.UserRepository, messageRepo database.MessageRepository, wsService WebsocketService) UserService {
	return &userService{
		userRepository:    userRepo,
		messageRepository: messageRepo,
		websocketService:  wsService,
	}
}

func (s *userService) GetUser(username string) (*models.User, error) {
	return s.userRepository.GetUser(username)
}

func (s *userService) GetAllUsers() ([]*models.User, error) {
	return s.userRepository.GetAllUsers()
}

func (s *userService) CreateUser(user *models.User) (*models.User, error) {
	if user.Username == "" {
		return nil, errors.New("username is required")
	}
	return s.userRepository.CreateUser(user)
}

func (s *userService) ListOnlineUsers() ([]*models.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	onlineUsers := make([]*models.User, 0)
	for username := range s.websocketService.GetClients() {
		user, err := s.userRepository.GetUser(username)
		if err != nil || user == nil {
			continue
		}
		onlineUsers = append(onlineUsers, user)
	}
	return onlineUsers, nil
}

func (s *userService) ListUserGroups(username string) ([]*models.Group, error) {
	return s.userRepository.GetUserGroups(username)
}

func (s *userService) GetDirectMessages(sender, receiver string) ([]*models.MessageDB, error) {
	if sender == "" || receiver == "" {
		return nil, errors.New("sender and receiver usernames are required")
	}
	return s.messageRepository.GetDirectMessages(sender, receiver)
}
