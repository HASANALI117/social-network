'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useForm, SubmitHandler } from 'react-hook-form';
import { useRequest } from '../../../hooks/useRequest';
import { Group } from '../../../types/Group';
import { User } from '../../../types/User';
import { GroupInvitation } from '../../../types/GroupInvitation';
import { GroupJoinRequest } from '../../../types/GroupJoinRequest';
import { useUserStore } from '../../../store/useUserStore';
import { Heading } from '../../../components/ui/heading';
import { Text } from '../../../components/ui/text';
import { Avatar } from '../../../components/ui/avatar';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Alert, AlertDescription, AlertTitle } from '../../../components/ui/alert';

interface InviteUserFormValues {
  userIdToInvite: string;
}

interface GroupJoinRequestsApiResponse {
  requests: GroupJoinRequest[];
  // Pagination fields if applicable
}

interface GroupInvitationsApiResponse { // Added for fetching specific invitations
  invitations: GroupInvitation[];
  // Pagination fields if applicable
}

interface GroupMembersApiResponse {
  members: User[];
  // Pagination fields if applicable
}

// New State Type for User's Group Status
type UserGroupStatus =
  | 'creator'
  | 'member'
  | 'pending_invitation' // User has an invitation to this group
  | 'pending_join_request' // User has requested to join this group
  | 'not_affiliated'
  | 'loading'
  | 'unknown_error';

