"use client"

import { useState } from 'react';
import { User } from '@/types/User';
import ProfileHeader from '@/components/profile/ProfileHeader';
import TabSwitcher from '@/components/profile/TabSwitcher';
import CreatePostForm from '@/components/profile/CreatePostForm';
import PostList from '@/components/profile/PostList';
import FollowersList from '@/components/profile/FollowersList';

const dummyUser: User = {
  id: '1',
  username: 'johndoe',
  email: 'john@example.com',
  first_name: 'John',
  last_name: 'Doe',
  avatar_url: null,
  about_me: 'Frontend developer passionate about creating beautiful user experiences',
  birth_date: '1990-01-01',
  created_at: '2024-01-01',
  updated_at: '2024-01-01'
};

export default function ProfilePage() {
  const [isPublic, setIsPublic] = useState(true);
  const [activeTab, setActiveTab] = useState('posts');
  const [posts, setPosts] = useState([
    {
      id: 1,
      content: 'Just launched my new portfolio website! 🚀',
      likes: 42,
      comments: 12,
      timestamp: '2024-03-01T10:00:00'
    },
    {
      id: 2,
      content: 'Learning something new everyday 💡 #coding',
      likes: 28,
      comments: 5,
      timestamp: '2024-02-28T15:30:00'
    }
  ]);

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
        timestamp: new Date().toISOString()
      },
      ...posts
    ]);
  };

  return (
    <div className="max-w-4xl mx-auto p-6 bg-gray-900 min-h-screen text-gray-100">
      <ProfileHeader
        user={dummyUser}
        isPublic={isPublic}
        onTogglePublic={() => setIsPublic(!isPublic)}
        onFollow={handleFollow}
      />

      <TabSwitcher
        activeTab={activeTab}
        onTabChange={setActiveTab}
      />

      {activeTab === 'posts' ? (
        <div>
          <CreatePostForm onSubmit={handleCreatePost} />
          <PostList posts={posts} user={dummyUser} />
        </div>
      ) : (
        <FollowersList />
      )}
    </div>
  );
}
