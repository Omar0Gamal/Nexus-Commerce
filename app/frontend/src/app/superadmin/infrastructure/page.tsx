import { DashboardShell } from "@/features/common/components/dashboard-shell"
import { InfraHealthCards } from "@/features/infrastructure/components/InfraHealthCards"
import { AiConfigurationForm } from "@/features/infrastructure/components/ai-configuration-form"

export default function InfrastructurePage() {
  return (
    <DashboardShell
      title="Infrastructure"
      description="Live platform health and secure configuration controls."
    >
      <InfraHealthCards />

      <section aria-label="AI configuration">
        <div className="mb-3 flex flex-col gap-1">
          <h2 className="text-sm font-semibold tracking-tight">Secure configuration</h2>
          <p className="text-xs text-muted-foreground">
            Sensitive changes are audit-logged and require explicit confirmation.
          </p>
        </div>
        <AiConfigurationForm />
      </section>
    </DashboardShell>
  )
}
