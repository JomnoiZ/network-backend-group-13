package database

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"google.golang.org/api/iterator"
)

type firestoreMessageRepository struct {
    client *firestore.Client
}

func NewFirestoreMessageRepository(client *firestore.Client) MessageRepository {
    return &firestoreMessageRepository{client: client}
}

func (r *firestoreMessageRepository) SaveMessage(message *models.MessageDB) error {
    ctx := context.Background()
    if message.ID == "" {
        message.ID = r.client.Collection("messages").NewDoc().ID
    }
    message.Timestamp = time.Now()
    _, err := r.client.Collection("messages").Doc(message.ID).Set(ctx, message)
    if err != nil {
        return err
    }
    return nil
}

func (r *firestoreMessageRepository) GetGroupMessages(groupID string) ([]*models.MessageDB, error) {
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

func (r *firestoreMessageRepository) GetDirectMessages(userID, targetID string) ([]*models.MessageDB, error) {
    ctx := context.Background()
    iter := r.client.Collection("messages").
        Where("group_id", "==", "").
        Where("sender_id", "in", []string{userID, targetID}).
        Where("receiver_id", "in", []string{userID, targetID}).
        OrderBy("timestamp", firestore.Asc).
        Documents(ctx)
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