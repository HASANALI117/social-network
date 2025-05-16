// frontend/src/components/chat/MessageInput.tsx
import { useState, FormEvent } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import Picker, { EmojiClickData, Theme } from 'emoji-picker-react';

interface MessageInputProps {
  onSendMessage: (content: string) => void;
  isSending: boolean;
  canSendMessage: boolean; // To disable input based on follow status/privacy (graceful handling)
}

export default function MessageInput({ onSendMessage, isSending, canSendMessage }: MessageInputProps) {
  const [message, setMessage] = useState('');
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);

  const onEmojiClick = (emojiData: EmojiClickData) => {
    setMessage(prevMessage => prevMessage + emojiData.emoji);
    // Optionally hide the picker after selection, or let user close it manually
    // setShowEmojiPicker(false);
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (message.trim() && canSendMessage) {
      onSendMessage(message.trim());
      setMessage('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="p-4 border-t border-gray-700 bg-gray-800 rounded-b-md relative">
      {showEmojiPicker && (
        <div style={{ position: 'absolute', bottom: '60px', right: '10px', zIndex: 10 }}>
          <Picker onEmojiClick={onEmojiClick} theme={Theme.DARK} />
        </div>
      )}
      <div className="flex items-center space-x-2">
        <Button
          type="button"
          plain={true}
          onClick={() => setShowEmojiPicker(!showEmojiPicker)}
          disabled={!canSendMessage}
          aria-label="Toggle emoji picker"
          className="p-2" // Added padding for better appearance as size="icon" is not available
        >
          😀
        </Button>
        <Input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder={canSendMessage ? "Type a message..." : "You cannot send messages to this user."}
          className="flex-grow bg-gray-700 border-gray-600 text-white placeholder-gray-400"
          disabled={isSending || !canSendMessage}
        />
        <Button type="submit" disabled={isSending || !message.trim() || !canSendMessage}>
          {isSending ? 'Sending...' : 'Send'}
        </Button>
      </div>
      {!canSendMessage && (
         <p className="text-xs text-red-400 mt-1 text-center">
           Messaging is restricted. You might need to follow this user or they need to follow you back, or their profile is private.
         </p>
      )}
    </form>
  );
}