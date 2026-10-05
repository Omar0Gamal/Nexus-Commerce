import type { UserResponse, ShopSettingsResponse, ShopBrief } from '@/api/model'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  user: UserResponse | null
  activeShop: ShopSettingsResponse | null
  shopBrief: ShopBrief | null
  setAuth: (token: string, user: UserResponse, activeShop?: ShopSettingsResponse, shopBrief?: ShopBrief) => void
  setBootstrapData: (user: UserResponse, activeShop?: ShopSettingsResponse) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      activeShop: null,
      shopBrief: null,
      setAuth: (token, user, activeShop, shopBrief) => {
        localStorage.setItem('admin_access_token', token);
        set({ token, user, activeShop: activeShop || null, shopBrief: shopBrief || null });
      },
      setBootstrapData: (user, activeShop) =>
        set({ user, activeShop: activeShop || null }),
      logout: () => {
        localStorage.removeItem('admin_access_token');
        set({ token: null, user: null, activeShop: null, shopBrief: null });
      }
    }),
    {
      name: 'nexus-auth-storage',
      // Persist only the access token to prevent stale profile state across sessions.
      partialize: (state) => ({ token: state.token }),
    }
  )
)
