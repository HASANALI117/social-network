'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm, SubmitHandler } from 'react-hook-form';
import { useRequest } from '../../../hooks/useRequest';
import { Group } from '../../../types/Group';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Textarea } from '../../../components/ui/textarea';
import { Heading } from '../../../components/ui/heading';
import { Text } from '../../../components/ui/text';
import { Alert, AlertDescription, AlertTitle } from '../../../components/ui/alert';
import { Avatar } from '../../../components/ui/avatar';
import { useUserStore } from '../../../store/useUserStore';
import { UserBasicInfo, User } from '../../../types/User'; // Added User for currentUser mapping
import Link from 'next/link';
import Image from 'next/image';
import { FiUploadCloud } from 'react-icons/fi';
import ImageCropperModal from '../../../components/common/ImageCropperModal';
import { uploadFileToMinio } from '../../../lib/minioUploader';
import 'react-image-crop/dist/ReactCrop.css';
import toast from 'react-hot-toast';
import GroupInviteManager from '../../../components/groups/GroupInviteManager'; // New Import

interface CreateGroupFormValues {
  name: string;
  description: string;
  avatar_url?: string;
}

export default function CreateGroupPage() {
  const router = useRouter();
  const { user } = useUserStore();

  // State for the two-step process
  const [creationStep, setCreationStep] = useState<'details' | 'invite'>('details');
  const [createdGroupId, setCreatedGroupId] = useState<string | null>(null);

  // Removed old invite UI state: searchTerm, searchResults, isActualSearchLoading, searchError, invitingUserId, invitedUserIds

  // State for avatar upload and cropping
  const [newAvatarFile, setNewAvatarFile] = useState<File | null>(null);
  const [newAvatarPreviewUrl, setNewAvatarPreviewUrl] = useState<string | null>(null);
  const [isCropperOpen, setIsCropperOpen] = useState(false);
  const [cropperImageSrc, setCropperImageSrc] = useState<string | null>(null);
  const [originalAvatarFileName, setOriginalAvatarFileName] = useState<string | null>(null);
  const [originalAvatarFileType, setOriginalAvatarFileType] = useState<string | null>(null);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);


  const [submissionError, setSubmissionError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors }
  } = useForm<CreateGroupFormValues>();

  const { post: createGroupRequest, isLoading, error: apiError } = useRequest<Group>();
  // Removed searchUsersRequestHook and sendInviteRequestHook as GroupInviteManager handles its own requests

  useEffect(() => {
    if (apiError) {
      setSubmissionError(apiError.message || 'An unexpected error occurred.');
    }
  }, [apiError]);

  // Removed useEffects for searchApiHookError and inviteApiHookError
  // Removed handleSearchUsers, debouncedSearchUsers, and handleSendInvite functions

  useEffect(() => {
    // Cleanup object URL for avatar preview
    return () => {
      if (newAvatarPreviewUrl) {
        URL.revokeObjectURL(newAvatarPreviewUrl);
      }
    };
  }, [newAvatarPreviewUrl]);

  const handleAvatarFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (!file.type.startsWith('image/')) {
        setSubmissionError('Please select an image file for the avatar.');
        return;
      }
      if (file.size > 5 * 1024 * 1024) { // 5MB limit
        setSubmissionError('Avatar image size should be less than 5MB.');
        return;
      }
      setSubmissionError(null);
      const reader = new FileReader();
      reader.onloadend = () => {
        setCropperImageSrc(reader.result as string);
        setOriginalAvatarFileName(file.name);
        setOriginalAvatarFileType(file.type);
        setIsCropperOpen(true);
      };
      reader.readAsDataURL(file);
      event.target.value = ''; // Allow selecting the same file again
    }
  };

  const handleAvatarCropComplete = (croppedImageBlob: Blob) => {
    if (newAvatarPreviewUrl) {
      URL.revokeObjectURL(newAvatarPreviewUrl);
    }
    const fileName = originalAvatarFileName || `group-avatar-${Date.now()}.png`;
    const fileType = originalAvatarFileType || croppedImageBlob.type || 'image/png';
    const croppedFile = new File([croppedImageBlob], fileName, { type: fileType });

    setNewAvatarFile(croppedFile);
    setNewAvatarPreviewUrl(URL.createObjectURL(croppedFile));
    setIsCropperOpen(false);
    setCropperImageSrc(null);
    setOriginalAvatarFileName(null);
    setOriginalAvatarFileType(null);
  };


  const onSubmit: SubmitHandler<CreateGroupFormValues> = async (formData) => {
    if (!user) {
      setSubmissionError("You must be logged in to create a group.");
      return;
    }
    setSubmissionError(null);
    setIsUploadingAvatar(false); // Reset upload status

    let uploadedAvatarUrl: string | undefined = undefined;

    if (newAvatarFile) {
      setIsUploadingAvatar(true);
      try {
        uploadedAvatarUrl = await uploadFileToMinio(newAvatarFile);
      } catch (uploadError) {
        console.error('Failed to upload group avatar:', uploadError);
        setSubmissionError('Failed to upload avatar. Please try again.');
        setIsUploadingAvatar(false);
        return; // Stop submission if avatar upload fails
      }
      setIsUploadingAvatar(false);
    }

    try {
      const payload: { name: string; description: string; avatar_url?: string } = {
        name: formData.name,
        description: formData.description,
      };
      // Use the uploaded Minio URL if available, otherwise use the URL from the form (if any, though this field will be removed)
      // Or, if we strictly use file upload, then only uploadedAvatarUrl matters.
      if (uploadedAvatarUrl) {
        payload.avatar_url = uploadedAvatarUrl;
      } else if (formData.avatar_url && !newAvatarFile) { // Keep existing URL if no new file and URL was somehow pre-filled
        payload.avatar_url = formData.avatar_url;
      }


      const newGroup = await createGroupRequest('/api/groups', payload);
      if (newGroup && newGroup.id) {
        setCreatedGroupId(newGroup.id);
        setCreationStep('invite');
      } else if (!apiError) {
        setSubmissionError('Failed to create group. Please try again.');
      }
    } catch (err) {
      console.error("Create group error:", err);
      if (!apiError) {
        setSubmissionError('An unexpected error occurred during submission.');
      }
    }
  };

  if (!user) {
    return (
      <div className="container mx-auto p-4 text-center text-white">
        <Heading level={2} className="mb-4">Access Denied</Heading>
        <Text className="mb-4">Please <Link href="/login" className="text-purple-400 hover:underline">log in</Link> to create a group.</Text>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4 text-white max-w-2xl">
      <Heading level={1} className="mb-8 text-center">Create a New Group</Heading>

      {creationStep === 'details' && (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 bg-gray-800 p-8 rounded-lg shadow-xl">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-300 mb-1">
              Group Name
            </label>
            <Input
              id="name"
              type="text"
              {...register('name', {
                required: 'Name is required.',
                minLength: { value: 3, message: 'Name must be at least 3 characters.' },
                maxLength: { value: 100, message: 'Name cannot exceed 100 characters.' }
              })}
              className="w-full bg-gray-700 border-gray-600 text-white placeholder-gray-400"
              placeholder="Enter group name"
            />
            {errors.name && <Text className="mt-1 text-sm text-red-400">{errors.name.message}</Text>}
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-300 mb-1">
              Group Description
            </label>
            <Textarea
              id="description"
              {...register('description', {
                required: 'Description is required.',
                minLength: { value: 10, message: 'Description must be at least 10 characters.' },
                maxLength: { value: 500, message: 'Description cannot exceed 500 characters.' }
              })}
              rows={4}
              className="w-full bg-gray-700 border-gray-600 text-white placeholder-gray-400"
              placeholder="Tell us about your group"
            />
            {errors.description && <Text className="mt-1 text-sm text-red-400">{errors.description.message}</Text>}
          </div>

          {/* Avatar Upload Section - Replaces URL input */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Group Avatar (Optional)
            </label>
            <div className="flex items-center gap-4">
              <div className="relative w-24 h-24 rounded-md overflow-hidden bg-gray-700 flex items-center justify-center">
                {newAvatarPreviewUrl ? (
                  <Image src={newAvatarPreviewUrl} alt="New Avatar Preview" fill style={{ objectFit: 'cover' }} />
                ) : (
                  <span className="text-gray-500 text-3xl">?</span> // Placeholder
                )}
              </div>
              <div>
                <input
                  type="file"
                  id="groupAvatarUpload"
                  accept="image/*"
                  onChange={handleAvatarFileChange}
                  className="hidden"
                  disabled={isLoading || isUploadingAvatar}
                />
                <label
                  htmlFor="groupAvatarUpload"
                  className={`cursor-pointer flex items-center gap-2 bg-gray-700 text-gray-100 px-4 py-2 rounded-lg hover:bg-gray-600 transition-colors ${ (isLoading || isUploadingAvatar) ? 'opacity-50 cursor-not-allowed' : ''}`}
                >
                  <FiUploadCloud />
                  {newAvatarFile ? 'Change Image' : 'Upload Image'}
                </label>
                {newAvatarFile && (
                  <p className="text-xs text-gray-400 mt-1">
                    Selected: {newAvatarFile.name}
                  </p>
                )}
                {isUploadingAvatar && <p className="text-sm text-purple-400 mt-1">Uploading avatar...</p>}
                 {/* Hidden input to clear react-hook-form's avatar_url if a file is chosen, or to pass existing if needed */}
                 <input type="hidden" {...register('avatar_url')} />
              </div>
            </div>
            {errors.avatar_url && <Text className="mt-1 text-sm text-red-400">{errors.avatar_url.message}</Text>}
          </div>
          {/* End Avatar Upload Section */}

          {submissionError && (
            <Alert open={!!submissionError} onClose={() => setSubmissionError(null)}>
              <AlertTitle>Error Creating Group</AlertTitle>
              <AlertDescription>{submissionError}</AlertDescription>
            </Alert>
          )}

          <div>
            <Button type="submit" disabled={isLoading || isUploadingAvatar} className="w-full">
              {isLoading || isUploadingAvatar ? 'Processing...' : 'Next: Invite Users'}
            </Button>
          </div>
        </form>
      )}

      {creationStep === 'invite' && (
        <div className="space-y-6 bg-gray-800 p-8 rounded-lg shadow-xl">
          <Heading level={2} className="mb-6 text-center">Invite Users to Your New Group</Heading>

          {createdGroupId && user && (
            <GroupInviteManager
              groupId={createdGroupId}
              currentUser={user ? { ...user, user_id: user.id } : null} // Map User to UserBasicInfo
              // onInviteSent and onInviteError can be handled here if needed, e.g., for analytics
            />
          )}

          <div className="flex flex-col sm:flex-row justify-between items-center space-y-3 sm:space-y-0 sm:space-x-4 pt-6">
            <Button
              onClick={() => router.push(`/groups/${createdGroupId}`)}
              disabled={!createdGroupId} // isLoading from createGroupRequest can also be used here if needed
              className="w-full sm:w-auto"
              color="indigo" // Example color
            >
              Finish & Go to Group
            </Button>
            <Button
              plain
              onClick={() => router.push(`/groups/${createdGroupId}`)}
              disabled={!createdGroupId}
              className="w-full sm:w-auto"
            >
              Skip Invites & Finish
            </Button>
          </div>
        </div>
      )}

      {cropperImageSrc && (
        <ImageCropperModal
          isOpen={isCropperOpen}
          onClose={() => {
            setIsCropperOpen(false);
            setCropperImageSrc(null);
            setOriginalAvatarFileName(null);
            setOriginalAvatarFileType(null);
          }}
          imageSrc={cropperImageSrc}
          onCropComplete={handleAvatarCropComplete}
          aspect={1} // Square aspect ratio for avatars
          circularCrop={false} // Or true, depending on desired avatar shape for groups
        />
      )}
    </div>
  );
}

// Add a cancel method to the debounced function type
interface DebouncedFunction<T extends (...args: any[]) => void> {
  (...args: Parameters<T>): void;
  cancel: () => void;
}

// Debounce function can be removed if not used elsewhere in this file after GroupInviteManager integration
// If it's a utility, it might be better placed in a utils file.
// For now, assuming GroupInviteManager handles its own debouncing.
// interface DebouncedFunction<T extends (...args: any[]) => void> {
//   (...args: Parameters<T>): void;
//   cancel: () => void;
// }

// const debounce = <T extends (...args: any[]) => void>(func: T, delay: number): DebouncedFunction<T> => {
//   let timeoutId: NodeJS.Timeout;
//
//   const debounced = ((...args: Parameters<T>) => {
//     clearTimeout(timeoutId);
//     timeoutId = setTimeout(() => {
//       func.apply(null, args);
//     }, delay);
//   }) as DebouncedFunction<T>;
//
//   debounced.cancel = () => {
//     clearTimeout(timeoutId);
//   };
//
//   return debounced;
// };