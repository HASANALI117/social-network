'use client';

import { useState, useEffect, useCallback } from 'react';
import ManageFollowRequestsSection from '@/components/notifications/ManageFollowRequestsSection';
import { GroupInvitation, GroupInvitationStatus } from '../../types/GroupInvitation';
import { Button } from '../../components/ui/button';
import { useUserStore } from '../../store/useUserStore';
import { Alert, AlertTitle, AlertDescription } from '../../components/ui/alert';
import { useRequest } from '../../hooks/useRequest';
import { Heading } from '../../components/ui/heading';
import { Text } from '../../components/ui/text';
import { Avatar } from '../../components/ui/avatar';

interface GroupInvitationsApiResponse {
  invitations: GroupInvitation[];
  // Potentially other fields like total, limit, offset if paginated
}

export default function NotificationsPage() {
  // State
  const [groupInvitations, setGroupInvitations] = useState<GroupInvitation[]>([]);
  const [isLoadingInvitations, setIsLoadingInvitations] = useState<boolean>(true);
  const [invitationsError, setInvitationsError] = useState<string | null>(null);
  const [actionFeedback, setActionFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  // Hooks
  const { user: currentUser } = useUserStore();
  const { get: fetchGroupInvitations, error: fetchInvitationsApiError, isLoading: apiIsFetchingInvitations } = useRequest<GroupInvitationsApiResponse>();
  const { post: handleInvitationActionRequest, isLoading: isUpdatingInvitation, error: updateInvitationApiError } = useRequest<GroupInvitation>();

  // Fetch Invitations Function
  const loadGroupInvitations = useCallback(async () => {
    if (!currentUser) {
      setIsLoadingInvitations(false);
      setGroupInvitations([]);
      return;
    }
    setIsLoadingInvitations(true);
    setInvitationsError(null); // Clear previous custom error
    setActionFeedback(null);

    const data = await fetchGroupInvitations(`/api/groups/invitations/pending`);
    if (data && data.invitations) {
      setGroupInvitations(data.invitations);
    }
    // Error handling will be done by the useEffect watching fetchInvitationsApiError
    setIsLoadingInvitations(false);
  }, [currentUser, fetchGroupInvitations]);

  // useEffect for Initial Load of Invitations
  useEffect(() => {
    if (currentUser) {
      loadGroupInvitations();
    } else {
      setIsLoadingInvitations(false);
      setGroupInvitations([]);
      setInvitationsError(null);
    }
  }, [currentUser, loadGroupInvitations]);

  // useEffect for API errors when fetching invitations
  useEffect(() => {
    if (fetchInvitationsApiError) {
      setInvitationsError(fetchInvitationsApiError.message || 'Failed to load group invitations.');
      setIsLoadingInvitations(false); // Ensure loading is stopped
    }
  }, [fetchInvitationsApiError]);

  // Handle Invitation Action Function
  const handleInvitationAction = async (invitationId: string, action: 'accept' | 'decline') => {
    setActionFeedback(null);
    const endpointAction = action === 'accept' ? 'accept' : 'reject';
    const apiUrl = `/api/groups/invitations/${invitationId}/${endpointAction}`;

    const updatedInvitation = await handleInvitationActionRequest(
      apiUrl,
      {} // POST requests usually have a body, even if empty for these actions
    );

    if (updatedInvitation && !updateInvitationApiError) { // Check hook's error state after call
      const successMessage = action === 'accept' ? 'Invitation accepted successfully!' : 'Invitation declined successfully!';
      setActionFeedback({ type: 'success', message: successMessage });
      loadGroupInvitations(); // Or optimistically update the list
    } else {
      setActionFeedback({ type: 'error', message: updateInvitationApiError?.message || `Failed to ${action} invitation.` });
    }
  };

  // Update JSX
  return (
    <div className="p-4 sm:p-6 bg-gray-900 min-h-screen text-gray-100">
      <header className="mb-6">
        <h1 className="text-3xl font-bold text-white">Notifications</h1>
      </header>
      
      <div className="mb-8"> {/* Follow Requests Section - Added mb-8 for spacing */}
        <h2 className="text-xl font-semibold text-gray-200 mb-3">Follow Requests</h2>
        <ManageFollowRequestsSection />
      </div>

      {/* Group Invitations Section */}
      <section className="mt-8">
        <Heading level={2} className="mb-4 text-xl font-semibold text-white">Group Invitations</Heading>
        
        {(isLoadingInvitations || apiIsFetchingInvitations) && <Text>Loading group invitations...</Text>}
        
        {invitationsError && !(isLoadingInvitations || apiIsFetchingInvitations) && (
          <Text className="text-red-500">Error: {invitationsError}</Text>
        )}
        
        {actionFeedback && (
          <Alert open={!!actionFeedback} onClose={() => setActionFeedback(null)} className="mb-4">
            <AlertTitle>{actionFeedback.type === 'success' ? 'Success!' : 'Error'}</AlertTitle>
            <AlertDescription>{actionFeedback.message}</AlertDescription>
          </Alert>
        )}

        {!(isLoadingInvitations || apiIsFetchingInvitations) && !invitationsError && groupInvitations.filter(inv => inv.status === "pending").length === 0 && (
          <Text className="text-gray-400">No pending group invitations.</Text>
        )}

        {!(isLoadingInvitations || apiIsFetchingInvitations) && !invitationsError && groupInvitations.filter(inv => inv.status === "pending").length > 0 && (
          <div className="space-y-4">
            {groupInvitations
              .filter(inv => inv.status === "pending") // Ensure we only show pending
              .map((invitation) => (
              <div key={invitation.id} className="p-4 bg-gray-800 rounded-lg shadow flex justify-between items-center">
                <div className="flex items-center"> {/* Wrapper for avatar + text */}
                  {/* Group Avatar */}
                  {invitation.group?.avatar_url && (
                    <Avatar src={invitation.group.avatar_url} initials={(invitation.group_name || invitation.group.name)?.charAt(0) || 'G'} alt={invitation.group_name || invitation.group.name || 'Group'} className="h-10 w-10 mr-3" />
                  )}
                  <div>
                    {/* Group Name */}
                    <Text className="font-semibold">
                      Invitation to join: <span className="text-blue-400">{invitation.group_name || invitation.group?.name || 'Unnamed Group'}</span>
                    </Text>
                    <div className="flex items-center text-sm text-gray-400 mt-1">
                      {/* Inviter Avatar */}
                      {invitation.inviter?.avatar_url && (
                        <Avatar src={invitation.inviter.avatar_url} initials={(invitation.inviter.first_name || invitation.inviter.username)?.charAt(0) || 'U'} alt={(invitation.inviter.first_name && invitation.inviter.last_name ? `${invitation.inviter.first_name} ${invitation.inviter.last_name}` : invitation.inviter.username) || 'Inviter'} className="h-5 w-5 mr-1.5" />
                      )}
                      {/* Inviter Name */}
                      Invited by: {(invitation.inviter?.first_name && invitation.inviter?.last_name ? `${invitation.inviter.first_name} ${invitation.inviter.last_name}` : invitation.inviter?.username) || 'A user'}
                    </div>
                  </div>
                </div>
                <div className="space-x-2">
                  <Button
                    outline
                    onClick={() => handleInvitationAction(invitation.id, 'accept')}
                    disabled={isUpdatingInvitation}
                  >
                    Accept
                  </Button>
                  <Button
                    color="red"
                    onClick={() => handleInvitationAction(invitation.id, 'decline')}
                    disabled={isUpdatingInvitation}
                  >
                    Decline
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
