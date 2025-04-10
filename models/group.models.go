package models

import "time"

type Group struct {
	ID        string    `firestore:"id"`
	Name      string    `firestore:"name"`
	Members   []string  `firestore:"members"`
	CreatedAt time.Time `firestore:"created_at"`
}
