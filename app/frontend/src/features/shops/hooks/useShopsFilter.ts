import { useState, useMemo } from "react"
import type { Shop } from "@/lib/shops"

interface UseShopsFilterProps {
  shops: Shop[]
}

export function useShopsFilter({ shops }: UseShopsFilterProps) {
  const [query, setQuery] = useState("")
  const [statusFilter, setStatusFilter] = useState<string>("")
  const [planFilter, setPlanFilter] = useState<string>("")

  const filteredShops = useMemo(() => {
    const q = query.trim().toLowerCase()
    return shops.filter((shop) => {
      const matchesSearch = !q || shop.name.toLowerCase().includes(q)
      const matchesStatus = !statusFilter || statusFilter === "all_statuses" || shop.status === statusFilter
      const matchesPlan = !planFilter || planFilter === "all_plans" || shop.plan_name?.toLowerCase() === planFilter.toLowerCase()
      return matchesSearch && matchesStatus && matchesPlan
    })
  }, [query, statusFilter, planFilter, shops])

  return {
    query,
    setQuery,
    statusFilter,
    setStatusFilter,
    planFilter,
    setPlanFilter,
    filteredShops,
  }
}
