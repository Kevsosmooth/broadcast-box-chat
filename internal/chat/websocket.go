package chat

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// WSHandler handles WebSocket connections for chat
type WSHandler struct {
	manager     *Manager
	rateLimiter *RateLimiter
	connections map[string]*Connection // userID -> connection
	connMux     sync.RWMutex
}

// Connection represents a WebSocket connection
type Connection struct {
	UserID     string
	Username   string
	StreamKey  string
	Conn       *websocket.Conn
	Send       chan WSMessage
	manager    *WSHandler
}

// NewWSHandler creates a new WebSocket handler
func NewWSHandler(manager *Manager, rateLimiter *RateLimiter) *WSHandler {
	return &WSHandler{
		manager:     manager,
		rateLimiter: rateLimiter,
		connections: make(map[string]*Connection),
	}
}

// HandleWebSocket handles incoming WebSocket connections
func (h *WSHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request, streamKey string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	connection := &Connection{
		Conn:      conn,
		StreamKey: streamKey,
		Send:      make(chan WSMessage, 256),
		manager:   h,
	}

	// Start goroutines for reading and writing
	go connection.writePump()
	go connection.readPump()
}

// readPump reads messages from the WebSocket connection
func (c *Connection) readPump() {
	defer func() {
		c.cleanup()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg map[string]interface{}
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(msg)
	}
}

// writePump writes messages to the WebSocket connection
func (c *Connection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages from the client
func (c *Connection) handleMessage(msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		c.sendError("Invalid message type")
		return
	}

	switch msgType {
	case "join":
		c.handleJoin(msg)
	case "message":
		c.handleChatMessage(msg)
	case "typing":
		c.handleTyping(msg)
	default:
		c.sendError("Unknown message type")
	}
}

// handleJoin handles a user joining the chat
func (c *Connection) handleJoin(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		c.sendError("Invalid join data")
		return
	}

	userID, _ := data["userId"].(string)
	username, _ := data["username"].(string)

	if userID == "" || username == "" {
		c.sendError("Missing userId or username")
		return
	}

	c.UserID = userID
	c.Username = username

	// Add user to manager
	err := c.manager.manager.AddUser(c.StreamKey, userID, username)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	// Register connection
	c.manager.connMux.Lock()
	c.manager.connections[userID] = c
	c.manager.connMux.Unlock()

	// Send message history
	messages := c.manager.manager.GetMessages(c.StreamKey, 100)
	c.Send <- WSMessage{
		Type:      "history",
		Data:      messages,
		Timestamp: time.Now(),
	}

	// Send user list
	users := c.manager.manager.GetUsers(c.StreamKey)
	c.Send <- WSMessage{
		Type:      "users",
		Data:      users,
		Timestamp: time.Now(),
	}

	// Check if user is timed out
	isTimedOut, duration := c.manager.rateLimiter.GetTimeoutStatus(userID)
	if isTimedOut {
		c.Send <- WSMessage{
			Type: "timeout",
			Data: map[string]interface{}{
				"duration": duration.Seconds(),
			},
			Timestamp: time.Now(),
		}
	}

	// Broadcast user joined
	c.broadcastToRoom(WSMessage{
		Type: "user_joined",
		Data: map[string]interface{}{
			"userId":   userID,
			"username": username,
		},
		Timestamp: time.Now(),
	})

	log.Printf("User %s (%s) joined chat for stream %s", username, userID, c.StreamKey)
}

// handleChatMessage handles a chat message from the user
func (c *Connection) handleChatMessage(msg map[string]interface{}) {
	if c.UserID == "" {
		c.sendError("Not joined to chat")
		return
	}

	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		c.sendError("Invalid message data")
		return
	}

	message, ok := data["message"].(string)
	if !ok || message == "" {
		c.sendError("Invalid message content")
		return
	}

	// Check rate limit
	allowed, rateLimitErr := c.manager.rateLimiter.CheckMessage(c.UserID, message)
	if !allowed {
		c.Send <- WSMessage{
			Type:      "rate_limit",
			Error:     rateLimitErr.Message,
			Timestamp: time.Now(),
		}
		return
	}

	// Add message to manager
	chatMsg, err := c.manager.manager.AddMessage(c.StreamKey, c.UserID, c.Username, message)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	// Broadcast to all users in the room
	c.broadcastToRoom(WSMessage{
		Type:      "message",
		Data:      chatMsg,
		Timestamp: time.Now(),
	})
}

