package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a chat message structure
type Message struct {
	Type      string    `json:"type"`      // "join", "leave", "message"
	Username  string    `json:"username"`  // Sender's username
	Content   string    `json:"content"`   // Message content
	Room      string    `json:"room"`      // Room identifier
	Timestamp time.Time `json:"timestamp"` // Message timestamp
}

// Client represents a connected WebSocket client
type Client struct {
	conn     *websocket.Conn
	username string
	room     string
	send     chan Message
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates and initializes a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s in room %s", client.username, client.room)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s", client.username)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Only send to clients in the same room
				if client.room == message.Room {
					select {
					case client.send <- message:
					default:
						// Client's send channel is full, remove them
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, check r.Header.Get("Origin") against allowed domains
		return true
	},
}

// readPump handles incoming messages from the client
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	// Configure read deadline and pong handler
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(messageData, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Set metadata
		msg.Username = c.username
		msg.Room = c.room
		msg.Timestamp = time.Now()

		// Broadcast to hub
		hub.broadcast <- msg
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message as JSON
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleWebSocket handles WebSocket connection requests
func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Get username and room from query parameters
	username := r.URL.Query().Get("username")
	room := r.URL.Query().Get("room")

	if username == "" {
		username = "Anonymous"
	}
	if room == "" {
		room = "general"
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create new client
	client := &Client{
		conn:     conn,
		username: username,
		room:     room,
		send:     make(chan Message, 256),
	}

	// Register client with hub
	hub.register <- client

	// Send join notification
	joinMsg := Message{
		Type:      "join",
		Username:  username,
		Content:   username + " joined the room",
		Room:      room,
		Timestamp: time.Now(),
	}
	hub.broadcast <- joinMsg

	// Start client goroutines
	go client.writePump()
	go client.readPump(hub)
}

// handleHealth provides a health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Create and start hub
	hub := NewHub()
	go hub.Run()

	// Setup routes
	http.HandleFunc("/ws", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	}))
	http.HandleFunc("/health", corsMiddleware(handleHealth))

	// Start server
	port := ":8080"
	log.Printf("Chat server starting on %s", port)
	log.Printf("WebSocket endpoint: ws://localhost%s/ws", port)
	log.Printf("Health check: http://localhost%s/health", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server error:", err)
	}
}