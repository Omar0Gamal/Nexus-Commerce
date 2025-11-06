"use client"

import {
  BarChart3,
  TrendingUp,
  CreditCard,
  Users,
  Activity
} from "lucide-react"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useRevenueData } from "../hooks/useRevenueData"
import { MRRGrowthChart } from "./MRRGrowthChart"
import { SubscriptionChurnChart } from "./SubscriptionChurnChart"
import { accentColors } from "@/features/common/constants/colors"
import { MetricCard } from "@/features/common/components/MetricCard"

export function AnalyticsDashboard() {
  const { revenueData, isLoading } = useRevenueData()

  if (isLoading) {
    return (
      <div className="flex flex-col gap-6">
        <section className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="card-elevated border-border/40">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="size-9 rounded-lg" />
              </CardHeader>
              <CardContent>
                <div className="flex flex-col gap-2 mt-1">
                  <Skeleton className="h-7 w-28" />
                  <Skeleton className="h-4 w-40" />
                </div>
              </CardContent>
            </Card>
          ))}
        </section>
        <section className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card className="col-span-1 h-[380px] card-elevated border-border/40"><CardContent className="h-full pt-6"><Skeleton className="h-full w-full rounded-lg" /></CardContent></Card>
          <Card className="col-span-1 h-[380px] card-elevated border-border/40"><CardContent className="h-full pt-6"><Skeleton className="h-full w-full rounded-lg" /></CardContent></Card>
        </section>
      </div>
    )
  }

  if (!revenueData || Object.keys(revenueData).length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-24 text-center border rounded-xl border-dashed bg-muted/10">
        <div className="flex size-12 items-center justify-center rounded-full bg-muted mb-4">
          <Activity className="size-6 text-muted-foreground" />
        </div>
        <h3 className="text-lg font-medium text-foreground">No analytics data</h3>
        <p className="text-sm text-muted-foreground mt-2 max-w-sm">
          Analytics will populate here once tenants begin subscribing and generating revenue.
        </p>
      </div>
    )
  }

  const metrics = [
    {
      title: "Monthly Recurring Revenue (MRR)",
      value: revenueData.mrr || "$0.00",
      icon: CreditCard,
      trend: revenueData.mrrGrowth || "0%",
      trendUp: revenueData.mrrGrowth?.startsWith("+"),
      description: "compared to last month",
      accent: accentColors.emerald,
    },
    {
      title: "Annual Recurring Revenue (ARR)",
      value: revenueData.arr || "$0.00",
      icon: TrendingUp,
      trend: revenueData.arrGrowth || "0%",
      trendUp: revenueData.arrGrowth?.startsWith("+"),
      description: "projected run rate",
      accent: accentColors.blue,
    },
    {
      title: "Active Subscriptions",
      value: revenueData.activeSubscriptions || 0,
      icon: Users,
      trend: revenueData.subscriptionsGrowth || "0",
      trendUp: revenueData.subscriptionsGrowth?.startsWith("+"),
      description: "net new this month",
      accent: accentColors.violet,
    },
    {
      title: "Churn Rate",
      value: revenueData.churnRate || "0%",
      icon: BarChart3,
      trend: revenueData.churnStatus === "good" ? "Healthy" : revenueData.churnStatus ? "Needs attention" : "No data",
      trendUp: revenueData.churnStatus === "good",
      description: "rolling 30-day average",
      accent: accentColors.amber,
    }
  ]

  const mrrHistory = revenueData.mrrHistory || []
  const churnHistory = revenueData.churnHistory || []

  return (
    <div className="flex flex-col gap-6">
      <section
        aria-label="Revenue summary"
        className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 stagger-children"
      >
        {metrics.map((metric) => (
          <MetricCard key={metric.title} {...metric} />
        ))}
      </section>

      <section aria-label="Revenue charts" className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <MRRGrowthChart mrrHistory={mrrHistory} />
        <SubscriptionChurnChart churnHistory={churnHistory} />
      </section>
    </div>
  )
}
