'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRequest } from '../../hooks/useRequest';
import { Avatar } from '../../components/ui/avatar';
import { useGlobalWebSocket } from '../../contexts/GlobalWebSocketContext';
// import { useGlobalWebSocket } from 'contexts/GlobalWebSocketContext'; // Avatar component handles this internally

interface ChatPartner {
  id: string;
  first_name: string;
  last_name: string;
  username: string;
  avatar_url: string;
  last_message: string;
  last_message_at: string; // ISO string date
}

export default function MessagesPage() {
  const [chatPartners, setChatPartners] = useState<ChatPartner[]>([]);
  const { get, isLoading, error } = useRequest<ChatPartner[]>();
  const { clearMessageCount } = useGlobalWebSocket();
  // const { onlineUserIds } = useGlobalWebSocket(); // Avatar component handles this internally

  useEffect(() => {
    const fetchConversations = async () => {
      const data = await get('/api/messages/conversations');
      if (data) {
        setChatPartners(data);
      }
    };
    fetchConversations();
  }, [get]);

  // Clear message count when user visits messages page
  useEffect(() => {
    clearMessageCount();
  }, [clearMessageCount]);

  if (isLoading) {
    return (
      <div className="p-6 text-center text-gray-300">
        Loading conversations...
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6 text-center text-red-500">
        Error loading conversations: {error.message}
      </div>
    );
  }

  return (
    <div className="p-4 md:p-6 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6 text-gray-100">Messages</h1>
      {chatPartners.length === 0 ? (
        <p className="text-gray-400 text-center">
          No conversations yet. Start a chat to see it here.
        </p>
      ) : (
        <div className="space-y-3">
          {chatPartners.map((partner) => (
            <Link
              href={`/chat/${partner.id}`}
              key={partner.id}
              className="block p-4 rounded-lg bg-gray-800 hover:bg-gray-700 transition-colors shadow"
            >
              <div className="flex items-center gap-4">
                <Avatar
                  src={
                    partner.avatar_url ||
                    `https://ui-avatars.com/api/?name=${partner.first_name}+${partner.last_name}&background=3b82f6&color=fff&bold=true`
                  }
                  alt={`${partner.first_name} ${partner.last_name}`}
                  userId={partner.id}
                  className="w-12 h-12"
                />
                <div className="flex-1 min-w-0">
                  <div className="flex justify-between items-start">
                    <p className="font-semibold text-gray-100 truncate">
                      {partner.first_name} {partner.last_name}
                    </p>
                    <p className="text-xs text-gray-500 whitespace-nowrap">
                      {new Date(partner.last_message_at).toLocaleTimeString(
                        [],
                        { hour: '2-digit', minute: '2-digit', hour12: true }
                      )}
                    </p>
                  </div>
                  <p className="text-sm text-gray-400 truncate mt-1">
                    {partner.last_message}
                  </p>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
