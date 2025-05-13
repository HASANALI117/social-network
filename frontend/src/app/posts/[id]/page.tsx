'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { useRequest } from '@/hooks/useRequest';
import { Post } from '@/types/Post';
import { Comment } from '@/types/Comment';
import Link from 'next/link';
import { Avatar } from '@/components/ui/avatar';
import CreateCommentForm from '@/components/posts/CreateCommentForm';


export default function PostDetailPage() {
  const params = useParams();
  const postId = params.id as string;

  const [post, setPost] = useState<Post | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  
  const { get: getPost, isLoading: isLoadingPostData, error: postFetchError } = useRequest<Post>();
  const { get: getCommentsList, isLoading: isLoadingCommentsData, error: commentsFetchError } = useRequest<{ comments: Comment[]; count: number; limit: number; offset: number }>();

  // States to manage overall loading/error for the page sections
  const [isLoadingPost, setIsLoadingPost] = useState(true);
  const [errorPost, setErrorPost] = useState<Error | null>(null);
  const [isLoadingComments, setIsLoadingComments] = useState(false); // Initially false, true when fetching
  const [errorComments, setErrorComments] = useState<Error | null>(null);


  const fetchPostDetails = useCallback(async () => {
    if (!postId) return;
    setIsLoadingPost(true);
    setErrorPost(null);
    try {
      const fetchedPost = await getPost(`/api/posts/${postId}`);
      if (fetchedPost) {
        setPost(fetchedPost);
      } else {
        // This case might be handled by postFetchError if getPost returns null on error
        setErrorPost(new Error('Post not found or could not be fetched.'));
      }
    } catch (err: any) {
      // This catch might be redundant if useRequest handles and sets postFetchError
      setErrorPost(err);
    } finally {
      setIsLoadingPost(false);
    }
  }, [postId, getPost]);

  const fetchComments = useCallback(async () => {
    if (!postId) return;
    setIsLoadingComments(true);
    setErrorComments(null);
    try {
      // Backend endpoint is GET /api/posts/{postId}/comments
      // It returns { comments: Comment[], limit: number, offset: number, count: number }
      const response = await getCommentsList(`/api/posts/${postId}/comments`); // Add query params for pagination later if needed
      if (response && response.comments) {
        setComments(response.comments);
      } else if (commentsFetchError) { // Check error from hook if response is not as expected
        setErrorComments(commentsFetchError);
        setComments([]);
      } else {
        // Fallback error if response is null/undefined but hook didn't set an error
        setErrorComments(new Error('Could not fetch comments.'));
        setComments([]);
      }
    } catch (err: any) {
      setErrorComments(err);
      setComments([]);
    } finally {
      setIsLoadingComments(false);
    }
  }, [postId, getCommentsList, commentsFetchError]);

  useEffect(() => {
    fetchPostDetails();
  }, [fetchPostDetails]);

  // Effect for postFetchError from useRequest
  useEffect(() => {
    if (postFetchError) {
      setErrorPost(postFetchError);
      setPost(null); // Clear post if fetch failed
      setIsLoadingPost(false); // Ensure loading is stopped
    }
  }, [postFetchError]);

  useEffect(() => {
    if (post) { // Fetch comments only if post details are successfully loaded
      fetchComments();
    }
  }, [post, fetchComments]); // Add fetchComments to dependency array

  useEffect(() => {
    if (commentsFetchError) {
      setErrorComments(commentsFetchError);
      setComments([]);
      setIsLoadingComments(false);
    }
  }, [commentsFetchError]);


  const handleNewComment = (newComment: Comment) => {
    setComments(prevComments => [newComment, ...prevComments]);
  };

  if (isLoadingPost) {
    return <div className="text-center py-10">Loading post...</div>;
  }

  if (errorPost) {
    return <div className="text-center py-10 text-red-500">Error: {errorPost.message}</div>;
  }

  if (!post) {
    return <div className="text-center py-10">Post not found.</div>;
  }

  return (
    <div className="max-w-2xl mx-auto p-4">
      {/* Display Post Details */}
      <div className="bg-gray-800 rounded-lg shadow p-6 mb-6">
        <h1 className="text-3xl font-bold text-white mb-2">{post.title}</h1>
        <div className="flex items-center gap-3 mb-4 text-sm text-gray-400">
          {/* Ideally, link to user profile */}
          <span>By: {post.user_first_name} {post.user_last_name}</span>
          <span>At: {new Date(post.created_at).toLocaleString()}</span>
        </div>
        {post.image_url && (
          <img src={post.image_url} alt={post.title} className="rounded-lg mb-4 w-full object-cover" />
        )}
        <p className="text-gray-200 whitespace-pre-wrap">{post.content}</p>
      </div>

      {/* Comments Section */}
      <div className="mt-8">
        <h2 className="text-2xl font-semibold text-white mb-4">Comments</h2>
        {isLoadingComments && <div className="text-gray-400">Loading comments...</div>}
        {errorComments && <div className="text-red-500">Error loading comments: {errorComments.message}</div>}
        {!isLoadingComments && !errorComments && comments.length === 0 && (
          <div className="text-gray-400">No comments yet.</div>
        )}
        {!isLoadingComments && !errorComments && comments.length > 0 && (
          <div className="space-y-4">
            {comments.map((comment) => (
              <div key={comment.id} className="bg-slate-800 p-4 rounded-lg shadow-md">
                <div className="flex items-start mb-3">
                  <Avatar
                    src={comment.user_avatar_url || `https://ui-avatars.com/api/?name=${encodeURIComponent(comment.user_first_name || 'U')}+${encodeURIComponent(comment.user_last_name || 'N')}&background=random&color=random&size=128&bold=true`}
                    alt={comment.user_first_name || `User ${comment.user_id.substring(0,4)}`}
                    className="w-10 h-10 mr-3 rounded-full flex-shrink-0"
                  />
                  <div className="flex-grow">
                    <div className="flex items-baseline gap-2">
                      <Link href={`/profile/${comment.user_id}`} className="font-semibold text-indigo-400 hover:text-indigo-300 hover:underline">
                        {comment.user_first_name && comment.user_last_name
                          ? `${comment.user_first_name} ${comment.user_last_name}`
                          : `User ${comment.user_id.substring(0, 8)}`}
                      </Link>
                      {/* Optional: Display username if available in Comment type and data */}
                      {/* {comment.username && <span className="text-sm text-gray-500">@{comment.username}</span>} */}
                    </div>
                    <p className="text-xs text-gray-500 mt-0.5">
                      {new Date(comment.created_at).toLocaleString()}
                    </p>
                  </div>
                </div>
                <p className="text-gray-300 whitespace-pre-wrap leading-relaxed">{comment.content}</p>
                {comment.image_url && (
                  <div className="mt-3">
                    <img
                      src={comment.image_url}
                      alt={`Comment image by ${comment.user_first_name || 'user'}`}
                      className="rounded-lg max-h-80 w-auto object-cover border border-gray-600"
                      // Consider adding an onError handler for broken image links
                      // onError={(e) => (e.currentTarget.style.display = 'none')} // Example: hide if broken
                    />
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Add Comment Form Placeholder */}
      <div className="mt-8">
        <h2 className="text-2xl font-semibold text-white mb-4">Add a Comment</h2>
        {post && postId && (
          <CreateCommentForm postId={postId} onCommentCreated={handleNewComment} />
        )}
      </div>
    </div>
  );
}