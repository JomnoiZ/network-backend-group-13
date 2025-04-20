package database

import (
	"context"
	"time"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(client *mongo.Client) UserRepository {
	collection := client.Database("chat").Collection("users")
	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		panic(err)
	}
	return &mongoUserRepository{collection: collection}
}

func (r *mongoUserRepository) GetUser(username string) (*models.User, error) {
	ctx := context.Background()
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mongoUserRepository) GetAllUsers() ([]*models.User, error) {
	ctx := context.Background()
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
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

func (r *mongoUserRepository) GetUserGroups(username string) ([]*models.Group, error) {
	ctx := context.Background()
	cursor, err := r.collection.Database().Collection("groups").Find(ctx, bson.M{"members": username})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	groups := []*models.Group{}
	for cursor.Next(ctx) {
		var group models.Group
		if err := cursor.Decode(&group); err != nil {
			return nil, err
		}
		groups = append(groups, &group)
	}
	return groups, nil
}
