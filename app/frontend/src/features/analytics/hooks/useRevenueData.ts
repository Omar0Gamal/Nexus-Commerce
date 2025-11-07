import { useGetApiV1PlatformAdminRevenue } from "@/api/endpoints/admin/admin"

export interface RevenueData {
  mrr?: string;
  arr?: string;
  mrrGrowth?: string;
  arrGrowth?: string;
  activeSubscriptions?: number;
  subscriptionsGrowth?: string;
  churnRate?: string;
  churnStatus?: string;
  mrrHistory?: { month: string; mrr: number }[];
  churnHistory?: { month: string; new: number; churned: number }[];
}

export function useRevenueData() {
  const { data: revenueEnvelope, isLoading, isError } = useGetApiV1PlatformAdminRevenue()
  const revenueData = revenueEnvelope?.data as unknown as RevenueData | undefined

  return {
    revenueData,
    isLoading,
    isError,
  }
}
