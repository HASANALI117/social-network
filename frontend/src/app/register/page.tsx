'use client';

import { useForm } from 'react-hook-form';
import { useState, useEffect, ChangeEvent } from 'react';
import ImageCropperModal from '@/components/common/ImageCropperModal';
import 'react-image-crop/dist/ReactCrop.css';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import Link from 'next/link';
import { Fieldset, Field, ErrorMessage } from '@/components/ui/fieldset';
import { useRequest } from '@/hooks/useRequest';
import { UserSignupData } from '@/types/User';
import { useRouter } from 'next/navigation';
import toast from 'react-hot-toast';
import { uploadFileToMinio } from '../../lib/minioUploader';

const formSchema = z.object({
  first_name: z.string().min(2, 'First name must be at least 2 characters'),
  last_name: z.string().min(2, 'Last name must be at least 2 characters'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  birth_date: z.string().refine((date) => {
    const inputDate = new Date(date);
    const today = new Date();
    return inputDate <= today;
  }, 'Date of birth cannot be in the future'),
  // avatar_url: z
  //   .custom<FileList>()
  //   .refine((files) => {
  //     if (!files || files.length === 0) return true;
  //     const file = files[0];
  //     return file.size <= 5 * 1024 * 1024; // 5MB
  //   }, 'File must be less than 5MB')
  //   .refine((files) => {
  //     if (!files || files.length === 0) return true;
  //     const file = files[0];
  //     return ['image/jpeg', 'image/png', 'image/webp'].includes(file.type);
  //   }, 'Only JPEG, PNG, and WEBP formats are allowed')
  //   .optional(),
  username: z
    .string()
    .min(3, 'Username must be at least 3 characters')
    .optional(),
  about_me: z
    .string()
    .max(500, 'Bio must be less than 500 characters')
    .optional(),
  terms: z.literal<boolean>(true, {
    errorMap: () => ({ message: 'You must accept the terms and conditions' }),
  }),
});

type FormValues = z.infer<typeof formSchema>;

export default function RegisterPage() {
  const [selectedAvatarFile, setSelectedAvatarFile] = useState<File | null>(null);
  const [avatarPreviewUrl, setAvatarPreviewUrl] = useState<string | null>(null);
  const [isCropperOpen, setIsCropperOpen] = useState(false);
  const [cropperImageSrc, setCropperImageSrc] = useState<string | null>(null);
  const [originalFileName, setOriginalFileName] = useState<string | null>(null);
  const [originalFileType, setOriginalFileType] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
  } = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      first_name: '',
      last_name: '',
      email: '',
      password: '',
      terms: false,
    },
  });

  const { isLoading, data, error, post } = useRequest<UserSignupData>();

  const router = useRouter();

  useEffect(() => {
    // Clean up the object URL when the component unmounts or the file changes
    return () => {
      if (avatarPreviewUrl) {
        URL.revokeObjectURL(avatarPreviewUrl);
      }
    };
  }, [avatarPreviewUrl]);

  const handleAvatarChange = (event: ChangeEvent<HTMLInputElement>) => {
    if (avatarPreviewUrl) {
      URL.revokeObjectURL(avatarPreviewUrl);
      setAvatarPreviewUrl(null);
    }
    const file = event.target.files?.[0];
    if (file) {
      if (file.size > 5 * 1024 * 1024) { // 5MB
        toast.error('File must be less than 5MB');
        setSelectedAvatarFile(null);
        if (event.target) event.target.value = ''; // Reset file input
        return;
      }
      if (!['image/jpeg', 'image/png', 'image/webp', 'image/gif'].includes(file.type)) {
        toast.error('Only JPEG, PNG, WEBP, and GIF formats are allowed');
        setSelectedAvatarFile(null);
        if (event.target) event.target.value = ''; // Reset file input
        return;
      }
      
      setOriginalFileName(file.name);
      setOriginalFileType(file.type);

      const reader = new FileReader();
      reader.onload = () => {
        setCropperImageSrc(reader.result as string);
        setIsCropperOpen(true);
      };
      reader.readAsDataURL(file);
      // Clear the file input so the same file can be selected again if needed after closing cropper without saving
      if (event.target) event.target.value = '';
    } else {
      setSelectedAvatarFile(null);
      setCropperImageSrc(null);
      setOriginalFileName(null);
      setOriginalFileType(null);
    }
  };

  const handleCropComplete = (croppedImageBlob: Blob | null) => {
    if (croppedImageBlob) {
      const fileName = originalFileName || 'avatar.png';
      const fileType = originalFileType || croppedImageBlob.type || 'image/png';
      
      // Ensure the filename has an appropriate extension
      let finalFileName = fileName;
      const currentExtension = fileName.split('.').pop()?.toLowerCase();
      const blobExtension = fileType.split('/')[1];

      if (blobExtension && currentExtension !== blobExtension) {
        finalFileName = `${fileName.substring(0, fileName.lastIndexOf('.')) || fileName}.${blobExtension}`;
      }


      const croppedFile = new File([croppedImageBlob], finalFileName, { type: fileType });
      setSelectedAvatarFile(croppedFile);

      if (avatarPreviewUrl) {
        URL.revokeObjectURL(avatarPreviewUrl);
      }
      setAvatarPreviewUrl(URL.createObjectURL(croppedFile));
    }
    setIsCropperOpen(false);
    setCropperImageSrc(null);
    // Reset original file info if not used or if cropper is simply closed
    if (!croppedImageBlob) {
        setOriginalFileName(null);
        setOriginalFileType(null);
    }
  };

  return (
    <>
      <div className="max-w-md mx-auto my-12 p-6 bg-white rounded-lg shadow-md dark:bg-zinc-900">
        <h1 className="text-2xl font-bold mb-6 text-center">Create Account</h1>
      <form
        onSubmit={handleSubmit(async (formData: FormValues) => {
          let avatarUrl = '';
          if (selectedAvatarFile) {
            try {
              // uploadFileToMinio returns the URL string directly or throws an error
              avatarUrl = await uploadFileToMinio(selectedAvatarFile);
              if (!avatarUrl) { // Should not happen if uploadFileToMinio resolves
                toast.error('Avatar upload failed to return a URL. Please try again.');
                return;
              }
            } catch (uploadError) {
              console.error('Avatar upload error:', uploadError);
              toast.error('Avatar upload failed. Please try again.');
              return;
            }
          }

          const registrationData = {
            ...formData,
            avatar_url: avatarUrl || null, // Send null or empty string if no avatar
          };

          post<UserSignupData>(
            '/api/users',
            registrationData,
            (userData: UserSignupData) => {
              toast.success('Account created successfully! Please log in.');
              console.log('User created:', userData);
              router.push('/login');
            }
            // Error handling is managed by the useRequest hook's 'error' state.
            // We can check 'error' from useRequest after this call if needed.
          );

          // Check for errors from useRequest hook after the post call
          if (error) {
            const message = error.message || 'Registration failed. Please try again.';
            toast.error(message);
            console.error('Registration error:', error);
          }

          console.log('Submitting registration data:', registrationData);
        })}
      >
        <Fieldset className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <Field>
              <label
                className="block text-sm font-medium mb-1"
                htmlFor="first-name"
              >
                First Name *
              </label>
              <Input
                id="first-name"
                placeholder="John"
                {...register('first_name')}
              />
              {errors.first_name?.message && (
                <ErrorMessage>{errors.first_name.message}</ErrorMessage>
              )}
            </Field>
            <Field>
              <label
                className="block text-sm font-medium mb-1"
                htmlFor="last-name"
              >
                Last Name *
              </label>
              <Input
                id="last-name"
                placeholder="Doe"
                {...register('last_name')}
              />
              {errors.last_name?.message && (
                <ErrorMessage>{errors.last_name.message}</ErrorMessage>
              )}
            </Field>
          </div>

          <Field>
            <label className="block text-sm font-medium mb-1" htmlFor="email">
              Email *
            </label>
            <Input
              id="email"
              type="email"
              placeholder="john@example.com"
              {...register('email')}
            />
            {errors.email?.message && (
              <ErrorMessage>{errors.email.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <label
              className="block text-sm font-medium mb-1"
              htmlFor="password"
            >
              Password *
            </label>
            <Input
              id="password"
              type="password"
              placeholder="••••••••"
              {...register('password')}
            />
            {errors.password?.message && (
              <ErrorMessage>{errors.password.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <label
              className="block text-sm font-medium mb-1"
              htmlFor="birth-date"
            >
              Date of Birth *
            </label>
            <Input
              id="birth-date"
              type="date"
              max={new Date().toISOString().split('T')[0]}
              {...register('birth_date')}
            />
            {errors.birth_date?.message && (
              <ErrorMessage>{errors.birth_date.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <label className="block text-sm font-medium mb-1" htmlFor="avatar">
              Profile Picture (Optional, max 5MB, JPG/PNG/WEBP/GIF)
            </label>
            {avatarPreviewUrl && (
              <div className="mt-2 mb-2">
                <img
                  src={avatarPreviewUrl}
                  alt="Avatar Preview"
                  className="w-24 h-24 rounded-full object-cover mx-auto"
                />
              </div>
            )}
            <Input
              id="avatar"
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              onChange={handleAvatarChange}
              className="block w-full text-sm text-gray-900 border border-gray-300 rounded-lg cursor-pointer bg-gray-50 dark:text-gray-400 focus:outline-none dark:bg-zinc-700 dark:border-zinc-600 dark:placeholder-zinc-400"
            />
            {/* No direct form error for avatar via react-hook-form, handled by toast */}
          </Field>

          <Field>
            <label
              className="block text-sm font-medium mb-1"
              htmlFor="username"
            >
              Username (Optional)
            </label>
            <Input
              id="username"
              placeholder="CoolDude123"
              {...register('username')}
            />
            {errors.username?.message && (
              <ErrorMessage>{errors.username.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <label
              className="block text-sm font-medium mb-1"
              htmlFor="about-me"
            >
              About Me (Optional)
            </label>
            <textarea
              id="about-me"
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-zinc-800 dark:border-zinc-700"
              rows={3}
              placeholder="Tell us something about yourself..."
              {...register('about_me')}
            />
            {errors.about_me?.message && (
              <ErrorMessage>{errors.about_me.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="terms"
                  {...register('terms')}
                  className="hidden"
                />
                <div
                  onClick={() => setValue('terms', !watch('terms'))}
                  className={`w-5 h-5 rounded border cursor-pointer ${
                    watch('terms')
                      ? 'bg-blue-600 border-blue-600'
                      : 'border-gray-300 dark:border-gray-600'
                  }`}
                >
                  {watch('terms') && (
                    <svg
                      className="w-4 h-4 text-white mx-auto"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                        clipRule="evenodd"
                      />
                    </svg>
                  )}
                </div>
                <label htmlFor="terms" className="text-sm">
                  I agree to the{' '}
                  <Link href="/terms" className="text-blue-600 hover:underline">
                    Terms of Service
                  </Link>
                </label>
              </div>
            </div>
            {errors.terms?.message && (
              <ErrorMessage>{errors.terms.message}</ErrorMessage>
            )}
          </Field>
        </Fieldset>

        <div className="mt-6">
          <Button
            type="submit"
            className="w-full bg-blue-600 hover:bg-blue-700"
          >
            Create Account
          </Button>
        </div>
      </form>

      <p className="mt-4 text-center text-sm">
        Already have an account?{' '}
        <Link href="/login" className="text-blue-600 hover:underline">
          Log in
        </Link>
      </p>
      </div>
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
    </>
  );
}
