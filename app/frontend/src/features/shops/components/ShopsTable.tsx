"use client"

import { useState } from "react"
import { Search, ChevronRight, ChevronLeft, ArrowUpDown, ArrowDown, ArrowUp, ChevronsLeft, ChevronsRight } from "lucide-react"

import type { Shop } from "@/lib/shops"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { Card, CardHeader, CardContent } from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

import { useShopsFilter } from "../hooks/useShopsFilter"
import { ShopCard } from "./ShopCard"
import { ShopDetailsSheet } from "./ShopDetailsSheet"

function GridSkeleton() {
  return (
    <>
      {Array.from({ length: 6 }).map((_, i) => (
        <Card key={i} className="flex flex-col justify-between h-[240px] card-elevated border-border/40">
          <CardHeader className="pb-4">
            <div className="flex items-start gap-3">
              <Skeleton className="size-10 rounded-xl" />
              <div className="flex flex-col gap-2 w-full mt-1">
                <Skeleton className="h-5 w-32" />
                <Skeleton className="h-3 w-24" />
              </div>
            </div>
          </CardHeader>
          <CardContent className="flex-1">
            <div className="flex items-center gap-2 mb-4">
              <Skeleton className="h-5 w-20 rounded-full" />
              <Skeleton className="h-5 w-16 rounded-full" />
            </div>
            <Skeleton className="h-16 w-full rounded-lg" />
          </CardContent>
          <div className="p-3 bg-muted/20 border-t border-border/40 flex items-center justify-between">
            <Skeleton className="h-7 w-20" />
            <Skeleton className="h-7 w-24 rounded-md" />
          </div>
        </Card>
      ))}
    </>
  )
}

export interface ShopsTableProps {
  shops: Shop[];
  isLoading?: boolean;
  totalCount?: number;
  page?: number;
  limit?: number;
  sort?: string;
  order?: "asc" | "desc";
  onPageChange?: (page: number) => void;
  onLimitChange?: (limit: number) => void;
  onSortChange?: (sort: string) => void;
  onOrderChange?: (order: "asc" | "desc") => void;
}

