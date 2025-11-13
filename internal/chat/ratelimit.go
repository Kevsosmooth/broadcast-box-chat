package chat

import (
	"strings"
	"sync"
	"time"
)

// RateLimiter handles rate limiting for chat messages
type RateLimiter struct {
	config      *ChatConfig
	userRecords map[string]*UserRateRecord
	mutex       sync.RWMutex
}

// UserRateRecord tracks rate limiting data for a user
type UserRateRecord struct {
	UserID           string
	Messages         []time.Time      // Timestamps of recent messages
	MessageContents  []string         // Content of recent messages for spam detection
	CharCountHistory []int            // Character counts
	TimeoutUntil     time.Time
	Violations       int
	LastCleanup      time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *ChatConfig) *RateLimiter {
	rl := &RateLimiter{
		config:      config,
		userRecords: make(map[string]*UserRateRecord),
	}

	// Start cleanup worker
	go rl.cleanupWorker()

	return rl
}

// CheckMessage checks if a message is allowed based on rate limits
func (rl *RateLimiter) CheckMessage(userID, message string) (bool, *ChatError) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	record := rl.getOrCreateRecord(userID)
	now := time.Now()

	// Check if user is timed out
	if now.Before(record.TimeoutUntil) {
		return false, &ChatError{
			Code:    "TIMEOUT",
			Message: "You are timed out. Please wait before sending messages.",
		}
	}

	// Clean old entries
	record.cleanup()

	messageLen := len(message)

	// Check message length
	if messageLen > rl.config.MaxCharactersPerMessage {
		return false, &ChatError{
			Code:    "MESSAGE_TOO_LONG",
			Message: "Message is too long. Maximum 500 characters.",
		}
	}

	// Tier 1: Basic frequency check (5 messages per 10 seconds)
	recentMessages := record.countMessagesInWindow(10 * time.Second)
	if recentMessages >= 5 {
		record.applyTimeout(30 * time.Second)
		record.Violations++
		return false, &ChatError{
			Code:    "RATE_LIMIT",
			Message: "Slow down! (30 second cooldown)",
		}
	}

	// Tier 2: Spam detection (10+ messages in 30 seconds)
	messagesIn30s := record.countMessagesInWindow(30 * time.Second)
	if messagesIn30s >= 10 {
		record.applyTimeout(2 * time.Minute)
		record.Violations++
		return false, &ChatError{
			Code:    "SPAM_DETECTED",
			Message: "Spam detected. (2 minute timeout)",
		}
	}

	// Tier 2.5: Heavy spam (20+ messages in 60 seconds)
	messagesIn60s := record.countMessagesInWindow(60 * time.Second)
	if messagesIn60s >= 20 {
		record.applyTimeout(5 * time.Minute)
		record.Violations += 2
		return false, &ChatError{
			Code:    "HEAVY_SPAM",
			Message: "Heavy spam detected. (5 minute timeout)",
		}
	}

	// Tier 3: Character-based rate limiting
	if messageLen > 300 {
		// Large messages (300-500 chars): max 1 per 10 seconds
		if recentMessages >= 1 {
			return false, &ChatError{
				Code:    "RATE_LIMIT_LONG_MESSAGE",
				Message: "Large messages limited to 1 per 10 seconds.",
			}
		}
	} else if messageLen > 100 {
		// Medium messages (100-300 chars): max 3 per 10 seconds
		if recentMessages >= 3 {
			return false, &ChatError{
				Code:    "RATE_LIMIT_MEDIUM_MESSAGE",
				Message: "Medium messages limited to 3 per 10 seconds.",
			}
		}
	}

	// Tier 4: Duplicate/similar message detection
	if record.isDuplicateSpam(message) {
		record.applyTimeout(5 * time.Minute)
		record.Violations++
		return false, &ChatError{
			Code:    "DUPLICATE_SPAM",
			Message: "Stop sending the same message repeatedly. (5 minute timeout)",
		}
	}

	// Tier 5: Heavy spam with long messages
	charsIn5Min := record.countCharsInWindow(5 * time.Minute)
	if messageLen >= 400 && charsIn5Min > 2000 {
		record.applyTimeout(10 * time.Minute)
		record.Violations += 2
		return false, &ChatError{
			Code:    "HEAVY_TEXT_SPAM",
			Message: "Too much text too quickly. (10 minute timeout)",
		}
	}

	// Escalating penalties for repeat offenders
	if record.Violations >= 5 {
		record.applyTimeout(30 * time.Minute)
		return false, &ChatError{
			Code:    "REPEAT_OFFENDER",
			Message: "Multiple violations. (30 minute timeout)",
		}
	} else if record.Violations >= 4 {
		record.applyTimeout(10 * time.Minute)
		return false, &ChatError{
			Code:    "REPEAT_OFFENDER",
			Message: "Multiple violations. (10 minute timeout)",
		}
	} else if record.Violations >= 3 {
		record.applyTimeout(5 * time.Minute)
		return false, &ChatError{
			Code:    "REPEAT_OFFENDER",
			Message: "Multiple violations. (5 minute timeout)",
		}
	}

	// Message is allowed - record it
	record.recordMessage(message, messageLen)

	return true, nil
}

