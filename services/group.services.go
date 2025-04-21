package services

import (
	"errors"
	"time"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
	"github.com/google/uuid"
)

type groupService struct {
	groupRepository   database.GroupRepository
	userRepository    database.UserRepository
	messageRepository database.MessageRepository
	websocketService  WebsocketService
}

type GroupService interface {
	GetAllGroups() ([]*models.Group, error)
	GetGroup(groupID string) (*models.Group, error)
	CreateGroup(name, owner string) (*models.Group, error)
	AddMember(groupID, username, requester string) error
	KickMember(groupID, username, requester string) error
	AddAdmin(groupID, username, requester string) error
	RemoveAdmin(groupID, username, requester string) error
	GetGroupMessages(groupID string) ([]*models.MessageDB, error)
}

func NewGroupService(groupRepo database.GroupRepository, userRepo database.UserRepository, messageRepo database.MessageRepository, wsService WebsocketService) GroupService {
	return &groupService{
		groupRepository:   groupRepo,
		userRepository:    userRepo,
		messageRepository: messageRepo,
		websocketService:  wsService,
	}
}

func (s *groupService) GetAllGroups() ([]*models.Group, error) {
	groups, err := s.groupRepository.GetAllGroups()
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *groupService) GetGroup(groupID string) (*models.Group, error) {
	return s.groupRepository.GetGroup(groupID)
}

func (s *groupService) CreateGroup(name, owner string) (*models.Group, error) {
	if name == "" || owner == "" {
		return nil, errors.New("name and owner are required")
	}
	_, err := s.userRepository.GetUser(owner)
	if err != nil {
		return nil, errors.New("owner not found")
	}
	group := &models.Group{
		ID:        uuid.New().String(),
		Name:      name,
		Owner:     owner,
		Admins:    []string{owner},
		Members:   []string{owner},
		CreatedAt: time.Now(),
	}
	createdGroup, err := s.groupRepository.CreateGroup(group)
	if err != nil {
		return nil, err
	}
	s.websocketService.AddToGroup(&models.Client{Username: owner}, group.ID)
	return createdGroup, nil
}

func (s *groupService) AddMember(groupID, username, requester string) error {
	group, err := s.groupRepository.GetGroup(groupID)
	if err != nil || group == nil {
		return errors.New("group not found")
	}
	// isMember := false
	// for _, m := range group.Members {
	// 	if m == requester {
	// 		isMember = true
	// 		break
	// 	}
	// }
	// if !isMember {
	// 	return errors.New("requester is not a group member")
	// }
	_, err = s.userRepository.GetUser(username)
	if err != nil {
		return errors.New("user not found")
	}
	for _, m := range group.Members {
		if m == username {
			return nil
			// return errors.New("user is already a member")
		}
	}
	group.Members = append(group.Members, username)
	err = s.groupRepository.UpdateGroup(group)
	if err != nil {
		return err
	}
	s.websocketService.AddToGroup(&models.Client{Username: username}, groupID)
	s.websocketService.NotifyGroupUpdate(groupID, "member_added", map[string]string{"username": username})
	return nil
}

func (s *groupService) KickMember(groupID, username, requester string) error {
	group, err := s.groupRepository.GetGroup(groupID)
	if err != nil || group == nil {
		return errors.New("group not found")
	}
	isAuthorized := group.Owner == requester
	if !isAuthorized {
		for _, admin := range group.Admins {
			if admin == requester {
				isAuthorized = true
				break
			}
		}
	}
	if !isAuthorized {
		return errors.New("unauthorized: only owner or admins can kick members")
	}
	if username == group.Owner {
		return errors.New("cannot kick group owner")
	}
	newMembers := []string{}
	wasMember := false
	for _, m := range group.Members {
		if m != username {
			newMembers = append(newMembers, m)
		} else {
			wasMember = true
		}
	}
	if !wasMember {
		return errors.New("user is not a group member")
	}
	group.Members = newMembers
	newAdmins := []string{}
	for _, a := range group.Admins {
		if a != username {
			newAdmins = append(newAdmins, a)
		}
	}
	group.Admins = newAdmins
	err = s.groupRepository.UpdateGroup(group)
	if err != nil {
		return err
	}
	s.websocketService.KickFromGroup(username, groupID)
	s.websocketService.NotifyGroupUpdate(groupID, "member_kicked", map[string]string{"username": username})
	return nil
}

func (s *groupService) AddAdmin(groupID, username, requester string) error {
	group, err := s.groupRepository.GetGroup(groupID)
	if err != nil || group == nil {
		return errors.New("group not found")
	}
	if group.Owner != requester {
		return errors.New("unauthorized: only owner can add admins")
	}
	isMember := false
	for _, m := range group.Members {
		if m == username {
			isMember = true
			break
		}
	}
	if !isMember {
		return errors.New("user is not a group member")
	}
	for _, a := range group.Admins {
		if a == username {
			return errors.New("user is already an admin")
		}
	}
	group.Admins = append(group.Admins, username)
	err = s.groupRepository.UpdateGroup(group)
	if err != nil {
		return err
	}
	s.websocketService.NotifyGroupUpdate(groupID, "admin_added", map[string]string{"username": username})
	return nil
}

func (s *groupService) RemoveAdmin(groupID, username, requester string) error {
	group, err := s.groupRepository.GetGroup(groupID)
	if err != nil || group == nil {
		return errors.New("group not found")
	}
	if group.Owner != requester {
		return errors.New("unauthorized: only owner can remove admins")
	}
	if username == group.Owner {
		return errors.New("cannot remove owner's admin status")
	}
	newAdmins := []string{}
	wasAdmin := false
	for _, a := range group.Admins {
		if a != username {
			newAdmins = append(newAdmins, a)
		} else {
			wasAdmin = true
		}
	}
	if !wasAdmin {
		return errors.New("user is not an admin")
	}
	group.Admins = newAdmins
	err = s.groupRepository.UpdateGroup(group)
	if err != nil {
		return err
	}
	s.websocketService.NotifyGroupUpdate(groupID, "admin_removed", map[string]string{"username": username})
	return nil
}

func (s *groupService) GetGroupMessages(groupID string) ([]*models.MessageDB, error) {
	return s.messageRepository.GetGroupMessages(groupID)
}
