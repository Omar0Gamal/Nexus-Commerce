import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { customInstance } from '@/lib/axios';

export interface CartItem {
  id: string | number;
  productId?: string;
  variantId?: string;
  title: string;
  price: number;
  quantity: number;
  image: string;
}

interface BackendCartResponse {
  status: number;
  data: {
    items: Array<{
      product_id: string;
      variant_id?: string;
      title: string;
      price: number;
      quantity: number;
      image_url?: string;
      slug?: string;
      variant_name?: string;
    }>;
    total_items: number;
    total_price: number;
  };
}

interface CartStore {
  items: CartItem[];
  isLoading: boolean;
  isInitialized: boolean;
  fetchCart: () => Promise<void>;
  addItem: (item: CartItem) => Promise<void>;
  removeItem: (id: string | number, variantId?: string) => Promise<void>;
  updateQuantity: (id: string | number, quantity: number, variantId?: string) => Promise<void>;
  clearCart: () => Promise<void>;
}

function mapBackendItem(item: BackendCartResponse['data']['items'][0]): CartItem {
  return {
    id: item.product_id,
    productId: item.product_id,
    variantId: item.variant_id || undefined,
    title: item.title,
    price: Number(item.price) || 0,
    quantity: Number(item.quantity) || 1,
    image: item.image_url || '',
  };
}

export const useCartStore = create<CartStore>()(
  persist(
    (set, get) => ({
      items: [],
      isLoading: false,
      isInitialized: false,

      fetchCart: async () => {
        set({ isLoading: true });
        try {
          const res = await customInstance<BackendCartResponse>({
            url: '/api/v1/cart',
            method: 'GET',
          });

          if (res?.data?.items) {
            set({
              items: res.data.items.map(mapBackendItem),
              isLoading: false,
              isInitialized: true,
            });
          } else {
            set({ isLoading: false, isInitialized: true });
          }
        } catch (err) {
          console.warn('Failed to fetch cart from backend:', err);
          set({ isLoading: false, isInitialized: true });
        }
      },

      addItem: async (item) => {
        const prodId = String(item.productId || item.id);
        const variantId = item.variantId;

        set((state) => {
          const existing = state.items.find(
            (i) => String(i.id) === prodId && i.variantId === variantId
          );
          if (existing) {
            return {
              items: state.items.map((i) =>
                String(i.id) === prodId && i.variantId === variantId
                  ? { ...i, quantity: i.quantity + item.quantity }
                  : i
              ),
            };
          }
          return {
            items: [
              ...state.items,
              {
                ...item,
                id: prodId,
                productId: prodId,
              },
            ],
          };
        });

        try {
          const res = await customInstance<BackendCartResponse>({
            url: '/api/v1/cart/items',
            method: 'POST',
            data: {
              product_id: prodId,
              variant_id: variantId || undefined,
              quantity: item.quantity,
            },
          });

          if (res?.data?.items) {
            set({ items: res.data.items.map(mapBackendItem) });
          }
        } catch (err) {
          console.error('Failed to sync added item with backend:', err);
          get().fetchCart();
        }
      },

      removeItem: async (id, variantId) => {
        const prodId = String(id);

        set((state) => ({
          items: state.items.filter(
            (i) => !(String(i.id) === prodId && (variantId ? i.variantId === variantId : true))
          ),
        }));

        try {
          const params = new URLSearchParams({ product_id: prodId });
          if (variantId) params.set('variant_id', variantId);

          const res = await customInstance<BackendCartResponse>({
            url: `/api/v1/cart/items?${params.toString()}`,
            method: 'DELETE',
          });

          if (res?.data?.items) {
            set({ items: res.data.items.map(mapBackendItem) });
          }
        } catch (err) {
          console.error('Failed to remove item from backend cart:', err);
          get().fetchCart();
        }
      },

      updateQuantity: async (id, quantity, variantId) => {
        const prodId = String(id);

        if (quantity <= 0) {
          return get().removeItem(prodId, variantId);
        }

        set((state) => ({
          items: state.items.map((i) =>
            String(i.id) === prodId && (variantId ? i.variantId === variantId : true)
              ? { ...i, quantity }
              : i
          ),
        }));

        try {
          const params = new URLSearchParams({ product_id: prodId });
          if (variantId) params.set('variant_id', variantId);

          const res = await customInstance<BackendCartResponse>({
            url: `/api/v1/cart/items?${params.toString()}`,
            method: 'PATCH',
            data: { quantity },
          });

          if (res?.data?.items) {
            set({ items: res.data.items.map(mapBackendItem) });
          }
        } catch (err) {
          console.error('Failed to update item quantity in backend cart:', err);
          get().fetchCart();
        }
      },

      clearCart: async () => {
        set({ items: [] });

        try {
          await customInstance({
            url: '/api/v1/cart',
            method: 'DELETE',
          });
        } catch (err) {
          console.error('Failed to clear backend cart:', err);
        }
      },
    }),
    {
      name: 'nexus-cart-storage',
      partialize: (state) => ({ items: state.items }),
    }
  )
);
