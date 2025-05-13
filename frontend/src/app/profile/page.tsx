'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { User, UserProfile } from '@/types/User';
import { Post } from '@/types/Post';
import ProfileHeader from '@/components/profile/ProfileHeader';
import EditProfileForm from '@/components/profile/EditProfileForm';
import TabSwitcher from '@/components/profile/TabSwitcher';
import CreatePostForm from '@/components/profile/CreatePostForm';
import PostList from '@/components/common/PostList';
import UserList from '@/components/profile/UserList';
import { useUserStore } from '@/store/useUserStore';
import { useRequest } from '@/hooks/useRequest';
import toast from 'react-hot-toast';

export default function ProfilePage() {
  const router = useRouter();
  const { user, isAuthenticated, update } = useUserStore();
  const { get, put } = useRequest<UserProfile>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [activeTab, setActiveTab] = useState('posts');
  const [isEditing, setIsEditing] = useState(false);

  useEffect(() => {
    const init = async () => {
      await useUserStore.persist.rehydrate();
      const { isAuthenticated: rehydratedIsAuthenticated, user: rehydratedUser } = useUserStore.getState();

      if (!rehydratedIsAuthenticated) {
        router.push('/login');
        return;
      }

      if (!rehydratedUser) {
        setIsLoading(false); // Set loading to false if user is not available after rehydration
        return;
      }

      try {
        // Use rehydratedUser.id for the API call
        get(`/api/users/${rehydratedUser.id}`, (data) => {
          setUserProfile(data);
          setError(null);
        });
      } catch (err) {
        setError(err as Error);
        toast.error('Failed to load profile');
      } finally {
        setIsLoading(false);
      }
    };

    init();
  }, [router, get]); // Removed isAuthenticated and user from dependencies as they are now fetched after rehydration

  const handleEdit = () => {
    setIsEditing(true);
  };

  const handleUpdateProfile = async (userData: Partial<User>) => {
    if (!user || !userProfile) return;
    
    await put(`/api/users/${user.id}`, userData, (data) => {
      toast.success('Profile updated successfully!');
      setUserProfile(data);
      update(data);
      setIsEditing(false);
    });
  };

  const handleTogglePrivacy = async () => {
    if (!user || !userProfile) return;

    const newPrivacyState = !userProfile.is_private;
    await put(`/api/users/${user.id}/privacy`, { is_private: newPrivacyState }, (data) => {
      setUserProfile(prevProfile => ({
        ...prevProfile,
        ...data
      }));
      update(data);
      toast.success(`Profile is now ${newPrivacyState ? 'private' : 'public'}`);
    });
  };

  const handleCreatePost = async (post: Post) => {
    if (!userProfile) return;
    
    try {
      setUserProfile({
        ...userProfile,
        latest_posts: [post, ...userProfile.latest_posts]
      });
      toast.success('Post created successfully!');
    } catch (error) {
      toast.error('Failed to create post');
      setUserProfile({
        ...userProfile,
        latest_posts: userProfile.latest_posts.filter(p => p.id !== post.id)
      });
    }
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
        <div className="text-white">Error loading profile: {error.message}</div>
      </div>
    );
  }

  if (!userProfile || !user) return null;

  return (
    <div className="max-w-4xl mx-auto p-6 bg-gray-900 min-h-screen text-gray-100">
      {isEditing ? (
        <EditProfileForm
          user={user}
          onSubmit={handleUpdateProfile}
          onCancel={() => setIsEditing(false)}
        />
      ) : (
        <ProfileHeader
          user={userProfile}
          onTogglePublic={handleTogglePrivacy}
          onEdit={handleEdit}
          pageType="own-static"
          currentUserId={user?.id}
        />
      )}

      <TabSwitcher activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'posts' && (
        <div>
          <CreatePostForm onSubmit={handleCreatePost} />
          <PostList posts={userProfile.latest_posts || []} />
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
