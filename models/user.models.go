package models

import "time"

type User struct {
    Username  string    `bson:"username" json:"username"`
    CreatedAt time.Time `bson:"created_at" json:"created_at"`
}