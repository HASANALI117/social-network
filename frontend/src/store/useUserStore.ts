import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { UserType } from '../types/User'

interface UserState {
  user: UserType | null
  isAuthenticated: boolean
  login: (user: UserType) => void
  logout: () => void
  update: (user: Partial<UserType>) => void
}

export const useUserStore = create<UserState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      login: (user) => set({ user, isAuthenticated: true }),
      logout: () => set({ user: null, isAuthenticated: false }),
      update: (updatedUser) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...updatedUser } : null,
        })),
    }),
    {
      name: 'user-storage',
      skipHydration: true, // Important for Next.js to prevent hydration mismatch
    }
  )
)