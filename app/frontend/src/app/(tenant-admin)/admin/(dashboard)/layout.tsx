import { TenantDashboardShell } from "@/features/tenant-admin/components/TenantDashboardShell"

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <TenantDashboardShell>
      {children}
    </TenantDashboardShell>
  )
}
