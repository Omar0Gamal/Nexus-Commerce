import { DashboardShell } from "@/features/common/components/dashboard-shell"
import { BillingDashboard } from "@/features/billing/components/BillingDashboard"

export default function BillingPage() {
  return (
    <DashboardShell
      title="Platform Billing"
      description="Manage platform fee collection, payouts, and Stripe Connect status."
    >
      <BillingDashboard />
    </DashboardShell>
  )
}
