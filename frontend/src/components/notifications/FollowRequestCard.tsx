import React from 'react';
import { FollowRequest, UserSummary } from '@/types/User';
import { Button } from '@/components/ui/button';
import { Avatar } from '@/components/ui/avatar'; // Updated import
import { FiCheck, FiX, FiTrash2 } from 'react-icons/fi';
import { formatDistanceToNow } from 'date-fns';

interface FollowRequestCardProps {
  request: FollowRequest;
  type: 'sent' | 'received';
  onAccept?: (requestId: string, requesterId: string) => void;
  onDecline?: (requestId: string, requesterId: string) => void;
  onCancel?: (requestId: string, targetId: string) => void;
  isLoading?: boolean; // To disable buttons during an action
}

const FollowRequestCard: React.FC<FollowRequestCardProps> = ({
  request,
  type,
  onAccept,
  onDecline,
  onCancel,
  isLoading = false,
}) => {
  const userToDisplay: UserSummary = type === 'sent' ? request.target : request.requester;

  const handleAccept = () => {
    if (onAccept) {
      onAccept(request.id, request.requester.id);
    }
  };

  const handleDecline = () => {
    if (onDecline) {
      onDecline(request.id, request.requester.id);
    }
  };

  const handleCancel = () => {
    if (onCancel) {
      onCancel(request.id, request.target.id);
    }
  };

  const timeAgo = formatDistanceToNow(new Date(request.created_at), { addSuffix: true });

  return (
    <div className="bg-gray-800 p-4 rounded-lg shadow flex items-center justify-between hover:bg-gray-750 transition-colors duration-150">
      <div className="flex items-center gap-4">
        <Avatar
          src={userToDisplay.avatar_url}
          initials={`${userToDisplay.first_name?.[0]?.toUpperCase() ?? ''}${userToDisplay.last_name?.[0]?.toUpperCase() ?? ''}`}
          alt={`${userToDisplay.first_name} ${userToDisplay.last_name}'s avatar`}
          className="h-12 w-12"
        />
        <div>
          <p className="font-semibold text-gray-100">
            {userToDisplay.first_name} {userToDisplay.last_name}
            <span className="text-gray-400 ml-1">@{userToDisplay.username}</span>
          </p>
          <p className="text-sm text-gray-400">{timeAgo}</p>
        </div>
      </div>
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