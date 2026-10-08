import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import type { ShopStatus } from "@/lib/shops"

export const statusLabels: Record<ShopStatus, string> = {
  active: "Active",
  trialing: "Trialing",
  past_due: "Past due",
  suspended: "Suspended",
}

export function StatusBadge({ status }: { status: ShopStatus }) {
  const styles = {
    active: "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/20",
    trialing: "bg-sky-500/10 text-sky-700 dark:text-sky-400 border-sky-500/20",
    past_due: "bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20",
    suspended: "bg-destructive/10 text-destructive border-destructive/20",
  }[status]

  const dotColor = {
    active: "bg-emerald-500",
    trialing: "bg-sky-500",
    past_due: "bg-amber-500",
    suspended: "bg-destructive",
  }[status]

  return (
    <Badge variant="outline" className={cn("gap-1.5 font-semibold text-[11px] shadow-sm", styles)}>
      <div className="relative flex h-2 w-2 items-center justify-center">
        {status === "active" && (
          <span className={cn("absolute inline-flex h-full w-full animate-ping rounded-full opacity-75", dotColor)} aria-hidden="true" />
        )}
        <span className={cn("relative inline-flex size-1.5 rounded-full", dotColor)} aria-hidden="true" />
      </div>
      {statusLabels[status]}
    </Badge>
  )
}
