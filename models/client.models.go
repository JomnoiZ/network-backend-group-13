package models

import "github.com/gorilla/websocket"

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Groups map[string]bool
}
