'use client';

import { UserType } from '@/types/User';
import { FiSave, FiX } from 'react-icons/fi';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';

const profileSchema = z.object({
  first_name: z.string().min(1, 'First name is required'),
  last_name: z.string().min(1, 'Last name is required'),
  username: z.string().optional(),
  about_me: z.string().optional(),
});

type ProfileFormData = z.infer<typeof profileSchema>;

interface EditProfileFormProps {
  user: UserType;
  onSubmit: (userData: Partial<UserType>) => void;
  onCancel: () => void;
}

export default function EditProfileForm({ user, onSubmit, onCancel }: EditProfileFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      first_name: user.first_name,
      last_name: user.last_name,
      username: user.username || '',
      about_me: user.about_me || '',
    },
  });

  return (
    <div className="bg-gray-800 rounded-lg shadow-lg p-6 mb-6">
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-semibold text-gray-100">Edit Profile</h2>
          <div className="flex gap-2">
            <button
              type="submit"
              disabled={isSubmitting}
              className="flex items-center gap-2 bg-purple-700 text-gray-100 px-4 py-2 rounded-lg hover:bg-purple-600 transition-colors disabled:opacity-50"
            >
              <FiSave />
              {isSubmitting ? 'Saving...' : 'Save'}
            </button>
            <button
              type="button"
              onClick={onCancel}
              disabled={isSubmitting}
              className="flex items-center gap-2 bg-gray-700 text-gray-100 px-4 py-2 rounded-lg hover:bg-gray-600 transition-colors disabled:opacity-50"
            >
              <FiX />
              Cancel
            </button>
          </div>
        </div>

        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
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
    </div>
  );
}