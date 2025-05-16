// frontend/src/components/chat/MessageItem.tsx
import { Message } from '@/types/Message';
import { Avatar } from '@/components/ui/avatar';
import { format } from 'date-fns'; // For formatting timestamps

interface MessageItemProps {
  message: Message;
  isCurrentUserSender: boolean;
}

export default function MessageItem({ message, isCurrentUserSender }: MessageItemProps) {
  const alignment = isCurrentUserSender ? 'justify-end' : 'justify-start';
  const bgColor = isCurrentUserSender ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-200';
  const senderName = isCurrentUserSender ? 'You' : message.sender_username || 'User';

  return (
    <div className={`flex ${alignment} mb-3`}>
      {!isCurrentUserSender && (
        <Avatar
          className="w-6 h-6 mr-2"
          src={message.sender_avatar_url}
          alt={senderName}
          initials={senderName.substring(0, 1).toUpperCase()}
        />
      )}
      <div className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg shadow ${bgColor}`}>
        {!isCurrentUserSender && (
          <p className="text-xs text-gray-400 mb-1">{senderName}</p>
        )}
        <p className="text-sm">{message.content}</p>
        <p className={`text-xs mt-1 ${isCurrentUserSender ? 'text-blue-200' : 'text-gray-400'} text-right`}>
          {format(new Date(message.created_at), 'p')}
        </p>
      </div>
      {isCurrentUserSender && (
        <Avatar
          className="w-6 h-6 ml-2"
          src={message.sender_avatar_url} // Assuming current user's avatar is part of the message object
          alt={senderName}
          initials="Y"
        />
      )}
    </div>
  );
}