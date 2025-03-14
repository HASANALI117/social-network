"use client";

import { FiPlus } from 'react-icons/fi';
import { useForm } from 'react-hook-form';
import { useRequest } from '@/hooks/useRequest';
import { useUserStore } from '@/store/useUserStore';
import { CreatePostFormValues, Post } from '@/types/Post';
import { useState } from 'react';

interface CreatePostFormProps {
  onSubmit?: (post: Post) => void;
}

export default function CreatePostForm({ onSubmit }: CreatePostFormProps) {
  const [error, setError] = useState<string | null>(null);
  const { user } = useUserStore();
  const { post: createPost, isLoading } = useRequest<Post>();

  const {
    register,
    handleSubmit,
    reset,
    formState: { isValid }
  } = useForm<CreatePostFormValues>({
    defaultValues: {
      title: '',
      content: '',
      privacy: 'public'
    }
  });

  const onSubmitForm = handleSubmit(async (data) => {
    if (!user) {
      setError('You must be logged in to create a post');
      return;
    }

    try {
      const result = await createPost('/api/posts/create', {
        ...data,
        userId: user.id
      });

      if (result) {
        reset();
        if (onSubmit) {
          onSubmit(result);
        }
      }
    } catch (err) {
      setError('Failed to create post. Please try again.');
    }
  });

  return (
    <form onSubmit={onSubmitForm} className="mb-6 bg-gray-800 rounded-lg shadow p-4">
      {error && (
        <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 text-red-500 rounded-lg">
          {error}
        </div>
      )}
      <input
        type="text"
        {...register('title', { required: true })}
        placeholder="Title"
        className="w-full p-4 border border-gray-700 bg-gray-900 text-gray-100 rounded-lg mb-4 focus:ring-2 focus:ring-purple-500 focus:border-transparent"
      />
      <textarea
        {...register('content', { required: true })}
        placeholder="What's on your mind?"
        className="w-full p-4 border border-gray-700 bg-gray-900 text-gray-100 rounded-lg mb-4 resize-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
        rows={3}
      />
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-4">
          <button
            type="button"
            className="text-purple-400 hover:text-purple-300 flex items-center gap-2"
          >
            <FiPlus />
            Add Image
          </button>
          <select
            {...register('privacy')}
            className="bg-gray-900 border border-gray-700 text-gray-100 rounded-lg px-3 py-2 focus:ring-2 focus:ring-purple-500 focus:border-transparent"
          >
            <option value="public">Public</option>
            <option value="friends">Friends Only</option>
            <option value="private">Private</option>
          </select>
        </div>
        <button
          type="submit"
          disabled={!isValid || isLoading || !user}
          className="bg-purple-700 text-gray-100 px-6 py-2 rounded-full hover:bg-purple-600 transition-colors disabled:opacity-50 disabled:hover:bg-purple-700"
        >
          {isLoading ? 'Posting...' : 'Post'}
        </button>
      </div>
    </form>
  );
}
