"use client";

import { FiPlus, FiX, FiImage } from 'react-icons/fi';
import { useForm } from 'react-hook-form';
import { useRequest } from '@/hooks/useRequest';
import { useUserStore } from '@/store/useUserStore';
import { CreatePostFormValues, Post } from '@/types/Post';
import { useState, useRef, useEffect, ChangeEvent } from 'react';
import toast from 'react-hot-toast';
import { uploadFileToMinio } from '../../lib/minioUploader';
import Image from 'next/image';

interface CreatePostFormProps {
  onSubmit?: (post: Post) => void;
}

const MAX_FILE_SIZE_MB = 5;
const MAX_FILE_SIZE_BYTES = MAX_FILE_SIZE_MB * 1024 * 1024;
const ALLOWED_FILE_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];

export default function CreatePostForm({ onSubmit }: CreatePostFormProps) {
  const { user } = useUserStore();
  const { post: createPostRequest, isLoading: isCreatingPost } = useRequest<Post>();
  const [isUploadingImage, setIsUploadingImage] = useState(false);
  const [selectedPostImageFile, setSelectedPostImageFile] = useState<File | null>(null);
  const [postImagePreviewUrl, setPostImagePreviewUrl] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { isValid }
  } = useForm<CreatePostFormValues>({
    defaultValues: {
      title: '',
      content: '',
      privacy: 'public',
      image_url: null
    }
  });

  useEffect(() => {
    // Clean up the object URL when the component unmounts or the preview URL changes
    return () => {
      if (postImagePreviewUrl) {
        URL.revokeObjectURL(postImagePreviewUrl);
      }
    };
  }, [postImagePreviewUrl]);

  const handleFileSelect = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (!ALLOWED_FILE_TYPES.includes(file.type)) {
        toast.error(`Invalid file type. Please select a JPG, PNG, GIF, or WEBP image.`);
        if (fileInputRef.current) fileInputRef.current.value = ''; // Reset file input
        return;
      }
      if (file.size > MAX_FILE_SIZE_BYTES) {
        toast.error(`File is too large. Maximum size is ${MAX_FILE_SIZE_MB}MB.`);
        if (fileInputRef.current) fileInputRef.current.value = ''; // Reset file input
        return;
      }
      setSelectedPostImageFile(file);
      if (postImagePreviewUrl) {
        URL.revokeObjectURL(postImagePreviewUrl); // Revoke old URL
      }
      setPostImagePreviewUrl(URL.createObjectURL(file));
    }
  };

  const handleRemoveImage = () => {
    setSelectedPostImageFile(null);
    if (postImagePreviewUrl) {
      URL.revokeObjectURL(postImagePreviewUrl);
    }
    setPostImagePreviewUrl(null);
    setValue('image_url', null);
    if (fileInputRef.current) {
      fileInputRef.current.value = ''; // Reset the file input
    }
  };

  const onSubmitForm = handleSubmit(async (data) => {
    if (!user) {
      toast.error('You must be logged in to create a post');
      return;
    }

    let imageUrl: string | null = null;

    if (selectedPostImageFile) {
      setIsUploadingImage(true);
      const uploadToastId = toast.loading('Uploading image...');
      try {
        imageUrl = await uploadFileToMinio(selectedPostImageFile);
        toast.success('Image uploaded successfully!', { id: uploadToastId });
      } catch (err) {
        console.error('Failed to upload image:', err);
        toast.error('Failed to upload image. Please try again.', { id: uploadToastId });
        setIsUploadingImage(false);
        return;
      } finally {
        setIsUploadingImage(false);
      }
    }

    const postData = {
      user_id: user.id,
      title: data.title,
      content: data.content,
      image_url: imageUrl,
      privacy: data.privacy
    };

    const createPostToastId = toast.loading('Creating post...');
    try {
      const result = await createPostRequest('/api/posts', postData);
      if (result) {
        toast.success('Post created successfully!', { id: createPostToastId });
        reset();
        handleRemoveImage(); // Clear image selection and preview
        if (onSubmit) {
          onSubmit(result);
        }
      } else {
        // This case might not be hit if useRequest throws on non-2xx
        toast.error('Failed to create post. Response was not OK.', { id: createPostToastId });
      }
    } catch (err) {
      console.error('Failed to create post:', err);
      toast.error('Failed to create post. Please try again.', { id: createPostToastId });
    }
  });

  const isLoading = isCreatingPost || isUploadingImage;

  return (
    <form onSubmit={onSubmitForm} className="mb-6 bg-gray-800 rounded-lg shadow p-4">
      <input
        type="text"
        {...register('title', { required: 'Title is required' })}
        placeholder="Title"
        className="w-full p-4 border border-gray-700 bg-gray-900 text-gray-100 rounded-lg mb-4 focus:ring-2 focus:ring-purple-500 focus:border-transparent"
      />
      <textarea
        {...register('content', { required: 'Content is required' })}
        placeholder="What's on your mind?"
        className="w-full p-4 border border-gray-700 bg-gray-900 text-gray-100 rounded-lg mb-4 resize-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
        rows={3}
      />

      {postImagePreviewUrl && (
        <div className="mb-4 relative">
          <Image
            src={postImagePreviewUrl}
            alt="Selected image preview"
            width={200}
            height={200}
            className="rounded-lg object-cover max-h-48 w-auto"
          />
          <button
            type="button"
            onClick={handleRemoveImage}
            className="absolute top-1 right-1 bg-red-600 hover:bg-red-700 text-white rounded-full p-1.5"
            aria-label="Remove image"
          >
            <FiX size={16} />
          </button>
        </div>
      )}

      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileSelect}
        accept={ALLOWED_FILE_TYPES.join(',')}
        className="hidden"
        id="post-image-upload"
      />

      <div className="flex justify-between items-center">
        <div className="flex items-center gap-4">
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            className="text-purple-400 hover:text-purple-300 flex items-center gap-2 px-3 py-2 rounded-md border border-gray-700 hover:border-purple-500"
            disabled={isLoading}
          >
            <FiImage />
            {selectedPostImageFile ? 'Change Image' : 'Add Image'}
          </button>
          <select
            {...register('privacy')}
            className="bg-gray-900 border border-gray-700 text-gray-100 rounded-lg px-3 py-2 focus:ring-2 focus:ring-purple-500 focus:border-transparent"
            disabled={isLoading}
          >
            <option value="public">Public</option>
            <option value="friends">Friends Only</option>
            <option value="private">Private</option>
          </select>
        </div>
        <button
          type="submit"
          disabled={!isValid || isLoading || !user}
          className="bg-purple-700 text-gray-100 px-6 py-2 rounded-full hover:bg-purple-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isUploadingImage ? 'Uploading...' : isCreatingPost ? 'Posting...' : 'Post'}
        </button>
      </div>
    </form>
  );
}
