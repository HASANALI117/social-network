export interface Post {
  id: string;
  user_id: string;
  title: string;
  content: string;
  image_url?: string;
  privacy: string;
  createdAt: Date;
  user_first_name?: string;
  user_last_name?: string;
  user_avatar_url?: string;
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
  user_first_name?: string;
  user_last_name?: string;
  user_avatar_url?: string;
}

export const transformPost = (post: PostResponse): Post => ({
  id: post.id,
  user_id: post.user_id,
  title: post.title,
  content: post.content,
  image_url: post.image_url,
  privacy: post.privacy,
  createdAt: new Date(post.created_at),
  user_first_name: post.user_first_name,
  user_last_name: post.user_last_name,
  user_avatar_url: post.user_avatar_url
});

export const transformPosts = (posts: PostResponse[]): Post[] =>
  posts.map(transformPost);