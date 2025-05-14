import React, { useEffect } from 'react';
import Link from 'next/link';
import { useRequest } from '../../../hooks/useRequest';
import { Post } from '../../../types/Post';
import PostCard from '../../common/PostCard';
import { Button } from '../../ui/button';
import { Text } from '../../ui/text';

interface GroupRecentPostsSummaryTabProps {
  groupId: string;
}

const GroupRecentPostsSummaryTab: React.FC<GroupRecentPostsSummaryTabProps> = ({ groupId }) => {
  const { data: posts, isLoading, error, get } = useRequest<Post[]>();

  useEffect(() => {
    get(`/api/groups/${groupId}/posts?limit=3&sort=recent`);
  }, [get, groupId]);

  if (isLoading) {
    return <Text>Loading recent posts...</Text>;
  }

  if (error) {
    return <Text color="red">Error loading posts: {error.message}</Text>;
  }

  if (!posts || !Array.isArray(posts) || posts.length === 0) {
    return (
      <div className="text-center py-4">
        <Text>No recent posts to display.</Text>
        <div className="mt-4">
          <Button href={`/groups/${groupId}/posts`} outline>
            View All Posts
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {posts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
      <div className="mt-6 text-center">
        <Button href={`/groups/${groupId}/posts`} color="purple">
          View All Posts
        </Button>
      </div>
    </div>
  );
};

export default GroupRecentPostsSummaryTab;