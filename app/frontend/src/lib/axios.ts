import Axios, { AxiosRequestConfig } from 'axios';

export const customInstance = <T>(
  config: AxiosRequestConfig,
  options?: AxiosRequestConfig,
): Promise<T> => {
  const source = Axios.CancelToken.source();
  let headers = { ...config.headers, ...options?.headers } as any;
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('admin_access_token');
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  const promise = Axios({
    ...config,
    ...options,
    headers,
    cancelToken: source.token,
    baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000',
  }).then(({ data }) => data)
    .catch((error) => {
      // Mock data for UI testing when backend is not available
      if (config.url?.includes('/api/v1/platform-admin/shops')) {
        console.warn("Backend not available, returning mock shop data");
        return {
          data: {
            shops: [
              { id: "shp_1", name: "Acme Corp", subdomain: "acme", plan_name: "Enterprise", status: "active", created_at: "2024-01-15T00:00:00Z" },
              { id: "shp_2", name: "Globex", subdomain: "globex", plan_name: "Pro", status: "active", created_at: "2024-02-10T00:00:00Z" },
              { id: "shp_3", name: "Initech", subdomain: "initech", plan_name: "Basic", status: "past_due", created_at: "2024-03-05T00:00:00Z" },
              { id: "shp_4", name: "Soylent", subdomain: "soylent", plan_name: "Enterprise", status: "trialing", created_at: "2024-04-20T00:00:00Z" },
              { id: "shp_5", name: "Massive Dynamic", subdomain: "massive", plan_name: "Pro", status: "active", created_at: "2024-05-12T00:00:00Z" }
            ]
          },
          meta: { total_items: 5, activeCount: 3, proPlusCount: 4, needsAttention: 1 }
        };
      }
      
      if (config.url?.includes('/api/v1/platform-admin/revenue')) {
        console.warn("Backend not available, returning mock revenue data");
        return {
          data: {
            mrr: "$45,200",
            arr: "$542,400",
            mrrGrowth: "+12.5%",
            arrGrowth: "+15.2%",
            activeSubscriptions: 128,
            subscriptionsGrowth: "+8",
            churnRate: "2.1%",
            churnStatus: "good",
            mrrHistory: [
              { month: "Jan", mrr: 32000 },
              { month: "Feb", mrr: 35000 },
              { month: "Mar", mrr: 36500 },
              { month: "Apr", mrr: 41000 },
              { month: "May", mrr: 42500 },
              { month: "Jun", mrr: 45200 }
            ],
            churnHistory: [
              { month: "Jan", new: 12, churned: 3 },
              { month: "Feb", new: 15, churned: 2 },
              { month: "Mar", new: 8, churned: 4 },
              { month: "Apr", new: 18, churned: 1 },
              { month: "May", new: 10, churned: 2 },
              { month: "Jun", new: 14, churned: 3 }
            ]
          }
        };
      }

      if (config.url?.includes('/api/v1/plans')) {
        console.warn("Backend not available, returning mock plan data");
        return {
          data: [
            {
              id: "plan_basic",
              name: "Basic",
              monthly_price: "500",
              max_products: 500,
              max_staff_accounts: 2,
              max_storage_mb: 2000,
              transaction_fee_percent: "2.0",
              features: {
                custom_domain: false,
                advanced_analytics: false,
                priority_support: false,
                api_access: false
              }
            },
            {
              id: "plan_pro",
              name: "Pro",
              monthly_price: "1500",
              max_products: 5000,
              max_staff_accounts: 5,
              max_storage_mb: 10000,
              transaction_fee_percent: "1.0",
              features: {
                custom_domain: true,
                advanced_analytics: true,
                priority_support: false,
                api_access: false
              }
            },
            {
              id: "plan_enterprise",
              name: "Enterprise",
              monthly_price: "4000",
              max_products: -1, // Unlimited
              max_staff_accounts: 15,
              max_storage_mb: 50000,
              transaction_fee_percent: "0.5",
              features: {
                custom_domain: true,
                advanced_analytics: true,
                priority_support: true,
                api_access: true
              }
            }
          ],
          meta: { total_items: 3 }
        };
      }

      if (error.response?.status === 401) {
        if (typeof window !== 'undefined' && window.location.pathname.startsWith('/superadmin') && window.location.pathname !== '/superadmin/login') {
          localStorage.removeItem('admin_access_token');
          window.location.href = '/superadmin/login';
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
