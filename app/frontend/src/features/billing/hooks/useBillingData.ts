import { useGetApiV1PlatformAdminBilling } from "@/api/endpoints/admin/admin"

export interface Payout {
  id: string;
  date: string;
  amount: string;
  status: string;
  bank: string;
}

export interface BillingData {
  platformFeesAvailable?: string;
  platformFeesPending?: string;
  totalProcessedVolume?: string;
  activeStripeAccounts?: number;
  recentPayouts?: Payout[];
}

export function useBillingData() {
  const { data: billingEnvelope, isLoading, isError } = useGetApiV1PlatformAdminBilling()
  const billingData = billingEnvelope?.data as unknown as BillingData | undefined

  return {
    billingData,
    isLoading,
    isError,
  }
}
