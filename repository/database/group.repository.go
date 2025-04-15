package database

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"google.golang.org/api/iterator"
)

type firestoreGroupRepository struct {
    client *firestore.Client
}

func NewFirestoreGroupRepository(client *firestore.Client) GroupRepository {
    return &firestoreGroupRepository{client: client}
}

func (r *firestoreGroupRepository) GetGroup(groupID string) (*models.Group, error) {
    ctx := context.Background()
    doc, err := r.client.Collection("groups").Doc(groupID).Get(ctx)
    if err != nil {
        return nil, err
    }
    var group models.Group
    if err := doc.DataTo(&group); err != nil {
        return nil, err
    }
    return &group, nil
}

func (r *firestoreGroupRepository) CreateGroup(group *models.Group) (*models.Group, error) {
    ctx := context.Background()
    group.CreatedAt = time.Now()
    _, err := r.client.Collection("groups").Doc(group.ID).Set(ctx, group)
    if err != nil {
        return nil, err
    }
    return group, nil
}

func (r *firestoreGroupRepository) UpdateGroup(group *models.Group) error {
    ctx := context.Background()
    _, err := r.client.Collection("groups").Doc(group.ID).Set(ctx, group)
    return err
}

func (r *firestoreGroupRepository) GetGroupMessages(groupID string) ([]*models.MessageDB, error) {
    ctx := context.Background()
    iter := r.client.Collection("messages").Where("group_id", "==", groupID).OrderBy("timestamp", firestore.Asc).Documents(ctx)
    messages := []*models.MessageDB{}
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, err
        }
        var msg models.MessageDB
        if err := doc.DataTo(&msg); err != nil {
            return nil, err
        }
        messages = append(messages, &msg)
    }
    return messages, nil
}