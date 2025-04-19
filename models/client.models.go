package models

import "github.com/gorilla/websocket"

type Client struct {
    Username string           `json:"username"`
    Conn     *websocket.Conn  `json:"-"`
    Send     chan []byte      `json:"-"`
    Groups   map[string]bool  `json:"groups"`
}