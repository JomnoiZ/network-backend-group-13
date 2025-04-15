package database

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/JomnoiZ/network-backend-group-13.git/models"
	"google.golang.org/api/iterator"
)

type firestoreUserRepository struct {
    client *firestore.Client
}

func NewFirestoreUserRepository(client *firestore.Client) UserRepository {
    return &firestoreUserRepository{client: client}
}

func (r *firestoreUserRepository) GetUser(userID string) (*models.User, error) {
    ctx := context.Background()
    doc, err := r.client.Collection("users").Doc(userID).Get(ctx)
    if err != nil {
        return nil, err
    }
    var user models.User
    if err := doc.DataTo(&user); err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *firestoreUserRepository) CreateUser(user *models.User) (*models.User, error) {
    ctx := context.Background()
    user.CreatedAt = time.Now()
    _, err := r.client.Collection("users").Doc(user.ID).Set(ctx, user)
    if err != nil {
        return nil, err
    }
    return user, nil
}

func (r *firestoreUserRepository) GetUserGroups(userID string) ([]*models.Group, error) {
    ctx := context.Background()
    iter := r.client.Collection("groups").Where("members", "array-contains", userID).Documents(ctx)
    groups := []*models.Group{}
    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, err
        }
        var group models.Group
        if err := doc.DataTo(&group); err != nil {
            return nil, err
        }
        groups = append(groups, &group)
    }
    return groups, nil
}