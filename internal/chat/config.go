package chat

import (
	"os"
	"strconv"
	"time"
)

// ChatConfig holds all configurable chat settings
type ChatConfig struct {
	// Memory limits
	MaxTotalMemoryMB     int // Default: 100 MB
	MaxMessagesPerStream int // Default: 500 messages
	MaxUsersPerStream    int // Default: 100 users

	// Time limits
	MessageRetentionMinutes int           // Default: 30 minutes
	CleanupIntervalMinutes  int           // Default: 5 minutes
	InactiveStreamTimeout   time.Duration // Default: 10 minutes

	// Rate limiting
	MaxMessagesPerMinute     int // Default: 10
	MaxCharactersPerMessage  int // Default: 500
	SpamThresholdMessages    int // Default: 20
	SpamTimeoutMinutes       int // Default: 5

	// Features
	EnableViewerList    bool // Default: true
	EnableMentions      bool // Default: true
	EnableTypingStatus  bool // Default: false
	EnableEmojis        bool // Default: true
}

// DefaultConfig returns the default chat configuration
func DefaultConfig() *ChatConfig {
	return &ChatConfig{
		// Memory limits
		MaxTotalMemoryMB:     100,
		MaxMessagesPerStream: 500,
		MaxUsersPerStream:    100,

		// Time limits
		MessageRetentionMinutes: 30,
		CleanupIntervalMinutes:  5,
		InactiveStreamTimeout:   10 * time.Minute,

		// Rate limiting
		MaxMessagesPerMinute:    10,
		MaxCharactersPerMessage: 500,
		SpamThresholdMessages:   20,
		SpamTimeoutMinutes:      5,

		// Features
		EnableViewerList:   true,
		EnableMentions:     true,
		EnableTypingStatus: false,
		EnableEmojis:       true,
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *ChatConfig {
	config := DefaultConfig()

	// Memory limits
	if val := os.Getenv("CHAT_MAX_MEMORY_MB"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxTotalMemoryMB = parsed
		}
	}

	if val := os.Getenv("CHAT_MAX_MESSAGES_PER_STREAM"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxMessagesPerStream = parsed
		}
	}

	if val := os.Getenv("CHAT_MAX_USERS_PER_STREAM"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxUsersPerStream = parsed
		}
	}

	// Time limits
	if val := os.Getenv("CHAT_MESSAGE_RETENTION_MINUTES"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MessageRetentionMinutes = parsed
		}
	}

	if val := os.Getenv("CHAT_CLEANUP_INTERVAL_MINUTES"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.CleanupIntervalMinutes = parsed
		}
	}

	// Rate limiting
	if val := os.Getenv("CHAT_MAX_MESSAGES_PER_MINUTE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxMessagesPerMinute = parsed
		}
	}

	if val := os.Getenv("CHAT_MAX_CHARACTERS_PER_MESSAGE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxCharactersPerMessage = parsed
		}
	}

	if val := os.Getenv("CHAT_SPAM_THRESHOLD_MESSAGES"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.SpamThresholdMessages = parsed
		}
	}

	if val := os.Getenv("CHAT_SPAM_TIMEOUT_MINUTES"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.SpamTimeoutMinutes = parsed
		}
	}

	// Features
	if val := os.Getenv("CHAT_ENABLE_VIEWER_LIST"); val != "" {
		config.EnableViewerList = val == "true"
	}

	if val := os.Getenv("CHAT_ENABLE_MENTIONS"); val != "" {
		config.EnableMentions = val == "true"
	}

	if val := os.Getenv("CHAT_ENABLE_TYPING_STATUS"); val != "" {
		config.EnableTypingStatus = val == "true"
	}

	if val := os.Getenv("CHAT_ENABLE_EMOJIS"); val != "" {
		config.EnableEmojis = val == "true"
	}

	return config
}

// CalculateCapacity returns estimated capacity based on config
func (c *ChatConfig) CalculateCapacity() map[string]interface{} {
	avgMessageSize := 500 // bytes
	avgUserSize := 200    // bytes

	totalMemoryBytes := c.MaxTotalMemoryMB * 1024 * 1024

	// Reserve memory for user data
	userMemoryPerStream := c.MaxUsersPerStream * avgUserSize

	// Calculate message capacity
	messageMemoryPerStream := c.MaxMessagesPerStream * avgMessageSize
	totalPerStream := messageMemoryPerStream + userMemoryPerStream

	maxStreams := totalMemoryBytes / totalPerStream

	return map[string]interface{}{
		"max_memory_mb":           c.MaxTotalMemoryMB,
		"max_messages_per_stream": c.MaxMessagesPerStream,
		"max_users_per_stream":    c.MaxUsersPerStream,
		"estimated_max_streams":   maxStreams,
		"total_message_capacity":  maxStreams * c.MaxMessagesPerStream,
		"avg_message_size_bytes":  avgMessageSize,
		"memory_per_stream_kb":    totalPerStream / 1024,
	}
}
