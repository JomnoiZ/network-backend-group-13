package services

import (
	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
)

type groupService struct {
	groupRepository database.GroupRepository
}

type GroupService interface {
	GetGroup(groupID string) (*models.Group, error)
	CreateGroup(group *models.Group) (*models.Group, error)
}

func NewGroupService(groupRepository database.GroupRepository) GroupService {
	return &groupService{
		groupRepository: groupRepository,
	}
}

func (s *groupService) GetGroup(groupID string) (*models.Group, error) {
	group, err := s.groupRepository.GetGroup(groupID)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (s *groupService) CreateGroup(group *models.Group) (*models.Group, error) {
	group, err := s.groupRepository.CreateGroup(group)
	if err != nil {
		return nil, err
	}
	return group, nil
}
