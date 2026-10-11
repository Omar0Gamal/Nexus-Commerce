"use client"

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useState } from "react"
import { ArrowDown, ArrowUp, ChevronsUpDown } from "lucide-react"
import { useGetApiV1AnalyticsProductsPerformance } from "@/api/endpoints/analytics/analytics"

type ProductData = {
  id: string
  name: string
  views: number
  conversionRate: number
  revenue: number
}

type SortField = keyof Omit<ProductData, "id">
type SortOrder = "asc" | "desc"

export function ProductPerformance() {
  const { data: performanceResponse, isLoading, isError } = useGetApiV1AnalyticsProductsPerformance()
  const [sortField, setSortField] = useState<SortField>("revenue")
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc")

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc")
    } else {
      setSortField(field)
      setSortOrder("desc")
    }
  }

  const rawEntries = performanceResponse?.data || []
  const products: ProductData[] = rawEntries.map((p, idx) => ({
    id: p.product_id || `prod-${idx}`,
    name: p.title || (p.product_id ? `Product (${p.product_id.slice(0, 8)})` : `Item #${idx + 1}`),
    views: Number(p.views || 0),
    conversionRate: Number(
      ((p.cart_to_purchase_pct ?? p.view_to_cart_pct ?? 0)).toFixed(1)
    ),
    revenue: (p.revenue_cents || 0) / 100,
  }))

  const sortedProducts = [...products].sort((a, b) => {
    let comparison = 0
    if (a[sortField] > b[sortField]) comparison = 1
    if (a[sortField] < b[sortField]) comparison = -1
    return sortOrder === "asc" ? comparison : -comparison
  })

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) return <ChevronsUpDown className="ml-1 size-3 text-muted-foreground" />
    return sortOrder === "asc" ? <ArrowUp className="ml-1 size-3" /> : <ArrowDown className="ml-1 size-3" />
  }

  return (
    <Card className="card-elevated border-border/40">
      <CardHeader>
        <CardTitle className="text-base font-medium">Product Performance</CardTitle>
        <CardDescription>Views, conversion rate, and revenue by product</CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-3 py-2">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        ) : isError ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            Failed to load product performance analytics from server.
          </div>
        ) : products.length === 0 ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            No product activity recorded yet in this reporting window.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="text-muted-foreground border-b border-border/50">
                <tr>
                  <th
                    className="font-medium pb-3 pr-4 cursor-pointer hover:text-foreground transition-colors"
                    onClick={() => handleSort("name")}
                  >
                    <div className="flex items-center">Product <SortIcon field="name" /></div>
                  </th>
                  <th
                    className="font-medium pb-3 px-4 text-right cursor-pointer hover:text-foreground transition-colors"
                    onClick={() => handleSort("views")}
                  >
                    <div className="flex items-center justify-end">Views <SortIcon field="views" /></div>
                  </th>
                  <th
                    className="font-medium pb-3 px-4 text-right cursor-pointer hover:text-foreground transition-colors"
                    onClick={() => handleSort("conversionRate")}
                  >
                    <div className="flex items-center justify-end">Conversion <SortIcon field="conversionRate" /></div>
                  </th>
                  <th
                    className="font-medium pb-3 pl-4 text-right cursor-pointer hover:text-foreground transition-colors"
                    onClick={() => handleSort("revenue")}
                  >
                    <div className="flex items-center justify-end">Revenue <SortIcon field="revenue" /></div>
                  </th>
                </tr>
              </thead>
              <tbody>
                {sortedProducts.map((product) => (
                  <tr key={product.id} className="border-b border-border/20 last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="py-3 pr-4 font-medium text-foreground">{product.name}</td>
                    <td className="py-3 px-4 text-right">{product.views.toLocaleString()}</td>
                    <td className="py-3 px-4 text-right">{product.conversionRate}%</td>
                    <td className="py-3 pl-4 text-right">${product.revenue.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
