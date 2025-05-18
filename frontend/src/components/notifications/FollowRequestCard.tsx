'use client';

import React from 'react';
import { UserSummary } from '@/types/User';
import { Button } from '@/components/ui/button';
import { Avatar } from '@/components/ui/avatar'; // Updated import
import { FiCheck, FiX, FiTrash2 } from 'react-icons/fi';
// import { formatDistanceToNow } from 'date-fns'; // Not used currently
import Link from 'next/link'; // Import Link for navigation

export interface FollowRequestCardProps { // Exporting for use if needed
  user: UserSummary;
  type: 'sent' | 'received';
  onAccept?: (requesterId: string) => void; // Operates on user.id
  onDecline?: (requesterId: string) => void; // Operates on user.id
  onCancel?: (targetId: string) => void; // Operates on user.id
  isLoading?: boolean;
}

const FollowRequestCard: React.FC<FollowRequestCardProps> = ({
  user, // Changed from request
  type,
  onAccept,
  onDecline,
  onCancel,
  isLoading = false,
}) => {
  // user prop directly represents the user involved in the request
  // If type is 'received', user is the requester.
  // If type is 'sent', user is the target.

  const handleAccept = () => {
    if (onAccept && type === 'received') { // Action for received requests
      onAccept(user.id);
    }
  };

  const handleDecline = () => {
    if (onDecline && type === 'received') { // Action for received requests
      onDecline(user.id);
    }
  };

  const handleCancel = () => {
    if (onCancel && type === 'sent') { // Action for sent requests
      onCancel(user.id);
    }
  };

  // The UserSummary from /api/users/me/follow-requests does not include the request timestamp.
  // We will use user.created_at (account creation) for now, or omit it.
  // For this iteration, let's omit the timestamp to avoid confusion,
  // as it's not the request creation time.
  // const timeAgo = user.created_at ? formatDistanceToNow(new Date(user.created_at), { addSuffix: true }) : '';

  return (
    <div className="bg-gray-800 p-4 rounded-lg shadow flex items-center justify-between hover:bg-gray-750 transition-colors duration-150">
      <Link href={`/profile/${user.id}`} className="flex items-center gap-4 group">
        <Avatar
          src={user.avatar_url}
          initials={`${user.first_name?.[0]?.toUpperCase() ?? ''}${user.last_name?.[0]?.toUpperCase() ?? ''}`}
          alt={`${user.first_name} ${user.last_name}'s avatar`}
          className="h-12 w-12 group-hover:opacity-80 transition-opacity"
          userId={user.id}
        />
        <div>
          <p className="font-semibold text-gray-100 group-hover:text-purple-300 transition-colors">
            {user.first_name} {user.last_name}
            <span className="text-gray-400 ml-1 group-hover:text-purple-400 transition-colors">@{user.username}</span>
          </p>
          {/* Timestamp omitted as per previous decision */}
        </div>
      </Link>
      <div className="flex gap-2">
        {type === 'received' && onAccept && onDecline && (
          <>
            <Button
              plain // Use 'plain' for a ghost-like appearance or define a specific style
              onClick={handleAccept}
              disabled={isLoading}
              className="text-green-400 hover:text-green-300 px-3 py-1 flex items-center" // Adjusted className
              aria-label="Accept follow request"
            >
              <FiCheck className="mr-1 h-4 w-4" /> Accept
            </Button>
            <Button
              plain // Use 'plain' for a ghost-like appearance
              onClick={handleDecline}
              disabled={isLoading}
              className="text-red-400 hover:text-red-300 px-3 py-1 flex items-center" // Adjusted className
              aria-label="Decline follow request"
            >
              <FiX className="mr-1 h-4 w-4" /> Decline
            </Button>
          </>
        )}
        {type === 'sent' && onCancel && (
          <Button
            plain // Use 'plain' for a ghost-like appearance
            onClick={handleCancel}
            disabled={isLoading}
            className="text-yellow-400 hover:text-yellow-300 px-3 py-1 flex items-center" // Adjusted className
            aria-label="Cancel follow request"
          >
            <FiTrash2 className="mr-1 h-4 w-4" /> Cancel Request
          </Button>
        )}
      </div>
    </div>
  );
};

export default FollowRequestCard;