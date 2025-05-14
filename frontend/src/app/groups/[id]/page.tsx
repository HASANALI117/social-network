'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
// import Link from 'next/link'; // No longer directly used here
// import { useForm, SubmitHandler } from 'react-hook-form'; // Removed for Phase 1
import { useRequest } from '../../../hooks/useRequest';
import { Group, PostSummary, EventSummary } from '../../../types/Group'; // PostSummary, EventSummary might not be directly needed here anymore
import { UserBasicInfo } from '../../../types/User';
// import { Post } from '../../../types/Post'; // Removed for Phase 1
// import { GroupInvitation } from '../../../types/GroupInvitation'; // Removed for Phase 1
// import { GroupJoinRequest } from '../../../types/GroupJoinRequest'; // Removed for Phase 1
import { useUserStore } from '../../../store/useUserStore';
// import Tabs from '../../../components/common/Tabs'; // No longer directly used here
import { Heading } from '../../../components/ui/heading'; // Potentially removable if not used directly
import { Text } from '../../../components/ui/text';
// import { Textarea } from '../../../components/ui/textarea'; // Removed for Phase 1
// import { Avatar } from '../../../components/ui/avatar'; // No longer directly used here
// import { Button } from '../../../components/ui/button'; // No longer directly used here
// import { Input } from '../../../components/ui/input'; // No longer directly used here
// import { Alert, AlertTitle, AlertDescription, AlertBody, AlertActions } from '../../../components/ui/alert'; // No longer directly used here
// import { format } from 'date-fns'; // No longer directly used here

// New Component Imports
import GroupDetailHeader from '../../../components/groups/detail/GroupDetailHeader';
import GroupNonMemberView from '../../../components/groups/detail/GroupNonMemberView';
import GroupMemberView from '../../../components/groups/detail/GroupMemberView';

