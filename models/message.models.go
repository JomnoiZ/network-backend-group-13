package models

import "time"

type Message struct {
	ID         string `json:"id"`
	Type       string `json:"type"` // message, typing, status, join_group
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
	Content    string `json:"content,omitempty"`
	Status     string `json:"status,omitempty"` // typing-start/stop or online/offline
}

type MessageDB struct {
	ID         string    `firestore:"id"`
	SenderID   string    `firestore:"sender_id"`
	ReceiverID string    `firestore:"receiver_id,omitempty"`
	GroupID    string    `firestore:"group_id,omitempty"`
	Content    string    `firestore:"content"`
	Timestamp  time.Time `firestore:"timestamp"`
}
