import { UserType } from '@/types/User';
import PostCard from './PostCard';

interface Post {
  id: number;
  content: string;
  likes: number;
  comments: number;
  timestamp: string;
}

interface PostListProps {
  posts: Post[];
  user: UserType;
}

export default function PostList({ posts, user }: PostListProps) {
  return (
    <div>
      {posts.map((post) => (
        <PostCard key={post.id} post={post} user={user} />
      ))}
    </div>
  );
}
