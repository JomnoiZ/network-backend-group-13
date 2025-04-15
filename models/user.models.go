package models

import "time"

type User struct {
    ID        string    `firestore:"id" json:"id"`
    Username  string    `firestore:"username" json:"username"`
    CreatedAt time.Time `firestore:"created_at" json:"created_at"`
}

