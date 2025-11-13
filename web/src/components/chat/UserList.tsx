import React, { useState } from 'react';
import { ChatUser } from '../../hooks/useChat';

interface UserListProps {
  users: ChatUser[];
  userCount: number;
  currentUserId: string;
  onUserClick?: (username: string) => void;
}

const UserList: React.FC<UserListProps> = ({
  users,
  userCount,
  currentUserId,
  onUserClick,
}) => {
  const [isCollapsed, setIsCollapsed] = useState(false);

  return (
    <div className="border-b border-blue-800 bg-blue-950">
      <button
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="w-full px-3 py-2 flex items-center justify-between hover:bg-blue-900 transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-blue-300">
            👥 Online
          </span>
          <span className="text-xs bg-blue-800 text-blue-200 px-2 py-0.5 rounded-full">
            {userCount}
          </span>
        </div>
        <span className="text-blue-400 text-xs">
          {isCollapsed ? '▼' : '▲'}
        </span>
      </button>

      {!isCollapsed && (
        <div className="max-h-32 md:max-h-48 overflow-y-auto px-3 py-2 space-y-1">
          {users.length === 0 ? (
            <div className="text-xs text-blue-400 text-center py-2">
              No users online
            </div>
          ) : (
            users.map((user) => {
              const isCurrentUser = user.userId === currentUserId;

              return (
                <div
                  key={user.userId}
                  onClick={() => !isCurrentUser && onUserClick?.(user.username)}
                  className={`
                    text-sm px-2 py-1 rounded
                    ${isCurrentUser ? 'bg-green-900/30 text-green-400' : 'text-blue-300 hover:bg-blue-900 cursor-pointer'}
                    transition-colors
                  `}
                >
                  <span className={isCurrentUser ? '⭐ ' : '• '}>
                    {user.username}
                  </span>
                  {isCurrentUser && (
                    <span className="text-xs text-green-500 ml-1">(You)</span>
                  )}
                </div>
              );
            })
          )}
        </div>
      )}
    </div>
  );
};

export default UserList;
