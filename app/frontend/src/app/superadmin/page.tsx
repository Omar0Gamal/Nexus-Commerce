import { DashboardShell } from "@/features/common/components/dashboard-shell"
import { OverviewDashboard } from "@/features/analytics/components/OverviewDashboard"

export default function OverviewPage() {
  return (
    <DashboardShell
      title="Overview"
      description="Your command center for platform metrics and recent activity."
    >
      <OverviewDashboard />
    </DashboardShell>
  )
}