export default function GroupDetailPage() {
  const params = useParams();
  const router = useRouter();
  const groupId = params.id as string;

  const [group, setGroup] = useState<Group | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isMember, setIsMember] = useState<boolean>(false);

  // State for User Search and Invite UI (Phase 3b)
  const [showInviteUI, setShowInviteUI] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [searchResults, setSearchResults] = useState<UserBasicInfo[]>([]);
  const [isActualSearchLoading, setIsActualSearchLoading] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);
  
  const [invitingUserId, setInvitingUserId] = useState<string | null>(null);
  const [inviteError, setInviteError] = useState<string | null>(null);
  const [inviteSuccess, setInviteSuccess] = useState<string | null>(null);

  const { user: currentUser } = useUserStore();
  const { get: fetchGroupRequest, error: groupApiError } = useRequest<Group>();
  const { get: searchUsersRequestHook, error: searchApiHookError, isLoading: isSearchHookLoading } = useRequest<UserBasicInfo[]>();
  const { post: sendInviteRequestHook, error: inviteApiHookError, isLoading: isInviteHookLoading } = useRequest<{ message: string }>();


  const loadGroupDetails = useCallback(async (id: string) => {
    setIsLoading(true);
    setError(null);
    setGroup(null);
    setIsMember(false);
    try {
      const groupData = await fetchGroupRequest(`/api/groups/${id}`);
      if (groupData) {
        const processedGroupData = {
          ...groupData,
          posts: groupData.posts || [],
          members: groupData.members || [],
          events: groupData.events || [],
        };
        setGroup(processedGroupData);
        if (groupData.members !== undefined && groupData.members !== null) {
          setIsMember(true);
        } else {
          setIsMember(false);
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
  }, [fetchGroupRequest, groupApiError]);

  useEffect(() => {
    if (groupId) {
      loadGroupDetails(groupId);
    } else {
      setError("Group ID is missing.");
      setIsLoading(false);
    }
  }, [groupId, loadGroupDetails]);
  
  useEffect(() => {
    if (groupApiError && !isLoading && !group) {
        setError(groupApiError.message || 'Failed to load group details.');
    }
  }, [groupApiError, isLoading, group]);

  const handleRequestToJoin = () => {
    console.log('Request to Join button clicked');
    alert('Request to Join functionality will be implemented later.');
  };

  const debounce = <F extends (...args: any[]) => any>(func: F, waitFor: number) => {
    let timeout: ReturnType<typeof setTimeout> | null = null;
    const debounced = (...args: Parameters<F>) => {
      if (timeout !== null) {
        clearTimeout(timeout);
        timeout = null;
      }
      timeout = setTimeout(() => func(...args), waitFor);
    };
    return debounced as (...args: Parameters<F>) => void;
  };

  const toggleInviteUI = () => {
    setShowInviteUI(prev => !prev);
    if (showInviteUI) {
      setSearchTerm('');
      setSearchResults([]);
      setSearchError(null);
      setInviteError(null);
      setInviteSuccess(null);
    }
  };

  const handleSearchUsers = useCallback(async (query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      setSearchError(null);
      setIsActualSearchLoading(false);
      return;
    }
    setIsActualSearchLoading(true);
    setSearchError(null);
    setInviteError(null);
    setInviteSuccess(null);
    try {
      const results = await searchUsersRequestHook(`/api/users/search?q=${encodeURIComponent(query)}`);
      if (results) {
        setSearchResults(results);
      } else if (searchApiHookError) {
        setSearchError(searchApiHookError.message || 'Failed to search users.');
        setSearchResults([]);
      } else {
        setSearchResults([]);
      }
    } catch (err: any) {
      setSearchError(err.message || 'An unexpected error occurred during search.');
      setSearchResults([]);
    } finally {
      setIsActualSearchLoading(false);
    }
  }, [searchUsersRequestHook, searchApiHookError]);

  const debouncedSearchUsers = useCallback(debounce(handleSearchUsers, 400), [handleSearchUsers]);

  useEffect(() => {
    if (searchTerm.trim()) {
      debouncedSearchUsers(searchTerm);
    } else {
      setSearchResults([]);
      setSearchError(null);
      if(isActualSearchLoading && !searchTerm.trim()){
          setIsActualSearchLoading(false);
      }
    }
  }, [searchTerm, debouncedSearchUsers, isActualSearchLoading]);

  const handleSendInvite = async (userIdToInvite: string) => {
    if (!groupId) return;
    setInvitingUserId(userIdToInvite);
    setInviteError(null);
    setInviteSuccess(null);
    try {
      const response = await sendInviteRequestHook(`/api/groups/${groupId}/invite`, { user_id: userIdToInvite });
      if (response) {
        const invitedUser = searchResults.find(u => u.user_id === userIdToInvite);
        setInviteSuccess(`Invitation sent to ${invitedUser ? `${invitedUser.first_name} ${invitedUser.last_name} (@${invitedUser.username})` : 'user'}!`);
      } else if (inviteApiHookError) {
        setInviteError(inviteApiHookError.message || 'Failed to send invitation.');
      }
    } catch (err: any) {
      setInviteError(err.message || 'An unexpected error occurred while sending invitation.');
    } finally {
      setInvitingUserId(null);
    }
  };

  const handleLeaveGroup = () => {
    console.log('Leave Group button clicked');
    alert('Leave Group functionality will be implemented later.');
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

  const {
    name,
    description,
    avatar_url,
    creator_info,
    created_at,
    members_count,
    posts_count,
    events_count,
  } = group;

  return (
    <div className="min-h-screen bg-gray-900 text-white p-4 md:p-8">
      <div className="container mx-auto max-w-4xl">
        <GroupDetailHeader
          name={name}
          description={description}
          avatar_url={avatar_url}
          creator_info={creator_info}
          created_at={created_at}
        />

        <section className="p-6 bg-gray-800 rounded-lg shadow-xl">
          {!isMember ? (
            <GroupNonMemberView
              members_count={members_count}
              posts_count={posts_count}
              events_count={events_count}
              handleRequestToJoin={handleRequestToJoin}
            />
          ) : (
            <GroupMemberView
              group={group} // Pass the full group object
              showInviteUI={showInviteUI}
              toggleInviteUI={toggleInviteUI}
              handleLeaveGroup={handleLeaveGroup}
              // Invite UI props
              searchTerm={searchTerm}
              setSearchTerm={setSearchTerm}
              isActualSearchLoading={isActualSearchLoading}
              searchError={searchError}
              setSearchError={setSearchError}
              inviteSuccess={inviteSuccess}
              setInviteSuccess={setInviteSuccess}
              inviteError={inviteError}
              setInviteError={setInviteError}
              searchResults={searchResults}
              handleSendInvite={handleSendInvite}
              isInviteHookLoading={isInviteHookLoading}
              invitingUserId={invitingUserId}
            />
          )}
        </section>
      </div>
    </div>
  );
}