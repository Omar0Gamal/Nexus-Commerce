"use client"

import * as React from "react"
import { useQueryState } from 'nuqs'
import Link from 'next/link'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Search, ChevronRight, Package, Truck, CheckCircle2, Clock, AlertCircle, RefreshCw } from "lucide-react"
import { Skeleton } from "@/components/ui/skeleton"
import { useGetApiV1Orders } from "@/api/endpoints/orders/orders"

export function OrdersTable() {
  const [search, setSearch] = useQueryState('search', { defaultValue: '' })
  const [statusPipeline, setStatusPipeline] = useQueryState('status', { defaultValue: 'all' })

  const { data: apiOrders, isError, isLoading } = useGetApiV1Orders()
  const rawOrders = (apiOrders as any)?.data || (apiOrders as any) || []
  const ordersToUse: any[] = Array.isArray(rawOrders) ? rawOrders : []

  const counts = React.useMemo(() => {
    return {
      all: ordersToUse.length,
      pending: ordersToUse.filter(o => (o.status || '').toLowerCase() === 'pending').length,
      paid: ordersToUse.filter(o => (o.status || '').toLowerCase() === 'paid').length,
      processing: ordersToUse.filter(o => (o.status || '').toLowerCase() === 'processing').length,
      shipped: ordersToUse.filter(o => (o.status || '').toLowerCase() === 'shipped').length,
      completed: ordersToUse.filter(o => (o.status || '').toLowerCase() === 'completed').length,
      refunded: ordersToUse.filter(o => (o.status || '').toLowerCase() === 'refunded').length,
    }
  }, [ordersToUse])

  const filteredOrders = React.useMemo(() => {
    return ordersToUse.filter((o: any) => {
      const orderNumStr = String(o.order_number ?? o.id ?? '').toLowerCase()
      const custNameStr = String(o.customer_name ?? '').toLowerCase()
      const q = search.toLowerCase().trim()
      const matchesSearch = !q || orderNumStr.includes(q) || custNameStr.includes(q)

      let matchesPipeline = true
      const s = (o.status || 'pending').toLowerCase()
      if (statusPipeline !== 'all') {
        matchesPipeline = s === statusPipeline
      }

      return matchesSearch && matchesPipeline
    })
  }, [ordersToUse, search, statusPipeline])

  const getStatusBadge = (status: string) => {
    const s = (status || 'pending').toLowerCase()
    switch (s) {
      case 'paid':
        return (
          <span className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <span className="size-1.5 rounded-full bg-emerald-400" />
            Paid
          </span>
        )
      case 'shipped':
        return (
          <span className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
            <Truck className="size-3" />
            Shipped
          </span>
        )
      case 'completed':
        return (
          <span className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium bg-purple-500/10 text-purple-400 border border-purple-500/20">
            <CheckCircle2 className="size-3" />
            Completed
          </span>
        )
      case 'processing':
        return (
          <span className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium bg-sky-500/10 text-sky-400 border border-sky-500/20">
            <Package className="size-3" />
            Processing
          </span>
        )
      case 'refunded':
        return (
          <span className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium bg-zinc-500/10 text-zinc-400 border border-zinc-500/20">
            <RefreshCw className="size-3" />
            Refunded
          </span>
        )
      default:
        return (
          <span className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <Clock className="size-3" />
            Pending
          </span>
        )
    }
  }

  const formatPrice = (price: number | string | undefined, currency: string | undefined) => {
    const num = Number(price || 0)
    return `EGP ${num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  }

  return (
    <div className="flex flex-col gap-6 animate-in fade-in duration-500">
      <div className="flex items-center gap-2 border-b border-white/[0.05] pb-2 overflow-x-auto">
        <button
          onClick={() => setStatusPipeline('all')}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
            statusPipeline === 'all'
              ? 'bg-zinc-800 text-zinc-100 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
          }`}
        >
          <span>All Orders</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.all}
          </span>
        </button>
        <button
          onClick={() => setStatusPipeline('pending')}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
            statusPipeline === 'pending'
              ? 'bg-amber-500/15 text-amber-400 border border-amber-500/20 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
          }`}
        >
          <Clock className="size-3 text-amber-400" />
          <span>Pending</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.pending}
          </span>
        </button>
        <button
          onClick={() => setStatusPipeline('paid')}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
            statusPipeline === 'paid'
              ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
          }`}
        >
          <span className="size-1.5 rounded-full bg-emerald-400" />
          <span>Paid</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.paid}
          </span>
        </button>
        <button
          onClick={() => setStatusPipeline('processing')}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
            statusPipeline === 'processing'
              ? 'bg-sky-500/15 text-sky-400 border border-sky-500/20 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
          }`}
        >
          <Package className="size-3 text-sky-400" />
          <span>Processing</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.processing}
          </span>
        </button>
        <button
          onClick={() => setStatusPipeline('shipped')}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
            statusPipeline === 'shipped'
              ? 'bg-blue-500/15 text-blue-400 border border-blue-500/20 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
          }`}
        >
          <Truck className="size-3 text-blue-400" />
          <span>Shipped</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.shipped}
          </span>
        </button>
        <button
          onClick={() => setStatusPipeline('completed')}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
            statusPipeline === 'completed'
              ? 'bg-purple-500/15 text-purple-400 border border-purple-500/20 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
          }`}
        >
          <CheckCircle2 className="size-3 text-purple-400" />
          <span>Completed</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.completed}
          </span>
        </button>
        <button
          onClick={() => setStatusPipeline('refunded')}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
            statusPipeline === 'refunded'
              ? 'bg-zinc-800 text-zinc-200 shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
          }`}
        >
          <RefreshCw className="size-3" />
          <span>Refunded</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.refunded}
          </span>
        </button>
      </div>
      <div className="flex items-center justify-between">
        <div className="relative group">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500 transition-colors group-focus-within:text-zinc-300" />
          <input 
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by order # or customer name..."
            className="h-9 pl-9 pr-4 bg-zinc-950 border border-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-zinc-700 focus:ring-1 focus:ring-zinc-700 w-80 rounded-xl text-xs transition-all"
          />
        </div>
      </div>

      <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
        <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] overflow-hidden">
          {isLoading ? (
            <div className="p-12 text-center text-zinc-500 text-xs">Loading orders...</div>
          ) : filteredOrders.length === 0 ? (
            <div className="p-16 text-center flex flex-col items-center justify-center">
              <h3 className="text-base font-semibold text-zinc-200 tracking-tight">No orders found</h3>
              <p className="text-xs mt-1 text-zinc-500">
                No orders match your search query or pipeline filter.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader className="border-b border-white/[0.05] bg-zinc-950/60">
                <TableRow className="border-none hover:bg-transparent">
                  <TableHead className="font-medium text-zinc-400 text-xs px-6 py-3.5">Order #</TableHead>
                  <TableHead className="font-medium text-zinc-400 text-xs px-6 py-3.5">Date</TableHead>
                  <TableHead className="font-medium text-zinc-400 text-xs px-6 py-3.5">Customer</TableHead>
                  <TableHead className="font-medium text-zinc-400 text-xs px-6 py-3.5">Pipeline Status</TableHead>
                  <TableHead className="font-medium text-zinc-400 text-xs px-6 py-3.5">Payment</TableHead>
                  <TableHead className="text-right font-medium text-zinc-400 text-xs px-6 py-3.5">Total</TableHead>
                  <TableHead className="w-[60px] px-6 py-3.5"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredOrders.map((order) => {
                  const createdAt = order.created_at ? new Date(order.created_at) : new Date()
                  const customerName = order.customer_name || 'Guest Checkout'
                  const customerEmail = order.customer_email || ''

                  return (
                    <TableRow key={order.id} className="border-b border-white/[0.04] last:border-0 hover:bg-white/[0.02] transition-colors group">
                      <TableCell className="px-6 py-3.5 font-mono text-zinc-100 font-semibold text-xs">
                        #{order.order_number || String(order.id).substring(0, 8)}
                      </TableCell>
                      <TableCell className="px-6 py-3.5 text-zinc-400 font-mono text-xs whitespace-nowrap">
                        {new Intl.DateTimeFormat('en-US', {
                          month: 'short',
                          day: 'numeric',
                          hour: 'numeric',
                          minute: '2-digit'
                        }).format(createdAt)}
                      </TableCell>
                      <TableCell className="px-6 py-3.5 text-xs">
                        <div className="flex flex-col">
                          <span className="text-zinc-200 font-medium">{customerName}</span>
                          {customerEmail && (
                            <span className="text-zinc-500 text-[10px]">{customerEmail}</span>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="px-6 py-3.5">
                        {getStatusBadge(order.status)}
                      </TableCell>
                      <TableCell className="px-6 py-3.5 text-xs text-zinc-300 capitalize">
                        {order.payment_method || order.payment_status || 'Paid'}
                      </TableCell>
                      <TableCell className="px-6 py-3.5 text-right font-mono text-xs text-zinc-100 font-medium">
                        {formatPrice(order.total_price || order.total_amount, order.currency)}
                      </TableCell>
                      <TableCell className="px-6 py-3.5 text-right">
                        <Link 
                          href={`/admin/orders/${order.id}`}
                          className="inline-flex h-7 w-7 items-center justify-center rounded-lg bg-zinc-800 text-zinc-400 hover:bg-zinc-700 hover:text-zinc-100 transition-all opacity-80 group-hover:opacity-100"
                        >
                          <ChevronRight className="h-3.5 w-3.5" />
                        </Link>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          )}
        </div>
      </div>
    </div>
  )
}
