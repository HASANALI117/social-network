import React from 'react';
import { Input } from '../../ui/input';
import { Button } from '../../ui/button';
import { Avatar } from '../../ui/avatar';
import { Heading } from '../../ui/heading';
import { Text } from '../../ui/text';
import { Alert, AlertTitle, AlertDescription, AlertActions } from '../../ui/alert';
import { UserBasicInfo } from '../../../types/User';

interface GroupInviteUserUIProps {
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  isActualSearchLoading: boolean;
  searchError: string | null;
  setSearchError: (error: string | null) => void;
  inviteSuccess: string | null;
  setInviteSuccess: (message: string | null) => void;
  inviteError: string | null;
  setInviteError: (error: string | null) => void;
  searchResults: UserBasicInfo[];
  handleSendInvite: (userIdToInvite: string) => void;
  isInviteHookLoading: boolean;
  invitingUserId: string | null;
}

export default function GroupInviteUserUI({
  searchTerm,
  setSearchTerm,
  isActualSearchLoading,
  searchError,
  setSearchError,
  inviteSuccess,
  setInviteSuccess,
  inviteError,
  setInviteError,
  searchResults,
  handleSendInvite,
  isInviteHookLoading,
  invitingUserId,
}: GroupInviteUserUIProps) {
  return (
    <div className="mt-6 p-4 bg-gray-700 rounded-lg shadow-md">
      <Heading level={3} className="text-xl mb-4 text-gray-100">Invite Users to Group</Heading>
      <Input
        type="text"
        placeholder="Search users by name or username..."
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        className="mb-4 w-full bg-gray-800 border-gray-600 placeholder-gray-500 text-white focus:ring-purple-500 focus:border-purple-500"
      />

      {isActualSearchLoading && <Text className="text-sm text-gray-400 my-2">Searching...</Text>}

      {searchError && (
        <Alert open={!!searchError} onClose={() => setSearchError(null)} size="sm">
          <AlertTitle className="text-red-600">Search Error</AlertTitle>
          <AlertDescription>{searchError}</AlertDescription>
          <AlertActions>
            <Button plain onClick={() => setSearchError(null)}>OK</Button>
          </AlertActions>
        </Alert>
      )}

      {inviteSuccess && (
        <Alert open={!!inviteSuccess} onClose={() => setInviteSuccess(null)} size="sm">
          <AlertTitle className="text-green-600">Success</AlertTitle>
          <AlertDescription>{inviteSuccess}</AlertDescription>
          <AlertActions>
            <Button plain onClick={() => setInviteSuccess(null)}>OK</Button>
          </AlertActions>
        </Alert>
      )}

      {inviteError && (
        <Alert open={!!inviteError} onClose={() => setInviteError(null)} size="sm">
          <AlertTitle className="text-red-600">Invite Error</AlertTitle>
          <AlertDescription>{inviteError}</AlertDescription>
          <AlertActions>
            <Button plain onClick={() => setInviteError(null)}>OK</Button>
          </AlertActions>
        </Alert>
      )}

      {searchResults.length > 0 && !isActualSearchLoading && (
        <div className="space-y-2 max-h-72 overflow-y-auto pr-2">
          {searchResults.map(userResult => (
            <div key={userResult.user_id} className="flex items-center justify-between p-3 bg-gray-600 rounded-md hover:bg-gray-500 transition-colors">
              <div className="flex items-center">
                <Avatar
                  src={userResult.avatar_url || null}
                  initials={!userResult.avatar_url && userResult.first_name && userResult.last_name ? `${userResult.first_name[0]}${userResult.last_name[0]}`.toUpperCase() : userResult.username ? userResult.username[0].toUpperCase() : 'U'}
                  alt={`${userResult.first_name} ${userResult.last_name}`}
                  className="h-10 w-10 mr-3 rounded-full border-2 border-gray-500"
                />
                <div>
                  <Text className="text-base font-semibold text-purple-300">{userResult.first_name} {userResult.last_name}</Text>
                  <Text className="text-xs text-gray-400">@{userResult.username}</Text>
                </div>
              </div>
              <Button
                onClick={() => handleSendInvite(userResult.user_id)}
                disabled={(isInviteHookLoading && invitingUserId === userResult.user_id) || (!!inviteSuccess && invitingUserId === userResult.user_id)} // Disable if inviting or successfully invited
                className={`text-white font-semibold py-1 px-3 rounded-md text-sm ${
                  (isInviteHookLoading && invitingUserId === userResult.user_id) ? 'bg-gray-500' :
                  (!!inviteSuccess && invitingUserId === userResult.user_id) ? 'bg-green-700 cursor-not-allowed' :
                  'bg-green-600 hover:bg-green-700'
                }`}
              >
                {(isInviteHookLoading && invitingUserId === userResult.user_id) ? 'Inviting...' :
                 (!!inviteSuccess && invitingUserId === userResult.user_id && !inviteError) ? 'Invited!' : 'Invite'}
              </Button>
            </div>
          ))}
        </div>
      )}
      {searchTerm && !isActualSearchLoading && searchResults.length === 0 && !searchError && (
        <Text className="text-sm text-gray-400 my-2">No users found matching your search.</Text>
      )}
    </div>
  );
}