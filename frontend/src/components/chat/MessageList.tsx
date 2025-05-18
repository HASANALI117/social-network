// frontend/src/components/chat/MessageList.tsx
import { Message } from '@/types/Message';
import MessageItem from './MessageItem';
import { Button } from '@/components/ui/button';
import { useEffect, useRef } from 'react';
import { isSameDay, format } from 'date-fns';

const formatDate = (date: Date) => {
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  if (isSameDay(date, today)) {
    return 'Today';
  } else if (isSameDay(date, yesterday)) {
    return 'Yesterday';
  }
  return format(date, 'MMM d, yyyy');
};

interface MessageListProps {
  messages: Message[];
  currentUserId: string;
  onLoadMore: () => void;
  hasMoreMessages: boolean;
  isLoadingMore: boolean;
  type?: 'direct' | 'group'; // Chat type
  emptyMessage?: string; // Custom empty state message
}

export default function MessageList({
  messages,
  currentUserId,
  onLoadMore,
  hasMoreMessages,
  isLoadingMore,
  type = 'direct',
  emptyMessage = type === 'group' ? 'No messages in this group yet. Start the conversation!' : 'No messages yet. Start the conversation!'
}: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages.length]); // Scroll to bottom when new messages are added

  return (
    <div ref={messagesContainerRef} className="flex-grow p-4 space-y-2 bg-gray-800 rounded-t-md">
      {hasMoreMessages && (
        <div className="text-center mb-4">
          <Button onClick={onLoadMore} disabled={isLoadingMore} outline>
            {isLoadingMore ? 'Loading...' : 'Load Older Messages'}
          </Button>
        </div>
      )}
      {messages.length === 0 && !isLoadingMore && (
        <div className="text-center text-gray-400">{emptyMessage}</div>
      )}
      {messages.map((msg, index) => {
        const isFirstMessageOfDay = index === 0 || !isSameDay(new Date(msg.created_at), new Date(messages[index - 1].created_at));
        const isFirstMessageFromUser = index === 0 || messages[index - 1].sender_id !== msg.sender_id;
        
        return (
          <div key={msg.id || `msg-${index}`}>
            {isFirstMessageOfDay && (
              <div className="flex justify-center my-4">
                <span className="px-4 py-1 text-xs text-gray-400 bg-gray-700 rounded-full">
                  {formatDate(new Date(msg.created_at))}
                </span>
              </div>
            )}
            <MessageItem
              message={msg}
              isCurrentUserSender={msg.sender_id === currentUserId}
              showSenderAlways={type === 'group' && isFirstMessageFromUser}
            />
          </div>
        );
      })}
      <div ref={messagesEndRef} />
    </div>
  );
}