import React from 'react';
import Link from 'next/link';
import { format } from 'date-fns';
import { Avatar } from '../../ui/avatar';
import { Text } from '../../ui/text';
import { PostSummary } from '../../../types/Group';

interface GroupPostsTabProps {
  posts: PostSummary[] | undefined;
}

export default function GroupPostsTab({ posts }: GroupPostsTabProps) {
  if (!posts || posts.length === 0) {
    return <Text className="text-center text-gray-400 py-4">No posts in this group yet.</Text>;
  }

  return (
    <div className="space-y-4">
      {posts.map((post: PostSummary) => {
        let formattedPostDate = 'Date not available';
        if (post.created_at) {
          try {
            const dateObj = new Date(post.created_at);
            if (!isNaN(dateObj.getTime())) {
              formattedPostDate = format(dateObj, 'PPpp');
            } else {
              console.error("Invalid date string for post:", post.created_at);
              formattedPostDate = 'Invalid date';
            }
          } catch (e) {
            console.error("Error parsing date string for post:", post.created_at, e);
            formattedPostDate = 'Error parsing date';
          }
        }
        return (
          <div key={post.id} className="p-4 bg-gray-700 rounded-lg shadow">
            <div className="flex items-center mb-2">
              <Avatar
                src={post.creator_avatar_url || null}
                initials={!post.creator_avatar_url && post.creator_name ? post.creator_name.substring(0, 1).toUpperCase() : undefined}
                alt={post.creator_name || 'Creator'}
                className="h-10 w-10 mr-3 rounded-full"
              />
              <div>
                <Text className="font-semibold text-purple-300">{post.creator_name}</Text>
                <Text className="text-xs text-gray-400">
                  {formattedPostDate}
                </Text>
              </div>
            </div>
            {/* Content preview was removed in original, keeping it that way */}
            {/* <Text className="text-gray-300 whitespace-pre-wrap">
              {post.content.length > 150 ? \`\${post.content.substring(0, 150)}...\` : post.content}
            </Text> */}
            {post.image_url && (
              <div className="mt-2">
                <img src={post.image_url} alt="Post image" className="max-h-60 rounded-md object-cover" />
              </div>
            )}
            <Link href={`/posts/${post.id}`} className="text-sm text-purple-400 hover:underline mt-2 inline-block">
              View Post
            </Link>
          </div>
        );
      })}
    </div>
  );
}