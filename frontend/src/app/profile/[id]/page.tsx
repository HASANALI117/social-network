'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { UserType } from '@/types/User';
import { useRequest } from '@/hooks/useRequest';
import ProfileHeader from '@/components/profile/ProfileHeader';
import TabSwitcher from '@/components/profile/TabSwitcher';
import PostList from '@/components/profile/PostList';
import FollowersList from '@/components/profile/FollowersList';

export default function UserProfilePage() {
  const params = useParams();
  const { get, post, isLoading, error } = useRequest<UserType>();
  const [user, setUser] = useState<UserType | null>(null);
  const [activeTab, setActiveTab] = useState('posts');
  const [posts] = useState([
    {
      id: 1,
      content: 'Just launched my new portfolio website! 🚀',
      likes: 42,
      comments: 12,
      timestamp: '2024-03-01T10:00:00',
    },
    {
      id: 2,
      content: 'Learning something new everyday 💡 #coding',
      likes: 28,
      comments: 5,
      timestamp: '2024-02-28T15:30:00',
    },
  ]);

  useEffect(() => {
    if (params.id) {
      get(`/api/users/get?id=${params.id}`, (userData) => {
        setUser(userData);
      });
    }
  }, [params.id, get]);

  const handleFollow = async () => {
    if (!user) return;
    // await post('/api/users/follow', { user_id: params.id });
    
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">Error loading user profile</div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">User not found</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6 bg-gray-900 min-h-screen text-gray-100">
      <ProfileHeader
        user={user}
        isPublic={true}
        onFollow={handleFollow}
        isPreview={true}
      />

      <TabSwitcher activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'posts' ? (
        <div className="space-y-4">
          <PostList posts={posts} user={user} />
        </div>
      ) : (
        <FollowersList />
      )}
    </div>
  );
}