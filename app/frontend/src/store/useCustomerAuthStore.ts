import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface CustomerUser {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
}

interface CustomerAuthState {
  customerToken: string | null;
  customerUser: CustomerUser | null;
  setCustomerAuth: (token: string, user?: CustomerUser | null) => void;
  setCustomerUser: (user: CustomerUser | null) => void;
  logoutCustomer: () => void;
}

export const useCustomerAuthStore = create<CustomerAuthState>()(
  persist(
    (set) => ({
      customerToken: null,
      customerUser: null,
      setCustomerAuth: (token, user) => {
        if (typeof window !== 'undefined') {
          localStorage.setItem('customer_access_token', token);
          document.cookie = `customer_access_token=${token}; path=/; max-age=2592000; SameSite=Lax;`;
        }
        set({ customerToken: token, customerUser: user || null });
      },
      setCustomerUser: (user) => set({ customerUser: user }),
      logoutCustomer: () => {
        if (typeof window !== 'undefined') {
          localStorage.removeItem('customer_access_token');
          document.cookie = 'customer_access_token=; path=/; max-age=0; SameSite=Lax;';
        }
        set({ customerToken: null, customerUser: null });
      },
    }),
    {
      name: 'nexus-customer-auth-storage',
      partialize: (state) => ({ customerToken: state.customerToken, customerUser: state.customerUser }),
    }
  )
);
