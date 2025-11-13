# Chat System Memory Capacity

## Default Configuration (100 MB)

### Per-Message Memory Usage
```
- ID (UUID):          36 bytes
- StreamKey:          50 bytes (avg)
- UserID:             36 bytes
- Username:           30 bytes (avg)
- Message text:      250 bytes (avg, max 500)
- Timestamp:           8 bytes
- JSON overhead:      50 bytes
- Go struct overhead: 40 bytes
─────────────────────────────
Total per message:   ~500 bytes
```

### Per-User Memory Usage
```
- UserID:            36 bytes
- Username:          30 bytes
- Timestamps:        24 bytes (3x time.Time)
- Counters:          32 bytes
- Flags:             16 bytes
- Overhead:          62 bytes
─────────────────────────────
Total per user:     ~200 bytes
```

### Per-Stream Memory Usage
```
With default limits (500 messages, 100 users):

Messages:  500 × 500 bytes  = 250 KB
Users:     100 × 200 bytes  =  20 KB
Overhead:  (maps, mutexes)  =   5 KB
─────────────────────────────────────
Total per stream:            ~275 KB
```

## Capacity Calculations

### Scenario 1: Low Traffic (100 active streams)
```
Total Memory Usage:
- 100 streams × 275 KB = 27.5 MB
- Available headroom: 72.5 MB (72%)

Messages per stream: 500
Total messages: 50,000
At 10 msg/min: 50 minutes of history ✅
```

### Scenario 2: Medium Traffic (300 active streams)
```
Total Memory Usage:
- 300 streams × 275 KB = 82.5 MB
- Available headroom: 17.5 MB (17%)

Messages per stream: 500
Total messages: 150,000
At 10 msg/min: 50 minutes of history ✅
```

### Scenario 3: High Traffic (500 active streams)
```
Total Memory Usage:
- 500 streams × 275 KB = 137.5 MB ⚠️ EXCEEDS LIMIT

Auto-adjustment kicks in:
- Reduce to 350 messages per stream
- 500 streams × 192 KB = 96 MB ✅
- 35 minutes of history still available
```

### Scenario 4: Very High Traffic (1000 active streams)
```
Total Memory Usage:
- 1000 streams × 275 KB = 275 MB ⚠️ EXCEEDS LIMIT

Auto-adjustment kicks in:
- Reduce to 175 messages per stream
- 1000 streams × 100 KB = 97.5 MB ✅
- 17 minutes of history (still acceptable for live chat)
```

## Messages per Stream by Configuration

| Memory Limit | Active Streams | Messages/Stream | History Duration* |
|--------------|----------------|-----------------|-------------------|
| 50 MB        | 100            | 250             | 25 minutes        |
| 50 MB        | 300            | 80              | 8 minutes         |
| 100 MB       | 100            | 500             | 50 minutes        |
| 100 MB       | 300            | 500             | 50 minutes        |
| 100 MB       | 500            | 350             | 35 minutes        |
| 100 MB       | 1000           | 175             | 17 minutes        |
| 200 MB       | 500            | 700             | 70 minutes        |
| 200 MB       | 1000           | 350             | 35 minutes        |

*Assumes 10 messages per minute average

## Total Message Capacity

```
At 500 bytes per message:

100 MB = 209,715 total messages

Distribution examples:
- 100 streams:  2,097 messages each (3.5 hours @ 10 msg/min)
- 300 streams:    699 messages each (1.2 hours @ 10 msg/min)
- 500 streams:    419 messages each (42 min @ 10 msg/min)
- 1000 streams:   209 messages each (21 min @ 10 msg/min)
```

## Circular Buffer Behavior

The circular buffer ensures memory never exceeds limits:

```
Stream with max 500 messages:

Message 1   → [1, _, _, ..., _]        (size: 1)
Message 2   → [1, 2, _, ..., _]        (size: 2)
...
Message 500 → [1, 2, 3, ..., 500]      (size: 500)
Message 501 → [501, 2, 3, ..., 500]    (size: 500, oldest dropped)
Message 502 → [501, 502, 3, ..., 500]  (size: 500, oldest dropped)

Memory usage: CONSTANT at 500 messages ✅
```

