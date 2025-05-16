// frontend/src/components/chat/MessageList.tsx
import { Message } from '@/types/Message';
import MessageItem from './MessageItem';
import { Button } from '@/components/ui/button';
import { useEffect, useRef } from 'react';

interface MessageListProps {
  messages: Message[];
  currentUserId: string;
  onLoadMore: () => void;
  hasMoreMessages: boolean;
  isLoadingMore: boolean;
}

export default function MessageList({
  messages,
  currentUserId,
  onLoadMore,
  hasMoreMessages,
  isLoadingMore,
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
    <div ref={messagesContainerRef} className="flex-grow overflow-y-auto p-4 space-y-2 bg-gray-800 rounded-t-md">
      {hasMoreMessages && (
        <div className="text-center mb-4">
          <Button onClick={onLoadMore} disabled={isLoadingMore} outline>
            {isLoadingMore ? 'Loading...' : 'Load Older Messages'}
          </Button>
        </div>
      )}
      {messages.length === 0 && !isLoadingMore && (
        <div className="text-center text-gray-400">No messages yet. Start the conversation!</div>
      )}
      {messages.map((msg, index) => (
        <MessageItem
          key={msg.id || `msg-${index}`} // Use backend ID if available, otherwise fallback
          message={msg}
          isCurrentUserSender={msg.sender_id === currentUserId}
        />
      ))}
      <div ref={messagesEndRef} />
    </div>
  );
}