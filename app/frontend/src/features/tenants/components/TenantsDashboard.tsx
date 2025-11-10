"use client"

import { useState } from "react"
import { useGetApiV1PlatformAdminShops } from "@/api/endpoints/admin/admin"
import type { Shop } from "@/lib/shops"

import { ShopsTable } from "@/features/shops/components/ShopsTable"
import { TenantsHero } from "./TenantsHero"
import { PlanDistributionChart } from "./PlanDistributionChart"
import { useTenantMetrics } from "../hooks/useTenantMetrics"

export function TenantsDashboard() {
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)
  const [sort, setSort] = useState("created_at")
  const [order, setOrder] = useState<"asc" | "desc">("desc")

  const { data: shopsEnvelope, isLoading, isError } = useGetApiV1PlatformAdminShops(
    { page, limit, sort, order }
  )

  const shops = (shopsEnvelope?.data?.shops as Shop[]) || []
  const totalCount = (shopsEnvelope?.meta?.total_items as number) || shops.length

  const metrics = useTenantMetrics(shops)

  return (
    <div className="flex flex-col gap-6">
      <section
        aria-label="Platform summary"
        className="grid grid-cols-1 lg:grid-cols-3 gap-4"
      >
        <TenantsHero 
          totalShops={shops.length}
          activeCount={shopsEnvelope?.meta?.activeCount as number || metrics.activeCount}
          proPlusCount={shopsEnvelope?.meta?.proPlusCount as number || metrics.proPlusCount}
          needsAttention={shopsEnvelope?.meta?.needsAttention as number || metrics.needsAttention}
          isLoading={isLoading}
        />
        
        <PlanDistributionChart 
          totalShops={shops.length}
          planData={metrics.planData}
          isLoading={isLoading}
        />
      </section>

      <section aria-label="All shops">
        {isError ? (
          <div className="flex h-32 items-center justify-center rounded-xl border bg-card text-destructive card-elevated">
            Failed to load shops
          </div>
        ) : (
          <ShopsTable 
            shops={shops} 
            isLoading={isLoading} 
            totalCount={totalCount}
            page={page}
            limit={limit}
            sort={sort}
            order={order}
            onPageChange={setPage}
            onLimitChange={setLimit}
            onSortChange={setSort}
            onOrderChange={setOrder}
          />
        )}
      </section>
    </div>
  )
}
