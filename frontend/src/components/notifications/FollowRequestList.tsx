import React from 'react';
import { FollowRequest } from '@/types/User';
import FollowRequestCard from './FollowRequestCard';

interface FollowRequestListProps {
  requests: FollowRequest[];
  type: 'sent' | 'received';
  onAccept?: (requestId: string, requesterId: string) => void;
  onDecline?: (requestId: string, requesterId: string) => void;
  onCancel?: (requestId: string, targetId: string) => void;
  isLoadingAction?: (requestId: string) => boolean; // To show loading on a specific card
  emptyMessage?: string;
}

const FollowRequestList: React.FC<FollowRequestListProps> = ({
  requests,
  type,
  onAccept,
  onDecline,
  onCancel,
  isLoadingAction,
  emptyMessage = "No requests here."
}) => {
  if (requests.length === 0) {
    return <p className="text-gray-400 text-center py-8">{emptyMessage}</p>;
  }

  return (
    <div className="space-y-3">
      {requests.map((request) => (
        <FollowRequestCard
          key={request.id}
          request={request}
          type={type}
          onAccept={onAccept}
          onDecline={onDecline}
          onCancel={onCancel}
          isLoading={isLoadingAction ? isLoadingAction(request.id) : false}
        />
      ))}
    </div>
  );
};

export default FollowRequestList;