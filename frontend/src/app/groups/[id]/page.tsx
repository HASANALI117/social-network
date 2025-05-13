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

interface GroupMembersApiResponse {
  members: User[];
  // Pagination fields if applicable
}

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

  // State for Join Request
  const [joinRequestStatus, setJoinRequestStatus] = useState<'idle' | 'pending' | 'requested' | 'member' | 'creator' | 'error'>('idle');
  const [joinRequestError, setJoinRequestError] = useState<string | null>(null);
  const [isProcessingJoin, setIsProcessingJoin] = useState(false);

  // State for managing join requests list
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
  const { get: fetchGroupRequest, error: groupApiError } = useRequest<Group>();
  const { get: fetchUserRequest, error: userApiError } = useRequest<User>(); // For creator

  // Hooks for fetching and updating join requests
  const { get: fetchJoinRequests, error: fetchJoinRequestsApiError } = useRequest<GroupJoinRequestsApiResponse>();
  const { put: updateJoinRequest, isLoading: isUpdatingJoinRequest, error: updateJoinRequestApiError } = useRequest<GroupJoinRequest>();
  const { get: fetchMembers, error: fetchMembersApiError } = useRequest<GroupMembersApiResponse>();

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
    if (groupId && currentUser && group) {
      if (currentUser.id === group.creator_id) {
        setJoinRequestStatus('creator');
        loadJoinRequests(groupId); // Load join requests if creator
      } else {
        // Placeholder: In a real app, fetch actual status (member, pending invite, pending request)
        // For now, if not creator, assume 'idle' to show the button.
        // This might need adjustment if the user is already a member or has a pending request.
        // We'll set it to 'idle' and let the button logic handle visibility.
        // A more robust check would involve an API call here.
        setJoinRequestStatus('idle');
      }
    }
  }, [groupId, currentUser, group, loadJoinRequests]); // Added loadJoinRequests
  
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
          setJoinRequestStatus('requested');
          // Optionally, display a success message to the user via a toast or state update
          console.log("Join request sent successfully:", newRequest);
        }
      );
    } catch (err: any) {
      // The useRequest hook's error (joinRequestApiError) should be populated
      setJoinRequestError(joinRequestApiError?.message || 'Failed to send join request.');
      setJoinRequestStatus('error');
      console.error("Failed to send join request:", err);
    } finally {
      setIsProcessingJoin(false);
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

      {/* Group Actions: Request to Join */}
      <section className="my-8 p-6 bg-gray-800 rounded-lg shadow-xl">
        <Heading level={2} className="mb-4">Group Actions</Heading>
        <div className="flex flex-col space-y-2">
          {currentUser && group && currentUser.id !== group.creator_id && joinRequestStatus === 'idle' && (
            <Button
              onClick={handleRequestToJoin}
              disabled={isProcessingJoin || isSendingJoin}
              className="w-full sm:w-auto"
            >
              {isProcessingJoin || isSendingJoin ? 'Sending Request...' : 'Request to Join Group'}
            </Button>
          )}
          {joinRequestStatus === 'requested' && (
            <Text className="text-green-400">Join request sent!</Text>
          )}
          {joinRequestStatus === 'member' && ( /* This status would be set by a more complete initial check */
            <Text className="text-blue-400">You are a member of this group.</Text>
          )}
          {currentUser && group && currentUser.id === group.creator_id && (
             <Text className="text-gray-400 italic">You are the creator of this group.</Text>
          )}
          {joinRequestError && <Text className="text-red-500">{joinRequestError}</Text>}
          {joinRequestStatus === 'error' && !joinRequestError && <Text className="text-red-500">An error occurred with your join request.</Text>}
        </div>
      </section>

      {/* Invite User Form */}
      {currentUser && group && (currentUser.id === group.creator_id) && (
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