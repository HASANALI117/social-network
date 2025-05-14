import React, { useState, useEffect, useCallback } from 'react';
import Tabs from '../../common/Tabs';
import { Text, Strong } from '../../ui/text';
import { Button } from '../../ui/button';
import { Avatar } from '../../ui/avatar';
import GroupPostsTab from './GroupPostsTab';
import GroupMembersTab from './GroupMembersTab';
import GroupEventsTab from './GroupEventsTab';
import GroupInviteManager from '../GroupInviteManager'; // New import
import { Group } from '../../../types/Group';
import { UserBasicInfo } from '../../../types/User';
import { GroupJoinRequest } from '../../../types/GroupJoinRequest'; // New import
import { useRequest } from '../../../hooks/useRequest'; // New import
import { toast } from 'react-hot-toast'; // New import

interface GroupMemberViewProps {
  group: Group;
  currentUser: UserBasicInfo | null;
  handleLeaveGroup: () => void;
}

export default function GroupMemberView({
  group,
  currentUser,
  handleLeaveGroup,
}: GroupMemberViewProps) {
  const { id: groupId, members_count, posts_count, events_count, posts, members, events, creator_info } = group;
  const [showInviteManager, setShowInviteManager] = useState(false);

  const [pendingJoinRequests, setPendingJoinRequests] = useState<GroupJoinRequest[]>([]);
  const [isLoadingJoinRequests, setIsLoadingJoinRequests] = useState(false);
  const [joinRequestsError, setJoinRequestsError] = useState<string | null>(null);
  const [processingRequestId, setProcessingRequestId] = useState<string | null>(null);

  const { get: fetchJoinRequests, isLoading: fetchJoinRequestsLoading, error: fetchJoinRequestsErr } = useRequest<GroupJoinRequest[]>();
  const { post: acceptJoinRequest, isLoading: acceptLoading, error: acceptErr } = useRequest<any>();
  const { post: rejectJoinRequest, isLoading: rejectLoading, error: rejectErr } = useRequest<any>();

  const isCurrentUserCreatorOrAdmin = currentUser?.user_id === creator_info.user_id; // Simplified admin check

  const loadPendingJoinRequests = useCallback(async () => {
    if (!groupId || !isCurrentUserCreatorOrAdmin) return;
    setIsLoadingJoinRequests(true);
    setJoinRequestsError(null);
    const data = await fetchJoinRequests(`/api/groups/${groupId}/requests/pending`);
    if (data) {
      setPendingJoinRequests(data);
    } else if (fetchJoinRequestsErr) {
      setJoinRequestsError(fetchJoinRequestsErr.message || 'Failed to load join requests.');
      toast.error(fetchJoinRequestsErr.message || 'Failed to load join requests.');
    }
    setIsLoadingJoinRequests(false);
  }, [groupId, fetchJoinRequests, fetchJoinRequestsErr, isCurrentUserCreatorOrAdmin]);

  useEffect(() => {
    if (isCurrentUserCreatorOrAdmin) {
      loadPendingJoinRequests();
    }
  }, [isCurrentUserCreatorOrAdmin, loadPendingJoinRequests]);

  const handleAcceptRequest = async (requestId: string) => {
    setProcessingRequestId(requestId);
    const response = await acceptJoinRequest(`/api/groups/requests/${requestId}/accept`, {});
    if (response && !acceptErr) {
      toast.success('Join request accepted.');
      setPendingJoinRequests(prev => prev.filter(req => req.id !== requestId));
      // Optionally, refresh group members or counts here
    } else {
      toast.error(acceptErr?.message || 'Failed to accept join request.');
    }
    setProcessingRequestId(null);
  };

  const handleRejectRequest = async (requestId: string) => {
    setProcessingRequestId(requestId);
    const response = await rejectJoinRequest(`/api/groups/requests/${requestId}/reject`, {});
    if (response && !rejectErr) {
      toast.success('Join request rejected.');
      setPendingJoinRequests(prev => prev.filter(req => req.id !== requestId));
    } else {
      toast.error(rejectErr?.message || 'Failed to reject join request.');
    }
    setProcessingRequestId(null);
  };
  
  const handleInviteSent = (invitedUserId: string, successMessage: string) => {
    // Optionally, update UI or state here, e.g., refetch members
    // For now, GroupInviteManager handles its own "Invited" state for the button
  };

  const handleInviteError = (errorMessage: string) => {
    // Error already toasted by GroupInviteManager
  };


  const invitationsTabContent = (
    <Tabs.Panel id="invitations" className="py-4 space-y-6">
      {isCurrentUserCreatorOrAdmin && (
        <div>
          <Strong className="text-lg block mb-2">Manage Join Requests</Strong>
          {isLoadingJoinRequests && <Text>Loading join requests...</Text>}
          {joinRequestsError && <Text className="text-red-500">{joinRequestsError}</Text>}
          {!isLoadingJoinRequests && !joinRequestsError && pendingJoinRequests.length === 0 && (
            <Text>No pending join requests.</Text>
          )}
          {pendingJoinRequests.length > 0 && (
            <ul className="space-y-3">
              {pendingJoinRequests.map(req => (
                <li key={req.id} className="flex items-center justify-between p-3 bg-gray-700 rounded-md">
                  <div className="flex items-center space-x-3">
                    <Avatar
                      src={req.requester_info?.avatar_url || null}
                      initials={`${req.requester_info?.first_name?.[0] || ''}${req.requester_info?.last_name?.[0] || ''}`}
                      alt={req.requester_info?.username || 'User'}
                      className="h-10 w-10"
                    />
                    <div>
                      <Strong>{req.requester_info?.first_name} {req.requester_info?.last_name}</Strong>
                      <Text className="text-sm text-gray-400">@{req.requester_info?.username}</Text>
                    </div>
                  </div>
                  <div className="space-x-2">
                    <Button
                      onClick={() => handleAcceptRequest(req.id)}
                      disabled={processingRequestId === req.id || acceptLoading || rejectLoading}
                      color="green"
                      className="text-xs px-2 py-1"
                    >
                      Accept
                    </Button>
                    <Button
                      onClick={() => handleRejectRequest(req.id)}
                      disabled={processingRequestId === req.id || acceptLoading || rejectLoading}
                      color="red"
                      className="text-xs px-2 py-1"
                    >
                      Reject
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      <div>
        <Button
          onClick={() => setShowInviteManager(prev => !prev)}
          className="mb-4 w-full sm:w-auto"
          color="blue"
        >
          {showInviteManager ? 'Hide Invite Form' : 'Invite Users to Group'}
        </Button>
        {showInviteManager && (
          <GroupInviteManager
            groupId={groupId}
            currentUser={currentUser}
            onInviteSent={handleInviteSent}
            onInviteError={handleInviteError}
            onClose={() => setShowInviteManager(false)} // Optional: allow GIM to close itself
          />
        )}
        {!isCurrentUserCreatorOrAdmin && !showInviteManager && pendingJoinRequests.length === 0 && (
           <Text className="text-gray-400">Group admins manage join requests. You can invite users.</Text>
        )}
      </div>
    </Tabs.Panel>
  );


  return (
    <div>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8 text-lg text-center">
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{members_count}</Text>
          <Text className="text-gray-300">Members</Text>
        </div>
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{posts_count}</Text>
          <Text className="text-gray-300">Posts</Text>
        </div>
        <div className="p-4 bg-gray-700 rounded-md">
          <Text className="font-semibold text-purple-400">{events_count}</Text>
          <Text className="text-gray-300">Events</Text>
        </div>
      </div>
      
      <div className="mb-6 flex justify-end">
         <Button
          onClick={handleLeaveGroup}
          outline // Secondary action style
          className="text-sm" // Adjust size as needed
        >
          Leave Group
        </Button>
      </div>

      <Tabs tabs={[
        { label: 'Posts', content: <Tabs.Panel id="posts" className="py-4"><GroupPostsTab posts={posts} /></Tabs.Panel> },
        { label: 'Members', content: <Tabs.Panel id="members" className="py-4"><GroupMembersTab members={members} /></Tabs.Panel> },
        { label: 'Events', content: <Tabs.Panel id="events" className="py-4"><GroupEventsTab events={events} /></Tabs.Panel> },
        { label: 'Invitations', content: invitationsTabContent }
      ]} />
      
      {/* Old invite UI and buttons are removed from here */}
    </div>
  );
}