'use client';

import { FiEdit, FiUsers, FiLock, FiUnlock, FiUserPlus, FiUserMinus, FiUserCheck, FiUserX, FiClock } from 'react-icons/fi';
import { User, UserProfile } from '@/types/User';
import { Button } from '@/components/ui/button'; // Assuming you have a Button component
import Link from 'next/link'; // Import Link
import { FiMessageSquare } from 'react-icons/fi'; // Import an icon for the chat button
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';

interface ProfileHeaderProps {
  user: UserProfile | User;
  currentUserId?: string;
  onTogglePublic?: () => void;
  onFollowAction?: (actionType: 'follow' | 'unfollow' | 'cancel_request' | 'accept_request' | 'decline_request') => void;
  onEdit?: () => void;
  isPreview?: boolean;
  isLoadingFollowAction?: boolean;
  pageType?: 'own-static' | 'dynamic';
  // Add is_followed for chat button logic, and is_following_viewer for the ideal scenario
  is_followed?: boolean;
  is_following_viewer?: boolean; // Ideal field, not currently available
}

export default function ProfileHeader({
  user,
  currentUserId,
  onTogglePublic,
  onFollowAction,
  onEdit,
  isPreview = false,
  isLoadingFollowAction = false,
  pageType = 'dynamic',
  is_followed, // Destructure new prop
  is_following_viewer, // Destructure new prop (ideal)
}: ProfileHeaderProps) {
  const { onlineUserIds } = useGlobalWebSocket();
  const isOwnProfile = user.id === currentUserId;
  const profileUser = user as UserProfile;

  // Use the passed is_followed prop, or fallback to profileUser.is_followed if not passed
  // This ensures compatibility if the prop isn't passed from all call sites immediately,
  // though for this specific task, UserProfilePage will pass it.
  const viewerIsFollowingTarget = typeof is_followed === 'boolean' ? is_followed : profileUser.is_followed;

  // Ideal condition:
  // const canChat = !isOwnProfile && (viewerIsFollowingTarget || is_following_viewer);
  // Current condition due to missing is_following_viewer:
  const canChat = !isOwnProfile && viewerIsFollowingTarget;


  const getFollowButton = () => {
    if (pageType === 'own-static') {
      return null;
    }
    
    if (isOwnProfile) {
      return (
        <Button
          className="flex items-center gap-2 bg-purple-700 text-gray-100 px-6 py-2 rounded-full opacity-50 cursor-not-allowed"
          disabled={true}
        >
          <FiUserPlus className="text-lg" />
          Follow
        </Button>
      );
    }

    if (profileUser.is_followed) {
      return (
        <Button
          onClick={() => onFollowAction?.('unfollow')}
          className="flex items-center gap-2 bg-red-600 text-gray-100 px-6 py-2 rounded-full hover:bg-red-500 transition-colors"
          disabled={isLoadingFollowAction}
        >
          <FiUserMinus className="text-lg" />
          Unfollow
        </Button>
      );
    } else {
      switch (profileUser.follow_request_state) {
        case 'SENT':
          return (
            <Button
              onClick={() => onFollowAction?.('cancel_request')}
              className="flex items-center gap-2 bg-yellow-500 text-gray-900 px-6 py-2 rounded-full hover:bg-yellow-400 transition-colors"
              disabled={isLoadingFollowAction}
            >
              <FiClock className="text-lg" />
              Requested
            </Button>
          );
        case 'RECEIVED':
          return (
            <>
              <Button
                onClick={() => onFollowAction?.('accept_request')}
                className="flex items-center gap-2 bg-green-500 text-gray-100 px-4 py-2 rounded-full hover:bg-green-400 transition-colors"
                disabled={isLoadingFollowAction}
              >
                <FiUserCheck className="text-lg" />
                Accept
              </Button>
              <Button
                onClick={() => onFollowAction?.('decline_request')}
                className="flex items-center gap-2 bg-gray-600 text-gray-100 px-4 py-2 rounded-full hover:bg-gray-500 transition-colors"
                disabled={isLoadingFollowAction}
              >
                <FiUserX className="text-lg" />
                Decline
              </Button>
            </>
          );
        default: // Includes empty or undefined follow_request_state
          return (
            <Button
              onClick={() => onFollowAction?.('follow')}
              className="flex items-center gap-2 bg-purple-700 text-gray-100 px-6 py-2 rounded-full hover:bg-purple-600 transition-colors"
              disabled={isLoadingFollowAction}
            >
              <FiUserPlus className="text-lg" />
              Follow
            </Button>
          );
      }
    }
  };

  return (
    <div className="bg-gray-800 rounded-lg shadow-lg p-6 mb-6">
      <div className="flex items-start gap-6">
        <div className="relative">
          <img
            src={
              user.avatar_url ||
              `https://ui-avatars.com/api/?name=${user.first_name}+${user.last_name}&background=3b82f6&color=fff&bold=true`
            }
            alt="Avatar"
            className="w-32 h-32 rounded-full border-4 border-purple-100"
          />
          {user.id && onlineUserIds.includes(user.id) && (
            <span className="absolute bottom-1 right-1 block h-4 w-4 rounded-full bg-green-500 ring-2 ring-white dark:ring-gray-800" />
          )}
        </div>
        <div className="flex-1">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-3xl font-bold text-gray-100">
                {user.first_name} {user.last_name}
              </h1>
              <p className="text-gray-400">@{user.username}</p>
            </div>
            <div className="flex items-center gap-4">
              {getFollowButton()}
              {/* Chat Button */}
              {canChat && (
                <Link href={`/chat/${user.id}`} passHref>
                  <Button outline className="flex items-center gap-2">
                    <FiMessageSquare className="text-lg" />
                    Message
                  </Button>
                </Link>
              )}
              {/* TODO: Add a comment here if is_following_viewer is not available:
                   "Ideally, chat is also enabled if userProfile.is_following_viewer is true."
                   This is now handled by the canChat logic and comments above.
              */}
              {!isPreview && onEdit && !isOwnProfile && ( // Edit button should not show on own profile if follow buttons are present
                <button
                  onClick={onEdit}
                  className="text-purple-400 hover:text-purple-300"
                >
                  <FiEdit className="text-2xl" />
                </button>
              )}
               {/* Show edit button on own profile regardless of follow status */}
              {!isPreview && onEdit && isOwnProfile && (
                <button
                  onClick={onEdit}
                  className="text-purple-400 hover:text-purple-300"
                >
                  <FiEdit className="text-2xl" />
                </button>
              )}
            </div>
          </div>

          <p className="text-gray-300 mb-4">{user.about_me}</p>

          <div className="flex items-center gap-6 text-gray-400">
            <div className="flex items-center gap-2">
              <FiUsers />
              <span>{('followers_count' in profileUser ? profileUser.followers_count : 0)} followers</span>
            </div>
            <div className="flex items-center gap-2">
              <FiUsers />
              <span>{('following_count' in profileUser ? profileUser.following_count : 0)} following</span>
            </div>
            {!isPreview && onTogglePublic && !isOwnProfile && ( // Toggle public only if not own profile and handler exists
              <button
                onClick={onTogglePublic}
                className="flex items-center gap-2 ml-auto text-sm px-4 py-2 rounded-full bg-gray-700 hover:bg-gray-600 text-gray-200"
              >
                {profileUser.is_private ? <FiLock /> : <FiUnlock />}
                {profileUser.is_private ? 'Private Profile' : 'Public Profile'}
              </button>
            )}
             {/* Show toggle public on own profile */}
            {!isPreview && onTogglePublic && isOwnProfile && (
              <button
                onClick={onTogglePublic}
                className="flex items-center gap-2 ml-auto text-sm px-4 py-2 rounded-full bg-gray-700 hover:bg-gray-600 text-gray-200"
              >
                {profileUser.is_private ? <FiLock /> : <FiUnlock />}
                {profileUser.is_private ? 'Private Profile' : 'Public Profile'}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
