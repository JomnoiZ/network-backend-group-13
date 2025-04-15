package database

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/JomnoiZ/network-backend-group-13.git/models"
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