import React, { useState, useRef, useEffect } from 'react';
import { formatTimeout } from '../../utils/time';

interface MessageInputProps {
  onSendMessage: (message: string) => void;
  onTyping?: (isTyping: boolean) => void;
  isTimeout: boolean;
  timeoutDuration: number;
  disabled?: boolean;
  maxLength?: number;
}

const MessageInput: React.FC<MessageInputProps> = ({
  onSendMessage,
  onTyping,
  isTimeout,
  timeoutDuration,
  disabled = false,
  maxLength = 500,
}) => {
  const [message, setMessage] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const typingTimeoutRef = useRef<number | undefined>(undefined);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;

    // Enforce max length
    if (value.length > maxLength) {
      return;
    }

    setMessage(value);

    // Send typing indicator
    if (onTyping && value.length > 0 && !isTyping) {
      setIsTyping(true);
      onTyping(true);
    }

    // Clear previous timeout
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    // Stop typing after 2 seconds of no input
    typingTimeoutRef.current = window.setTimeout(() => {
      if (isTyping) {
        setIsTyping(false);
        onTyping?.(false);
      }
    }, 2000);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (isTimeout || disabled || !message.trim()) {
      return;
    }

    onSendMessage(message);
    setMessage('');

    // Stop typing indicator
    if (isTyping) {
      setIsTyping(false);
      onTyping?.(false);
    }

    // Clear typing timeout
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Send on Enter (but not Shift+Enter for mobile compatibility)
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e as any);
    }
  };

  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, []);

  const charCount = message.length;
  const isNearLimit = charCount > maxLength * 0.8;
  const isAtLimit = charCount >= maxLength;

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-1 p-2 bg-blue-950 border-t border-blue-800">
      {isTimeout && (
        <div className="text-xs text-red-400 px-2 py-1 bg-red-900/30 rounded">
          Timed out. You can chat again in {formatTimeout(timeoutDuration)}
        </div>
      )}

      <div className="flex gap-2">
        <input
          ref={inputRef}
          type="text"
          value={message}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={isTimeout ? 'Timed out...' : 'Type a message...'}
          disabled={disabled || isTimeout}
          className={`
            flex-1 px-3 py-2 md:py-1.5
            bg-blue-900 text-white rounded
            border border-blue-700
            focus:outline-none focus:border-blue-500
            placeholder-blue-400
            disabled:opacity-50 disabled:cursor-not-allowed
            text-sm
          `}
          maxLength={maxLength}
        />

        <button
          type="submit"
          disabled={disabled || isTimeout || !message.trim()}
          className={`
            px-4 py-2 md:py-1.5
            bg-blue-600 hover:bg-blue-500
            disabled:bg-blue-800 disabled:cursor-not-allowed disabled:opacity-50
            text-white font-medium rounded
            transition-colors
            text-sm
            min-w-[60px]
          `}
        >
          Send
        </button>
      </div>

      <div className="flex justify-between items-center px-2">
        <div className="text-xs text-blue-400">
          {isTyping && '✏️ Typing...'}
        </div>
        <div className={`text-xs ${isAtLimit ? 'text-red-400' : isNearLimit ? 'text-yellow-400' : 'text-blue-400'}`}>
          {charCount}/{maxLength}
        </div>
      </div>
    </form>
  );
};

export default MessageInput;
