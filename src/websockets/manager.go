package websockets

import (
	"encoding/json"
	"log"
	"sync"

	ws "github.com/gofiber/websocket/v2"
)

type Client struct {
	UserID string
	Conn   *ws.Conn
}

type Message struct {
	Event string      `json:"event"`
	To    string      `json:"to,omitempty"`
	Data  interface{} `json:"data"`
}

var (
	clients   = make(map[string]*Client)
	clientsMu sync.RWMutex
)

func AddClient(userID string, conn *ws.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if old, exists := clients[userID]; exists {
		old.Conn.WriteMessage(ws.CloseMessage, []byte{})
		old.Conn.Close()
	}

	clients[userID] = &Client{UserID: userID, Conn: conn}
	log.Printf("🟢 %s connected", userID)
}

func RemoveClient(userID string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if client, exists := clients[userID]; exists {
		client.Conn.WriteMessage(ws.CloseMessage, []byte{})
		client.Conn.Close()
		delete(clients, userID)
		log.Printf("🔴 %s disconnected", userID)
	}
}

func SendToUser(userID string, event string, data interface{}) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	if client, ok := clients[userID]; ok {
		msg := Message{Event: event, Data: data}
		send(client.Conn, msg)
	}
}

func Broadcast(event string, data interface{}) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	for _, client := range clients {
		msg := Message{Event: event, Data: data}
		send(client.Conn, msg)
	}
}

func send(conn *ws.Conn, msg Message) {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("marshal error:", err)
		return
	}

	if err := conn.WriteMessage(ws.TextMessage, b); err != nil {
		log.Println("write error:", err)
		conn.Close()
	}
}