export function ShopsTable({
  shops,
  isLoading = false,
  totalCount = 0,
  page = 1,
  limit = 10,
  sort = "created_at",
  order = "desc",
  onPageChange,
  onLimitChange,
  onSortChange,
  onOrderChange
}: ShopsTableProps) {
  const [selectedShop, setSelectedShop] = useState<Shop | null>(null)

  const {
    query, setQuery,
    statusFilter, setStatusFilter,
    planFilter, setPlanFilter,
    filteredShops
  } = useShopsFilter({ shops })

  const handleSort = (field: string) => {
    if (!onSortChange || !onOrderChange) return;
    if (sort === field) {
      onOrderChange(order === "asc" ? "desc" : "asc")
    } else {
      onSortChange(field)
      onOrderChange("desc")
    }
  }

  const renderSortIcon = (field: string) => {
    if (sort !== field) return <ArrowUpDown className="ml-2 size-4 text-muted-foreground/30" />
    return order === "asc" ? <ArrowUp className="ml-2 size-4" /> : <ArrowDown className="ml-2 size-4" />
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-1 flex-col sm:flex-row gap-3 items-center">
          <div className="relative w-full sm:max-w-xs">
            <Search
              className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <Input
              type="search"
              placeholder="Search by shop name..."
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              className="pl-9 bg-card"
              aria-label="Search shops by name"
            />
          </div>

          <div className="flex gap-3 w-full sm:w-auto">
            <Select value={statusFilter} onValueChange={(val) => setStatusFilter(val || '')}>
              <SelectTrigger className="w-[140px] bg-card">
                <SelectValue placeholder="Status">
                {statusFilter === "all_statuses" ? "All Statuses" :
                 statusFilter === "active" ? "Active" :
                 statusFilter === "trialing" ? "Trialing" :
                 statusFilter === "past_due" ? "Past Due" :
                 statusFilter === "suspended" ? "Suspended" : undefined}
              </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all_statuses" label="All Statuses">All Statuses</SelectItem>
                <SelectItem value="active" label="Active">Active</SelectItem>
                <SelectItem value="trialing" label="Trialing">Trialing</SelectItem>
                <SelectItem value="past_due" label="Past Due">Past Due</SelectItem>
                <SelectItem value="suspended" label="Suspended">Suspended</SelectItem>
              </SelectContent>
            </Select>

            <Select value={planFilter} onValueChange={(val) => setPlanFilter(val || '')}>
              <SelectTrigger className="w-[140px] bg-card">
                <SelectValue placeholder="Plan">
                {planFilter === "all_plans" ? "All Plans" :
                 planFilter === "basic" ? "Basic" :
                 planFilter === "pro" ? "Pro" :
                 planFilter === "enterprise" ? "Enterprise" : undefined}
              </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all_plans" >All Plans</SelectItem>
                <SelectItem value="basic" label="Basic">Basic</SelectItem>
                <SelectItem value="pro" label="Pro">Pro</SelectItem>
                <SelectItem value="enterprise" label="Enterprise">Enterprise</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="text-sm text-muted-foreground whitespace-nowrap hidden lg:block">
          {isLoading ? (
            <Skeleton className="h-4 w-24 inline-block align-middle" />
          ) : (
            `${filteredShops.length} of ${shops.length} shops`
          )}
        </div>
      </div>

      <div className="flex flex-col gap-5 mt-2">
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            <GridSkeleton />
          </div>
        ) : filteredShops.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-3 text-muted-foreground p-12 border border-border/40 border-dashed rounded-xl bg-card/50 shadow-sm">
            <div className="flex size-14 items-center justify-center rounded-full bg-muted">
              <Search className="size-7 text-muted-foreground/50" />
            </div>
            <div className="text-center mt-2">
              <p className="text-base font-semibold text-foreground">No shops found</p>
              <p className="text-sm mt-1 max-w-sm mx-auto">
                {query ? `We couldn't find any shops matching "${query}". Try adjusting your filters.` : "There are currently no active shops on the platform."}
              </p>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5 stagger-children">
            {filteredShops.map((shop) => (
              <ShopCard key={shop.id} shop={shop} onSelect={setSelectedShop} />
            ))}
          </div>
        )}

        {/* Pagination Controls */}
        <div className="flex flex-col sm:flex-row gap-4 items-center justify-between p-4 rounded-xl border border-border/40 bg-card/50 shadow-sm mt-2">
          <div className="text-sm text-muted-foreground text-center sm:text-left">
            {isLoading ? (
              <Skeleton className="h-4 w-32" />
            ) : (
              <>
                Showing{" "}
                <span className="font-medium text-foreground">
                  {Math.min((page - 1) * limit + 1, totalCount)}
                </span>{" "}
                to{" "}
                <span className="font-medium text-foreground">
                  {Math.min(page * limit, totalCount)}
                </span>{" "}
                of <span className="font-medium text-foreground">{totalCount}</span> shops
              </>
            )}
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="icon"
              className="size-9 bg-card border-border/50"
              onClick={() => onPageChange?.(1)}
              disabled={page === 1 || isLoading}
            >
              <ChevronsLeft className="size-4" />
              <span className="sr-only">First page</span>
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="size-9 bg-card border-border/50"
              onClick={() => onPageChange?.(page - 1)}
              disabled={page === 1 || isLoading}
            >
              <ChevronLeft className="size-4" />
              <span className="sr-only">Previous page</span>
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="size-9 bg-card border-border/50"
              onClick={() => onPageChange?.(page + 1)}
              disabled={page * limit >= totalCount || isLoading}
            >
              <ChevronRight className="size-4" />
              <span className="sr-only">Next page</span>
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="size-9 bg-card border-border/50"
              onClick={() => onPageChange?.(Math.ceil(totalCount / limit))}
              disabled={page * limit >= totalCount || isLoading || totalCount === 0}
            >
              <ChevronsRight className="size-4" />
              <span className="sr-only">Last page</span>
            </Button>
          </div>
        </div>
      </div>

      <ShopDetailsSheet
        shop={selectedShop}
        open={!!selectedShop}
        onOpenChange={(open) => !open && setSelectedShop(null)}
      />
    </div>
  )
}
