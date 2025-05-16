'use client';

// frontend/src/components/chat/MessageItem.tsx
import { Message } from '@/types/Message';
import { Avatar } from '@/components/ui/avatar';
import { format } from 'date-fns'; // For formatting timestamps
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';

interface MessageItemProps {
  message: Message;
  isCurrentUserSender: boolean;
}

export default function MessageItem({ message, isCurrentUserSender }: MessageItemProps) {
  const { onlineUserIds } = useGlobalWebSocket();
  const alignment = isCurrentUserSender ? 'justify-end' : 'justify-start';
  const bgColor = isCurrentUserSender ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-200';
  const senderName = isCurrentUserSender ? 'You' : message.sender_username || 'User';

  return (
    <div className={`flex ${alignment} mb-3`}>
      {!isCurrentUserSender && (
        <div className="relative">
          <Avatar
            className="w-6 h-6 mr-2"
            src={message.sender_avatar_url}
            alt={senderName}
            initials={senderName.substring(0, 1).toUpperCase()}
            userId={message.sender_id}
          />
        </div>
      )}
      <div className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg shadow ${bgColor}`}>
        {!isCurrentUserSender && (
          <p className="text-xs text-gray-400 mb-1">{senderName}</p>
        )}
        <p className="text-lg">{message.content}</p>
        <p className={`text-xs mt-1 ${isCurrentUserSender ? 'text-blue-200' : 'text-gray-400'} text-right`}>
          {format(new Date(message.created_at), 'p')}
        </p>
      </div>
      {isCurrentUserSender && (
         <div className="relative">
          <Avatar
            className="w-6 h-6 ml-2"
            src={message.sender_avatar_url} // Assuming current user's avatar is part of the message object
            alt={senderName}
            initials="Y"
            userId={message.sender_id}
          />
          {/* Current user's own avatar in chat doesn't need an indicator based on onlineUserIds,
              as that list is for *other* users. If we want to show current user as "online",
              it would be based on connectionStatus, but that's usually for a global avatar (like in a dropdown),
              not for every message they send. For simplicity, we'll omit the dot for the current sender's messages.
           */}
        </div>
      )}
    </div>
  );
}