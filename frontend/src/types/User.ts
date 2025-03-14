// Global type declarations for the frontend
export interface UserType {
  id: string;
  username?: string;
  email: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
  about_me?: string;
  birth_date: string;
  created_at: string;
  updated_at: string;
}

export interface UserSignupData {
  first_name: string;
  last_name: string;
  email: string;
  password: string;
  birth_date: string;
  username?: string;
  about_me?: string;
  avatar_url?: File | null;
}
