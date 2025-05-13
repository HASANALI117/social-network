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
import ImageCropperModal from '@/components/common/ImageCropperModal';

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
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [imageToCrop, setImageToCrop] = useState<string | null>(null);
  const [croppedImageFile, setCroppedImageFile] = useState<File | null>(null);
  const [showCropperModal, setShowCropperModal] = useState(false);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onloadend = () => {
        setImageToCrop(reader.result as string);
        setShowCropperModal(true);
      };
      reader.readAsDataURL(file);
    } else {
      setImageToCrop(null);
      setImagePreview(null);
      setCroppedImageFile(null);
      setValue('image', undefined);
    }
  };

  const handleCropComplete = useCallback((croppedBlob: Blob | null) => {
    setShowCropperModal(false);
    if (croppedBlob) {
      const file = new File([croppedBlob], `cropped-image-${Date.now()}.png`, { type: 'image/png' });
      setCroppedImageFile(file);
      setImagePreview(URL.createObjectURL(file));
      // We don't directly set react-hook-form's FileList here as we'll use croppedImageFile for upload
    } else {
      // Crop was cancelled or failed in modal
      setImageToCrop(null);
      // Keep existing preview/file if user cancels, or clear if needed
      // For now, let's clear if crop is explicitly cancelled by passing null
      // setImagePreview(null);
      // setCroppedImageFile(null);
      // setValue('image', undefined);
    }
  }, [setValue]);

  const handleCancelCrop = useCallback(() => {
    setShowCropperModal(false);
    setImageToCrop(null);
    // Reset the file input if crop is cancelled
    const fileInput = document.getElementById('comment-image-upload') as HTMLInputElement;
    if (fileInput) {
        fileInput.value = ''; // Clears the selected file in the input
    }
    setValue('image', undefined); // Clear RHF value
    // Don't clear existing preview or croppedImageFile here, user might want to keep it if they just closed modal
  }, [setValue]);


  const onSubmit: SubmitHandler<CommentFormValues> = async (data) => {
    if (!user) {
      setSubmissionError("Please log in to comment.");
      return;
    }

    setSubmissionError(null);
    let imageUrl: string | undefined = undefined;

    try {
      if (croppedImageFile) {
        const uploadedUrl = await uploadFileToMinio(croppedImageFile);
        if (!uploadedUrl) {
          setSubmissionError("Failed to upload image. Please try again.");
          return;
        }
        imageUrl = uploadedUrl;
      }

      const payload = {
        content: data.content,
        ...(imageUrl && { image_url: imageUrl }),
      };

      await submitComment(`/api/posts/${postId}/comments`, payload, (newComment: Comment) => {
        onCommentCreated(newComment);
        reset();
        setImagePreview(null);
        setImageToCrop(null);
        setCroppedImageFile(null);
        const fileInput = document.getElementById('comment-image-upload') as HTMLInputElement;
        if (fileInput) {
            fileInput.value = '';
        }
      });

      if (commentPostError) { // Check error state from useRequest after await
        setSubmissionError(commentPostError.message || "Failed to post comment.");
      }

    } catch (error: any) {
      console.error("Comment submission error:", error);
      setSubmissionError(error.message || "An unexpected error occurred while posting your comment.");
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
          {...register('image')}
          onChange={handleFileChange}
          className="mt-1"
        />
      </div>

      {imagePreview && (
        <div className="mt-2">
          <img src={imagePreview} alt="Selected preview" className="max-h-40 rounded-md border" />
          <Button type="button" outline={true} onClick={() => {
            setImagePreview(null);
            setCroppedImageFile(null);
            setImageToCrop(null);
            setValue('image', undefined);
            const fileInput = document.getElementById('comment-image-upload') as HTMLInputElement;
            if (fileInput) {
                fileInput.value = '';
            }
          }} className="mt-1 text-xs">
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

      {showCropperModal && imageToCrop && (
        <ImageCropperModal
          imageSrc={imageToCrop}
          onCropComplete={handleCropComplete}
          onClose={handleCancelCrop}
          aspect={1}
          isOpen={showCropperModal}
        />
      )}
    </form>
  );
};

export default CreateCommentForm;