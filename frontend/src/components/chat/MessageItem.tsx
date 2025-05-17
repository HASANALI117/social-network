'use client';

// frontend/src/components/chat/MessageItem.tsx
import { Message } from '@/types/Message';
import { Avatar } from '@/components/ui/avatar';
import { format } from 'date-fns'; // For formatting timestamps
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';

interface MessageItemProps {
  message: Message;
  isCurrentUserSender: boolean;
  showSenderAlways?: boolean; // For group chats, we always show sender info
}

export default function MessageItem({ message, isCurrentUserSender, showSenderAlways = false }: MessageItemProps) {
  const { onlineUserIds } = useGlobalWebSocket();
  const alignment = isCurrentUserSender ? 'justify-end' : 'justify-start';
  const bgColor = isCurrentUserSender ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-200';
  const senderName = isCurrentUserSender ? 'You' : message.sender_username || 'Unknown User';
  const isGroupMessage = message.type === 'group';

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
        {/* Show sender name for group messages or when not current user */}
        {(showSenderAlways || (!isCurrentUserSender && isGroupMessage)) && (
          <div className="flex items-center gap-1 mb-1">
            <p className={`text-sm ${isCurrentUserSender ? 'text-blue-200' : 'text-gray-400'} font-medium`}>
              {senderName}
            </p>
            {message.sender_id && onlineUserIds.includes(message.sender_id) && !isCurrentUserSender && (
              <span className="h-2 w-2 bg-green-500 rounded-full"/>
            )}
          </div>
        )}
        <p className="text-base break-words">{message.content}</p>
        <div className="flex items-center justify-end gap-1 mt-1">
          <p className={`text-xs ${isCurrentUserSender ? 'text-blue-200' : 'text-gray-400'}`}>
            {format(new Date(message.created_at), 'p')}
          </p>
        </div>
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