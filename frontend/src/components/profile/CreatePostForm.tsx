"use client";

import { FiPlus, FiX, FiImage } from 'react-icons/fi';
import { useForm } from 'react-hook-form';
import { useRequest } from '@/hooks/useRequest';
import { useUserStore } from '@/store/useUserStore';
import { Post } from '@/types/Post';
import { User } from '@/types/User'; // Added User import
import { useState, useRef, useEffect, ChangeEvent } from 'react';
import toast from 'react-hot-toast';
import { uploadFileToMinio } from '../../lib/minioUploader';
import Image from 'next/image';

// Define CreatePostFormValues locally
interface CreatePostFormValues {
  title: string;
  content: string;
  image_url: string | null;
  privacy: 'public' | 'semi_private' | 'private'; // Match Post type
  allowed_user_ids?: string[]; // Added for private posts
  group_id?: string; // Added for group posts
}

interface CreatePostFormProps {
  onSubmit?: (post: Post) => void;
  groupId?: string; // Optional groupId
}

const MAX_FILE_SIZE_MB = 5;
const MAX_FILE_SIZE_BYTES = MAX_FILE_SIZE_MB * 1024 * 1024;
const ALLOWED_FILE_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];

export default function CreatePostForm({ onSubmit, groupId }: CreatePostFormProps) {
  const { user } = useUserStore();
  const { post: createPostRequest, isLoading: isCreatingPost } = useRequest<Post>();
  // Dedicated useRequest instance for followers
  const {
    data: followersDataFromHook,
    error: followersErrorFromHook,
    isLoading: isLoadingFollowersFromHook, // Renamed from isLoadingFollowers
    get: getFollowers, // Renamed from getFollowersRequest
  } = useRequest<User[]>();
  const [isUploadingImage, setIsUploadingImage] = useState(false);
  const [selectedPostImageFile, setSelectedPostImageFile] = useState<File | null>(null);
  const [postImagePreviewUrl, setPostImagePreviewUrl] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [followers, setFollowers] = useState<User[]>([]);
  const [selectedFollowerIds, setSelectedFollowerIds] = useState<string[]>([]);

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch, // Added watch
    formState: { isValid }
  } = useForm<CreatePostFormValues>({
    defaultValues: {
      title: '',
      content: '',
      privacy: groupId ? 'public' : 'public', // Default to public if in group context
      image_url: null,
      allowed_user_ids: [],
      group_id: groupId
    }
  });

  const privacyValue = watch('privacy'); // Watch privacy field

  useEffect(() => {
    // Clean up the object URL when the component unmounts or the preview URL changes
    return () => {
      if (postImagePreviewUrl) {
        URL.revokeObjectURL(postImagePreviewUrl);
      }
    };
  }, [postImagePreviewUrl]);

  // Effect to trigger follower fetch
  useEffect(() => {
    if (!groupId && privacyValue === 'private' && user?.id) { // Only fetch followers if not in group context and privacy is private
      getFollowers(`/api/users/${user.id}/followers`); // Use getFollowers
    } else {
      setFollowers([]);
      setSelectedFollowerIds([]);
      setValue('allowed_user_ids', []);
    }
  }, [privacyValue, user?.id, getFollowers, setValue, groupId]); // Dependency: getFollowers, groupId

  // Effect to process fetched followers data
  useEffect(() => {
    if (followersErrorFromHook) {
      console.error('Failed to fetch followers:', followersErrorFromHook);
      toast.error('Failed to fetch followers.');
      setFollowers([]);
      return;
    }

    if (followersDataFromHook) {
      let extractedList: any[] = [];
      if (Array.isArray(followersDataFromHook)) {
        extractedList = followersDataFromHook;
      } else if (typeof followersDataFromHook === 'object' && followersDataFromHook !== null) {
        const potentialKeys = ['followers', 'data', 'users', 'items'];
        for (const key of potentialKeys) {
          if (Array.isArray((followersDataFromHook as any)[key])) {
            extractedList = (followersDataFromHook as any)[key];
            break;
          }
        }
      }
      const validFollowers = extractedList.filter(
        item => typeof item === 'object' && item !== null &&
                'id' in item && 'first_name' in item && 'last_name' in item
      ) as User[];
      setFollowers(validFollowers);
    } else if (!isLoadingFollowersFromHook) { // If not loading and no data/error
      setFollowers([]);
    }
  }, [followersDataFromHook, followersErrorFromHook, isLoadingFollowersFromHook]);

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

  const handleFollowerSelectionChange = (followerId: string) => {
    setSelectedFollowerIds(prevSelected => {
      const newSelected = prevSelected.includes(followerId)
        ? prevSelected.filter(id => id !== followerId)
        : [...prevSelected, followerId];
      setValue('allowed_user_ids', newSelected);
      return newSelected;
    });
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

    const postData: any = { // Use 'any' for flexibility or define a more specific type
      user_id: user.id,
      title: data.title,
      content: data.content,
      image_url: imageUrl,
      privacy: groupId ? 'public' : data.privacy // If in group, force public (within group context)
    };

    if (groupId) {
      postData.group_id = groupId;
      // For group posts, allowed_user_ids might not be relevant or handled differently by backend
      // Assuming group posts are visible to all group members, so not setting allowed_user_ids
      postData.allowed_user_ids = [];
    } else if (data.privacy === 'private') {
      postData.allowed_user_ids = selectedFollowerIds;
    } else {
      postData.allowed_user_ids = [];
    }

    const endpoint = '/api/posts'; // Always use /api/posts
    const createPostToastId = toast.loading('Creating post...');
    try {
      const result = await createPostRequest(endpoint, postData);
      if (result) {
        toast.success('Post created successfully!', { id: createPostToastId });
        reset(); // Resets form to defaultValues
        handleRemoveImage(); // Clear image selection and preview
        setSelectedFollowerIds([]); // Clear selected followers
        setFollowers([]); // Clear fetched followers list
        // setValue('allowed_user_ids', []); // Already handled by reset if in defaultValues
        if (onSubmit) {
          onSubmit(result);
        }
      } else {
        toast.error('Failed to create post. Response was not OK.', { id: createPostToastId });
      }
    } catch (err) {
      console.error('Failed to create post:', err);
      toast.error('Failed to create post. Please try again.', { id: createPostToastId });
    }
  });

  const isLoading = isCreatingPost || isUploadingImage || isLoadingFollowersFromHook; // Use isLoadingFollowersFromHook

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
    
      {!groupId && privacyValue === 'private' && ( // Only show follower selection if not in group context and privacy is private
        <div className="mb-4">
          <h3 className="text-lg font-semibold text-gray-100 mb-2">Select Allowed Followers:</h3>
          {isLoadingFollowersFromHook && <p className="text-gray-400">Loading followers...</p>}
          {!isLoadingFollowersFromHook && followersErrorFromHook && <p className="text-red-400">Error loading followers.</p>}
          {!isLoadingFollowersFromHook && !followersErrorFromHook && followers.length === 0 && (
            <p className="text-gray-400">You have no followers to select, or none were found.</p>
          )}
          {!isLoadingFollowersFromHook && !followersErrorFromHook && followers.length > 0 && (
            <div className="max-h-60 overflow-y-auto border border-gray-700 rounded-lg p-2 bg-gray-900">
              {followers.map(follower => (
                <label key={follower.id} className="flex items-center space-x-3 p-2 hover:bg-gray-700 rounded-md cursor-pointer">
                  <input
                    type="checkbox"
                    className="form-checkbox h-5 w-5 text-purple-600 bg-gray-800 border-gray-600 rounded focus:ring-purple-500"
                    checked={selectedFollowerIds.includes(follower.id)}
                    onChange={() => handleFollowerSelectionChange(follower.id)}
                    disabled={isLoading}
                  />
                  {follower.avatar_url && (
                    <Image src={follower.avatar_url} alt={`${follower.first_name} ${follower.last_name}`} width={32} height={32} className="rounded-full" />
                  )}
                  <span className="text-gray-200">{follower.first_name} {follower.last_name}</span>
                </label>
              ))}
            </div>
          )}
        </div>
      )}
    
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
          {!groupId && ( // Only show privacy dropdown if not in group context
            <select
              {...register('privacy')}
              className="bg-gray-900 border border-gray-700 text-gray-100 rounded-lg px-3 py-2 focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              disabled={isLoading}
            >
              <option value="public">Public</option>
              <option value="semi_private">Followers Only</option>
              <option value="private">Private</option>
            </select>
          )}
        </div>
        <button
          type="submit"
          disabled={!isValid || isLoading || !user || (!groupId && privacyValue === 'private' && selectedFollowerIds.length === 0 && followers.length > 0) }
          className="bg-purple-700 text-gray-100 px-6 py-2 rounded-full hover:bg-purple-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isUploadingImage ? 'Uploading...' : isCreatingPost ? 'Posting...' : 'Post'}
        </button>
      </div>
    </form>
  );
}
