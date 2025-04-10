package database

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/JomnoiZ/network-backend-group-13.git/models"
)

type userRepository struct {
	firestoreClient *firestore.Client
}

type UserRepository interface {
	GetUser(userID string) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
}

func NewUserRepository(firestoreClient *firestore.Client) UserRepository {
	return &userRepository{
		firestoreClient: firestoreClient,
	}
}

func (r *userRepository) GetUser(userID string) (*models.User, error) {
	doc, err := r.firestoreClient.Collection("users").Doc(userID).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var user models.User
	doc.DataTo(&user)
	return &user, nil
}

func (r *userRepository) CreateUser(user *models.User) (*models.User, error) {
	user.CreatedAt = time.Now()
	_, err := r.firestoreClient.Collection("users").Doc(user.ID).Set(context.Background(), user)
	if err != nil {
		return nil, err
	}
	return user, err
}
