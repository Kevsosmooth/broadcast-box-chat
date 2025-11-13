package chat

import (
	"sync"
	"time"
)

// ChatMessage represents a single chat message
type ChatMessage struct {
	ID        string    `json:"id"`
	StreamKey string    `json:"streamKey"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatUser represents a user in the chat
type ChatUser struct {
	UserID       string
	Username     string
	ConnectedAt  time.Time
	LastMessage  time.Time
	MessageCount int
	CharCount    int64
	TimeoutUntil time.Time
	Violations   int
	IsActive     bool
}

// CircularBuffer implements a fixed-size ring buffer for messages
type CircularBuffer struct {
	data    []ChatMessage
	maxSize int
	head    int
	tail    int
	size    int
	mutex   sync.RWMutex
}

// NewCircularBuffer creates a new circular buffer with the specified size
func NewCircularBuffer(maxSize int) *CircularBuffer {
	return &CircularBuffer{
		data:    make([]ChatMessage, maxSize),
		maxSize: maxSize,
		head:    0,
		tail:    0,
		size:    0,
	}
}

// Add adds a message to the buffer
func (cb *CircularBuffer) Add(msg ChatMessage) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.data[cb.tail] = msg
	cb.tail = (cb.tail + 1) % cb.maxSize

	if cb.size < cb.maxSize {
		cb.size++
	} else {
		// Buffer is full, move head forward
		cb.head = (cb.head + 1) % cb.maxSize
	}
}

// GetAll returns all messages in the buffer (newest first)
func (cb *CircularBuffer) GetAll() []ChatMessage {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.size == 0 {
		return []ChatMessage{}
	}

	result := make([]ChatMessage, cb.size)
	for i := 0; i < cb.size; i++ {
		idx := (cb.head + i) % cb.maxSize
		result[i] = cb.data[idx]
	}

	return result
}

// GetRecent returns the N most recent messages
func (cb *CircularBuffer) GetRecent(n int) []ChatMessage {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.size == 0 {
		return []ChatMessage{}
	}

	count := n
	if count > cb.size {
		count = cb.size
	}

	result := make([]ChatMessage, count)
	startIdx := cb.size - count

	for i := 0; i < count; i++ {
		idx := (cb.head + startIdx + i) % cb.maxSize
		result[i] = cb.data[idx]
	}

	return result
}

// Size returns the current number of messages in the buffer
func (cb *CircularBuffer) Size() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.size
}

// Clear removes all messages from the buffer
func (cb *CircularBuffer) Clear() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.head = 0
	cb.tail = 0
	cb.size = 0
}

// RemoveOlderThan removes messages older than the specified duration
func (cb *CircularBuffer) RemoveOlderThan(duration time.Duration) int {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.size == 0 {
		return 0
	}

	cutoff := time.Now().Add(-duration)
	removed := 0

	// Remove old messages from the head
	for cb.size > 0 {
		msg := cb.data[cb.head]
		if msg.Timestamp.After(cutoff) {
			break
		}

		cb.head = (cb.head + 1) % cb.maxSize
		cb.size--
		removed++
	}

	return removed
}

// ChatRoom represents a chat room for a specific stream
type ChatRoom struct {
	StreamKey    string
	Messages     *CircularBuffer
	Users        map[string]*ChatUser
	LastActivity time.Time
	MessageCount int64
	BytesUsed    int64
	MessagesMux  sync.RWMutex
	UsersMux     sync.RWMutex
}

// NewChatRoom creates a new chat room
func NewChatRoom(streamKey string, maxMessages int) *ChatRoom {
	return &ChatRoom{
		StreamKey:    streamKey,
		Messages:     NewCircularBuffer(maxMessages),
		Users:        make(map[string]*ChatUser),
		LastActivity: time.Now(),
		MessageCount: 0,
		BytesUsed:    0,
	}
}

// AddMessage adds a message to the room
func (cr *ChatRoom) AddMessage(msg ChatMessage) {
	cr.MessagesMux.Lock()
	defer cr.MessagesMux.Unlock()

	cr.Messages.Add(msg)
	cr.LastActivity = time.Now()
	cr.MessageCount++

	// Estimate memory usage
	msgSize := len(msg.ID) + len(msg.StreamKey) + len(msg.UserID) +
		len(msg.Username) + len(msg.Message) + 100 // overhead
	cr.BytesUsed += int64(msgSize)
}

// GetMessages returns all messages or recent N messages
func (cr *ChatRoom) GetMessages(recentN int) []ChatMessage {
	cr.MessagesMux.RLock()
	defer cr.MessagesMux.RUnlock()

	if recentN > 0 {
		return cr.Messages.GetRecent(recentN)
	}
	return cr.Messages.GetAll()
}

// AddUser adds or updates a user in the room
func (cr *ChatRoom) AddUser(user *ChatUser) {
	cr.UsersMux.Lock()
	defer cr.UsersMux.Unlock()

	cr.Users[user.UserID] = user
	cr.LastActivity = time.Now()
}

// RemoveUser removes a user from the room
func (cr *ChatRoom) RemoveUser(userID string) {
	cr.UsersMux.Lock()
	defer cr.UsersMux.Unlock()

	delete(cr.Users, userID)
}

// GetUser returns a user by ID
func (cr *ChatRoom) GetUser(userID string) (*ChatUser, bool) {
	cr.UsersMux.RLock()
	defer cr.UsersMux.RUnlock()

	user, exists := cr.Users[userID]
	return user, exists
}

// GetAllUsers returns all users in the room
func (cr *ChatRoom) GetAllUsers() []*ChatUser {
	cr.UsersMux.RLock()
	defer cr.UsersMux.RUnlock()

	users := make([]*ChatUser, 0, len(cr.Users))
	for _, user := range cr.Users {
		users = append(users, user)
	}
	return users
}

// UserCount returns the number of users in the room
func (cr *ChatRoom) UserCount() int {
	cr.UsersMux.RLock()
	defer cr.UsersMux.RUnlock()

	return len(cr.Users)
}

// CleanupOldMessages removes messages older than the retention period
func (cr *ChatRoom) CleanupOldMessages(retention time.Duration) int {
	cr.MessagesMux.Lock()
	defer cr.MessagesMux.Unlock()

	removed := cr.Messages.RemoveOlderThan(retention)

	// Update byte count estimate
	if removed > 0 {
		avgMessageSize := int64(500)
		cr.BytesUsed -= int64(removed) * avgMessageSize
		if cr.BytesUsed < 0 {
			cr.BytesUsed = 0
		}
	}

	return removed
}

// MemoryTracker tracks global memory usage
type MemoryTracker struct {
	TotalBytes    int64
	TotalMessages int64
	ActiveStreams int
	MaxBytes      int64
	mutex         sync.RWMutex
}

// NewMemoryTracker creates a new memory tracker
func NewMemoryTracker(maxMemoryMB int) *MemoryTracker {
	return &MemoryTracker{
		MaxBytes: int64(maxMemoryMB * 1024 * 1024),
	}
}

// Update updates the memory tracker with current stats
func (mt *MemoryTracker) Update(bytesUsed int64, messageCount int64, streamCount int) {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	mt.TotalBytes = bytesUsed
	mt.TotalMessages = messageCount
	mt.ActiveStreams = streamCount
}

// GetStats returns current memory statistics
func (mt *MemoryTracker) GetStats() map[string]interface{} {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	usagePercent := float64(mt.TotalBytes) / float64(mt.MaxBytes) * 100

	return map[string]interface{}{
		"total_bytes":       mt.TotalBytes,
		"total_mb":          float64(mt.TotalBytes) / 1024 / 1024,
		"max_bytes":         mt.MaxBytes,
		"max_mb":            float64(mt.MaxBytes) / 1024 / 1024,
		"usage_percent":     usagePercent,
		"total_messages":    mt.TotalMessages,
		"active_streams":    mt.ActiveStreams,
		"is_near_limit":     usagePercent > 80,
		"is_critical":       usagePercent > 90,
	}
}

// IsNearLimit returns true if memory usage is above 80%
func (mt *MemoryTracker) IsNearLimit() bool {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	return float64(mt.TotalBytes) / float64(mt.MaxBytes) > 0.8
}

// IsCritical returns true if memory usage is above 90%
func (mt *MemoryTracker) IsCritical() bool {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	return float64(mt.TotalBytes) / float64(mt.MaxBytes) > 0.9
}
