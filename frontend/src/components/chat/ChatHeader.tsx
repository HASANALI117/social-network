// frontend/src/components/chat/ChatHeader.tsx
import { UserProfile } from '@/types/User';
import { Group } from '@/types/Group';
import { Avatar } from '@/components/ui/avatar';
import Link from 'next/link';
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';
import { UsersIcon } from '@heroicons/react/24/outline';

interface ChatHeaderProps {
  type: 'direct' | 'group';
  target: UserProfile | Group | null;
  onlineMemberCount?: number;
  totalMemberCount?: number;
}

export default function ChatHeader({ type, target, onlineMemberCount, totalMemberCount }: ChatHeaderProps) {
  const { onlineUserIds } = useGlobalWebSocket();
  
  if (!target) {
    return (
      <div className="sticky top-0 z-10 p-4 border-b border-gray-700 bg-gray-900 text-white text-center">
        Loading {type === 'group' ? 'group' : 'user'}...
      </div>
    );
  }

  const isGroup = type === 'group';
  const group = isGroup ? target as Group : null;
  const user = !isGroup ? target as UserProfile : null;

  const displayName = isGroup
    ? group?.name || 'Unnamed Group'
    : user ? `${user.first_name} ${user.last_name}` : 'Unknown User';

  const getInitials = () => {
    if (isGroup) {
      return (group?.name || '').charAt(0).toUpperCase();
    }
    return user ? `${user.first_name?.[0]?.toUpperCase() ?? ''}${user.last_name?.[0]?.toUpperCase() ?? ''}` : '';
  };

  return (
    <div className="sticky top-0 z-10 bg-gray-900 p-4 flex items-center border-b border-gray-700 space-x-3">
      <Avatar
        className="w-10 h-10"
        src={isGroup ? group?.avatar_url : user?.avatar_url}
        alt={displayName}
        initials={getInitials()}
      />
      <div className="flex-1">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Link
              href={isGroup ? `/groups/${group?.id}` : `/profile/${user?.id}`}
              className="hover:underline"
            >
              <h2 className="font-semibold text-lg text-white">{displayName}</h2>
            </Link>
            {!isGroup && user?.id && onlineUserIds.includes(user.id) && (
              <div className="flex items-center">
                <span className="h-2.5 w-2.5 bg-green-500 rounded-full mr-1.5 flex-shrink-0"></span>
                <span className="text-xs text-green-400">Online</span>
              </div>
            )}
          </div>
          <div className="flex items-center space-x-3">
            {isGroup && typeof onlineMemberCount === 'number' && typeof totalMemberCount === 'number' && (
              <div className="flex items-center text-gray-400">
                <span className="h-2.5 w-2.5 bg-green-500 rounded-full mr-1.5 flex-shrink-0"></span>
                <span className="text-sm">Online: {onlineMemberCount}/{totalMemberCount}</span>
              </div>
            )}
            {isGroup && group?.members_count && typeof onlineMemberCount !== 'number' && ( // Show total members if online count is not available
              <div className="flex items-center text-gray-400">
                <UsersIcon className="w-5 h-5 mr-1" />
                <span className="text-sm">{group.members_count} members</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}