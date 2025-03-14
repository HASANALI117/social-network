'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { UserType } from '@/types/User';
import ProfileHeader from '@/components/profile/ProfileHeader';
import TabSwitcher from '@/components/profile/TabSwitcher';
import CreatePostForm from '@/components/profile/CreatePostForm';
import PostList from '@/components/profile/PostList';
import FollowersList from '@/components/profile/FollowersList';
import { useUserStore } from '@/store/useUserStore';

export default function ProfilePage() {
  const router = useRouter();
  const { user, isAuthenticated } = useUserStore();
  const [isLoading, setIsLoading] = useState(true);
  const [isPublic, setIsPublic] = useState(true);
  const [activeTab, setActiveTab] = useState('posts');
  const [posts, setPosts] = useState([
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
    // Handle store hydration
    useUserStore.persist.rehydrate()
    
    // Check authentication after hydration
    if (!isAuthenticated) {
      router.push('/login');
    } else {
      setIsLoading(false);
    }
  }, [isAuthenticated, router]);

  const handleFollow = () => {
    // TODO: Implement follow logic
  };

  const handleCreatePost = (content: string) => {
    setPosts([
      {
        id: posts.length + 1,
        content,
        likes: 0,
        comments: 0,
        timestamp: new Date().toISOString(),
      },
      ...posts,
    ]);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-white">Loading...</div>
      </div>
    );
  }

  if (!user) return null;

  return (
    <div className="max-w-4xl mx-auto p-6 bg-gray-900 min-h-screen text-gray-100">
      <ProfileHeader
        user={user}
        isPublic={isPublic}
        onTogglePublic={() => setIsPublic(!isPublic)}
        onFollow={handleFollow}
      />

      <TabSwitcher activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'posts' ? (
        <div>
          <CreatePostForm onSubmit={handleCreatePost} />
          <PostList posts={posts} user={user} />
        </div>
      ) : (
        <FollowersList />
      )}
    </div>
  );
}
