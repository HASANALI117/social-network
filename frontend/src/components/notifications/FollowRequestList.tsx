import React from 'react';
import { UserSummary } from '@/types/User';
import FollowRequestCard from './FollowRequestCard';

export interface FollowRequestListProps { // Exporting for use in ManageFollowRequestsSection if needed for typing
  users: UserSummary[];
  type: 'sent' | 'received';
  onAccept?: (requesterId: string) => void; // Changed from (requestId, requesterId)
  onDecline?: (requesterId: string) => void; // Changed from (requestId, requesterId)
  onCancel?: (targetId: string) => void; // Changed from (requestId, targetId)
  isLoadingAction?: (userId: string) => boolean; // Changed from (requestId)
  emptyMessage?: string;
}

const FollowRequestList: React.FC<FollowRequestListProps> = ({
  users,
  type,
  onAccept,
  onDecline,
  onCancel,
  isLoadingAction,
  emptyMessage = "No requests here."
}) => {
  if (users.length === 0) {
    return <p className="text-gray-400 text-center py-8">{emptyMessage}</p>;
  }

  return (
    <div className="space-y-3">
      {users.map((user) => (
        <FollowRequestCard
          key={user.id} // Assuming UserSummary has an id
          user={user}
          type={type}
          onAccept={onAccept ? () => onAccept(user.id) : undefined}
          onDecline={onDecline ? () => onDecline(user.id) : undefined}
          onCancel={onCancel ? () => onCancel(user.id) : undefined}
          isLoading={isLoadingAction ? isLoadingAction(user.id) : false}
        />
      ))}
    </div>
  );
};

export default FollowRequestList;