// Time formatting utilities for chat messages

/**
 * Formats a timestamp to a readable format
 * Returns formats like "3:45 PM" or "2m ago"
 */
export function formatTimestamp(timestamp: string | Date, relative: boolean = false): string {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);

  // If less than 1 minute ago and relative mode
  if (relative && diffSeconds < 60) {
    return 'just now';
  }

  // If less than 1 hour ago and relative mode
  if (relative && diffMinutes < 60) {
    return `${diffMinutes}m ago`;
  }

  // If less than 24 hours ago and relative mode
  if (relative && diffHours < 24) {
    return `${diffHours}h ago`;
  }

  // Otherwise return time in 12-hour format
  const hours = date.getHours();
  const minutes = date.getMinutes();
  const ampm = hours >= 12 ? 'PM' : 'AM';
  const displayHours = hours % 12 || 12;
  const displayMinutes = minutes < 10 ? `0${minutes}` : minutes;

  return `${displayHours}:${displayMinutes} ${ampm}`;
}

/**
 * Formats a duration in seconds to readable format
 * Examples: "2:30", "1:05:45"
 */
export function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }

  return `${minutes}:${secs.toString().padStart(2, '0')}`;
}

/**
 * Formats a timeout duration to readable format
 * Examples: "2m 30s", "1h 5m"
 */
export function formatTimeout(seconds: number): string {
  if (seconds < 60) {
    return `${Math.floor(seconds)}s`;
  }

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = Math.floor(seconds % 60);

  if (minutes < 60) {
    if (remainingSeconds > 0) {
      return `${minutes}m ${remainingSeconds}s`;
    }
    return `${minutes}m`;
  }

  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;

  if (remainingMinutes > 0) {
    return `${hours}h ${remainingMinutes}m`;
  }

  return `${hours}h`;
}
