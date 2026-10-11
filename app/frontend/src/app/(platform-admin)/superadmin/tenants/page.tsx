import { DashboardShell } from "@/features/common/components/dashboard-shell"
import { TenantsDashboard } from "@/features/tenants/components/TenantsDashboard"

export default function Page() {

  return (
    <DashboardShell
      title="Platform Tenants"
      description="Monitor and manage all shops on the platform."
    >
      <TenantsDashboard />
    </DashboardShell>
  )
}
