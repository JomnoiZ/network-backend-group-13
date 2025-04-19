package models

import "time"

type Message struct {
    ID        string      `json:"id"`
    Type      string      `json:"type"` 
    Sender    string      `json:"sender"` 
    Receiver  string      `json:"receiver,omitempty"` 
    GroupID   string      `json:"group_id,omitempty"`
    Content   string      `json:"content,omitempty"`
    Status    string      `json:"status,omitempty"`
    Data      interface{} `json:"data,omitempty"`
}

type MessageDB struct {
    ID        string    `bson:"id" json:"id"`
    Sender    string    `bson:"sender" json:"sender"` 
    Receiver  string    `bson:"receiver" json:"receiver,omitempty"`
    GroupID   string    `bson:"group_id" json:"group_id,omitempty"`
    Content   string    `bson:"content" json:"content"`
    Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}