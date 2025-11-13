import { useEffect, useRef, useState, useCallback } from 'react';
import { getUserId, getUsername } from '../utils/username';

export interface ChatMessage {
  id: string;
  streamKey: string;
  userId: string;
  username: string;
  message: string;
  timestamp: string;
}

export interface ChatUser {
  userId: string;
  username: string;
  connectedAt: string;
  isActive: boolean;
}

interface UseChatOptions {
  streamKey: string;
  enabled?: boolean;
}

export function useChat({ streamKey, enabled = true }: UseChatOptions) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [users, setUsers] = useState<ChatUser[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isTimeout, setIsTimeout] = useState(false);
  const [timeoutDuration, setTimeoutDuration] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | undefined>(undefined);
  const reconnectAttemptsRef = useRef(0);

  const userId = getUserId();
  const username = getUsername();

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (!enabled || !streamKey) return;

    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/api/chat?streamKey=${streamKey}`;

      console.log('Connecting to chat:', wsUrl);
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('Chat connected');
        setIsConnected(true);
        setError(null);
        reconnectAttemptsRef.current = 0;

        // Join the chat
        ws.send(JSON.stringify({
          type: 'join',
          data: {
            userId,
            username,
          },
        }));
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          handleMessage(data);
        } catch (e) {
          console.error('Failed to parse message:', e);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setError('Connection error');
      };

      ws.onclose = () => {
        console.log('Chat disconnected');
        setIsConnected(false);

        // Attempt to reconnect with exponential backoff
        if (enabled && reconnectAttemptsRef.current < 5) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 10000);
          console.log(`Reconnecting in ${delay}ms...`);

          reconnectTimeoutRef.current = window.setTimeout(() => {
            reconnectAttemptsRef.current++;
            connect();
          }, delay);
        }
      };

      wsRef.current = ws;
    } catch (e) {
      console.error('Failed to connect to chat:', e);
      setError('Failed to connect');
    }
  }, [enabled, streamKey, userId, username]);

  // Handle incoming messages
  const handleMessage = (data: any) => {
    switch (data.type) {
      case 'history':
        // Received message history on connect
        setMessages(data.data || []);
        break;

      case 'message':
        // New message received
        setMessages((prev) => [...prev, data.data]);
        break;

      case 'users':
        // User list updated
        setUsers(data.data || []);
        break;

      case 'user_joined':
        // User joined
        setUsers((prev) => {
          const exists = prev.some(u => u.userId === data.data.userId);
          if (exists) return prev;
          return [...prev, {
            userId: data.data.userId,
            username: data.data.username,
            connectedAt: new Date().toISOString(),
            isActive: true,
          }];
        });
        break;

      case 'user_left':
        // User left
        setUsers((prev) => prev.filter(u => u.userId !== data.data.userId));
        break;

      case 'timeout':
        // User is timed out
        setIsTimeout(true);
        setTimeoutDuration(data.data.duration || 0);
        break;

      case 'rate_limit':
        // Rate limited
        setError(data.error || 'You are sending messages too quickly');
        setTimeout(() => setError(null), 3000);
        break;

      case 'error':
        // Error message
        setError(data.error || 'An error occurred');
        setTimeout(() => setError(null), 5000);
        break;

      case 'system':
        // System message
        console.log('System:', data.data.message);
        break;

      default:
        console.log('Unknown message type:', data.type);
    }
  };

  // Send a chat message
  const sendMessage = useCallback((message: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      setError('Not connected to chat');
      return false;
    }

    if (!message.trim()) {
      return false;
    }

    try {
      wsRef.current.send(JSON.stringify({
        type: 'message',
        data: { message: message.trim() },
      }));
      return true;
    } catch (e) {
      console.error('Failed to send message:', e);
      setError('Failed to send message');
      return false;
    }
  }, []);

  // Send typing indicator
  const sendTyping = useCallback((isTyping: boolean) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      return;
    }

    try {
      wsRef.current.send(JSON.stringify({
        type: 'typing',
        data: { isTyping },
      }));
    } catch (e) {
      console.error('Failed to send typing indicator:', e);
    }
  }, []);

  // Connect on mount
  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect]);

  // Handle timeout countdown
  useEffect(() => {
    if (isTimeout && timeoutDuration > 0) {
      const interval = setInterval(() => {
        setTimeoutDuration((prev) => {
          if (prev <= 1) {
            setIsTimeout(false);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);

      return () => clearInterval(interval);
    }
  }, [isTimeout, timeoutDuration]);

  return {
    messages,
    users,
    userCount: users.length,
    isConnected,
    isTimeout,
    timeoutDuration,
    error,
    sendMessage,
    sendTyping,
    currentUserId: userId,
    currentUsername: username,
  };
}
