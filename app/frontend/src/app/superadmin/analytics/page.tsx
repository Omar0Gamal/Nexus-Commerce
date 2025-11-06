import { DashboardShell } from "@/features/common/components/dashboard-shell"
import { AnalyticsDashboard } from "@/features/analytics/components/AnalyticsDashboard"

export default function AnalyticsPage() {
  return (
    <DashboardShell
      title="Platform Analytics"
      description="Track platform revenue, subscriptions, and financial growth."
    >
      <AnalyticsDashboard />
    </DashboardShell>
  )
}
