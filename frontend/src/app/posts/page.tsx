"use client"

import { useEffect, useState } from 'react';
import CreatePostForm from "@/components/profile/CreatePostForm"
import { Post, PostResponse, transformPosts } from "@/types/Post"
import { useRequest } from '@/hooks/useRequest';

export default function PostsPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const { get: getPosts, isLoading } = useRequest<{ posts: PostResponse[] }>();

  useEffect(() => {
    loadPosts();
  }, []);

  const loadPosts = async () => {
    const result = await getPosts('/api/posts');
    if (result) {
      setPosts(transformPosts(result.posts));
    }
  };

  const handlePostCreated = (newPost: Post) => {
    setPosts(prevPosts => [newPost, ...prevPosts]);
  };

  return (
    <div className="p-6 min-h-screen bg-gray-900 text-white">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-2xl font-bold mb-4">Posts</h1>
        
        <div className="mb-8">
          <CreatePostForm onSubmit={handlePostCreated} />
        </div>

        {isLoading ? (
          <div className="text-center text-gray-400">Loading posts...</div>
        ) : (
          <div className="space-y-6">
            {posts.map(post => (
              <div key={post.id} className="p-6 bg-gray-800 rounded-lg">
                <h2 className="text-xl font-semibold mb-2">{post.title}</h2>
                {post.image_url && (
                  <img
                    src={post.image_url}
                    alt={post.title}
                    className="mb-4 rounded-lg"
                  />
                )}
                <p className="text-gray-300">{post.content}</p>
                <div className="mt-4 text-sm text-gray-400">
                  Posted on {new Date(post.createdAt).toLocaleDateString()}
                </div>
              </div>
            ))}
            {!isLoading && posts.length === 0 && (
              <div className="text-center text-gray-400">No posts yet</div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
