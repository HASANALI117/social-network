'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { UserProfile } from '@/types/User';
import { useRequest } from '@/hooks/useRequest';
import ProfileHeader from '@/components/profile/ProfileHeader';
import TabSwitcher from '@/components/profile/TabSwitcher';
import PostList from '@/components/profile/PostList';
import UserList from '@/components/profile/UserList';

export default function UserProfilePage() {
  const params = useParams();
  const { get, post, isLoading, error } = useRequest<UserProfile>();
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [activeTab, setActiveTab] = useState('posts');

  useEffect(() => {
    if (params.id) {
      get(`/api/users/${params.id}`, (userData) => {
        setUserProfile(userData);
      });
    }
  }, [params.id, get]);

  const handleFollow = async () => {
    if (!userProfile) return;
    await post('/api/users/follow', { user_id: params.id });
    // Refetch user data to update followers count
    get(`/api/users/${params.id}`, (userData) => {
      setUserProfile(userData);
    });
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">Loading...</div>
      </div>
    );
  }

  if (error) {
    const errorMessage = error.message.includes('403')
      ? "This profile is private. Follow this user to see their details and posts."
      : error.message.includes('404')
      ? "User not found"
      : "Error loading user profile";

    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">{errorMessage}</div>
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

  return (
    <div className="max-w-4xl mx-auto p-6 bg-gray-900 min-h-screen text-gray-100">
      <ProfileHeader
        user={userProfile}
        isPublic={!userProfile.is_private}
        onFollow={handleFollow}
        isPreview={true}
      />

      <div className="mb-6 text-gray-300">
        <p className="mb-2">{userProfile.about_me}</p>
        <div className="text-sm text-gray-400">
          <p>Email: {userProfile.email}</p>
          <p>Birth date: {new Date(userProfile.birth_date).toLocaleDateString()}</p>
          <p>Joined: {new Date(userProfile.created_at).toLocaleDateString()}</p>
        </div>
      </div>

      <TabSwitcher activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'posts' && (
        <div className="space-y-4">
          <PostList posts={userProfile.latest_posts} />
        </div>
      )}

      {activeTab === 'followers' && (
        <UserList 
          users={userProfile.latest_followers}
          emptyMessage="No followers yet"
        />
      )}

      {activeTab === 'following' && (
        <UserList 
          users={userProfile.latest_following}
          emptyMessage="Not following anyone yet"
        />
      )}
    </div>
  );
}