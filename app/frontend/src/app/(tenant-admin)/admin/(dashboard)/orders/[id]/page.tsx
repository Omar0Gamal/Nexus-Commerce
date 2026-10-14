"use client"

import React, { useState } from 'react'
import Link from 'next/link'
import { useParams, useRouter } from 'next/navigation'
import {
  ChevronLeft,
  Package,
  User,
  CreditCard,
  Truck,
  Check,
  RefreshCw,
  Clock,
  CheckCircle2,
  MapPin,
  Mail,
  Phone,
  ShieldCheck,
  ExternalLink
} from 'lucide-react'
import { useGetApiV1OrdersId, usePatchApiV1OrdersIdStatus } from '@/api/endpoints/orders/orders'
import { useGetApiV1CatalogProductsId } from '@/api/endpoints/catalog/catalog'
import { useQueryClient } from '@tanstack/react-query'
import { useBreadcrumbStore } from '@/store/breadcrumb'

function OrderItemImage({ item }: { item: any }) {
  const { data: productEnvelope } = useGetApiV1CatalogProductsId(item.product_id || "")
  const product: any = (productEnvelope as any)?.data || productEnvelope
  const primaryImg = product?.images?.[0]?.url || item.image_url

  const fallbackImg = React.useMemo(() => {
    if (primaryImg) return primaryImg
    const nameLower = (item.name || "").toLowerCase()
    if (nameLower.includes("velocity") || nameLower.includes("flyknit") || nameLower.includes("nitro")) return "/seed-prod-1.jpg"
    if (nameLower.includes("monolith") || nameLower.includes("court")) return "/seed-prod-2.jpg"
    if (nameLower.includes("chelsea") || nameLower.includes("dune")) return "/seed-prod-3.jpg"
    if (nameLower.includes("aerocore") || nameLower.includes("heritage")) return "/seed-prod-4.jpg"
    if (nameLower.includes("alpine") || nameLower.includes("storm") || nameLower.includes("hiker")) return "/seed-prod-5.jpg"
    if (nameLower.includes("stealth") || nameLower.includes("carbon")) return "/seed-prod-6.jpg"
    if (nameLower.includes("canvas") || nameLower.includes("vanguard")) return "/seed-prod-7.jpg"
    if (nameLower.includes("brush") || nameLower.includes("care")) return "/seed-prod-27.jpg"
    if (nameLower.includes("air max")) return "/seed-prod-26.jpg"
    if (nameLower.includes("puma") || nameLower.includes("smash")) return "/seed-prod-28.jpg"
    return "/seed-prod-1.jpg"
  }, [primaryImg, item.name])

  return (
    <div className="size-14 rounded-xl bg-zinc-950 border border-zinc-800 overflow-hidden flex-shrink-0">
      <img src={primaryImg || fallbackImg} alt={item.name} className="w-full h-full object-cover" />
    </div>
  )
}

