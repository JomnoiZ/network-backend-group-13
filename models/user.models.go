package models

import "time"

type User struct {
	ID        string    `firestore:"id"`
	Username  string    `firestore:"username"`
	CreatedAt time.Time `firestore:"created_at"`
}
