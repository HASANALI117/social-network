'use client';

import Link from 'next/link';
import { User } from '@/types/User';
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';
import { useEffect } from 'react';

interface UserCardProps {
  user: User;
  className?: string; // Add className as an optional prop
}

export default function UserCard({ user, className }: UserCardProps) {
  const { onlineUserIds } = useGlobalWebSocket();
  // Combine default classes with any provided className
  const combinedClassName = `flex items-center gap-3 p-3 rounded-lg hover:bg-gray-800 transition-colors ${className || ''}`;

  useEffect(() => {
    console.log({onlineUserIds});
  }, [onlineUserIds]);
    
  return (
    <Link
      href={`/profile/${user.id}`}
      className={combinedClassName.trim()} // Apply combined classes
    >
      <img
        src={user.avatar_url || `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=3b82f6&color=fff&bold=true`}
        alt={`${user.first_name} ${user.last_name}`}
        className="w-10 h-10 rounded-full"
      />
      <div>
        <div className="flex items-center">
          <div className="font-medium text-gray-100">
            {user.first_name} {user.last_name}
          </div>
          {onlineUserIds.includes(user.id) && (
            <span className="h-2 w-2 bg-green-500 rounded-full ml-2 flex-shrink-0"></span>
          )}
        </div>
        <div className="text-sm text-gray-400">@{user.username}</div>
      </div>
    </Link>
  );
}