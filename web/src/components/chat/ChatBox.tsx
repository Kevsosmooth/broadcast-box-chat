import React, { useState, useEffect } from 'react';
import { useChat } from '../../hooks/useChat';
import MessageList from './MessageList';
import MessageInput from './MessageInput';
import UserList from './UserList';

interface ChatBoxProps {
  streamKey: string;
  isFullscreen?: boolean;
  className?: string;
}

const ChatBox: React.FC<ChatBoxProps> = ({
  streamKey,
  isFullscreen = false,
  className = '',
}) => {
  const [isVisible, setIsVisible] = useState(!isFullscreen);
  const [inputMessage, setInputMessage] = useState('');

  const {
    messages,
    users,
    userCount,
    isConnected,
    isTimeout,
    timeoutDuration,
    error,
    sendMessage,
    sendTyping,
    currentUserId,
    currentUsername,
  } = useChat({ streamKey, enabled: true });

  // Show/hide chat in fullscreen mode
  useEffect(() => {
    if (isFullscreen) {
      // In fullscreen, chat starts hidden
      setIsVisible(false);
    } else {
      // In normal mode, always visible
      setIsVisible(true);
    }
  }, [isFullscreen]);

  const handleSendMessage = (message: string) => {
    sendMessage(message);
  };

  const handleMentionClick = (username: string) => {
    setInputMessage(`@${username} `);
  };

  // Fullscreen overlay mode (transparent background)
  if (isFullscreen) {
    return (
      <>
        {/* Toggle button */}
        <button
          onClick={() => setIsVisible(!isVisible)}
          className="
            fixed top-4 right-4 z-50
            bg-blue-600/80 hover:bg-blue-500/80
            backdrop-blur-sm
            text-white px-3 py-2 rounded-lg
            shadow-lg
            transition-all
            text-sm font-medium
          "
        >
          💬 Chat {isVisible ? '×' : `(${userCount})`}
        </button>

        {/* Chat overlay */}
        {isVisible && (
          <div
            className={`
              fixed bottom-0 right-0 z-40
              w-full md:w-96 h-2/3 md:h-1/2
              bg-gradient-to-t from-black/90 via-black/80 to-black/60
              backdrop-blur-md
              border-t md:border-l border-white/10
              flex flex-col
              animate-slide-up
              ${className}
            `}
          >
            {/* Header */}
            <div className="p-3 border-b border-white/10 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className={`
                  w-2 h-2 rounded-full
                  ${isConnected ? 'bg-green-500' : 'bg-red-500'}
                  animate-pulse
                `} />
                <span className="text-white font-semibold text-sm">
                  Chat
                </span>
                <span className="text-xs text-white/60">
                  {userCount} online
                </span>
              </div>
              <button
                onClick={() => setIsVisible(false)}
                className="text-white/60 hover:text-white text-xl leading-none"
              >
                ×
              </button>
            </div>

            {/* Error banner */}
            {error && (
              <div className="bg-red-500/20 border-b border-red-500/30 px-3 py-2 text-red-200 text-xs">
                {error}
              </div>
            )}

            {/* Messages */}
            <div className="flex-1 overflow-hidden">
              <MessageList
                messages={messages}
                currentUserId={currentUserId}
                onMentionClick={handleMentionClick}
              />
            </div>

            {/* Input */}
            <MessageInput
              onSendMessage={handleSendMessage}
              onTyping={sendTyping}
              isTimeout={isTimeout}
              timeoutDuration={timeoutDuration}
              disabled={!isConnected}
            />
          </div>
        )}
      </>
    );
  }

  // Desktop/Mobile normal mode (side panel)
  return (
    <div
      className={`
        flex flex-col
        bg-slate-900 border border-blue-800 rounded-lg overflow-hidden
        ${className}
      `}
    >
      {/* Header */}
      <div className="p-3 bg-blue-950 border-b border-blue-800 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className={`
            w-2 h-2 rounded-full
            ${isConnected ? 'bg-green-500' : 'bg-red-500'}
            ${isConnected && 'animate-pulse'}
          `} />
          <span className="text-white font-semibold text-sm">
            Stream Chat
          </span>
        </div>
        <span className="text-xs text-blue-400">
          {currentUsername}
        </span>
      </div>

      {/* Error banner */}
      {error && (
        <div className="bg-red-900/30 border-b border-red-800 px-3 py-2 text-red-200 text-xs">
          {error}
        </div>
      )}

      {/* User list */}
      <UserList
        users={users}
        userCount={userCount}
        currentUserId={currentUserId}
        onUserClick={handleMentionClick}
      />

      {/* Messages */}
      <div className="flex-1 overflow-hidden flex flex-col min-h-0">
        <MessageList
          messages={messages}
          currentUserId={currentUserId}
          onMentionClick={handleMentionClick}
        />
      </div>

      {/* Input */}
      <MessageInput
        onSendMessage={handleSendMessage}
        onTyping={sendTyping}
        isTimeout={isTimeout}
        timeoutDuration={timeoutDuration}
        disabled={!isConnected}
      />
    </div>
  );
};

export default ChatBox;
