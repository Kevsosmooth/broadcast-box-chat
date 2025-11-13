package chat

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager handles all chat rooms and global operations
type Manager struct {
	config       *ChatConfig
	rooms        map[string]*ChatRoom
	roomsMux     sync.RWMutex
	memTracker   *MemoryTracker
	stopCleanup  chan bool
	stopMonitor  chan bool
}

// NewManager creates a new chat manager
func NewManager(config *ChatConfig) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	manager := &Manager{
		config:      config,
		rooms:       make(map[string]*ChatRoom),
		memTracker:  NewMemoryTracker(config.MaxTotalMemoryMB),
		stopCleanup: make(chan bool),
		stopMonitor: make(chan bool),
	}

	// Start background jobs
	go manager.cleanupWorker()
	go manager.monitorWorker()

	return manager
}

// GetOrCreateRoom gets an existing room or creates a new one
func (m *Manager) GetOrCreateRoom(streamKey string) *ChatRoom {
	m.roomsMux.Lock()
	defer m.roomsMux.Unlock()

	if room, exists := m.rooms[streamKey]; exists {
		return room
	}

	room := NewChatRoom(streamKey, m.config.MaxMessagesPerStream)
	m.rooms[streamKey] = room

	log.Printf("Created chat room for stream: %s", streamKey)
	return room
}

// GetRoom gets an existing room
func (m *Manager) GetRoom(streamKey string) (*ChatRoom, bool) {
	m.roomsMux.RLock()
	defer m.roomsMux.RUnlock()

	room, exists := m.rooms[streamKey]
	return room, exists
}

// AddMessage adds a message to a room
func (m *Manager) AddMessage(streamKey, userID, username, message string) (*ChatMessage, error) {
	room := m.GetOrCreateRoom(streamKey)

	msg := &ChatMessage{
		ID:        uuid.New().String(),
		StreamKey: streamKey,
		UserID:    userID,
		Username:  username,
		Message:   message,
		Timestamp: time.Now(),
	}

	room.AddMessage(*msg)
	return msg, nil
}

// AddUser adds a user to a room
func (m *Manager) AddUser(streamKey, userID, username string) error {
	room := m.GetOrCreateRoom(streamKey)

	// Check user limit
	if room.UserCount() >= m.config.MaxUsersPerStream {
		return ErrRoomFull
	}

	user := &ChatUser{
		UserID:      userID,
		Username:    username,
		ConnectedAt: time.Now(),
		IsActive:    true,
	}

	room.AddUser(user)
	log.Printf("User %s (%s) joined room: %s", username, userID, streamKey)
	return nil
}

// RemoveUser removes a user from a room
func (m *Manager) RemoveUser(streamKey, userID string) {
	room, exists := m.GetRoom(streamKey)
	if !exists {
		return
	}

	room.RemoveUser(userID)
	log.Printf("User %s left room: %s", userID, streamKey)
}

// GetMessages gets messages from a room
func (m *Manager) GetMessages(streamKey string, recentN int) []ChatMessage {
	room, exists := m.GetRoom(streamKey)
	if !exists {
		return []ChatMessage{}
	}

	return room.GetMessages(recentN)
}

// GetUsers gets all users in a room
func (m *Manager) GetUsers(streamKey string) []*ChatUser {
	room, exists := m.GetRoom(streamKey)
	if !exists {
		return []*ChatUser{}
	}

	return room.GetAllUsers()
}

// GetUserCount gets the number of users in a room
func (m *Manager) GetUserCount(streamKey string) int {
	room, exists := m.GetRoom(streamKey)
	if !exists {
		return 0
	}

	return room.UserCount()
}

// cleanupWorker runs periodic cleanup tasks
func (m *Manager) cleanupWorker() {
	ticker := time.NewTicker(time.Duration(m.config.CleanupIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performCleanup()
		case <-m.stopCleanup:
			return
		}
	}
}

