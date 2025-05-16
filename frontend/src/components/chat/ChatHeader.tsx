// frontend/src/components/chat/ChatHeader.tsx
import { UserProfile } from '@/types/User'; // Assuming UserProfile contains necessary details
import { Avatar } from '@/components/ui/avatar';
import Link from 'next/link';

interface ChatHeaderProps {
  targetUser: UserProfile | null; // Or a simpler User type if full profile isn't needed
}

export default function ChatHeader({ targetUser }: ChatHeaderProps) {
  if (!targetUser) {
    return (
      <div className="p-4 border-b border-gray-700 bg-gray-800 text-white text-center">
        Loading user...
      </div>
    );
  }

  const displayName = `${targetUser.first_name} ${targetUser.last_name}`;

  return (
    <div className="p-3 border-b border-gray-700 bg-gray-800 flex items-center space-x-3">
      <Avatar
        className="w-10 h-10"
        src={targetUser.avatar_url}
        alt={displayName}
        initials={`${targetUser.first_name?.[0]?.toUpperCase() ?? ''}${targetUser.last_name?.[0]?.toUpperCase() ?? ''}`}
      />
      <div>
        <Link href={`/profile/${targetUser.id}`} className="hover:underline">
          <h2 className="font-semibold text-lg text-white">{displayName}</h2>
        </Link>
        {/* Optionally, display online status if available */}
        {/* <p className="text-xs text-green-400">Online</p> */}
      </div>
    </div>
  );
}