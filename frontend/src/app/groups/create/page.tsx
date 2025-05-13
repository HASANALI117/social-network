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
import { useUserStore } from '../../../store/useUserStore';
import Link from 'next/link';

interface CreateGroupFormValues {
  title: string;
  description: string;
}

export default function CreateGroupPage() {
  const router = useRouter();
  const { user } = useUserStore();
  const [submissionError, setSubmissionError] = useState<string | null>(null);

  const { 
    register, 
    handleSubmit, 
    formState: { errors } 
  } = useForm<CreateGroupFormValues>();

  const { post: createGroupRequest, isLoading, error: apiError } = useRequest<Group>();

  useEffect(() => {
    if (apiError) {
      setSubmissionError(apiError.message || 'An unexpected error occurred.');
    }
  }, [apiError]);

  const onSubmit: SubmitHandler<CreateGroupFormValues> = async (formData) => {
    if (!user) {
      setSubmissionError("You must be logged in to create a group.");
      // Optionally redirect to login: router.push('/login');
      return;
    }
    setSubmissionError(null);

    try {
      const newGroup = await createGroupRequest('/api/groups', formData);
      if (newGroup && newGroup.id) {
        // Optionally: show success toast
        router.push(`/groups/${newGroup.id}`);
      } else if (!apiError) { // If newGroup is null/undefined but no apiError, set a generic error
        setSubmissionError('Failed to create group. Please try again.');
      }
    } catch (err) {
      // This catch block might be redundant if useRequest handles errors and sets apiError
      console.error("Create group error:", err);
      setSubmissionError('An unexpected error occurred during submission.');
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
      
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 bg-gray-800 p-8 rounded-lg shadow-xl">
        <div>
          <label htmlFor="title" className="block text-sm font-medium text-gray-300 mb-1">
            Group Title
          </label>
          <Input
            id="title"
            type="text"
            {...register('title', { 
              required: 'Title is required.',
              minLength: { value: 3, message: 'Title must be at least 3 characters.' },
              maxLength: { value: 100, message: 'Title cannot exceed 100 characters.' }
            })}
            className="w-full bg-gray-700 border-gray-600 text-white placeholder-gray-400"
            placeholder="Enter group title"
          />
          {errors.title && <Text className="mt-1 text-sm text-red-400">{errors.title.message}</Text>}
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

        {submissionError && (
          <Alert open={!!submissionError} onClose={() => setSubmissionError(null)}>
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{submissionError}</AlertDescription>
          </Alert>
        )}

        <div>
          <Button type="submit" disabled={isLoading} className="w-full">
            {isLoading ? 'Creating Group...' : 'Create Group'}
          </Button>
        </div>
      </form>
    </div>
  );
}