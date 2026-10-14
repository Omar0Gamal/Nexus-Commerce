"use client"

import * as React from "react"
import { BarChart3, TrendingUp, Users, ShoppingBag } from "lucide-react"
import { MetricCard } from "@/features/common/components/MetricCard"
import { useGetApiV1AnalyticsDashboardSummary, useGetApiV1AnalyticsTrend } from "@/api/endpoints/analytics/analytics";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";
import { ProductPerformance } from "@/features/analytics/components/ProductPerformance";
import { Heatmap } from "@/features/analytics/components/Heatmap";
import { CohortRetention } from "@/features/analytics/components/CohortRetention";

export default function AnalyticsPage() {
  const { data: summaryResponse, isLoading } = useGetApiV1AnalyticsDashboardSummary()
  const { data: trendData } = useGetApiV1AnalyticsTrend()
  
  const metrics: any = summaryResponse?.data
  
  const chartData = trendData?.data?.map((p: any) => ({
    name: p.date ? new Date(p.date).toLocaleDateString('en-US', { weekday: 'short' }) : '',
    visits: p.visits || 0
  })) || [];

  return (
    <div className="flex flex-col gap-8 w-full max-w-6xl">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-3xl font-semibold tracking-tight text-zinc-50">Analytics</h1>
          <p className="text-zinc-400">Track your store's performance and revenue.</p>
        </div>
        <select className="h-10 w-40 rounded-full bg-zinc-950 border border-zinc-800 px-4 text-sm font-medium text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700">
          <option>Last 7 Days</option>
          <option>Last 30 Days</option>
          <option>Last Year</option>
          <option>All Time</option>
        </select>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {[1,2,3,4].map(i => <div key={i} className="h-32 rounded-2xl bg-zinc-800/50 animate-pulse border border-white/[0.05]" />)}
        </div>
      ) : !metrics ? (
        <div className="text-center py-12 text-zinc-500 bg-zinc-900/50 rounded-2xl border border-white/[0.05]">No data found</div>
      ) : (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <MetricCard
            title="Total Revenue"
            value={metrics.totalRevenue}
            trend={metrics.revenueChange}
            trendUp={metrics.revenueIsPositive}
            icon={TrendingUp}
            accent={{ bar: "bg-emerald-500", iconBg: "bg-emerald-500/10 dark:bg-emerald-500/15", iconText: "text-emerald-600 dark:text-emerald-400", trend: "text-emerald-600 dark:text-emerald-400", ring: "ring-emerald-500/20", stroke: "#10b981" }}
          />
          <MetricCard
            title="Total Orders"
            value={metrics.totalOrders}
            trend={metrics.ordersChange}
            trendUp={metrics.ordersIsPositive}
            icon={ShoppingBag}
            accent={{ bar: "bg-blue-500", iconBg: "bg-blue-500/10 dark:bg-blue-500/15", iconText: "text-blue-600 dark:text-blue-400", trend: "text-blue-600 dark:text-blue-400", ring: "ring-blue-500/20", stroke: "#3b82f6" }}
          />
          <MetricCard
            title="Active Customers"
            value={metrics.activeCustomers}
            trend={metrics.customersChange}
            trendUp={metrics.customersIsPositive}
            icon={Users}
            accent={{ bar: "bg-violet-500", iconBg: "bg-violet-500/10 dark:bg-violet-500/15", iconText: "text-violet-600 dark:text-violet-400", trend: "text-violet-600 dark:text-violet-400", ring: "ring-violet-500/20", stroke: "#8b5cf6" }}
          />
          <MetricCard
            title="Conversion Rate"
            value={metrics.conversionRate}
            trend={metrics.conversionChange}
            trendUp={metrics.conversionIsPositive}
            icon={BarChart3}
            accent={{ bar: "bg-amber-500", iconBg: "bg-amber-500/10 dark:bg-amber-500/15", iconText: "text-amber-600 dark:text-amber-400", trend: "text-amber-600 dark:text-amber-400", ring: "ring-amber-500/20", stroke: "#f59e0b" }}
          />
        </div>
      )}

      <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
        <div className="rounded-2xl bg-zinc-900/50 p-6 lg:p-8 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)]">
          <div className="flex flex-col gap-2 mb-8">
            <h2 className="text-lg font-medium text-zinc-100">Visits Overview</h2>
            <p className="text-sm text-zinc-500">Daily store visits for the last 7 days.</p>
          </div>
          
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
                <XAxis dataKey="name" fontSize={12} tickLine={false} axisLine={false} />
                <YAxis fontSize={12} tickLine={false} axisLine={false} />
                <Tooltip 
                  contentStyle={{ borderRadius: '10px', border: '1px solid hsl(var(--border))', backgroundColor: 'hsl(var(--card))' }}
                  cursor={{ fill: 'rgba(255,255,255,0.05)' }}
                />
                <Bar dataKey="visits" fill="#10b981" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <ProductPerformance />
        <Heatmap title="Storefront Activity" description="Daily visits and interactions over the last 30 days" />
      </div>

      <div className="w-full">
        <CohortRetention />
      </div>
    </div>
  )
}