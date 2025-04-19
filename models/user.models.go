package models

import "time"

type User struct {
	ID        string    `bson:"id" json:"id"`
	Username  string    `bson:"username" json:"username"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}