package services

import (
	"errors"
	"time"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/JomnoiZ/network-backend-group-13.git/repository/database"
	"github.com/google/uuid"
)

type groupService struct {
    groupRepository  database.GroupRepository
    userRepository   database.UserRepository
    websocketService WebsocketService
}

type GroupService interface {
    GetGroup(groupID string) (*models.Group, error)
    CreateGroup(name, ownerID string) (*models.Group, error)
    AddMember(groupID, userID, requesterID string) error
    KickMember(groupID, userID, requesterID string) error
    AddAdmin(groupID, userID, requesterID string) error
    RemoveAdmin(groupID, userID, requesterID string) error
    GetGroupMessages(groupID string) ([]*models.MessageDB, error)
}

func NewGroupService(groupRepo database.GroupRepository, userRepo database.UserRepository, wsService WebsocketService) GroupService {
    return &groupService{
        groupRepository:  groupRepo,
        userRepository:   userRepo,
        websocketService: wsService,
    }
}

func (s *groupService) GetGroup(groupID string) (*models.Group, error) {
    return s.groupRepository.GetGroup(groupID)
}

func (s *groupService) CreateGroup(name, ownerID string) (*models.Group, error) {
    if name == "" || ownerID == "" {
        return nil, errors.New("name and owner ID are required")
    }
    _, err := s.userRepository.GetUser(ownerID)
    if err != nil {
        return nil, errors.New("owner not found")
    }
    group := &models.Group{
        ID:        uuid.New().String(),
        Name:      name,
        OwnerID:   ownerID,
        Admins:    []string{ownerID},
        Members:   []string{ownerID},
        CreatedAt: time.Now(),
    }
    createdGroup, err := s.groupRepository.CreateGroup(group)
    if err != nil {
        return nil, err
    }
    s.websocketService.AddToGroup(&models.Client{ID: ownerID}, group.ID)
    return createdGroup, nil
}

func (s *groupService) AddMember(groupID, userID, requesterID string) error {
    group, err := s.groupRepository.GetGroup(groupID)
    if err != nil || group == nil {
        return errors.New("group not found")
    }
    isMember := false
    for _, m := range group.Members {
        if m == requesterID {
            isMember = true
            break
        }
    }
    if !isMember {
        return errors.New("requester is not a group member")
    }
    _, err = s.userRepository.GetUser(userID)
    if err != nil {
        return errors.New("user not found")
    }
    for _, m := range group.Members {
        if m == userID {
            return errors.New("user is already a member")
        }
    }
    group.Members = append(group.Members, userID)
    err = s.groupRepository.UpdateGroup(group)
    if err != nil {
        return err
    }
    s.websocketService.AddToGroup(&models.Client{ID: userID}, groupID)
    s.websocketService.NotifyGroupUpdate(groupID, "member_added", map[string]string{"user_id": userID})
    return nil
}

func (s *groupService) KickMember(groupID, userID, requesterID string) error {
    group, err := s.groupRepository.GetGroup(groupID)
    if err != nil || group == nil {
        return errors.New("group not found")
    }
    isAuthorized := group.OwnerID == requesterID
    if !isAuthorized {
        for _, admin := range group.Admins {
            if admin == requesterID {
                isAuthorized = true
                break
            }
        }
    }
    if !isAuthorized {
        return errors.New("unauthorized: only owner or admins can kick members")
    }
    if userID == group.OwnerID {
        return errors.New("cannot kick group owner")
    }
    newMembers := []string{}
    wasMember := false
    for _, m := range group.Members {
        if m != userID {
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
        if a != userID {
            newAdmins = append(newAdmins, a)
        }
    }
    group.Admins = newAdmins
    err = s.groupRepository.UpdateGroup(group)
    if err != nil {
        return err
    }
    s.websocketService.KickFromGroup(userID, groupID)
    s.websocketService.NotifyGroupUpdate(groupID, "member_kicked", map[string]string{"user_id": userID})
    return nil
}

func (s *groupService) AddAdmin(groupID, userID, requesterID string) error {
    group, err := s.groupRepository.GetGroup(groupID)
    if err != nil || group == nil {
        return errors.New("group not found")
    }
    if group.OwnerID != requesterID {
        return errors.New("unauthorized: only owner can add admins")
    }
    isMember := false
    for _, m := range group.Members {
        if m == userID {
            isMember = true
            break
        }
    }
    if !isMember {
        return errors.New("user is not a group member")
    }
    for _, a := range group.Admins {
        if a == userID {
            return errors.New("user is already an admin")
        }
    }
    group.Admins = append(group.Admins, userID)
    err = s.groupRepository.UpdateGroup(group)
    if err != nil {
        return err
    }
    s.websocketService.NotifyGroupUpdate(groupID, "admin_added", map[string]string{"user_id": userID})
    return nil
}

func (s *groupService) RemoveAdmin(groupID, userID, requesterID string) error {
    group, err := s.groupRepository.GetGroup(groupID)
    if err != nil || group == nil {
        return errors.New("group not found")
    }
    if group.OwnerID != requesterID {
        return errors.New("unauthorized: only owner can remove admins")
    }
    if userID == group.OwnerID {
        return errors.New("cannot remove owner's admin status")
    }
    newAdmins := []string{}
    wasAdmin := false
    for _, a := range group.Admins {
        if a != userID {
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
    s.websocketService.NotifyGroupUpdate(groupID, "admin_removed", map[string]string{"user_id": userID})
    return nil
}

func (s *groupService) GetGroupMessages(groupID string) ([]*models.MessageDB, error) {
    return s.groupRepository.GetGroupMessages(groupID)
}