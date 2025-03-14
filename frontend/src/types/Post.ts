export interface Post {
  id: string;
  userId: string;
  title: string;
  content: string;
  imageUrl?: string;
  privacy: string;
  createdAt: Date;
  likes: number;
  comments: number;
}

export interface CreatePostFormValues {
  title: string;
  content: string;
  imageUrl?: string;
  privacy: string;
}