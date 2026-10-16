"use client"

import * as React from "react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import { CheckCircle2, ArrowRight, Package, User, Truck, ShieldCheck, Mail } from "lucide-react"

export default function OrderConfirmationPage() {
  const searchParams = useSearchParams()
  const orderId = searchParams.get("order_id") || "ord-confirmed"

  return (
    <div className="min-h-[75vh] flex flex-col items-center justify-center py-16 px-4 animate-in fade-in duration-500">
      <div className="max-w-lg w-full bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-white/[0.08] rounded-3xl p-8 sm:p-10 text-center shadow-xl">
        <div className="size-16 bg-emerald-500/10 border border-emerald-500/20 rounded-3xl flex items-center justify-center mx-auto mb-6 text-emerald-500">
          <CheckCircle2 className="size-8 stroke-[2]" />
        </div>

        <span className="text-xs font-mono uppercase tracking-widest text-emerald-500 font-semibold">
          Order Verified & Confirmed
        </span>

        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50 mt-1 mb-3">
          Thank you for your order!
        </h1>
        <p className="text-xs sm:text-sm text-zinc-500 dark:text-zinc-400 mb-6 leading-relaxed">
          Your order has been logged into our Cairo fulfillment queue. We have dispatched a confirmation email and SMS tracking notifications.
        </p>

        <div className="bg-zinc-50 dark:bg-zinc-950/60 rounded-2xl p-5 mb-6 border border-zinc-200 dark:border-white/[0.06] text-left flex flex-col gap-3">
          <div className="flex justify-between items-center text-xs text-zinc-500">
            <span>Order Reference</span>
            <span>Fulfillment Status</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100 truncate max-w-[220px]">
              #{orderId}
            </span>
            <span className="rounded-full bg-emerald-500/10 border border-emerald-500/20 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-500 flex items-center gap-1">
              <span className="size-1.5 rounded-full bg-emerald-500 animate-pulse" />
              Confirmed
            </span>
          </div>
          <div className="pt-2 border-t border-zinc-200 dark:border-white/[0.04] flex items-center justify-between text-[11px] text-zinc-500">
            <span className="flex items-center gap-1">
              <Truck className="size-3 text-zinc-400" />
              Estimated Delivery:
            </span>
            <span className="font-medium text-zinc-700 dark:text-zinc-300">
              Within 24-48 Hours (Greater Cairo)
            </span>
          </div>
        </div>

        <div className="flex flex-col gap-3">
          <Link
            href="/account"
            className="w-full flex items-center justify-center gap-2 h-11 rounded-xl bg-zinc-900 dark:bg-zinc-100 text-zinc-50 dark:text-zinc-900 font-semibold text-xs hover:opacity-90 transition-opacity shadow-sm"
          >
            <User className="size-3.5" />
            <span>View Order History</span>
          </Link>
          <Link
            href="/"
            className="w-full flex items-center justify-center gap-2 h-11 rounded-xl border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-300 font-medium text-xs hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors"
          >
            <span>Continue Shopping</span>
            <ArrowRight className="size-3.5" />
          </Link>
        </div>
      </div>
    </div>
  )
}
