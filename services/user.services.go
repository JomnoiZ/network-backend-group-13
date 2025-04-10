package services

import (
	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
)

type userService struct {
	userRepository database.UserRepository
}

type UserService interface {
	GetUser(userID string) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
}

func NewUserService(userRepository database.UserRepository) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (s *userService) GetUser(userID string) (*models.User, error) {
	user, err := s.userRepository.GetUser(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) CreateUser(user *models.User) (*models.User, error) {
	user, err := s.userRepository.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
