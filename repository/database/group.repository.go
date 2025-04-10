package database

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/JomnoiZ/network-backend-group-13.git/models"
)

type groupRepository struct {
	firestoreClient *firestore.Client
}

type GroupRepository interface {
	GetGroup(groupID string) (*models.Group, error)
	CreateGroup(group *models.Group) (*models.Group, error)
}

func NewGroupRepository(firestoreClient *firestore.Client) GroupRepository {
	return &groupRepository{
		firestoreClient: firestoreClient,
	}
}

func (r *groupRepository) GetGroup(groupID string) (*models.Group, error) {
	doc, err := r.firestoreClient.Collection("groups").Doc(groupID).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var group models.Group
	doc.DataTo(&group)
	return &group, nil
}

func (r *groupRepository) CreateGroup(group *models.Group) (*models.Group, error) {
	group.CreatedAt = time.Now()
	_, err := r.firestoreClient.Collection("groups").Doc(group.ID).Set(context.Background(), group)
	if err != nil {
		return nil, err
	}
	return group, nil
}
