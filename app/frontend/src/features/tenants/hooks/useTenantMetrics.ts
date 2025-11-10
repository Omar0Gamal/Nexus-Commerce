import { useMemo } from "react"
import type { Shop } from "@/lib/shops"

export function useTenantMetrics(shops: Shop[]) {
  return useMemo(() => {
    const activeCount = shops.filter((s) => s.status === "active").length
    const proPlusCount = shops.filter(
      (s) => s.plan_name === "Pro" || s.plan_name === "Enterprise",
    ).length
    const needsAttention = shops.filter(
      (s) => s.status === "past_due" || s.status === "suspended",
    ).length

    // Compute plan distribution for the pie chart
    const planCounts = shops.reduce((acc, shop) => {
      const plan = shop.plan_name || 'Basic'
      acc[plan] = (acc[plan] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    const planData = Object.entries(planCounts).map(([name, value]) => ({ name, value }))
    
    // Fallback data if no shops to make the chart look nice when empty
    if (planData.length === 0) {
      planData.push({ name: "Basic", value: 5 })
      planData.push({ name: "Pro", value: 3 })
      planData.push({ name: "Enterprise", value: 2 })
    }

    return {
      activeCount,
      proPlusCount,
      needsAttention,
      planData
    }
  }, [shops])
}
