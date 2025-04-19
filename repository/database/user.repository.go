package database

import (
	"context"
	"time"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(client *mongo.Client) UserRepository {
	collection := client.Database("chat").Collection("users")
	return &mongoUserRepository{collection: collection}
}

func (r *mongoUserRepository) GetUser(userID string) (*models.User, error) {
	ctx := context.Background()
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mongoUserRepository) CreateUser(user *models.User) (*models.User, error) {
	ctx := context.Background()
	user.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *mongoUserRepository) GetUserGroups(userID string) ([]*models.Group, error) {
	ctx := context.Background()
	cursor, err := r.collection.Database().Collection("groups").Find(ctx, bson.M{"members": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []*models.Group
	for cursor.Next(ctx) {
		var group models.Group
		if err := cursor.Decode(&group); err != nil {
			return nil, err
		}
		groups = append(groups, &group)
	}
	return groups, nil
}