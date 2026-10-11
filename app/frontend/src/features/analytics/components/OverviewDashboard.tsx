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
import { useGetApiV1PlatformAdminShops } from "@/api/endpoints/admin/admin";

export function OverviewDashboard() {
  const { revenueData, isLoading: isRevenueLoading } = useRevenueData()
  const { data: shopsEnvelope, isLoading: isShopsLoading } = useGetApiV1PlatformAdminShops({ limit: 5 })
  
  const isLoading = isRevenueLoading || isShopsLoading
  const activeTenants = shopsEnvelope?.meta?.total_items || (shopsEnvelope?.data?.shops as any[])?.length || 0
  const topTenants = ((shopsEnvelope?.data?.shops as any[]) || []).map((shop: any) => ({
    name: shop.name || "Unknown Shop",
    revenue: "$0.00",
    plan: shop.settings?.plan || "Pro",
  }))

  if (isLoading) {
    return <div className="h-40 flex items-center justify-center text-muted-foreground">Loading overview...</div>
  }

  const mrrSparkline = revenueData?.mrrHistory?.map(h => ({ v: h.mrr })) || [];
  const subsSparkline = revenueData?.churnHistory?.map(h => ({ v: h.new })) || [];

  const metrics = [
    {
      title: "Monthly Recurring Revenue",
      value: revenueData?.mrr || "$0.00",
      icon: CreditCard,
      trend: revenueData?.mrrGrowth || "0%",
      accent: accentColors.cobalt,
      sparklineData: mrrSparkline.length > 0 ? mrrSparkline : undefined,
    },
    {
      title: "Active Tenants",
      value: activeTenants.toString(),
      icon: Store,
      trend: "Current",
      accent: accentColors.cobalt,
      sparklineData: undefined,
    },
    {
      title: "Platform Subscriptions",
      value: revenueData?.activeSubscriptions || 0,
      icon: Users,
      trend: "+8",
      accent: accentColors.cobalt,
      sparklineData: subsSparkline.length > 0 ? subsSparkline : undefined,
    },
    {
      title: "System Status",
      value: "Healthy",
      icon: Activity,
      trend: "99.9% uptime",
      accent: accentColors.cobalt,
      sparklineData: undefined,
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
        <RecentActivityList activities={[]} />
        <TopTenantsList tenants={topTenants} />
      </section>
    </div>
  )
}
