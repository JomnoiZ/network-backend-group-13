package models

import "time"

type Message struct {
    ID         string      `json:"id"`
    Type       string      `json:"type"` // message, typing, status, join_group, group_update
    SenderID   string      `json:"sender_id"`
    ReceiverID string      `json:"receiver_id,omitempty"`
    GroupID    string      `json:"group_id,omitempty"`
    Content    string      `json:"content,omitempty"`
    Status     string      `json:"status,omitempty"`
    Data       interface{} `json:"data,omitempty"`
}

type MessageDB struct {
    ID         string    `firestore:"id" json:"id"`
    SenderID   string    `firestore:"sender_id" json:"sender_id"`
    ReceiverID string    `firestore:"receiver_id" json:"receiver_id,omitempty"`
    GroupID    string    `firestore:"group_id" json:"group_id,omitempty"`
    Content    string    `firestore:"content" json:"content"`
    Timestamp  time.Time `firestore:"timestamp" json:"timestamp"`
}
