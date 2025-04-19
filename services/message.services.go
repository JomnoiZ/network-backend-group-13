package services

import (
	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
)

type MessageService interface {
    GetGroupMessages(groupID string) ([]*models.MessageDB, error)
    GetDirectMessages(userID, targetID string) ([]*models.MessageDB, error)
}

type messageService struct {
    messageRepo database.MessageRepository
}

func NewMessageService(messageRepo database.MessageRepository) MessageService {
    return &messageService{messageRepo: messageRepo}
}

func (s *messageService) GetGroupMessages(groupID string) ([]*models.MessageDB, error) {
    return s.messageRepo.GetGroupMessages(groupID)
}

func (s *messageService) GetDirectMessages(userID, targetID string) ([]*models.MessageDB, error) {
    return s.messageRepo.GetDirectMessages(userID, targetID)
}