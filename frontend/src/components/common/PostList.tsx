import { Post } from '@/types/Post';
import PostCard from './PostCard'; // This will be correct once PostCard is in common

interface PostListProps {
  posts: Post[];
}

export default function PostList({ posts }: PostListProps) {
  return (
    <div className="space-y-6">
      {posts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
      {posts.length === 0 && (
        <p className="text-center text-gray-500">No posts yet</p>
      )}
    </div>
  );
}