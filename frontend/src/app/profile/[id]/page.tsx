'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { UserProfile } from '@/types/User';
import { useRequest } from '@/hooks/useRequest';
import { useUserStore } from '@/store/useUserStore'; // To get current user ID
import toast from 'react-hot-toast';
import ProfileHeader from '@/components/profile/ProfileHeader';
import TabSwitcher from '@/components/profile/TabSwitcher';
import PostList from '@/components/profile/PostList';
import UserList from '@/components/profile/UserList';

export default function UserProfilePage() {
  const params = useParams();
  const { get, post, del, isLoading: isLoadingProfile, error: profileError } = useRequest<UserProfile>();
  // Separate loading state for follow actions
  const { post: postFollow, del: delFollow, isLoading: isLoadingFollowAction, error: followError } = useRequest();
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [activeTab, setActiveTab] = useState('posts');
  const currentUser = useUserStore((state) => state.user);
  const hydrated = useUserStore((state) => state.hydrated);
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);


  // Manually trigger rehydration if not using onRehydrateStorage effectively
  // This is a common workaround for `skipHydration: true`
  useEffect(() => {
    useUserStore.persist.rehydrate();
  }, []);

  const fetchProfileData = useCallback(() => {
    if (params.id) {
      // The backend should now return follow_status and is_private
      get(`/api/users/${params.id}`, (userData) => {
        // console.log('[UserProfilePage] API userData:', userData); // Commented out API response log
        setUserProfile(userData);
      });
    }
  }, [params.id, get]);

  useEffect(() => {
    fetchProfileData();
  }, [fetchProfileData]);

  const handleFollowAction = async (actionType: 'follow' | 'unfollow' | 'cancel_request' | 'accept_request' | 'decline_request') => {
    if (!userProfile || !params.id) return;
    const targetId = params.id as string; // ID of the profile being viewed

    try {
      switch (actionType) {
        case 'follow':
          await postFollow(`/api/users/${targetId}/follow`, {});
          // The backend will determine the new state (is_followed, follow_request_state)
          // e.g. if user is private, follow_request_state becomes 'SENT'
          // if public, is_followed becomes true.
          // We will show a generic message here, and the UI will update based on fetchProfileData
          toast.success('Follow action processed. Profile updating...');
          break;
        case 'unfollow':
          await delFollow(`/api/users/${targetId}/unfollow`);
          toast.success('Unfollowed successfully. Profile updating...');
          break;
        case 'cancel_request':
          // API for cancelling a request sent by current user to targetId
          await delFollow(`/api/users/${targetId}/cancel-follow-request`);
          toast.success('Follow request cancelled. Profile updating...');
          break;
        case 'accept_request':
          // Current user accepts a request from userProfile.id (the profile owner)
          await postFollow(`/api/users/${userProfile.id}/accept`, {});
          toast.success('Follow request accepted. Profile updating...');
          break;
        case 'decline_request':
          // Current user declines a request from userProfile.id (the profile owner)
          await delFollow(`/api/users/${userProfile.id}/reject`);
          toast.success('Follow request declined. Profile updating...');
          break;
        default:
          console.warn("Unhandled follow action type:", actionType);
          toast.error("Unknown action type.");
          return; // Do not proceed to fetchProfileData for unhandled types
      }
      fetchProfileData(); // Refetch to get the authoritative state from the backend
    } catch (e: any) {
      console.error("Follow action failed:", e);
      const errorMessage = e?.response?.data?.error || e.message || 'An unknown error occurred.';
      switch (actionType) {
        case 'follow':
          toast.error(`Failed to follow: ${errorMessage}`);
          break;
        case 'unfollow':
          toast.error(`Failed to unfollow: ${errorMessage}`);
          break;
        case 'cancel_request':
          toast.error(`Failed to cancel request: ${errorMessage}`);
          break;
        case 'accept_request':
          toast.error(`Failed to accept request: ${errorMessage}`);
          break;
        case 'decline_request':
          toast.error(`Failed to decline request: ${errorMessage}`);
          break;
        default:
          toast.error(`Action failed: ${errorMessage}`);
      }
      // Error is handled by useRequest hook (followError), but we show a toast here.
      // Consider if fetchProfileData() should be called here too to ensure UI consistency even on error.
      // For now, relying on the existing logic where fetchProfileData is called after the switch.
    }
  };

  // Wait for Zustand to be hydrated before rendering content that depends on auth state
  if (!hydrated) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">Loading user session...</div>
      </div>
    );
  }

  if (isLoadingProfile && !userProfile) { // Show loading only if userProfile is not yet fetched
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">Loading profile...</div>
      </div>
    );
  }

  if (profileError) {
    // The "private profile" scenario (previously a 403) is now handled by rendering a minimal profile.
    // This block handles other errors like 404, 500, or other unexpected 403s.
    const errorMessage = profileError.message.includes('404')
      ? "User not found"
      : `Error loading user profile: ${profileError.message}`; // Provide more error context

    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">{errorMessage}</div>
        {followError && <p className="text-red-500 mt-2">Follow action failed: {followError.message}</p>}
      </div>
    );
  }

  if (!userProfile) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">User not found</div>
      </div>
    );
  }

  // Determine if this is the current user's own profile
  const isOwnProfile = isAuthenticated && currentUser?.id === userProfile.id;

  // Determine if the viewer has full access
  // Full access if:
  // 1. Profile is not private OR
  // 2. Profile is private BUT viewer is the owner OR
  // 3. Profile is private BUT viewer is already following
  const hasFullAccess = !userProfile.is_private || isOwnProfile || userProfile.is_followed;

  // The restricted view applies if the profile is private AND the user doesn't have full access.
  const isRestrictedView = userProfile.is_private && !hasFullAccess;

  return (
    <div className="max-w-4xl mx-auto p-6 bg-gray-900 min-h-screen text-gray-100">
      {/* {userProfile && (console.log('[UserProfilePage] Props to ProfileHeader:', { user: userProfile, currentUserId: isAuthenticated ? currentUser?.id : undefined, isAuthenticated }), null)} // Commented out ProfileHeader props log */}
      <ProfileHeader
        user={userProfile}
        currentUserId={isAuthenticated ? currentUser?.id : undefined}
        onFollowAction={handleFollowAction}
        isPreview={false}
        isLoadingFollowAction={isLoadingFollowAction}
        pageType="dynamic"
      />
      {followError && !isLoadingFollowAction && <p className="text-red-500 mt-2 text-center">Error: {followError.message}</p>}

      {isRestrictedView ? (
        <div className="text-center py-10">
          <h2 className="text-xl font-semibold text-gray-300">This account is private.</h2>
          <p className="text-gray-400 mt-2">Follow them to see their posts and full profile.</p>
        </div>
      ) : (
        <>
          {/* Full Profile Content */}
          <div className="mb-6 text-gray-300">
            {userProfile.about_me && <p className="mb-2">{userProfile.about_me}</p>}
            <div className="text-sm text-gray-400">
              {userProfile.email && <p>Email: {userProfile.email}</p>}
              {userProfile.birth_date && ( // Only display if birth_date is meaningful
                <p>Birth date: {new Date(userProfile.birth_date).toLocaleDateString()}</p>
              )}
              {userProfile.created_at && (
                <p>Joined: {new Date(userProfile.created_at).toLocaleDateString()}</p>
              )}
            </div>
          </div>

          <TabSwitcher activeTab={activeTab} onTabChange={setActiveTab} />

          {activeTab === 'posts' && (
            <div className="space-y-4">
              <PostList posts={userProfile.latest_posts || []} />
            </div>
          )}

          {activeTab === 'followers' && (
            <UserList
              users={userProfile.latest_followers || []}
              emptyMessage="No followers yet"
            />
          )}

          {activeTab === 'following' && (
            <UserList
              users={userProfile.latest_following || []}
              emptyMessage="Not following anyone yet"
            />
          )}
        </>
      )}
    </div>
  );
}