export default function GroupDetailPage() {
  const params = useParams();
  const router = useRouter();
  const groupId = params.id as string;

  const [group, setGroup] = useState<Group | null>(null);
  const [creator, setCreator] = useState<User | null>(null); // If fetching creator separately
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [inviteError, setInviteError] = useState<string | null>(null);
  const [inviteSuccess, setInviteSuccess] = useState<string | null>(null);

  // New state for comprehensive user group status
  const [currentUserGroupStatus, setCurrentUserGroupStatus] = useState<UserGroupStatus>('loading');
  // const [joinRequestStatus, setJoinRequestStatus] = useState<'idle' | 'pending' | 'requested' | 'member' | 'creator' | 'error'>('idle'); // To be removed or integrated
  const [joinRequestError, setJoinRequestError] = useState<string | null>(null); // May still be useful for specific join action errors
  const [isProcessingJoin, setIsProcessingJoin] = useState(false); // For join/cancel join request button
  const [pendingJoinRequestId, setPendingJoinRequestId] = useState<string | null>(null); // To store ID for cancellation


  // State for managing join requests list (for creator view)
  const [joinRequests, setJoinRequests] = useState<GroupJoinRequest[]>([]);
  const [isLoadingJoinRequests, setIsLoadingJoinRequests] = useState(false);
  const [joinRequestsError, setJoinRequestsError] = useState<string | null>(null);
  const [requestActionFeedback, setRequestActionFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  // State for Group Members
  const [members, setMembers] = useState<User[]>([]);
  const [isLoadingMembers, setIsLoadingMembers] = useState(false);
  const [membersError, setMembersError] = useState<string | null>(null);

  const { user: currentUser } = useUserStore();
  const { register: registerInvite, handleSubmit: handleSubmitInvite, formState: { errors: inviteFormErrors }, reset: resetInviteForm } = useForm<InviteUserFormValues>();
  const { post: sendInviteRequest, isLoading: isSendingInvite, error: sendInviteApiError } = useRequest<GroupInvitation>();
  const { post: sendJoinRequest, isLoading: isSendingJoin, error: joinRequestApiError } = useRequest<GroupJoinRequest>();
  const { del: cancelJoinRequest, isLoading: isCancelingJoinRequest, error: cancelJoinRequestApiError } = useRequest<void>(); // For cancelling join request
  const { del: leaveGroupRequest, isLoading: isLeavingGroup, error: leaveGroupApiError } = useRequest<void>(); // For leaving group
  const { get: fetchGroupRequest, error: groupApiError } = useRequest<Group>();
  const { get: fetchUserRequest, error: userApiError } = useRequest<User>(); // For creator

  // Hooks for fetching and updating join requests (for creator view)
  const { get: fetchJoinRequests, error: fetchJoinRequestsApiError } = useRequest<GroupJoinRequestsApiResponse>();
  const { put: updateJoinRequest, isLoading: isUpdatingJoinRequest, error: updateJoinRequestApiError } = useRequest<GroupJoinRequest>();
  
  // Hooks for fetching members and specific user status data
  const { get: fetchMembers, error: fetchMembersApiError } = useRequest<GroupMembersApiResponse>();
  const { get: fetchMyGroupInvitation, error: fetchMyGroupInvitationError } = useRequest<GroupInvitationsApiResponse>();
  const { get: fetchMyGroupJoinRequest, error: fetchMyGroupJoinRequestError } = useRequest<GroupJoinRequestsApiResponse>();


  const loadGroupDetails = useCallback(async (id: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const groupData = await fetchGroupRequest(`/api/groups/${id}`);
      if (groupData) {
        setGroup(groupData);
        // If creator details are not embedded, fetch them
        if (groupData.creator_id) {
          const userData = await fetchUserRequest(`/api/users/${groupData.creator_id}`);
          if (userData) {
            setCreator(userData);
          } else if (userApiError) {
            console.error("Failed to fetch creator details:", userApiError.message);
            // Set a fallback or handle partial data
          }
        }
      } else if (groupApiError) {
        setError(groupApiError.message || 'Failed to load group details.');
      } else {
        setError('Group not found or failed to load.');
      }
    } catch (err: any) {
      setError(err.message || 'An unexpected error occurred.');
    } finally {
      setIsLoading(false);
    }
  }, [fetchGroupRequest, fetchUserRequest, groupApiError, userApiError]);

  useEffect(() => {
    if (groupId) {
      loadGroupDetails(groupId);
    } else {
      // Handle case where groupId is not available, though Next.js routing should ensure it
      setError("Group ID is missing.");
      setIsLoading(false);
    }
  }, [groupId, loadGroupDetails]);


  const loadMembers = useCallback(async (gId: string) => {
    setIsLoadingMembers(true);
    setMembersError(null);
    try {
      const data = await fetchMembers(`/api/groups/${gId}/members`);
      if (data && data.members) {
        setMembers(data.members);
      } else if (fetchMembersApiError) {
        setMembersError(fetchMembersApiError.message || 'Failed to load members.');
      } else {
        setMembersError('Failed to load members.');
      }
    } catch (err: any) {
      setMembersError(err.message || 'An unexpected error occurred while fetching members.');
    } finally {
      setIsLoadingMembers(false);
    }
  }, [fetchMembers, fetchMembersApiError]);

  useEffect(() => {
    if (groupId) {
      loadMembers(groupId);
    }
  }, [groupId, loadMembers]);

  const loadJoinRequests = useCallback(async (gId: string) => {
    if (!currentUser || !group || currentUser.id !== group.creator_id) {
      return;
    }
    setIsLoadingJoinRequests(true);
    setJoinRequestsError(null);
    setRequestActionFeedback(null); // Clear previous feedback
    try {
      const data = await fetchJoinRequests(`/api/groups/${gId}/join-requests?status=pending`);
      if (data && data.requests) {
        setJoinRequests(data.requests);
      } else if (fetchJoinRequestsApiError) {
        setJoinRequestsError(fetchJoinRequestsApiError.message || 'Failed to load join requests.');
      } else {
        setJoinRequestsError('Failed to load join requests.');
      }
    } catch (err: any) {
      setJoinRequestsError(err.message || 'An unexpected error occurred while fetching join requests.');
    } finally {
      setIsLoadingJoinRequests(false);
    }
  }, [currentUser, group, fetchJoinRequests, fetchJoinRequestsApiError]);

  useEffect(() => {
    if (groupId && currentUser && group && group.creator_id) { // Ensure group.creator_id is available
        loadJoinRequests(groupId); // Load join requests if creator (logic inside loadJoinRequests checks for creator status)
    }
  }, [groupId, currentUser, group, loadJoinRequests]);


  // Main useEffect to determine currentUserGroupStatus
  useEffect(() => {
    const determineStatus = async () => {
      if (!currentUser || !group || !groupId || !members) { // Added members to dependency
        setCurrentUserGroupStatus('loading');
        return;
      }

      setCurrentUserGroupStatus('loading');

      if (currentUser.id === group.creator_id) {
        setCurrentUserGroupStatus('creator');
        return;
      }

      if (!isLoadingMembers && members.some(m => m.id === currentUser.id)) {
        setCurrentUserGroupStatus('member');
        return;
      }

      try {
        // Fetch pending invitation for this group
        const invData = await fetchMyGroupInvitation(`/api/users/me/group-invitations?group_id=${groupId}&status=pending`);
        if (invData && invData.invitations && invData.invitations.length > 0) {
          setCurrentUserGroupStatus('pending_invitation');
          return;
        }

        // Fetch pending join request for this group
        const reqData = await fetchMyGroupJoinRequest(`/api/users/me/group-join-requests?group_id=${groupId}&status=pending`);
        if (reqData && reqData.requests && reqData.requests.length > 0) {
          setPendingJoinRequestId(reqData.requests[0].id); // Store request ID for cancellation
          setCurrentUserGroupStatus('pending_join_request');
          return;
        }
        
        setCurrentUserGroupStatus('not_affiliated');

      } catch (error) {
        console.error("Error determining user group status:", error);
        // Check specific API errors if available
        if (fetchMyGroupInvitationError) console.error("Invitation fetch error:", fetchMyGroupInvitationError.message);
        if (fetchMyGroupJoinRequestError) console.error("Join request fetch error:", fetchMyGroupJoinRequestError.message);
        setCurrentUserGroupStatus('unknown_error');
      }
    };

    if (group && currentUser && groupId && members) { // Ensure essential data is present
        determineStatus();
    }
  // Dependencies: currentUser, group, groupId, members, isLoadingMembers, fetchMyGroupInvitation, fetchMyGroupJoinRequest
  }, [currentUser, group, groupId, members, isLoadingMembers, fetchMyGroupInvitation, fetchMyGroupJoinRequest, fetchMyGroupInvitationError, fetchMyGroupJoinRequestError]);
  

  // Consolidate error display
  useEffect(() => {
    if (groupApiError && !group) setError(groupApiError.message || 'Failed to load group details.');
    // userApiError is handled within loadGroupDetails for now
  }, [groupApiError, group]);

  const handleJoinRequestAction = async (requestId: string, action: 'accept' | 'decline') => {
    if (!groupId) return;
    setRequestActionFeedback(null);
    try {
      await updateJoinRequest(
        `/api/groups/${groupId}/join-requests/${requestId}/${action}`,
        {}, // Empty body for PUT, or include payload if API requires
        (updatedRequest) => {
          setRequestActionFeedback({ type: 'success', message: `Request ${action}ed successfully.` });
          loadJoinRequests(groupId); // Re-fetch to update the list
          console.log(`Join request ${action}ed:`, updatedRequest);
        }
      );
    } catch (err: any) {
      setRequestActionFeedback({ type: 'error', message: updateJoinRequestApiError?.message || `Failed to ${action} request.` });
      console.error(`Failed to ${action} join request:`, err);
    }
  };

  const handleRequestToJoin = async () => {
    if (!currentUser || !group || !groupId) {
      setJoinRequestError("Cannot send request: Missing user or group information.");
      return;
    }
    if (currentUser.id === group.creator_id) {
        setJoinRequestError("Creator cannot request to join their own group.");
        return;
    }

    setIsProcessingJoin(true);
    setJoinRequestError(null);

    try {
      await sendJoinRequest(
        `/api/groups/${groupId}/join-requests`,
        {}, // Empty body for POST, or include a message if your API supports it
        (newRequest) => {
          // setJoinRequestStatus('requested'); // Replaced by currentUserGroupStatus
          setCurrentUserGroupStatus('pending_join_request');
          if(newRequest.id) setPendingJoinRequestId(newRequest.id);
          // Optionally, display a success message to the user via a toast or state update
          console.log("Join request sent successfully:", newRequest);
        }
      );
    } catch (err: any) {
      // The useRequest hook's error (joinRequestApiError) should be populated
      setJoinRequestError(joinRequestApiError?.message || 'Failed to send join request.');
      // setJoinRequestStatus('error'); // Replaced by currentUserGroupStatus
      setCurrentUserGroupStatus('unknown_error'); // Or a more specific error state
      console.error("Failed to send join request:", err);
    } finally {
      setIsProcessingJoin(false);
    }
  };

  const handleCancelJoinRequest = async () => {
    if (!groupId || !pendingJoinRequestId || !currentUser) {
      setJoinRequestError("Cannot cancel request: Missing group or request ID.");
      return;
    }
    setIsProcessingJoin(true); // Reuse for loading state
    setJoinRequestError(null);
    try {
      await cancelJoinRequest(`/api/groups/${groupId}/join-requests/${pendingJoinRequestId}`);
      setCurrentUserGroupStatus('not_affiliated');
      setPendingJoinRequestId(null);
      // Optionally, display success message
      console.log("Join request cancelled successfully.");
    } catch (err: any) {
      setJoinRequestError(cancelJoinRequestApiError?.message || 'Failed to cancel join request.');
      // Keep current status or set to unknown_error, as the request might still exist
      console.error("Failed to cancel join request:", err);
    } finally {
      setIsProcessingJoin(false);
    }
  };

  const handleLeaveGroup = async () => {
    if (!groupId || !currentUser) {
      // setSomeError("Cannot leave group: Missing information."); // Define a generic error state if needed
      console.error("Cannot leave group: Missing information.");
      return;
    }
    // Consider adding a confirmation dialog here
    setIsProcessingJoin(true); // Reuse for loading state, or create a new one e.g., setIsProcessingLeave
    try {
      await leaveGroupRequest(`/api/groups/${groupId}/members/me`); // Assuming 'me' resolves to current user on backend
      setCurrentUserGroupStatus('not_affiliated');
      // Optionally, display success message and redirect or refresh data
      console.log("Successfully left the group.");
      // router.push('/groups'); or refresh group data
      loadMembers(groupId); // Refresh members list
    } catch (err: any) {
      // setSomeError(leaveGroupApiError?.message || 'Failed to leave group.');
      console.error("Failed to leave group:", leaveGroupApiError?.message || err);
    } finally {
      setIsProcessingJoin(false); // Or setIsProcessingLeave(false)
    }
  };

  const onInviteUserSubmit: SubmitHandler<InviteUserFormValues> = async (data) => {
    setInviteError(null);
    setInviteSuccess(null);

    if (!groupId || !currentUser || !group) {
      setInviteError("Cannot send invite: Missing group or user information.");
      return;
    }

    // Basic check: only group creator can invite (can be expanded later)
    // if (currentUser.id !== group.creator_id) {
    //   setInviteError("Only the group creator can invite users.");
    //   return;
    // }

    try {
      await sendInviteRequest(
        `/api/groups/${groupId}/invitations`,
        { invitee_id: data.userIdToInvite /* inviter_id will be set by backend */ },
        (newInvitation: GroupInvitation) => {
          setInviteSuccess(`Invitation sent to user ${data.userIdToInvite}!`);
          resetInviteForm();
          // TODO: Optionally, update a list of pending invitations or group members
          console.log('Invitation sent:', newInvitation);
        }
      );
    } catch (err: any) {
      // error from useRequest hook is captured in sendInviteApiError
      // This catch block is for other potential errors during the submit process itself
      console.error("Invitation submission error:", err);
      setInviteError(sendInviteApiError?.message || err.message || 'Failed to send invitation.');
    }
  };


  if (isLoading) {
    return <div className="container mx-auto p-4 text-center text-white"><Text>Loading group details...</Text></div>;
  }

  if (error) {
    return <div className="container mx-auto p-4 text-center text-red-500"><Text>Error: {error}</Text></div>;
  }

  if (!group) {
    return <div className="container mx-auto p-4 text-center text-white"><Text>Group not found.</Text></div>;
  }

  const creatorName = creator ? `${creator.first_name} ${creator.last_name}` : (group.creator_first_name && group.creator_last_name ? `${group.creator_first_name} ${group.creator_last_name}`: `User ID: ${group.creator_id.substring(0,8)}...`);
  const creatorAvatar = creator?.avatar_url || ''; // Fallback if creator details are embedded or not found

  return (
    <div className="container mx-auto p-4 text-white">
      <header className="mb-8 p-6 bg-gray-800 rounded-lg shadow-xl">
        <Heading level={1} className="mb-2">{group.title}</Heading>
        <Text className="text-gray-300 mb-4">{group.description}</Text>
        <div className="flex items-center text-sm text-gray-400">
          <Avatar
            src={creatorAvatar || null}
            initials={!creatorAvatar ? creatorName.substring(0,1) : undefined}
            alt={creatorName}
            className="h-8 w-8 mr-2"
          />
          Created by:
          {creator ? (
            <Link href={`/profile/${creator.id}`} className="ml-1 text-purple-400 hover:underline">
              {creatorName}
            </Link>
          ) : (
            <span className="ml-1">{creatorName}</span>
          )}
          <span className="mx-2">|</span>
          Created on: {new Date(group.created_at).toLocaleDateString()}
        </div>
      </header>

      {/* Group Actions Section */}
      <section className="my-8 p-6 bg-gray-800 rounded-lg shadow-xl">
        <Heading level={2} className="mb-4">Group Actions</Heading>
        <div className="flex flex-col space-y-3">
          {currentUserGroupStatus === 'loading' && <Text className="text-gray-400">Loading status...</Text>}

          {currentUserGroupStatus === 'not_affiliated' && (
            <Button
              onClick={handleRequestToJoin}
              disabled={isProcessingJoin || isSendingJoin}
              className="w-full sm:w-auto"
            >
              {isProcessingJoin || isSendingJoin ? 'Sending Request...' : 'Request to Join Group'}
            </Button>
          )}

          {currentUserGroupStatus === 'pending_join_request' && (
            <>
              <Text className="text-yellow-400">Your request to join this group is pending.</Text>
              <Button
                onClick={handleCancelJoinRequest}
                disabled={isProcessingJoin || isCancelingJoinRequest}
                color="red"
                className="w-full sm:w-auto"
              >
                {isProcessingJoin || isCancelingJoinRequest ? 'Cancelling...' : 'Cancel Join Request'}
              </Button>
            </>
          )}

          {currentUserGroupStatus === 'pending_invitation' && (
            <Text className="text-blue-400">
              You have a pending invitation to this group. Check your <Link href="/notifications" className="underline hover:text-blue-300">notifications</Link> to respond.
            </Text>
          )}

          {currentUserGroupStatus === 'member' && (
            <>
              <Text className="text-green-400">You are a member of this group.</Text>
              <Button
                onClick={handleLeaveGroup}
                disabled={isProcessingJoin || isLeavingGroup} // Re-use isProcessingJoin or use isLeavingGroup
                color="red" // Or a suitable color for leaving
                className="w-full sm:w-auto"
              >
                {isProcessingJoin || isLeavingGroup ? 'Leaving...' : 'Leave Group'}
              </Button>
            </>
          )}
          
          {currentUserGroupStatus === 'creator' && (
             <Text className="text-purple-400 italic">You are the creator of this group.</Text>
          )}

          {joinRequestError && <Text className="text-red-500 mt-2">{joinRequestError}</Text>}
          {currentUserGroupStatus === 'unknown_error' && !joinRequestError && (
            <Text className="text-red-500">An error occurred determining your group status or with your last action.</Text>
          )}
        </div>
      </section>

      {/* Invite User Form */}
      {currentUser && group && !isLoadingMembers && (currentUser.id === group.creator_id || members.some(member => member.id === currentUser.id)) && (
        <section className="p-6 bg-gray-800 rounded-lg shadow-xl">
          <Heading level={2} className="mb-4">Invite User to Group</Heading>
          <form onSubmit={handleSubmitInvite(onInviteUserSubmit)} className="space-y-4">
            <div>
              <label htmlFor="userIdToInvite" className="block text-sm font-medium text-gray-300 mb-1">
                User ID to Invite
              </label>
              <Input
                id="userIdToInvite"
                type="text"
                {...registerInvite('userIdToInvite', { required: 'User ID is required.' })}
                className="w-full bg-gray-700 border-gray-600 text-white placeholder-gray-400"
                placeholder="Enter exact User ID"
              />
              {inviteFormErrors.userIdToInvite && <Text className="mt-1 text-sm text-red-400">{inviteFormErrors.userIdToInvite.message}</Text>}
            </div>

            {inviteSuccess && <Text className="text-sm text-green-400">{inviteSuccess}</Text>}
            {inviteError && <Text className="text-sm text-red-400">{inviteError}</Text>}
            {sendInviteApiError && !inviteError && <Text className="text-sm text-red-400">Error: {sendInviteApiError.message}</Text>}

            <Button type="submit" disabled={isSendingInvite} className="w-full sm:w-auto">
              {isSendingInvite ? 'Sending Invite...' : 'Send Invitation'}
            </Button>
          </form>
        </section>
      )}

      {/* Pending Join Requests Section */}
      {currentUser && group && currentUser.id === group.creator_id && (
        <section className="mt-8 p-6 bg-gray-800 rounded-lg shadow-xl">
          <Heading level={2} className="mb-4">Pending Join Requests</Heading>
          {isLoadingJoinRequests && <Text>Loading join requests...</Text>}
          {joinRequestsError && <Text className="text-red-500">Error: {joinRequestsError}</Text>}
          
          {requestActionFeedback && (
            <Alert open={!!requestActionFeedback} onClose={() => setRequestActionFeedback(null)} className="mb-4">
              <AlertTitle>{requestActionFeedback.type === 'success' ? 'Success' : 'Error'}</AlertTitle>
              <AlertDescription>{requestActionFeedback.message}</AlertDescription>
              <div className="mt-4 flex justify-end">
                <Button onClick={() => setRequestActionFeedback(null)}>OK</Button>
              </div>
            </Alert>
          )}

          {!isLoadingJoinRequests && !joinRequestsError && joinRequests.filter(r => r.status === 'pending').length === 0 && (
            <Text className="text-gray-400">No pending join requests.</Text>
          )}

          {!isLoadingJoinRequests && !joinRequestsError && joinRequests.filter(r => r.status === 'pending').length > 0 && (
            <div className="space-y-4">
              {joinRequests.filter(r => r.status === 'pending').map((request) => (
                <div key={request.id} className="p-4 bg-gray-700 rounded-lg shadow flex justify-between items-center">
                  <div className="flex items-center">
                    <Avatar
                      className="h-10 w-10 mr-3"
                      initials={request.user_id ? request.user_id.substring(0, 1).toUpperCase() : "U"}
                      alt={`User ${request.user_id ? request.user_id.substring(0,8) : 'Unknown'}`}
                    />
                    <div>
                      <Text className="font-semibold">
                        {`User ID: ${request.user_id ? request.user_id.substring(0,8) : 'Unknown'}...`}
                      </Text>
                      <Text className="text-xs text-gray-400">Wants to join this group</Text>
                    </div>
                  </div>
                  <div className="space-x-2">
                    <Button
                      outline
                      onClick={() => handleJoinRequestAction(request.id, 'accept')}
                      disabled={isUpdatingJoinRequest}
                    >
                      Accept
                    </Button>
                    <Button
                      color="red"
                      onClick={() => handleJoinRequestAction(request.id, 'decline')}
                      disabled={isUpdatingJoinRequest}
                    >
                      Decline
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Placeholder for Group Content (Posts) */}
      <section className="my-8 p-6 bg-gray-800 rounded-lg shadow-xl"> {/* Added margin and styling like other sections */}
        <Heading level={2} className="mb-4">Group Posts</Heading>
        {/* Logic to fetch and display group posts */}
        <div className="bg-gray-700 p-4 rounded-lg"> {/* Consistent with other inner divs */}
          <Text className="text-gray-400 italic">Group posts will be displayed here.</Text>
        </div>
      </section>

      {/* Group Members Section */}
      <section className="mt-8 p-6 bg-gray-800 rounded-lg shadow-xl">
        <Heading level={2} className="mb-4">
          Members ({members.length > 0 ? members.length : (group?.member_count || 0)})
        </Heading>
        {isLoadingMembers && <Text>Loading members...</Text>}
        {membersError && <Text className="text-red-500">Error: {membersError}</Text>}
        
        {!isLoadingMembers && !membersError && members.length === 0 && (
          <Text className="text-gray-400">No members in this group yet.</Text>
        )}

        {!isLoadingMembers && !membersError && members.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {members.map((member) => (
              <Link key={member.id} href={`/profile/${member.id}`} passHref>
                <div className="bg-gray-700 p-4 rounded-lg shadow hover:bg-gray-600 transition-colors cursor-pointer flex flex-col items-center text-center">
                  <Avatar
                    className="h-16 w-16 mb-2"
                    src={member.avatar_url || undefined}
                    alt={`${member.first_name} ${member.last_name}`}
                    initials={!member.avatar_url ? `${member.first_name?.substring(0,1)}${member.last_name?.substring(0,1)}` : undefined}
                  />
                  <Text className="font-semibold truncate w-full">{member.first_name} {member.last_name}</Text>
                  <Text className="text-xs text-gray-400 truncate w-full">@{member.username}</Text>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}