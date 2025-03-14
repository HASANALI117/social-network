export interface Post {
  id: string;
  user_id: string;
  title: string;
  content: string;
  image_url?: string;
  privacy: string;
  createdAt: Date;
}

export interface CreatePostFormValues {
  title: string;
  content: string;
  imageUrl?: string;
  privacy: string;
}

export interface PostResponse {
  id: string;
  user_id: string;
  title: string;
  content: string;
  image_url?: string;
  privacy: string;
  created_at: string;
}

export const transformPost = (post: PostResponse): Post => ({
  id: post.id,
  user_id: post.user_id,
  title: post.title,
  content: post.content,
  image_url: post.image_url,
  privacy: post.privacy,
  createdAt: new Date(post.created_at)
});

export const transformPosts = (posts: PostResponse[]): Post[] =>
  posts.map(transformPost);