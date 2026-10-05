import Axios, { AxiosRequestConfig } from 'axios';

export const customInstance = <T>(
  config: AxiosRequestConfig,
  options?: AxiosRequestConfig,
): Promise<T> => {
  const source = Axios.CancelToken.source();
  const headers = { ...config.headers, ...options?.headers } as Record<string, string>;

  if (typeof window !== 'undefined') {
    const pathname = window.location.pathname;
    const isPlatformOrTenantAdmin =
      pathname.startsWith('/superadmin') || pathname.startsWith('/admin');

    if (!headers['Authorization']) {
      const token = isPlatformOrTenantAdmin
        ? localStorage.getItem('admin_access_token')
        : (localStorage.getItem('customer_access_token') || localStorage.getItem('admin_access_token'));
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
    }

    if (!headers['X-Shop-ID']) {
      const shopId = localStorage.getItem('nexus_shop_id');
      if (shopId) {
        headers['X-Shop-ID'] = shopId;
      }
    }

    if (!headers['X-Cart-Token']) {
      let cartToken = localStorage.getItem('nexus_cart_token');
      if (!cartToken) {
        cartToken =
          typeof crypto !== 'undefined' && crypto.randomUUID
            ? crypto.randomUUID()
            : 'cart-' + Math.random().toString(36).substring(2) + Date.now().toString(36);
        localStorage.setItem('nexus_cart_token', cartToken);
      }
      headers['X-Cart-Token'] = cartToken;
    }
  }

  const promise = Axios({
    ...config,
    ...options,
    headers,
    cancelToken: source.token,
    baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  })
    .then((response) => {
      if (typeof window !== 'undefined' && response.headers?.['x-cart-token']) {
        localStorage.setItem('nexus_cart_token', response.headers['x-cart-token']);
      }
      return response.data;
    })
    .catch((error) => {
      if (error.response?.status === 401) {
        if (typeof window !== 'undefined') {
          const pathname = window.location.pathname;
          if (pathname.startsWith('/superadmin') && pathname !== '/superadmin/login') {
            localStorage.removeItem('admin_access_token');
            window.location.href = '/superadmin/login';
          } else if (pathname.startsWith('/admin') && pathname !== '/admin/login') {
            localStorage.removeItem('admin_access_token');
            window.location.href = '/admin/login';
          } else if (pathname.endsWith('/account')) {
            localStorage.removeItem('customer_access_token');
            window.location.href = '/login';
          }
        }
      }
      throw error;
    });

  // @ts-expect-error adding cancel property to promise
  promise.cancel = () => {
    source.cancel('Query was cancelled');
  };

  return promise;
};

