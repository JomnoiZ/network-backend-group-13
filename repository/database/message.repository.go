package database

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"google.golang.org/api/iterator"
)

type messageRepository struct {
	firestoreClient *firestore.Client
}

type MessageRepository interface {
	SaveMessage(msg *models.MessageDB)
	GetMessagesForUser(userID string) ([]models.MessageDB, error)
}

func NewMessageRepository(firestoreClient *firestore.Client) MessageRepository {
	return &messageRepository{
		firestoreClient: firestoreClient,
	}
}

func (r *messageRepository) SaveMessage(msg *models.MessageDB) {
	msg.Timestamp = time.Now()
	_, _, err := r.firestoreClient.Collection("messages").Add(context.Background(), msg)
	if err != nil {
		log.Printf("❌ Failed to save message: %v", err)
	}
}

func (r *messageRepository) GetMessagesForUser(userID string) ([]models.MessageDB, error) {
	var messages []models.MessageDB

	iter := r.firestoreClient.Collection("messages").
		Where("receiver_id", "==", userID).
		OrderBy("timestamp", firestore.Asc).
		Documents(context.Background())

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var m models.MessageDB
		doc.DataTo(&m)
		messages = append(messages, m)
	}
	return messages, nil
}
