"use client"

import {
  CreditCard,
  Users,
  Store,
  Activity
} from "lucide-react"
import { useRevenueData } from "../hooks/useRevenueData"
import { RecentActivityList } from "./RecentActivityList"
import { TopTenantsList } from "./TopTenantsList"
import { accentColors } from "@/features/common/constants/colors"
import { MetricCard } from "@/features/common/components/MetricCard"
import { mockSparklineData, mockActivities, mockTopTenants } from "@/mocks/analytics"

export function OverviewDashboard() {
  const { revenueData, isLoading } = useRevenueData()

  if (isLoading) {
    return <div className="h-40 flex items-center justify-center text-muted-foreground">Loading overview...</div>
  }

  const metrics = [
    {
      title: "Monthly Recurring Revenue",
      value: revenueData?.mrr || "$0.00",
      icon: CreditCard,
      trend: revenueData?.mrrGrowth || "0%",
      accent: accentColors.emerald,
      sparklineData: mockSparklineData.revenue,
    },
    {
      title: "Active Tenants",
      value: "5",
      icon: Store,
      trend: "+2",
      accent: accentColors.blue,
      sparklineData: mockSparklineData.tenants,
    },
    {
      title: "Platform Subscriptions",
      value: revenueData?.activeSubscriptions || 0,
      icon: Users,
      trend: "+8",
      accent: accentColors.violet,
      sparklineData: mockSparklineData.subs,
    },
    {
      title: "System Status",
      value: "Healthy",
      icon: Activity,
      trend: "99.9% uptime",
      accent: accentColors.amber,
      sparklineData: mockSparklineData.health,
    }
  ]

  return (
    <div className="flex flex-col gap-4">
      <section className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4 stagger-children">
        {metrics.map((metric) => (
          <MetricCard key={metric.title} {...metric} />
        ))}
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <RecentActivityList activities={mockActivities} />
        <TopTenantsList tenants={mockTopTenants} />
      </section>
    </div>
  )
}
