'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { UserType } from '@/types/User';
import { Post, PostResponse, transformPosts } from '@/types/Post';
import ProfileHeader from '@/components/profile/ProfileHeader';
import EditProfileForm from '@/components/profile/EditProfileForm';
import TabSwitcher from '@/components/profile/TabSwitcher';
import CreatePostForm from '@/components/profile/CreatePostForm';
import PostList from '@/components/profile/PostList';
import FollowersList from '@/components/profile/FollowersList';
import { useUserStore } from '@/store/useUserStore';
import { useRequest } from '@/hooks/useRequest';
import toast from 'react-hot-toast';

export default function ProfilePage() {
  const router = useRouter();
  const { user, isAuthenticated, update } = useUserStore();
  const { put } = useRequest<UserType>();
  const { get: getPosts } = useRequest<{ posts: PostResponse[] }>();
  const [isLoading, setIsLoading] = useState(true);
  const [isPublic, setIsPublic] = useState(true);
  const [activeTab, setActiveTab] = useState('posts');
  const [isEditing, setIsEditing] = useState(false);
  const [posts, setPosts] = useState<Post[]>([]);

  useEffect(() => {
    // Handle store hydration
    useUserStore.persist.rehydrate();
    
    const init = async () => {
      // Check authentication after hydration
      if (!isAuthenticated) {
        router.push('/login');
        return;
      }

      try {
        // Load user's posts
        const result = await getPosts(`/api/posts/user?id=${user?.id}`);
        if (result?.posts) {
          setPosts(transformPosts(result.posts));
        }
      } catch (error) {
        toast.error('Failed to load posts');
      } finally {
        setIsLoading(false);
      }
    };

    init();
  }, [isAuthenticated, router, user?.id, getPosts]);

  const handleEdit = () => {
    setIsEditing(true);
  };

  const handleUpdateProfile = async (userData: Partial<UserType>) => {
    if (!user) return;
    
    await put(`/api/users/update?id=${user.id}`, userData, (data) => {
      console.log('User updated:', userData);
      toast.success('Profile updated successfully!');
      update(data);
      setIsEditing(false);
    });
  };

  const handleFollow = () => {
    // TODO: Implement follow logic
  };

  const handleCreatePost = async (post: Post) => {
    try {
      setPosts(prevPosts => [post, ...prevPosts]);
      toast.success('Post created successfully!');
    } catch (error) {
      toast.error('Failed to create post');
      setPosts(prevPosts => prevPosts.filter(p => p.id !== post.id));
    }
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
      {isEditing ? (
        <EditProfileForm
          user={user}
          onSubmit={handleUpdateProfile}
          onCancel={() => setIsEditing(false)}
        />
      ) : (
        <ProfileHeader
          user={user}
          isPublic={isPublic}
          onTogglePublic={() => setIsPublic(!isPublic)}
          onFollow={handleFollow}
          onEdit={handleEdit}
        />
      )}

      <TabSwitcher activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'posts' ? (
        <div>
          <CreatePostForm onSubmit={handleCreatePost} />
          <PostList posts={posts} />
        </div>
      ) : (
        <FollowersList />
      )}
    </div>
  );
}
