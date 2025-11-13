import React, { useEffect, useRef } from 'react';
import { ChatMessage } from '../../hooks/useChat';
import { formatTimestamp } from '../../utils/time';

interface MessageListProps {
  messages: ChatMessage[];
  currentUserId: string;
  onMentionClick?: (username: string) => void;
}

const MessageList: React.FC<MessageListProps> = ({
  messages,
  currentUserId,
  onMentionClick,
}) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = React.useState(true);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (autoScroll) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, autoScroll]);

  // Check if user is scrolled to bottom
  const handleScroll = () => {
    if (!containerRef.current) return;

    const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;

    setAutoScroll(isAtBottom);
  };

  // Parse message for @mentions
  const renderMessage = (text: string, isMentioned: boolean) => {
    const mentionRegex = /@(\w+)/g;
    const parts = text.split(mentionRegex);

    return parts.map((part, index) => {
      // Odd indices are the captured groups (usernames)
      if (index % 2 === 1) {
        return (
          <span
            key={index}
            onClick={() => onMentionClick?.(part)}
            className="text-blue-300 hover:text-blue-200 cursor-pointer font-semibold"
          >
            @{part}
          </span>
        );
      }
      return <span key={index}>{part}</span>;
    });
  };

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center text-blue-400 text-sm">
        No messages yet. Be the first to say something!
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      onScroll={handleScroll}
      className="flex-1 overflow-y-auto px-3 py-2 space-y-2"
    >
      {messages.map((msg) => {
        const isOwnMessage = msg.userId === currentUserId;
        const isMentioned = msg.message.includes(`@${currentUserId}`);

        return (
          <div
            key={msg.id}
            className={`
              text-sm break-words
              ${isMentioned ? 'bg-yellow-900/20 border-l-2 border-yellow-500 pl-2' : ''}
              ${isOwnMessage ? 'opacity-90' : ''}
            `}
          >
            <div className="flex items-baseline gap-2">
              <span className="text-xs text-blue-400">
                {formatTimestamp(msg.timestamp, false)}
              </span>
              <span
                className={`
                  font-semibold
                  ${isOwnMessage ? 'text-green-400' : 'text-blue-300'}
                `}
              >
                {msg.username}:
              </span>
            </div>
            <div className="text-white mt-0.5">
              {renderMessage(msg.message, isMentioned)}
            </div>
          </div>
        );
      })}
      <div ref={messagesEndRef} />

      {!autoScroll && (
        <button
          onClick={() => {
            setAutoScroll(true);
            messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
          }}
          className="
            fixed bottom-20 right-6
            bg-blue-600 hover:bg-blue-500
            text-white text-xs
            px-3 py-1.5 rounded-full
            shadow-lg
            transition-colors
          "
        >
          ↓ New messages
        </button>
      )}
    </div>
  );
};

export default MessageList;
