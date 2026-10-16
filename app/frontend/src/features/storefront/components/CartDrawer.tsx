"use client"

import * as React from "react"
import { ShoppingBag, X, Minus, Plus, Trash2, Tag, Truck, Check, ArrowRight } from "lucide-react"
import Image from "next/image"
import Link from "next/link"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { useCartStore } from "@/store/cart"

export function CartDrawer() {
  const [open, setOpen] = React.useState(false)
  const [promoCode, setPromoCode] = React.useState("")
  const [appliedDiscountPercent, setAppliedDiscountPercent] = React.useState(0)
  const [promoError, setPromoError] = React.useState("")

  const items = useCartStore((state) => state.items)
  const updateQuantity = useCartStore((state) => state.updateQuantity)
  const removeItem = useCartStore((state) => state.removeItem)
  const fetchCart = useCartStore((state) => state.fetchCart)
  const isInitialized = useCartStore((state) => state.isInitialized)

  React.useEffect(() => {
    if (!isInitialized) {
      fetchCart()
    }
  }, [isInitialized, fetchCart])

  const onOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    if (isOpen) {
      fetchCart()
    }
  }

  const totalCount = items.reduce((acc, item) => acc + item.quantity, 0)
  const subtotal = items.reduce((acc, item) => acc + item.price * item.quantity, 0)
  const freeShippingThreshold = 3000
  const progressToFreeShipping = Math.min(100, Math.round((subtotal / freeShippingThreshold) * 100))
  const remainingForFreeShipping = Math.max(0, freeShippingThreshold - subtotal)

  const discountAmount = Math.round((subtotal * appliedDiscountPercent) / 100)
  const finalTotal = subtotal - discountAmount

  const handleApplyPromo = () => {
    setPromoError("")
    const code = promoCode.trim().toUpperCase()
    if (code === "WELCOME10") {
      setAppliedDiscountPercent(10)
    } else if (code === "CAIRO20") {
      setAppliedDiscountPercent(20)
    } else if (code === "FREESHIP") {
      setAppliedDiscountPercent(5)
    } else {
      setPromoError("Invalid promotional code")
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetTrigger className="flex items-center gap-2 text-xs font-semibold text-zinc-900 dark:text-zinc-100 bg-zinc-100 dark:bg-white/[0.05] px-3.5 py-2 rounded-full hover:bg-zinc-200 dark:hover:bg-white/[0.1] transition-colors cursor-pointer border border-zinc-200 dark:border-white/10">
        <ShoppingBag className="w-3.5 h-3.5" />
        <span>Bag ({totalCount})</span>
      </SheetTrigger>
      <SheetContent className="w-full sm:max-w-md border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-950 p-0 flex flex-col">
        <SheetHeader className="p-5 border-b border-zinc-200 dark:border-white/[0.05]">
          <div className="flex items-center justify-between">
            <SheetTitle className="text-lg font-bold text-zinc-900 dark:text-zinc-50 tracking-tight flex items-center gap-2">
              <ShoppingBag className="size-4 text-emerald-500" />
              <span>Shopping Bag ({totalCount})</span>
            </SheetTitle>
          </div>
        </SheetHeader>

        {items.length > 0 && (
          <div className="px-5 py-3 bg-zinc-50 dark:bg-zinc-900/60 border-b border-zinc-200 dark:border-white/[0.05] text-xs">
            <div className="flex items-center justify-between mb-1.5 font-medium">
              <span className="flex items-center gap-1.5 text-zinc-700 dark:text-zinc-300">
                <Truck className="size-3.5 text-emerald-500" />
                {remainingForFreeShipping > 0 ? (
                  <span>Add <strong className="text-zinc-900 dark:text-white font-mono">EGP {remainingForFreeShipping.toLocaleString()}</strong> for Free Cairo Express Shipping</span>
                ) : (
                  <span className="text-emerald-500 font-semibold">🎉 You have unlocked Free Next-Day Cairo Shipping!</span>
                )}
              </span>
              <span className="font-mono text-zinc-500">{progressToFreeShipping}%</span>
            </div>
            <div className="w-full h-1.5 bg-zinc-200 dark:bg-zinc-800 rounded-full overflow-hidden">
              <div
                className="h-full bg-emerald-500 rounded-full transition-all duration-500"
                style={{ width: `${progressToFreeShipping}%` }}
              />
            </div>
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-5 flex flex-col gap-4">
          {items.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center py-20 gap-3 text-zinc-500">
              <div className="size-16 rounded-3xl bg-zinc-100 dark:bg-zinc-900 flex items-center justify-center">
                <ShoppingBag className="size-8 opacity-40 stroke-[1.5]" />
              </div>
              <p className="font-semibold text-zinc-800 dark:text-zinc-200 text-sm">Your shopping bag is empty</p>
              <p className="text-xs max-w-xs text-zinc-500">Discover authenticated sneakers and add your size to begin checkout.</p>
            </div>
          ) : (
            items.map((item) => (
              <div key={`${item.id}-${item.variantId || 'base'}`} className="flex gap-3.5 items-center border-b border-zinc-100 dark:border-zinc-800/60 pb-4 last:border-b-0">
                <div className="w-16 h-16 rounded-xl bg-zinc-100 dark:bg-zinc-900 overflow-hidden relative border border-zinc-200 dark:border-white/[0.05] shrink-0">
                  {item.image ? (
                    <Image src={item.image} alt={item.title} fill className="object-cover" />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-[10px] text-zinc-400">No Photo</div>
                  )}
                </div>
                <div className="flex-1 min-w-0 flex flex-col gap-0.5">
                  <div className="flex items-start justify-between gap-2">
                    <h3 className="font-medium text-xs text-zinc-900 dark:text-zinc-100 truncate">{item.title}</h3>
                    <button
                      onClick={() => removeItem(item.id, item.variantId)}
                      className="text-zinc-400 hover:text-red-500 transition-colors p-0.5"
                      title="Remove item"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                  <span className="text-xs font-mono font-semibold text-zinc-900 dark:text-zinc-200">
                    EGP {Number(item.price).toLocaleString()}
                  </span>
                  <div className="flex items-center gap-2 mt-1.5">
                    <button
                      onClick={() => {
                        if (item.quantity > 1) updateQuantity(item.id, item.quantity - 1, item.variantId)
                        else removeItem(item.id, item.variantId)
                      }}
                      className="size-6 flex items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-300 hover:bg-zinc-200 transition-colors"
                    >
                      <Minus className="w-3 h-3" />
                    </button>
                    <span className="text-xs font-mono font-bold w-4 text-center">{item.quantity}</span>
                    <button
                      onClick={() => updateQuantity(item.id, item.quantity + 1, item.variantId)}
                      className="size-6 flex items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-300 hover:bg-zinc-200 transition-colors"
                    >
                      <Plus className="w-3 h-3" />
                    </button>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>

        {items.length > 0 && (
          <div className="p-5 border-t border-zinc-200 dark:border-white/[0.05] bg-zinc-50 dark:bg-zinc-900/50 flex flex-col gap-3.5">
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <Tag className="absolute left-3 top-1/2 -translate-y-1/2 size-3 text-zinc-400" />
                <input
                  type="text"
                  value={promoCode}
                  onChange={(e) => setPromoCode(e.target.value)}
                  placeholder="Promo (e.g. WELCOME10)"
                  className="w-full h-8 pl-8 pr-3 rounded-lg bg-white dark:bg-zinc-950 border border-zinc-200 dark:border-zinc-800 text-xs font-mono uppercase text-zinc-900 dark:text-zinc-100 placeholder:normal-case placeholder:text-zinc-400 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                />
              </div>
              <button
                type="button"
                onClick={handleApplyPromo}
                className="h-8 px-3 rounded-lg bg-zinc-900 dark:bg-zinc-100 text-zinc-100 dark:text-zinc-900 text-xs font-semibold hover:opacity-90 transition-opacity"
              >
                Apply
              </button>
            </div>
            {promoError && <p className="text-[10px] text-red-400">{promoError}</p>}
            {appliedDiscountPercent > 0 && (
              <p className="text-[10px] text-emerald-400 font-medium">
                ✓ Coupon applied: {appliedDiscountPercent}% discount applied!
              </p>
            )}

            <div className="flex flex-col gap-1.5 pt-1 text-xs">
              <div className="flex justify-between text-zinc-500">
                <span>Subtotal</span>
                <span className="font-mono">EGP {subtotal.toLocaleString()}</span>
              </div>
              {discountAmount > 0 && (
                <div className="flex justify-between text-emerald-500">
                  <span>Promotional Discount</span>
                  <span className="font-mono">-EGP {discountAmount.toLocaleString()}</span>
                </div>
              )}
              <div className="flex justify-between text-sm font-bold text-zinc-900 dark:text-zinc-50 pt-1 border-t border-zinc-200 dark:border-white/[0.05]">
                <span>Total Amount</span>
                <span className="font-mono text-emerald-500">EGP {finalTotal.toLocaleString()}</span>
              </div>
            </div>

            <Link
              href="/checkout"
              onClick={() => setOpen(false)}
              className="w-full flex items-center justify-center gap-2 h-12 rounded-xl bg-zinc-900 dark:bg-zinc-100 text-zinc-50 dark:text-zinc-900 font-semibold text-xs shadow-md hover:scale-[1.01] active:scale-[0.99] transition-all cursor-pointer"
            >
              <span>Proceed to Checkout</span>
              <ArrowRight className="size-3.5" />
            </Link>
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
