package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	nickname string
}

type Message struct {
	Type       string `json:"type"`
	From       string `json:"from"`
	To         string `json:"to,omitempty"`
	Content    string `json:"content"`
	IsPrivate  bool   `json:"isPrivate"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

var hub = Hub{
	broadcast:  make(chan Message),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.broadcastUserList()
			log.Printf("%s joined", client.nickname)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.broadcastUserList()
				log.Printf("%s left", client.nickname)
			}
		case message := <-h.broadcast:
			if message.IsPrivate {
				for client := range h.clients {
					if client.nickname == message.To {
						client.send <- encodeMessage(message)
						break
					}
				}
			} else {
				for client := range h.clients {
					client.send <- encodeMessage(message)
				}
			}
		}
	}
}

func (h *Hub) broadcastUserList() {
	var nicknames []string
	for client := range h.clients {
		nicknames = append(nicknames, client.nickname)
	}
	userList, _ := json.Marshal(map[string]interface{}{
		"type":     "userlist",
		"nicknames": nicknames,
	})
	for client := range h.clients {
		client.send <- userList
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	client := &Client{
		conn: conn,
		send: make(chan []byte),
	}
	hub.register <- client

	go handleMessages(client)
	go sendMessages(client)
}

func handleMessages(client *Client) {
	defer func() {
		hub.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("Read Error: %v", err)
			break
		}

		var incoming map[string]interface{}
		if err := json.Unmarshal(msg, &incoming); err == nil {
			if incoming["type"] == "setname" {
				client.nickname = incoming["nickname"].(string)
				hub.broadcastUserList()
				continue
			}
		}

		var message Message
		if err := json.Unmarshal(msg, &message); err == nil {
			hub.broadcast <- message
		}
	}
}

func sendMessages(client *Client) {
	for msg := range client.send {
		err := client.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("Write Error: %v", err)
			break
		}
	}
}

func encodeMessage(msg Message) []byte {
	encoded, _ := json.Marshal(msg)
	return encoded
}

func main() {
	go hub.run()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)
	fmt.Println("Server started on http://localhost:8765")
	err := http.ListenAndServe(":8765", nil)
	if err != nil {
		log.Fatal("ListenAndServe Error:", err)
	}
}