## Memory Safety Features

### Layer 1: Per-Stream Limits
- Max 500 messages per stream (configurable)
- Max 100 users per stream (configurable)
- Circular buffer automatically drops oldest

### Layer 2: Global Memory Tracking
- Monitors total memory usage
- Calculates: total_bytes / max_bytes
- Reports usage percentage

### Layer 3: Cleanup Jobs
- Every 5 minutes: Remove messages > 30 min old
- Every 10 minutes: Close streams with 0 users
- Recalculate memory usage after cleanup

### Layer 4: Auto-Adjustment
- If memory > 80%: Reduce message retention to 20 min
- If memory > 90%: Aggressive cleanup (10 min retention)
- If memory > 95%: Emergency mode (5 min retention)

### Layer 5: Hard Limits
- Cannot exceed configured memory limit
- New messages rejected if at capacity
- Users see "Chat is full" message

## Configuration Examples

### Small Deployment (50 MB)
```bash
CHAT_MAX_MEMORY_MB=50
CHAT_MAX_MESSAGES_PER_STREAM=250
CHAT_MAX_USERS_PER_STREAM=50
```
Good for: <100 concurrent streams

### Default (100 MB)
```bash
CHAT_MAX_MEMORY_MB=100
CHAT_MAX_MESSAGES_PER_STREAM=500
CHAT_MAX_USERS_PER_STREAM=100
```
Good for: 100-500 concurrent streams

### Large Deployment (200 MB)
```bash
CHAT_MAX_MEMORY_MB=200
CHAT_MAX_MESSAGES_PER_STREAM=1000
CHAT_MAX_USERS_PER_STREAM=200
```
Good for: 500-1000 concurrent streams

### Enterprise (500 MB)
```bash
CHAT_MAX_MEMORY_MB=500
CHAT_MAX_MESSAGES_PER_STREAM=2000
CHAT_MAX_USERS_PER_STREAM=500
```
Good for: 1000+ concurrent streams

## Monitoring Endpoints

The chat system will expose these metrics:

```json
GET /api/chat/stats

{
  "memory": {
    "total_mb": 45.2,
    "max_mb": 100,
    "usage_percent": 45.2,
    "is_near_limit": false,
    "is_critical": false
  },
  "messages": {
    "total_count": 95124,
    "total_streams": 287,
    "avg_per_stream": 331
  },
  "users": {
    "total_connected": 1523,
    "avg_per_stream": 5.3
  }
}
```

## Recommendations

### For most deployments:
- **100 MB** is plenty for up to 500 concurrent streams
- Most streams have <50 active viewers
- Average chat rate: 5-15 messages/minute
- 500 messages = 30-50 minutes of history

### When to increase memory:
- Consistently hitting >80% memory usage
- Multiple very active streams (100+ viewers each)
- Want longer message history (>1 hour)

### When to decrease memory:
- Running on resource-constrained server
- Few concurrent streams (<50)
- Want to conserve memory for video processing

## Real-World Examples

### Twitch-like Platform
```
Expected: 1000 active streams
Peak viewers per stream: 50
Peak message rate: 20 msg/min

Configuration:
CHAT_MAX_MEMORY_MB=200
CHAT_MAX_MESSAGES_PER_STREAM=500

Result: 25 minutes of history, comfortable headroom
```

### Small Community Platform
```
Expected: 50 active streams
Peak viewers per stream: 20
Peak message rate: 10 msg/min

Configuration:
CHAT_MAX_MEMORY_MB=50
CHAT_MAX_MESSAGES_PER_STREAM=500

Result: 50 minutes of history, excellent performance
```

### Personal Streaming Server
```
Expected: 1-5 active streams
Peak viewers per stream: 10
Peak message rate: 5 msg/min

Configuration:
CHAT_MAX_MEMORY_MB=25
CHAT_MAX_MESSAGES_PER_STREAM=250

Result: 50 minutes of history, minimal resource usage
```
