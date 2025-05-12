import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { User } from '../types/User'

interface UserState {
  user: User | null
  isAuthenticated: boolean
  hydrated: boolean; // New state to track hydration
  login: (user: User) => void
  logout: () => void
  update: (user: Partial<User>) => void
  setHydrated: () => void; // Action to set hydrated
}

export const useUserStore = create<UserState>()(
  persist(
    (set, get) => ({ // Added get here
      user: null,
      isAuthenticated: false,
      hydrated: false, // Initialize hydrated as false
      login: (user) => set({ user, isAuthenticated: true }),
      logout: () => set({ user: null, isAuthenticated: false }),
      update: (updatedUser) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...updatedUser } : null,
        })),
      setHydrated: () => set({ hydrated: true }),
    }),
    {
      name: 'user-storage',
      skipHydration: true,
      onRehydrateStorage: () => {
        // This is called when rehydration is done
        return (state, error) => {
          if (state) {
            state.setHydrated();
          }
        }
      }
    }
  )
)