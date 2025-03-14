import { UserType } from '@/types/User';
import { Post } from '@/types/Post';
import { formatDistanceToNow } from 'date-fns';

interface PostListProps {
  posts: Post[];
  user: UserType;
}

export default function PostList({ posts, user }: PostListProps) {
  return (
    <div className="space-y-6">
      {posts.map((post) => (
        <article key={post.id} className="bg-gray-800 rounded-lg p-6">
          <div className="mb-4">
            <h2 className="text-xl font-semibold mb-2">{post.title}</h2>
            <p className="text-gray-300">{post.content}</p>
            {post.imageUrl && (
              <img
                src={post.imageUrl}
                alt={post.title}
                className="mt-4 rounded-lg max-h-96 object-cover"
              />
            )}
          </div>
          <div className="flex items-center justify-between text-sm text-gray-400">
            <div className="flex gap-4">
              <span>{post.likes} likes</span>
              <span>{post.comments} comments</span>
            </div>
            <time dateTime={post.createdAt.toString()}>
              {formatDistanceToNow(new Date(post.createdAt), { addSuffix: true })}
            </time>
          </div>
        </article>
      ))}
    </div>
  );
}
