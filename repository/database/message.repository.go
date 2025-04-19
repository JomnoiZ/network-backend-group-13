package database

import (
	"context"
	"time"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoMessageRepository struct {
    collection *mongo.Collection
}

func NewMongoMessageRepository(client *mongo.Client) MessageRepository {
    collection := client.Database("chat").Collection("messages")
    return &mongoMessageRepository{collection: collection}
}

func (r *mongoMessageRepository) SaveMessage(message *models.MessageDB) error {
    ctx := context.Background()
    if message.ID == "" {
        message.ID = uuid.New().String()
    }
    message.Timestamp = time.Now()
    _, err := r.collection.InsertOne(ctx, message)
    return err
}

func (r *mongoMessageRepository) GetGroupMessages(groupID string) ([]*models.MessageDB, error) {
    ctx := context.Background()
    cursor, err := r.collection.Find(ctx, bson.M{"group_id": groupID}, options.Find().SetSort(bson.M{"timestamp": 1}))
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var messages []*models.MessageDB
    for cursor.Next(ctx) {
        var msg models.MessageDB
        if err := cursor.Decode(&msg); err != nil {
            return nil, err
        }
        messages = append(messages, &msg)
    }
    return messages, nil
}

func (r *mongoMessageRepository) GetDirectMessages(sender, receiver string) ([]*models.MessageDB, error) {
    ctx := context.Background()
    filter := bson.M{
        "group_id": "",
        "$or": []bson.M{
            {"sender": sender, "receiver": receiver},
            {"sender": receiver, "receiver": sender},
        },
    }
    cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.M{"timestamp": 1}))
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var messages []*models.MessageDB
    for cursor.Next(ctx) {
        var msg models.MessageDB
        if err := cursor.Decode(&msg); err != nil {
            return nil, err
        }
        messages = append(messages, &msg)
    }
    return messages, nil
}