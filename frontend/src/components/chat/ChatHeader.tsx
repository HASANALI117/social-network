// frontend/src/components/chat/ChatHeader.tsx
import { UserProfile } from '@/types/User'; // Assuming UserProfile contains necessary details
import { Avatar } from '@/components/ui/avatar';
import Link from 'next/link';
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';

interface ChatHeaderProps {
  targetUser: UserProfile | null; // Or a simpler User type if full profile isn't needed
}

export default function ChatHeader({ targetUser }: ChatHeaderProps) {
  const { onlineUserIds } = useGlobalWebSocket();
  if (!targetUser) {
    return (
      <div className="sticky top-0 z-10 p-4 border-b border-gray-700 bg-gray-900 text-white text-center">
        Loading user...
      </div>
    );
  }

  const displayName = `${targetUser.first_name} ${targetUser.last_name}`;

  return (
    <div className="sticky top-0 z-10 bg-gray-900 p-4 flex items-center border-b border-gray-700 space-x-3">
      <Avatar
        className="w-10 h-10"
        src={targetUser.avatar_url}
        alt={displayName}
        initials={`${targetUser.first_name?.[0]?.toUpperCase() ?? ''}${targetUser.last_name?.[0]?.toUpperCase() ?? ''}`}
      />
      <div>
        <div className="flex items-center">
          <Link href={`/profile/${targetUser.id}`} className="hover:underline">
            <h2 className="font-semibold text-lg text-white">{displayName}</h2>
          </Link>
          {targetUser && onlineUserIds.includes(targetUser.id) && (
            <div className="flex items-center ml-2">
              <span className="h-2.5 w-2.5 bg-green-500 rounded-full mr-1.5 flex-shrink-0"></span>
              <span className="text-xs text-green-400">Online</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}