// handleTyping handles typing indicator
func (c *Connection) handleTyping(msg map[string]interface{}) {
	if c.UserID == "" {
		return
	}

	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return
	}

	isTyping, _ := data["isTyping"].(bool)

	// Broadcast typing status to room (excluding sender)
	c.broadcastToRoomExcept(WSMessage{
		Type: "typing",
		Data: map[string]interface{}{
			"userId":   c.UserID,
			"username": c.Username,
			"isTyping": isTyping,
		},
		Timestamp: time.Now(),
	}, c.UserID)
}

// broadcastToRoom broadcasts a message to all users in the room
func (c *Connection) broadcastToRoom(msg WSMessage) {
	c.manager.connMux.RLock()
	defer c.manager.connMux.RUnlock()

	for _, conn := range c.manager.connections {
		if conn.StreamKey == c.StreamKey {
			select {
			case conn.Send <- msg:
			default:
				// Channel full, skip
			}
		}
	}
}

// broadcastToRoomExcept broadcasts to all users except one
func (c *Connection) broadcastToRoomExcept(msg WSMessage, exceptUserID string) {
	c.manager.connMux.RLock()
	defer c.manager.connMux.RUnlock()

	for _, conn := range c.manager.connections {
		if conn.StreamKey == c.StreamKey && conn.UserID != exceptUserID {
			select {
			case conn.Send <- msg:
			default:
				// Channel full, skip
			}
		}
	}
}

// sendError sends an error message to the client
func (c *Connection) sendError(errorMsg string) {
	c.Send <- WSMessage{
		Type:      "error",
		Error:     errorMsg,
		Timestamp: time.Now(),
	}
}

// cleanup cleans up the connection
func (c *Connection) cleanup() {
	// Remove from manager
	if c.UserID != "" {
		c.manager.manager.RemoveUser(c.StreamKey, c.UserID)

		c.manager.connMux.Lock()
		delete(c.manager.connections, c.UserID)
		c.manager.connMux.Unlock()

		// Broadcast user left
		c.broadcastToRoom(WSMessage{
			Type: "user_left",
			Data: map[string]interface{}{
				"userId":   c.UserID,
				"username": c.Username,
			},
			Timestamp: time.Now(),
		})

		log.Printf("User %s (%s) left chat for stream %s", c.Username, c.UserID, c.StreamKey)
	}

	close(c.Send)
	c.Conn.Close()
}

// HTTPHandler returns an HTTP handler function for WebSocket connections
func (h *WSHandler) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	// Extract stream key from URL path or query
	streamKey := r.URL.Query().Get("streamKey")
	if streamKey == "" {
		http.Error(w, "Missing streamKey parameter", http.StatusBadRequest)
		return
	}

	h.HandleWebSocket(w, r, streamKey)
}

// GetRoomStats returns statistics for a specific room
func (h *WSHandler) GetRoomStats(streamKey string) map[string]interface{} {
	h.connMux.RLock()
	defer h.connMux.RUnlock()

	connectedUsers := 0
	for _, conn := range h.connections {
		if conn.StreamKey == streamKey {
			connectedUsers++
		}
	}

	messages := h.manager.GetMessages(streamKey, 0)
	users := h.manager.GetUsers(streamKey)

	return map[string]interface{}{
		"stream_key":      streamKey,
		"connected_users": connectedUsers,
		"total_users":     len(users),
		"message_count":   len(messages),
		"users":           users,
	}
}

// BroadcastSystemMessage broadcasts a system message to a room
func (h *WSHandler) BroadcastSystemMessage(streamKey, message string) {
	h.connMux.RLock()
	defer h.connMux.RUnlock()

	msg := WSMessage{
		Type: "system",
		Data: map[string]interface{}{
			"message": message,
		},
		Timestamp: time.Now(),
	}

	for _, conn := range h.connections {
		if conn.StreamKey == streamKey {
			select {
			case conn.Send <- msg:
			default:
			}
		}
	}
}