// getOrCreateRecord gets or creates a rate record for a user
func (rl *RateLimiter) getOrCreateRecord(userID string) *UserRateRecord {
	if record, exists := rl.userRecords[userID]; exists {
		return record
	}

	record := &UserRateRecord{
		UserID:           userID,
		Messages:         make([]time.Time, 0),
		MessageContents:  make([]string, 0),
		CharCountHistory: make([]int, 0),
		LastCleanup:      time.Now(),
	}

	rl.userRecords[userID] = record
	return record
}

// recordMessage records a sent message
func (r *UserRateRecord) recordMessage(content string, charCount int) {
	now := time.Now()
	r.Messages = append(r.Messages, now)
	r.MessageContents = append(r.MessageContents, content)
	r.CharCountHistory = append(r.CharCountHistory, charCount)
}

// countMessagesInWindow counts messages within a time window
func (r *UserRateRecord) countMessagesInWindow(window time.Duration) int {
	cutoff := time.Now().Add(-window)
	count := 0

	for _, timestamp := range r.Messages {
		if timestamp.After(cutoff) {
			count++
		}
	}

	return count
}

// countCharsInWindow counts total characters sent within a time window
func (r *UserRateRecord) countCharsInWindow(window time.Duration) int {
	cutoff := time.Now().Add(-window)
	totalChars := 0

	for i, timestamp := range r.Messages {
		if timestamp.After(cutoff) && i < len(r.CharCountHistory) {
			totalChars += r.CharCountHistory[i]
		}
	}

	return totalChars
}

// isDuplicateSpam checks if message is a duplicate/similar to recent messages
func (r *UserRateRecord) isDuplicateSpam(message string) bool {
	if len(r.MessageContents) < 3 {
		return false
	}

	// Check last 5 messages
	recentCount := len(r.MessageContents)
	if recentCount > 5 {
		recentCount = 5
	}

	duplicateCount := 0
	normalizedMessage := strings.ToLower(strings.TrimSpace(message))

	for i := len(r.MessageContents) - recentCount; i < len(r.MessageContents); i++ {
		recentMsg := strings.ToLower(strings.TrimSpace(r.MessageContents[i]))

		// Exact match
		if recentMsg == normalizedMessage {
			duplicateCount++
		} else if similarity(recentMsg, normalizedMessage) > 0.8 {
			// 80% similar
			duplicateCount++
		}
	}

	// 3 or more similar messages in recent history = spam
	return duplicateCount >= 3
}

// similarity calculates simple similarity between two strings (0.0 to 1.0)
func similarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple character-based similarity
	longer := s1
	shorter := s2
	if len(s2) > len(s1) {
		longer = s2
		shorter = s1
	}

	matches := 0
	for i := 0; i < len(shorter); i++ {
		if i < len(longer) && shorter[i] == longer[i] {
			matches++
		}
	}

	return float64(matches) / float64(len(longer))
}

// applyTimeout applies a timeout to the user
func (r *UserRateRecord) applyTimeout(duration time.Duration) {
	r.TimeoutUntil = time.Now().Add(duration)
}

// cleanup removes old message records
func (r *UserRateRecord) cleanup() {
	now := time.Now()

	// Only cleanup if it's been more than 1 minute since last cleanup
	if now.Sub(r.LastCleanup) < time.Minute {
		return
	}

	cutoff := now.Add(-5 * time.Minute)

	// Remove messages older than 5 minutes
	newMessages := make([]time.Time, 0)
	newContents := make([]string, 0)
	newCharCounts := make([]int, 0)

	for i, timestamp := range r.Messages {
		if timestamp.After(cutoff) {
			newMessages = append(newMessages, timestamp)
			if i < len(r.MessageContents) {
				newContents = append(newContents, r.MessageContents[i])
			}
			if i < len(r.CharCountHistory) {
				newCharCounts = append(newCharCounts, r.CharCountHistory[i])
			}
		}
	}

	r.Messages = newMessages
	r.MessageContents = newContents
	r.CharCountHistory = newCharCounts
	r.LastCleanup = now
}

// cleanupWorker periodically cleans up old user records
func (rl *RateLimiter) cleanupWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.performCleanup()
	}
}

// performCleanup removes inactive user records
func (rl *RateLimiter) performCleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	toDelete := []string{}

	for userID, record := range rl.userRecords {
		// Remove users inactive for more than 30 minutes
		if len(record.Messages) == 0 ||
		   (len(record.Messages) > 0 && now.Sub(record.Messages[len(record.Messages)-1]) > 30*time.Minute) {
			toDelete = append(toDelete, userID)
		}
	}

	for _, userID := range toDelete {
		delete(rl.userRecords, userID)
	}
}

// GetTimeoutStatus returns the timeout status for a user
func (rl *RateLimiter) GetTimeoutStatus(userID string) (bool, time.Duration) {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	record, exists := rl.userRecords[userID]
	if !exists {
		return false, 0
	}

	now := time.Now()
	if now.Before(record.TimeoutUntil) {
		remaining := record.TimeoutUntil.Sub(now)
		return true, remaining
	}

	return false, 0
}