// performCleanup cleans up old messages and inactive rooms
func (m *Manager) performCleanup() {
	m.roomsMux.Lock()
	defer m.roomsMux.Unlock()

	retention := time.Duration(m.config.MessageRetentionMinutes) * time.Minute
	totalRemoved := 0
	roomsToDelete := []string{}

	for streamKey, room := range m.rooms {
		// Clean old messages
		removed := room.CleanupOldMessages(retention)
		totalRemoved += removed

		// Mark empty rooms for deletion
		if room.UserCount() == 0 && time.Since(room.LastActivity) > m.config.InactiveStreamTimeout {
			roomsToDelete = append(roomsToDelete, streamKey)
		}
	}

	// Delete inactive rooms
	for _, streamKey := range roomsToDelete {
		delete(m.rooms, streamKey)
		log.Printf("Deleted inactive room: %s", streamKey)
	}

	if totalRemoved > 0 || len(roomsToDelete) > 0 {
		log.Printf("Cleanup: Removed %d messages, deleted %d rooms", totalRemoved, len(roomsToDelete))
	}
}

// monitorWorker monitors memory usage
func (m *Manager) monitorWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.updateMemoryStats()
		case <-m.stopMonitor:
			return
		}
	}
}

// updateMemoryStats updates memory tracking statistics
func (m *Manager) updateMemoryStats() {
	m.roomsMux.RLock()
	defer m.roomsMux.RUnlock()

	var totalBytes int64
	var totalMessages int64

	for _, room := range m.rooms {
		totalBytes += room.BytesUsed
		totalMessages += room.MessageCount
	}

	m.memTracker.Update(totalBytes, totalMessages, len(m.rooms))

	// Log warnings if approaching limits
	if m.memTracker.IsCritical() {
		log.Printf("⚠️ CRITICAL: Chat memory usage at %.1f%%",
			float64(totalBytes)/float64(m.memTracker.MaxBytes)*100)
		m.performEmergencyCleanup()
	} else if m.memTracker.IsNearLimit() {
		log.Printf("⚠️ WARNING: Chat memory usage at %.1f%%",
			float64(totalBytes)/float64(m.memTracker.MaxBytes)*100)
	}
}

// performEmergencyCleanup performs aggressive cleanup when memory is critical
func (m *Manager) performEmergencyCleanup() {
	m.roomsMux.Lock()
	defer m.roomsMux.Unlock()

	log.Println("Performing emergency cleanup...")

	// Reduce retention to 10 minutes for all rooms
	emergencyRetention := 10 * time.Minute
	totalRemoved := 0

	for _, room := range m.rooms {
		removed := room.CleanupOldMessages(emergencyRetention)
		totalRemoved += removed
	}

	log.Printf("Emergency cleanup: Removed %d messages", totalRemoved)
}

// GetStats returns current chat statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.roomsMux.RLock()
	defer m.roomsMux.RUnlock()

	totalUsers := 0
	totalMessages := 0

	for _, room := range m.rooms {
		totalUsers += room.UserCount()
		totalMessages += room.Messages.Size()
	}

	stats := map[string]interface{}{
		"total_rooms":    len(m.rooms),
		"total_users":    totalUsers,
		"total_messages": totalMessages,
		"memory":         m.memTracker.GetStats(),
		"config":         m.config.CalculateCapacity(),
	}

	return stats
}

// Stop stops all background workers
func (m *Manager) Stop() {
	close(m.stopCleanup)
	close(m.stopMonitor)
	log.Println("Chat manager stopped")
}

// Error definitions
var (
	ErrRoomFull = &ChatError{Code: "ROOM_FULL", Message: "Chat room is full"}
	ErrTimeout  = &ChatError{Code: "TIMEOUT", Message: "You are timed out from chat"}
	ErrRateLimit = &ChatError{Code: "RATE_LIMIT", Message: "You are sending messages too quickly"}
)

// ChatError represents a chat error
type ChatError struct {
	Code    string
	Message string
}

func (e *ChatError) Error() string {
	return e.Message
}
