'use client';

import React, { useState, useCallback } from 'react';
import { useForm, SubmitHandler } from 'react-hook-form';
import { useRequest } from '@/hooks/useRequest';
import { Comment } from '@/types/Comment';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Alert } from '@/components/ui/alert';
import { useUserStore } from '@/store/useUserStore';
import { uploadFileToMinio } from '@/lib/minioUploader';

interface CommentFormValues {
  content: string;
  image?: FileList;
}

interface CreateCommentFormProps {
  postId: string;
  onCommentCreated: (newComment: Comment) => void;
}

const CreateCommentForm: React.FC<CreateCommentFormProps> = ({ postId, onCommentCreated }) => {
  const { user } = useUserStore();
  const { register, handleSubmit, reset, formState: { errors }, watch, setValue } = useForm<CommentFormValues>();
  const { post: submitComment, isLoading: isPostingComment, error: commentPostError } = useRequest<Comment>();

  const [submissionError, setSubmissionError] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    } else {
      setSelectedFile(null);
      setImagePreview(null);
      setValue('image', undefined);
    }
  };

  const onSubmit: SubmitHandler<CommentFormValues> = async (data) => {
    if (!user) {
      setSubmissionError("Please log in to comment.");
      return;
    }

    setSubmissionError(null);
    let imageUrl: string | undefined = undefined;

    try {
      if (selectedFile) { // Use selectedFile instead of croppedImageFile
        const uploadedUrl = await uploadFileToMinio(selectedFile); // Removed second argument
        if (uploadedUrl) { // Assuming uploadFileToMinio returns string URL directly or null/undefined on failure
          imageUrl = uploadedUrl;
        } else {
          throw new Error('Failed to get image URL from upload.');
        }
      }

      const payload = {
        content: data.content,
        ...(imageUrl && { image_url: imageUrl }),
      };

      await submitComment(`/api/posts/${postId}/comments`, payload, (newComment: Comment) => {
        onCommentCreated(newComment);
        reset(); // Clears react-hook-form fields
        setSelectedFile(null);
        setImagePreview(null);
        // Clear the file input in react-hook-form if it's registered
        const fileInput = document.getElementById('comment-image-upload') as HTMLInputElement;
        if (fileInput) {
            fileInput.value = ''; // Clears the selected file in the input element
        }
        setValue('image', undefined); // Ensure RHF state is also cleared
      });

      if (commentPostError) { // Check error state from useRequest after await
        setSubmissionError(commentPostError.message || "Failed to post comment.");
      }

    } catch (error: any) {
      console.error("Comment submission error:", error);
      setSubmissionError((error as Error).message || "An unexpected error occurred while posting your comment.");
    }
  };

  if (!user) {
    return <p className="text-sm text-gray-600 dark:text-gray-400">Please log in to comment.</p>;
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <Textarea
          placeholder="Write a comment..."
          {...register('content', { required: 'Comment content cannot be empty.' })}
          className={errors.content ? 'border-red-500' : ''}
        />
        {errors.content && <p className="text-xs text-red-500 mt-1">{errors.content.message}</p>}
      </div>

      <div>
        <label htmlFor="comment-image-upload" className="text-sm font-medium text-gray-700 dark:text-gray-300">
          Add an image (optional)
        </label>
        <Input
          id="comment-image-upload"
          type="file"
          accept="image/*"
          {...register("image")}
          onChange={handleFileChange}
          className="mt-2 block w-full text-sm text-gray-400 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-purple-600 file:text-white hover:file:bg-purple-700"
        />
      </div>

      {imagePreview && (
        <div className="mt-2">
          <img src={imagePreview} alt="Selected image" className="max-h-40 rounded-md border border-gray-600" />
          <Button type="button" onClick={() => { // Removed variant="ghost" and size="sm"
            setImagePreview(null);
            setSelectedFile(null);
            setValue("image", undefined); // Clear the file input in react-hook-form
            const fileInput = document.getElementById('comment-image-upload') as HTMLInputElement;
            if (fileInput) {
                fileInput.value = ''; // Also clear the native file input
            }
          }} className="mt-1 text-xs text-red-400">
            Remove Image
          </Button>
        </div>
      )}

      {submissionError && (
        <div className="p-3 my-2 text-sm text-red-700 bg-red-100 border border-red-300 rounded-md dark:bg-red-200 dark:text-red-800 dark:border-red-600" role="alert">
          <p>{submissionError}</p>
        </div>
      )}
      
      {commentPostError && !submissionError && ( // Display useRequest error if not already handled by submissionError
        <div className="p-3 my-2 text-sm text-red-700 bg-red-100 border border-red-300 rounded-md dark:bg-red-200 dark:text-red-800 dark:border-red-600" role="alert">
          <p>{commentPostError.message || "Failed to post comment."}</p>
        </div>
      )}

      <Button type="submit" disabled={isPostingComment}>
        {isPostingComment ? 'Posting...' : 'Post Comment'}
      </Button>
    </form>
  );
};

export default CreateCommentForm;