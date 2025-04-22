package database

import (
	"context"
	"time"

	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoGroupRepository struct {
	collection *mongo.Collection
}

func NewMongoGroupRepository(client *mongo.Client) GroupRepository {
	collection := client.Database("chat").Collection("groups")
	return &mongoGroupRepository{collection: collection}
}

func (r *mongoGroupRepository) GetAllGroups() ([]*models.Group, error) {
	ctx := context.Background()
	cursor, err := r.collection.Find(ctx, bson.M{})
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
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *mongoGroupRepository) GetGroup(groupID string) (*models.Group, error) {
	ctx := context.Background()
	var group models.Group
	err := r.collection.FindOne(ctx, bson.M{"id": groupID}).Decode(&group)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *mongoGroupRepository) CreateGroup(group *models.Group) (*models.Group, error) {
	ctx := context.Background()
	group.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *mongoGroupRepository) UpdateGroup(group *models.Group) error {
	ctx := context.Background()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"id": group.ID}, group)
	return err
}
