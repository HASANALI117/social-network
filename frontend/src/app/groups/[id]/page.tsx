'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
// import Link from 'next/link'; // No longer directly used here
// import { useForm, SubmitHandler } from 'react-hook-form'; // Removed for Phase 1
import { useRequest } from '../../../hooks/useRequest';
import { Group, PostSummary, EventSummary } from '../../../types/Group'; // PostSummary, EventSummary might not be directly needed here anymore
import { UserBasicInfo } from '../../../types/User';
// import { Post } from '../../../types/Post';
// import { GroupInvitation } from '../../../types/GroupInvitation';
// import { GroupJoinRequest } from '../../../types/GroupJoinRequest';
import { useUserStore } from '../../../store/useUserStore';
// import Tabs from '../../../components/common/Tabs';
import { Heading } from '../../../components/ui/heading';
import { Text } from '../../../components/ui/text';
// import { Textarea } from '../../../components/ui/textarea';
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

  const { user: currentUser } = useUserStore();
  const { get: fetchGroupRequest, error: groupApiError } = useRequest<Group>();
  // Removed old invite UI state and hooks

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

  // Removed debounce, toggleInviteUI, handleSearchUsers, debouncedSearchUsers, handleSendInvite, and related useEffects
  // as this logic is now encapsulated in GroupInviteManager.tsx or will be in GroupMemberView.tsx

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
              group={group}
              currentUser={currentUser ? { ...currentUser, user_id: currentUser.id } : null}
              handleLeaveGroup={handleLeaveGroup}
              // Old invite props removed
            />
          )}
        </section>
      </div>
    </div>
  );
}