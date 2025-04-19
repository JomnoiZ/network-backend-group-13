package models

import "time"

type Group struct {
	ID        string    `bson:"id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	OwnerID   string    `bson:"owner_id" json:"owner_id"`
	Admins    []string  `bson:"admins" json:"admins"`
	Members   []string  `bson:"members" json:"members"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}