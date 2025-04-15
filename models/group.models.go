package models

import "time"

type Group struct {
    ID        string    `firestore:"id" json:"id"`
    Name      string    `firestore:"name" json:"name"`
    OwnerID   string    `firestore:"owner_id" json:"owner_id"`
    Admins    []string  `firestore:"admins" json:"admins"`
    Members   []string  `firestore:"members" json:"members"`
    CreatedAt time.Time `firestore:"created_at" json:"created_at"`
}