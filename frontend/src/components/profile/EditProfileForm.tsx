'use client';

import { User as UserType, UpdateUserProfileData } from '@/types/User'; // Changed UserType to User
import { FiSave, FiX, FiUploadCloud } from 'react-icons/fi';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useState, useEffect, ChangeEvent } from 'react';
import Image from 'next/image';
import { uploadFileToMinio } from '@/lib/minioUploader'; // Adjusted path
import ImageCropperModal from '@/components/common/ImageCropperModal'; // Added import
import 'react-image-crop/dist/ReactCrop.css'; // Ensure CSS is imported
// import { toast } from 'react-hot-toast'; // Assuming you have a toast library

const profileSchema = z.object({
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  username: z.string().optional(),
  about_me: z.string().optional(),
  // avatar_url is handled separately
});

type ProfileFormData = Omit<z.infer<typeof profileSchema>, 'avatar_url'>;

interface EditProfileFormProps {
  user: UserType;
  onSubmit: (userData: UpdateUserProfileData) => void; // Changed to UpdateUserProfileData
  onCancel: () => void;
}

export default function EditProfileForm({ user, onSubmit, onCancel }: EditProfileFormProps) {
  const [newAvatarFile, setNewAvatarFile] = useState<File | null>(null);
  const [newAvatarPreviewUrl, setNewAvatarPreviewUrl] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [isCropperOpen, setIsCropperOpen] = useState(false);
  const [cropperImageSrc, setCropperImageSrc] = useState<string | null>(null);
  const [originalFileName, setOriginalFileName] = useState<string | null>(null);
  const [originalFileType, setOriginalFileType] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting: isFormSubmitting },
    setValue, // To potentially set avatar_url if needed, though we handle it separately
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      first_name: user.first_name,
      last_name: user.last_name,
      username: user.username || '',
      about_me: user.about_me || '',
    },
  });

  useEffect(() => {
    // Cleanup object URL
    return () => {
      if (newAvatarPreviewUrl) {
        URL.revokeObjectURL(newAvatarPreviewUrl);
      }
    };
  }, [newAvatarPreviewUrl]);

  const handleAvatarChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      // Basic validation (can be expanded)
      if (!file.type.startsWith('image/')) {
        // toast.error('Please select an image file.');
        console.error('Please select an image file.');
        return;
      }
      if (file.size > 5 * 1024 * 1024) { // 5MB limit
        // toast.error('Image size should be less than 5MB.');
        console.error('Image size should be less than 5MB.');
        return;
      }

      const reader = new FileReader();
      reader.onloadend = () => {
        setCropperImageSrc(reader.result as string);
        setOriginalFileName(file.name);
        setOriginalFileType(file.type);
        setIsCropperOpen(true);
      };
      reader.readAsDataURL(file);
      // Clear the input value to allow selecting the same file again if needed
      event.target.value = '';
    }
  };

  const handleCropComplete = (croppedImageBlob: Blob) => {
    if (newAvatarPreviewUrl) {
      URL.revokeObjectURL(newAvatarPreviewUrl);
    }
    const fileName = originalFileName || `avatar-${Date.now()}.png`;
    const fileType = originalFileType || croppedImageBlob.type || 'image/png';
    const croppedFile = new File([croppedImageBlob], fileName, { type: fileType });

    setNewAvatarFile(croppedFile);
    setNewAvatarPreviewUrl(URL.createObjectURL(croppedFile));
    setIsCropperOpen(false);
    setCropperImageSrc(null);
    setOriginalFileName(null);
    setOriginalFileType(null);
  };

  const handleFormSubmit = async (data: ProfileFormData) => {
    setIsUploading(true);
    let newAvatarMinioUrl: string | undefined = undefined;

    if (newAvatarFile) {
      try {
        newAvatarMinioUrl = await uploadFileToMinio(newAvatarFile);
      } catch (error) {
        console.error('Failed to upload avatar:', error);
        // toast.error('Failed to upload avatar. Please try again.');
        setIsUploading(false);
        return; // Stop submission if avatar upload fails
      }
    }

    const payload: UpdateUserProfileData = { ...data };
    if (newAvatarMinioUrl) {
      payload.avatar_url = newAvatarMinioUrl;
    }
    // If no new avatar, avatar_url will not be in payload, assuming backend handles partial update.
    // If backend requires it, send user.avatar_url:
    // else if (user.avatar_url && !newAvatarFile) {
    //  payload.avatar_url = user.avatar_url;
    // }


    onSubmit(payload);
    setIsUploading(false);
  };
  
  const isSubmitting = isFormSubmitting || isUploading;

  return (
    <div className="bg-gray-800 rounded-lg shadow-lg p-6 mb-6">
      <form onSubmit={handleSubmit(handleFormSubmit)}>
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-semibold text-gray-100">Edit Profile</h2>
          <div className="flex gap-2">
            <button
              type="submit"
              disabled={isSubmitting || isUploading}
              className="flex items-center gap-2 bg-purple-700 text-gray-100 px-4 py-2 rounded-lg hover:bg-purple-600 transition-colors disabled:opacity-50"
            >
              <FiSave />
              {isSubmitting ? 'Saving...' : 'Save Changes'}
            </button>
            <button
              type="button"
              onClick={onCancel}
              disabled={isSubmitting || isUploading}
              className="flex items-center gap-2 bg-gray-700 text-gray-100 px-4 py-2 rounded-lg hover:bg-gray-600 transition-colors disabled:opacity-50"
            >
              <FiX />
              Cancel
            </button>
          </div>
        </div>

        <div className="space-y-6"> {/* Increased spacing */}
          {/* Avatar Upload Section */}
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-2">
              Profile Picture
            </label>
            <div className="flex items-center gap-4">
              <div className="relative w-24 h-24 rounded-full overflow-hidden bg-gray-700 flex items-center justify-center">
                {newAvatarPreviewUrl ? (
                  <Image src={newAvatarPreviewUrl} alt="New Avatar Preview" fill style={{ objectFit: 'cover' }} />
                ) : user.avatar_url ? (
                  <Image src={user.avatar_url} alt={user.username || 'User Avatar'} fill style={{ objectFit: 'cover' }} />
                ) : (
                  <span className="text-gray-500 text-3xl">?</span>
                )}
              </div>
              <div>
                <input
                  type="file"
                  id="avatarUpload"
                  accept="image/*"
                  onChange={handleAvatarChange}
                  className="hidden"
                  disabled={isSubmitting || isUploading}
                />
                <label
                  htmlFor="avatarUpload"
                  className="cursor-pointer flex items-center gap-2 bg-gray-700 text-gray-100 px-4 py-2 rounded-lg hover:bg-gray-600 transition-colors disabled:opacity-50"
                >
                  <FiUploadCloud />
                  {newAvatarFile ? 'Change Image' : 'Upload Image'}
                </label>
                {newAvatarFile && (
                  <p className="text-xs text-gray-400 mt-1">
                    Selected: {newAvatarFile.name}
                  </p>
                )}
                 {isUploading && <p className="text-sm text-purple-400 mt-1">Uploading avatar...</p>}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4"> {/* Responsive grid */}
            <div>
              <label htmlFor="first_name" className="block text-sm font-medium text-gray-400 mb-1">
                First Name
              </label>
              <input
                type="text"
                id="first_name"
                {...register('first_name')}
                className="w-full bg-gray-700 text-gray-100 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-purple-500"
              />
              {errors.first_name && (
                <p className="text-red-500 text-sm mt-1">{errors.first_name.message}</p>
              )}
            </div>
            <div>
              <label htmlFor="last_name" className="block text-sm font-medium text-gray-400 mb-1">
                Last Name
              </label>
              <input
                type="text"
                id="last_name"
                {...register('last_name')}
                className="w-full bg-gray-700 text-gray-100 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-purple-500"
              />
              {errors.last_name && (
                <p className="text-red-500 text-sm mt-1">{errors.last_name.message}</p>
              )}
            </div>
          </div>

          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-400 mb-1">
              Username
            </label>
            <input
              type="text"
              id="username"
              {...register('username')}
              className="w-full bg-gray-700 text-gray-100 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-purple-500"
            />
            {errors.username && (
              <p className="text-red-500 text-sm mt-1">{errors.username.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="about_me" className="block text-sm font-medium text-gray-400 mb-1">
              About Me
            </label>
            <textarea
              id="about_me"
              {...register('about_me')}
              rows={4}
              className="w-full bg-gray-700 text-gray-100 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-purple-500"
            />
            {errors.about_me && (
              <p className="text-red-500 text-sm mt-1">{errors.about_me.message}</p>
            )}
          </div>
        </div>
      </form>
      {cropperImageSrc && (
        <ImageCropperModal
          isOpen={isCropperOpen}
          onClose={() => {
            setIsCropperOpen(false);
            setCropperImageSrc(null);
            setOriginalFileName(null);
            setOriginalFileType(null);
          }}
          imageSrc={cropperImageSrc}
          onCropComplete={handleCropComplete}
          aspect={1}
          circularCrop={true}
        />
      )}
    </div>
  );
}