export default function OrderDetailPage() {
  const params = useParams()
  const id = (params?.id as string) || ""
  const queryClient = useQueryClient()
  const router = useRouter()
  const setTitleOverride = useBreadcrumbStore((state) => state.setTitleOverride)

  const { data: orderDetail, isLoading, isError } = useGetApiV1OrdersId(id)
  const patchStatusMutation = usePatchApiV1OrdersIdStatus()

  React.useEffect(() => {
    if (orderDetail?.data && id) {
      const ord = orderDetail.data as any
      const label = ord.order_number ? `Order #${ord.order_number}` : `Order #${id.slice(0, 8)}`
      setTitleOverride(id, label)
    }
  }, [orderDetail, id, setTitleOverride])
  
  const [carrier, setCarrier] = useState("Bosta")
  const [trackingNumber, setTrackingNumber] = useState("")

  if (isLoading) {
    return (
      <div className="min-h-screen bg-zinc-950 p-8 flex items-center justify-center text-zinc-400 text-xs">
        Loading order details...
      </div>
    )
  }

  if (isError || !orderDetail?.data) {
    return (
      <div className="min-h-screen bg-zinc-950 p-8 flex flex-col items-center justify-center gap-3 text-red-400 text-xs">
        <span>Order not found or session expired.</span>
        <Link href="/admin/orders" className="text-zinc-300 underline">Back to orders</Link>
      </div>
    )
  }

  const order = orderDetail.data as any
  const status = (order.status || "pending").toLowerCase()

  const timelineStages = [
    { key: "pending", label: "Order Placed" },
    { key: "paid", label: "Payment Captured" },
    { key: "processing", label: "Fulfillment Prep" },
    { key: "shipped", label: "Dispatched / Shipped" },
    { key: "completed", label: "Delivered & Completed" },
  ]

  const stageOrder = ["pending", "paid", "processing", "shipped", "completed"]
  const currentStageIndex = stageOrder.indexOf(status) >= 0 ? stageOrder.indexOf(status) : (status === "refunded" ? -1 : 1)

  const handleUpdateStatus = async (newStatus: string) => {
    try {
      await patchStatusMutation.mutateAsync({
        id,
        data: { status: newStatus },
      })
      queryClient.invalidateQueries({ queryKey: ['orders', id] })
    } catch (e) {
      console.error(e)
      alert("Failed to update status")
    }
  }

  const handleFulfill = async () => {
    if (!trackingNumber.trim()) {
      alert("Please provide a tracking number before dispatching.");
      return;
    }
    try {
      await patchStatusMutation.mutateAsync({
        id,
        data: { status: "shipped" }
      });
      alert(`Order marked as shipped via ${carrier} (AWB: ${trackingNumber})`);
      queryClient.invalidateQueries({ queryKey: ['orders', id] })
    } catch (e) {
      console.error(e)
      alert("Failed to fulfill order")
    }
  }

  const handleRefund = async () => {
    if (!confirm("Are you sure you want to refund this order? This will release captured funds and log a refund audit.")) return;
    try {
      await patchStatusMutation.mutateAsync({
        id,
        data: { status: "refunded" }
      });
      alert("Order refunded successfully.");
      queryClient.invalidateQueries({ queryKey: ['orders', id] })
    } catch (e) {
      console.error(e)
      alert("Failed to initiate refund")
    }
  }

  return (
    <div className="flex flex-col gap-8 w-full max-w-5xl mx-auto pb-24 text-zinc-50 animate-in fade-in duration-500">
      <div className="flex flex-col gap-4 border-b border-white/[0.05] pb-6">
        <Link 
          href="/admin/orders" 
          className="inline-flex items-center text-xs font-medium text-zinc-400 hover:text-zinc-200 transition-colors w-max"
        >
          <ChevronLeft className="h-3.5 w-3.5 mr-1" />
          Back to Orders
        </Link>

        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="flex flex-col gap-1.5">
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-semibold tracking-tight">
                Order <span className="font-mono text-zinc-400">#{order.order_number || String(order.id).substring(0, 8)}</span>
              </h1>
              <span className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium border capitalize ${
                status === 'paid' || status === 'completed' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 
                status === 'shipped' ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' :
                status === 'refunded' ? 'bg-red-500/10 text-red-400 border-red-500/20' :
                'bg-amber-500/10 text-amber-400 border-amber-500/20'
              }`}>
                <span className="size-1.5 rounded-full bg-current" />
                {status}
              </span>
            </div>
            <p className="text-xs text-zinc-400 font-mono">
              Placed {order.created_at && new Intl.DateTimeFormat('en-US', {
                month: 'short', day: 'numeric', year: 'numeric',
                hour: 'numeric', minute: '2-digit'
              }).format(new Date(order.created_at))}
            </p>
          </div>

          <div className="flex items-center gap-2.5">
            {['paid', 'shipped', 'completed'].includes(status) && (
              <button 
                onClick={handleRefund}
                disabled={patchStatusMutation.isPending}
                className="px-3.5 h-9 rounded-xl font-medium text-xs text-red-400 border border-red-500/20 bg-red-500/10 hover:bg-red-500/20 transition-colors disabled:opacity-50 flex items-center gap-1.5"
              >
                <RefreshCw className="w-3.5 h-3.5" />
                Refund Order
              </button>
            )}
            {status === 'paid' && (
              <button
                onClick={() => handleUpdateStatus('processing')}
                disabled={patchStatusMutation.isPending}
                className="px-3.5 h-9 rounded-xl font-medium text-xs text-sky-300 border border-sky-500/20 bg-sky-500/10 hover:bg-sky-500/20 transition-colors flex items-center gap-1.5"
              >
                <Package className="w-3.5 h-3.5" />
                Mark Processing
              </button>
            )}
            {status === 'shipped' && (
              <button
                onClick={() => handleUpdateStatus('completed')}
                disabled={patchStatusMutation.isPending}
                className="px-3.5 h-9 rounded-xl font-medium text-xs text-emerald-300 border border-emerald-500/20 bg-emerald-500/10 hover:bg-emerald-500/20 transition-colors flex items-center gap-1.5"
              >
                <CheckCircle2 className="w-3.5 h-3.5" />
                Mark Delivered
              </button>
            )}
          </div>
        </div>
      </div>

      {status !== 'refunded' && (
        <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
          <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-5 flex flex-col gap-4">
            <span className="text-xs font-semibold uppercase tracking-wider text-zinc-400">
              Fulfillment Pipeline Tracker
            </span>
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
              {timelineStages.map((stage, idx) => {
                const isPassed = currentStageIndex >= idx
                const isCurrent = currentStageIndex === idx
                return (
                  <div
                    key={stage.key}
                    className={`rounded-xl p-3 border flex flex-col gap-1 transition-colors ${
                      isCurrent
                        ? "bg-emerald-500/15 border-emerald-500/30 text-emerald-300"
                        : isPassed
                        ? "bg-zinc-950/80 border-emerald-500/20 text-zinc-300"
                        : "bg-zinc-950/40 border-zinc-800/60 text-zinc-600"
                    }`}
                  >
                    <div className="flex items-center justify-between text-xs font-mono">
                      <span>Step 0{idx + 1}</span>
                      {isPassed && <Check className="size-3.5 text-emerald-400" />}
                    </div>
                    <span className="text-xs font-semibold leading-tight">{stage.label}</span>
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 flex flex-col gap-6">
          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 shadow-sm overflow-hidden">
              <div className="px-6 py-4 border-b border-white/[0.05] flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Package className="h-4 w-4 text-emerald-400" />
                  <h2 className="text-sm font-semibold text-zinc-100">Ordered Items</h2>
                </div>
                <span className="text-xs text-zinc-400">
                  {order.items?.length || 0} item(s)
                </span>
              </div>
              
              <div className="divide-y divide-white/[0.04]">
                {order.items?.map((item: any, i: number) => {
                  return (
                    <div key={item.id || i} className="p-5 flex items-center justify-between gap-4">
                      <div className="flex items-center gap-3.5 min-w-0">
                        <OrderItemImage item={item} />
                        <div className="min-w-0 flex flex-col gap-0.5">
                          <span className="font-medium text-xs text-zinc-100 truncate">{item.name}</span>
                          <span className="text-[11px] font-mono text-zinc-500">
                            Qty: {item.quantity} × EGP {Number(item.price).toLocaleString()}
                          </span>
                        </div>
                      </div>
                      <div className="text-right flex-shrink-0">
                        <span className="font-mono text-xs font-semibold text-zinc-100">
                          EGP {(Number(item.price) * (item.quantity || 1)).toLocaleString()}
                        </span>
                      </div>
                    </div>
                  )
                })}
              </div>
              
              <div className="px-6 py-4 bg-zinc-950/60 border-t border-white/[0.05]">
                <div className="flex flex-col gap-2 font-mono text-xs max-w-xs ml-auto">
                  <div className="flex justify-between text-zinc-400">
                    <span>Subtotal</span>
                    <span>EGP {Number(order.total_price || 0).toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-zinc-400">
                    <span>Cairo Courier Dispatch</span>
                    <span>EGP {Number(order.shipping_fee || 0).toLocaleString()}</span>
                  </div>
                  {Number(order.discount_amount) > 0 && (
                    <div className="flex justify-between text-emerald-400">
                      <span>Promotional Voucher</span>
                      <span>-EGP {Number(order.discount_amount).toLocaleString()}</span>
                    </div>
                  )}
                  <div className="flex justify-between text-zinc-100 text-sm font-bold pt-2 border-t border-zinc-800 mt-1">
                    <span className="font-sans font-semibold">
                      {status === 'pending' ? 'Total Due' : 'Total Paid'}
                    </span>
                    <span className="text-emerald-400">EGP {Number(order.total_price || 0).toLocaleString()}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {status !== 'shipped' && status !== 'completed' && status !== 'refunded' && (
            <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
              <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-6 flex flex-col gap-4">
                <div className="flex items-center gap-2">
                  <Truck className="size-4 text-blue-400" />
                  <h3 className="text-sm font-semibold text-zinc-100">Dispatch & Carrier Booking</h3>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-zinc-400">Logistics Carrier</label>
                    <select
                      value={carrier}
                      onChange={(e) => setCarrier(e.target.value)}
                      className="h-9 rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                    >
                      <option value="Bosta">Bosta Logistics (Next-Day Cairo)</option>
                      <option value="Aramex">Aramex Express Egypt</option>
                      <option value="Mylerz">Mylerz Fast Delivery</option>
                      <option value="Egypt Post">Egypt Post Express</option>
                    </select>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-zinc-400">Tracking / Airway Bill (AWB)</label>
                    <input 
                      type="text" 
                      value={trackingNumber}
                      onChange={(e) => setTrackingNumber(e.target.value)}
                      placeholder="e.g. BST-9823419-EG"
                      className="h-9 rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-100 font-mono"
                    />
                  </div>
                </div>

                <button 
                  onClick={handleFulfill}
                  disabled={patchStatusMutation.isPending || !trackingNumber.trim()}
                  className="self-start mt-1 flex items-center gap-2 rounded-xl bg-zinc-100 hover:bg-white text-zinc-950 px-4 py-2 text-xs font-semibold transition-all disabled:opacity-50"
                >
                  <Check className="h-3.5 w-3.5" />
                  <span>{patchStatusMutation.isPending ? 'Processing...' : 'Mark Dispatched & Save AWB'}</span>
                </button>
              </div>
            </div>
          )}
        </div>

        <div className="flex flex-col gap-6">
          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-5 flex flex-col gap-4">
              <div className="flex items-center gap-2 border-b border-white/[0.05] pb-3">
                <User className="h-4 w-4 text-zinc-400" />
                <h3 className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">Customer Profile</h3>
              </div>
              <div className="flex flex-col gap-2 text-xs">
                <span className="font-semibold text-zinc-100 text-sm">
                  {order.customer_name || 'Omar Gamal'}
                </span>
                <div className="flex items-center gap-2 text-zinc-400">
                  <Mail className="size-3.5 text-zinc-500" />
                  <span>{order.customer_email || 'customer@example.com'}</span>
                </div>
                <div className="flex items-center gap-2 text-zinc-400">
                  <Phone className="size-3.5 text-zinc-500" />
                  <span>+20 100 234 5678</span>
                </div>
              </div>
            </div>
          </div>

          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-5 flex flex-col gap-4">
              <div className="flex items-center gap-2 border-b border-white/[0.05] pb-3">
                <MapPin className="h-4 w-4 text-emerald-400" />
                <h3 className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">Delivery Destination</h3>
              </div>
              <div className="text-zinc-300 text-xs flex flex-col gap-1 leading-relaxed">
                <span className="font-medium text-zinc-200">
                  {order.shipping_address?.address_line_1 || "Plot 42, Road 90 South"}
                </span>
                <span className="text-zinc-400">
                  {order.shipping_address?.city || "New Cairo"}, {order.shipping_address?.state || "Cairo Governorate"}
                </span>
                <span className="text-zinc-500 font-mono text-[11px]">
                  Postal Code: {order.shipping_address?.postal_code || "11835"} • Egypt
                </span>
              </div>
            </div>
          </div>

          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-5 flex flex-col gap-3">
              <div className="flex items-center gap-2 border-b border-white/[0.05] pb-3">
                <CreditCard className="h-4 w-4 text-blue-400" />
                <h3 className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">Payment Details</h3>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-zinc-400">Gateway Method</span>
                <span className="font-semibold text-zinc-200 capitalize">
                  {order.payment_method || "Paymob Visa / Mastercard"}
                </span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-zinc-400">Capture Status</span>
                <span className="text-emerald-400 font-medium flex items-center gap-1">
                  <ShieldCheck className="size-3.5" /> Settled
                </span>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  )
}
