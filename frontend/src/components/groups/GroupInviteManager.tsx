"use client";

import React, { useState, useEffect, useCallback } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useRequest } from '@/hooks/useRequest';
import { UserBasicInfo } from '@/types/User';
// import { useUserStore } from '@/store/useUserStore'; // Not used directly in this component based on current logic
import { Avatar } from '@/components/ui/avatar';
import { Text, Strong } from '@/components/ui/text';
import { toast } from 'react-hot-toast';
import { debounce } from 'lodash';

interface GroupInviteManagerProps {
  groupId: string;
  currentUser: UserBasicInfo | null;
  onInviteSent?: (invitedUserId: string, successMessage: string) => void;
  onInviteError?: (errorMessage: string) => void;
  onClose?: () => void;
}

const GroupInviteManager: React.FC<GroupInviteManagerProps> = ({
  groupId,
  currentUser,
  onInviteSent,
  onInviteError,
  onClose,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [searchResults, setSearchResults] = useState<UserBasicInfo[]>([]);
  const [invitedUserIds, setInvitedUserIds] = useState<Set<string>>(new Set());

  const {
    data: searchData,
    isLoading: searchIsLoading,
    error: searchErrorActual,
    get: searchUsersGet,
  } = useRequest<UserBasicInfo[]>();

  const {
    data: inviteData, // Added to potentially use response from invite
    isLoading: inviteIsLoading,
    error: inviteErrorActual,
    post: sendInvitePost,
  } = useRequest<any>();


  const debouncedSearch = useCallback(
    debounce((query: string) => {
      if (query.trim().length > 0) {
        searchUsersGet(`/api/users/search?q=${query}`);
      } else {
        setSearchResults([]);
      }
    }, 500),
    [searchUsersGet]
  );

  useEffect(() => {
    debouncedSearch(searchTerm);
    return () => debouncedSearch.cancel();
  }, [searchTerm, debouncedSearch]);

  useEffect(() => {
    if (searchData) {
      setSearchResults(searchData.filter(user => user.user_id !== currentUser?.user_id));
    }
  }, [searchData, currentUser]);

  useEffect(() => {
    if (searchErrorActual) {
      toast.error(searchErrorActual.message || 'Failed to search users.');
      if (onInviteError) {
        onInviteError(searchErrorActual.message || 'Failed to search users.');
      }
    }
  }, [searchErrorActual, onInviteError]);

  useEffect(() => {
    if (inviteErrorActual) {
      toast.error(inviteErrorActual.message || 'Failed to send invite.');
      if (onInviteError) {
        onInviteError(inviteErrorActual.message || 'Failed to send invite.');
      }
    }
  }, [inviteErrorActual, onInviteError]);


  const handleInvite = async (userIdToInvite: string) => {
    const response = await sendInvitePost(`/api/groups/${groupId}/invitations`, { invitee_id: userIdToInvite });
    if (response && !inviteErrorActual) { // Check inviteErrorActual from the hook state
      toast.success('Invite sent successfully!');
      setInvitedUserIds(prev => new Set(prev).add(userIdToInvite));
      if (onInviteSent) {
        onInviteSent(userIdToInvite, 'Invite sent successfully!');
      }
    }
  };

  return (
    <div className="p-4 border rounded-lg shadow-sm bg-white dark:bg-gray-800">
      <div className="flex justify-between items-center mb-4">
        <Strong className="text-lg">Invite Users to Group</Strong>
        {onClose && (
          <Button plain onClick={onClose} className="text-sm"> {/* Use plain for ghost-like, className for size */}
            Close
          </Button>
        )}
      </div>
      <Input
        type="text"
        placeholder="Search users by name or username..."
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        className="mb-4"
      />
      {searchIsLoading && <Text className="text-sm text-gray-500">Searching...</Text>}
      {!searchIsLoading && searchResults.length === 0 && searchTerm.length > 0 && (
        <Text className="text-sm text-gray-500">No users found.</Text>
      )}
      {!searchIsLoading && searchResults.length > 0 && (
        <ul className="space-y-2 max-h-60 overflow-y-auto">
          {searchResults.map((user) => (
            <li key={user.user_id} className="flex items-center justify-between p-2 border-b dark:border-gray-700">
              <div className="flex items-center space-x-2">
                <Avatar
                  src={user.avatar_url || null}
                  initials={`${user.first_name?.[0] || ''}${user.last_name?.[0] || ''}`}
                  alt={user.username}
                  className="h-8 w-8"
                />
                <div>
                  <Strong className="text-sm">{user.first_name} {user.last_name}</Strong>
                  <Text className="text-xs text-gray-500">@{user.username}</Text>
                </div>
              </div>
              <Button
                onClick={() => handleInvite(user.user_id)}
                disabled={inviteIsLoading || invitedUserIds.has(user.user_id)}
                className="text-sm py-1 px-2" // Adjust padding for a smaller button feel
              >
                {invitedUserIds.has(user.user_id) ? 'Invited' : 'Invite'}
              </Button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default GroupInviteManager;