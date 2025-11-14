import { Wallet, Landmark, ArrowRightLeft, Receipt } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { BillingData } from "../hooks/useBillingData"

const cardConfigs = [
  {
    title: "Available Platform Fees",
    key: "platformFeesAvailable" as const,
    icon: Wallet,
    description: "Ready to be paid out to your bank account",
    bar: "bg-emerald-500",
    iconBg: "bg-emerald-500/10 dark:bg-emerald-500/15",
    iconText: "text-emerald-600 dark:text-emerald-400",
    ring: "ring-emerald-500/20",
  },
  {
    title: "Pending Platform Fees",
    key: "platformFeesPending" as const,
    icon: ArrowRightLeft,
    description: "Currently clearing from Stripe",
    bar: "bg-amber-500",
    iconBg: "bg-amber-500/10 dark:bg-amber-500/15",
    iconText: "text-amber-600 dark:text-amber-400",
    ring: "ring-amber-500/20",
  },
  {
    title: "Total Processed Volume",
    key: "totalProcessedVolume" as const,
    icon: Receipt,
    description: "All time GMV processed across tenants",
    bar: "bg-blue-500",
    iconBg: "bg-blue-500/10 dark:bg-blue-500/15",
    iconText: "text-blue-600 dark:text-blue-400",
    ring: "ring-blue-500/20",
  },
  {
    title: "Active Connected Accounts",
    key: "activeStripeAccounts" as const,
    icon: Landmark,
    description: "Tenants fully onboarded to Stripe Connect",
    bar: "bg-violet-500",
    iconBg: "bg-violet-500/10 dark:bg-violet-500/15",
    iconText: "text-violet-600 dark:text-violet-400",
    ring: "ring-violet-500/20",
  },
] as const

interface BillingHeroCardsProps {
  billingData: BillingData
}

export function BillingHeroCards({ billingData }: BillingHeroCardsProps) {
  return (
    <section
      aria-label="Billing summary"
      className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 stagger-children"
    >
      {cardConfigs.map((config) => {
        const value = config.key === "activeStripeAccounts"
          ? billingData[config.key] || 0
          : billingData[config.key] || "$0.00"

        return (
          <Card
            key={config.title}
            className="group relative overflow-hidden card-elevated border-border/40 dark:border-border/30 bg-card hover:-translate-y-0.5 transition-all duration-300"
          >
            {/* Left accent bar */}
            <div className={`absolute left-0 top-0 bottom-0 w-[3px] ${config.bar} rounded-l-lg opacity-70 group-hover:opacity-100 transition-opacity`} />

            <CardHeader className="flex flex-row items-center justify-between pb-2 pl-5">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {config.title}
              </CardTitle>
              <div className={`flex size-9 items-center justify-center rounded-lg ${config.iconBg} ring-1 ${config.ring} transition-transform duration-300 group-hover:scale-110`}>
                <config.icon className={`size-4 ${config.iconText}`} />
              </div>
            </CardHeader>
            <CardContent className="pl-5">
              <div className="text-2xl font-bold tabular-nums tracking-tight">{value}</div>
              <p className="text-xs text-muted-foreground/70 mt-1.5">
                {config.description}
              </p>
            </CardContent>
          </Card>
        )
      })}
    </section>
  )
}
