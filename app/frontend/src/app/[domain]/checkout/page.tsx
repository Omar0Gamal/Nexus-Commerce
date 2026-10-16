"use client"

import * as React from "react"
import Link from "next/link"
import Image from "next/image"
import { useParams, useRouter } from "next/navigation"
import {
  Lock,
  CreditCard,
  Loader2,
  ArrowLeft,
  AlertCircle,
  ShoppingBag,
  CheckCircle2,
  Banknote,
  ShieldCheck,
  Truck
} from "lucide-react"
import { useCartStore } from "@/store/cart"
import { customInstance } from "@/lib/axios"
import { Alert, AlertDescription } from "@/components/ui/alert"

export default function CheckoutPage() {
  const router = useRouter()
  const params = useParams()
  const domain = (params?.domain as string) || ""

  const items = useCartStore((state) => state.items)
  const clearCart = useCartStore((state) => state.clearCart)
  const fetchCart = useCartStore((state) => state.fetchCart)
  const isInitialized = useCartStore((state) => state.isInitialized)

  const [shopName, setShopName] = React.useState(domain)

  const [email, setEmail] = React.useState("customer@example.com")
  const [firstName, setFirstName] = React.useState("Omar")
  const [lastName, setLastName] = React.useState("Gamal")
  const [address1, setAddress1] = React.useState("Plot 42, Road 90 South, 5th Settlement")
  const [address2, setAddress2] = React.useState("Apartment 3B")
  const [city, setCity] = React.useState("New Cairo")
  const [governorate, setGovernorate] = React.useState("Cairo")
  const [country, setCountry] = React.useState("Egypt")
  const [zipCode, setZipCode] = React.useState("11835")
  const [phone, setPhone] = React.useState("+20 100 234 5678")

  const [paymentMethod, setPaymentMethod] = React.useState<"test" | "cod" | "paymob">("test")

  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [errorMessage, setErrorMessage] = React.useState<string | null>(null)

  React.useEffect(() => {
    if (!isInitialized) {
      fetchCart()
    }

    if (typeof window !== "undefined") {
      const storedShopName = localStorage.getItem("nexus_shop_name")
      if (storedShopName) setShopName(storedShopName)
    }
  }, [isInitialized, fetchCart])

  const subtotal = items.reduce((acc, item) => acc + item.price * item.quantity, 0)
  const shippingFee = subtotal >= 3000 || subtotal === 0 ? 0 : 50
  const grandTotal = subtotal + shippingFee

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setErrorMessage(null)

    if (items.length === 0) {
      setErrorMessage("Your cart is empty. Please add items before checking out.")
      return
    }

    if (!email.trim() || !firstName.trim() || !lastName.trim() || !address1.trim() || !city.trim()) {
      setErrorMessage("Please complete all required shipping fields.")
      return
    }

    setIsSubmitting(true)

    try {
      let cartToken = typeof window !== "undefined" ? localStorage.getItem("nexus_cart_token") : null
      if (!cartToken) {
        cartToken =
          typeof crypto !== "undefined" && crypto.randomUUID
            ? crypto.randomUUID()
            : "cart-" + Math.random().toString(36).substring(2) + Date.now().toString(36)
        if (typeof window !== "undefined") {
          localStorage.setItem("nexus_cart_token", cartToken)
        }
      }

      const intentResponse = await customInstance<{
        status: number
        data: { intent_id: string; payment_url: string; order_id?: string }
      }>({
        url: "/api/v1/checkout/intent",
        method: "POST",
        data: {
          cart_token: cartToken,
          email: email.trim(),
          payment_method: paymentMethod,
          shipping_address: {
            first_name: firstName.trim(),
            last_name: lastName.trim(),
            address1: address1.trim(),
            address2: address2.trim() || undefined,
            city: city.trim(),
            state: governorate.trim(),
            country: country.trim() || "Egypt",
            postal_code: zipCode.trim() || "11835",
            phone: phone.trim() || "+20 100 234 5678",
          },
        },
      })

      await clearCart()

      const paymentUrl = intentResponse?.data?.payment_url
      if (paymentUrl) {
        if (paymentUrl.startsWith("http://") || paymentUrl.startsWith("https://")) {
          window.location.href = paymentUrl
        } else {
          router.push(paymentUrl)
        }
      } else {
        const orderRef = intentResponse?.data?.order_id || intentResponse?.data?.intent_id || "confirmed"
        router.push(`/order-confirmation?order_id=${orderRef}`)
      }
    } catch (err: any) {
      console.error("Checkout submission failed:", err)
      const msg =
        err?.response?.data?.message ||
        err?.response?.data?.error ||
        "Payment processing failed. Please verify your address and retry."
      setErrorMessage(msg)
      setIsSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col lg:flex-row bg-zinc-50 dark:bg-zinc-950 font-sans selection:bg-zinc-900 selection:text-white pb-20">
      <div className="flex-1 flex flex-col pt-8 pb-16 px-6 sm:px-10 lg:px-16 max-w-3xl">
        <header className="mb-8">
          <Link
            href="/"
            className="text-lg font-bold tracking-tight text-zinc-900 dark:text-zinc-50 flex items-center gap-2 hover:opacity-80 transition-opacity"
          >
            <ArrowLeft className="w-4 h-4 text-zinc-400" />
            <span>{shopName || "Kicks & Co. Cairo"}</span>
          </Link>

          <div className="flex items-center gap-2 mt-3 text-xs font-medium text-zinc-500">
            <span>Bag</span>
            <span>&rsaquo;</span>
            <span className="text-zinc-900 dark:text-zinc-100 font-semibold">Delivery & Billing</span>
            <span>&rsaquo;</span>
            <span>Confirmation</span>
          </div>
        </header>

        {errorMessage && (
          <Alert variant="destructive" className="mb-6 rounded-2xl bg-red-500/10 border-red-500/20 text-red-400">
            <AlertCircle className="w-4 h-4 mr-2" />
            <AlertDescription className="text-xs">{errorMessage}</AlertDescription>
          </Alert>
        )}

        {items.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center bg-white dark:bg-zinc-900 rounded-3xl border border-zinc-200 dark:border-zinc-800">
            <ShoppingBag className="w-12 h-12 text-zinc-400 mb-3" />
            <h2 className="text-base font-semibold text-zinc-900 dark:text-zinc-100">Your bag is empty</h2>
            <p className="text-xs text-zinc-500 mt-1 mb-6">Select a pair of shoes to proceed with checkout.</p>
            <Link
              href="/"
              className="inline-flex h-10 items-center justify-center rounded-xl bg-zinc-900 dark:bg-zinc-100 px-5 text-xs font-semibold text-zinc-50 dark:text-zinc-900 shadow hover:opacity-90 transition-opacity"
            >
              Return to Catalog
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="flex flex-col gap-8">
            <section className="flex flex-col gap-3">
              <h2 className="text-sm font-bold uppercase tracking-wider text-zinc-900 dark:text-zinc-100">
                1. Contact Information
              </h2>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">Email Address (for serialized receipt)</label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="omar.gamal@example.com"
                  className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3.5 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                />
              </div>
            </section>

            <section className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-bold uppercase tracking-wider text-zinc-900 dark:text-zinc-100">
                  2. Egyptian Delivery Destination
                </h2>
                <span className="text-[11px] text-emerald-500 font-medium flex items-center gap-1">
                  <Truck className="size-3" /> Same-day dispatch available
                </span>
              </div>

              <div className="flex flex-col gap-3">
                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1">
                    <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">First Name</label>
                    <input
                      type="text"
                      required
                      value={firstName}
                      onChange={(e) => setFirstName(e.target.value)}
                      className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3.5 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                    />
                  </div>
                  <div className="flex flex-col gap-1">
                    <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">Last Name</label>
                    <input
                      type="text"
                      required
                      value={lastName}
                      onChange={(e) => setLastName(e.target.value)}
                      className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3.5 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                    />
                  </div>
                </div>

                <div className="flex flex-col gap-1">
                  <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">Street Address & Building</label>
                  <input
                    type="text"
                    required
                    value={address1}
                    onChange={(e) => setAddress1(e.target.value)}
                    placeholder="Plot 42, Road 90 South"
                    className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3.5 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                  />
                </div>

                <div className="flex flex-col gap-1">
                  <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">Apartment / Floor / Landmark (optional)</label>
                  <input
                    type="text"
                    value={address2}
                    onChange={(e) => setAddress2(e.target.value)}
                    placeholder="Building 12, Floor 3, Apt 3B"
                    className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3.5 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                  />
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  <div className="flex flex-col gap-1">
                    <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">City / District</label>
                    <input
                      type="text"
                      required
                      value={city}
                      onChange={(e) => setCity(e.target.value)}
                      placeholder="New Cairo"
                      className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3.5 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                    />
                  </div>
                  <div className="flex flex-col gap-1">
                    <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">Governorate</label>
                    <select
                      value={governorate}
                      onChange={(e) => setGovernorate(e.target.value)}
                      className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3 text-xs text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                    >
                      <option value="Cairo">Cairo (القاهرة)</option>
                      <option value="Giza">Giza (الجيزة)</option>
                      <option value="Alexandria">Alexandria (الإسكندرية)</option>
                      <option value="Qalyubia">Qalyubia (القليوبية)</option>
                      <option value="Sharqia">Sharqia (الشرقية)</option>
                      <option value="Red Sea">Red Sea (البحر الأحمر)</option>
                    </select>
                  </div>
                  <div className="flex flex-col gap-1">
                    <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">Postal Code</label>
                    <input
                      type="text"
                      value={zipCode}
                      onChange={(e) => setZipCode(e.target.value)}
                      placeholder="11835"
                      className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3.5 text-xs text-zinc-900 dark:text-zinc-100 font-mono focus:outline-none focus:ring-1 focus:ring-zinc-700"
                    />
                  </div>
                </div>

                <div className="flex flex-col gap-1">
                  <label className="text-xs font-medium text-zinc-600 dark:text-zinc-400">Mobile Phone (for courier SMS updates)</label>
                  <input
                    type="tel"
                    required
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    placeholder="+20 100 234 5678"
                    className="h-10 w-full rounded-xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 px-3.5 text-xs text-zinc-900 dark:text-zinc-100 font-mono focus:outline-none focus:ring-1 focus:ring-zinc-700"
                  />
                </div>
              </div>
            </section>

            <section className="flex flex-col gap-3">
              <h2 className="text-sm font-bold uppercase tracking-wider text-zinc-900 dark:text-zinc-100">
                3. Payment Method
              </h2>

              <div className="flex flex-col gap-2.5">
                <label
                  onClick={() => setPaymentMethod("test")}
                  className={`p-4 rounded-2xl border transition-all cursor-pointer flex items-start gap-3.5 ${
                    paymentMethod === "test"
                      ? "border-emerald-500 bg-emerald-500/10 dark:bg-emerald-500/10 shadow-sm"
                      : "border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 hover:border-zinc-300"
                  }`}
                >
                  <input
                    type="radio"
                    name="payment"
                    checked={paymentMethod === "test"}
                    onChange={() => setPaymentMethod("test")}
                    className="mt-1 accent-emerald-500"
                  />
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-1.5">
                        <CreditCard className="size-3.5 text-emerald-500" />
                        Instant Test Card (Portfolio Demo 1-Click Approval)
                      </span>
                      <span className="rounded-full bg-emerald-500/20 px-2 py-0.5 text-[10px] font-bold text-emerald-400 font-mono">
                        Instant
                      </span>
                    </div>
                    <p className="text-[11px] text-zinc-500 mt-1 leading-relaxed">
                      Simulates immediate Visa/Mastercard payment settlement using test card <code className="font-mono text-zinc-400">•••• 4242</code> without external gateway redirections.
                    </p>
                  </div>
                </label>

                <label
                  onClick={() => setPaymentMethod("cod")}
                  className={`p-4 rounded-2xl border transition-all cursor-pointer flex items-start gap-3.5 ${
                    paymentMethod === "cod"
                      ? "border-emerald-500 bg-emerald-500/10 dark:bg-emerald-500/10 shadow-sm"
                      : "border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 hover:border-zinc-300"
                  }`}
                >
                  <input
                    type="radio"
                    name="payment"
                    checked={paymentMethod === "cod"}
                    onChange={() => setPaymentMethod("cod")}
                    className="mt-1 accent-emerald-500"
                  />
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-1.5">
                        <Banknote className="size-3.5 text-amber-500" />
                        Cash on Delivery (الدفع عند الاستلام)
                      </span>
                      <span className="rounded-full bg-amber-500/20 px-2 py-0.5 text-[10px] font-bold text-amber-400">
                        COD
                      </span>
                    </div>
                    <p className="text-[11px] text-zinc-500 mt-1 leading-relaxed">
                      Inspect your shoes upon arrival with the Cairo courier. Pay in cash or via mobile POS machine upon receipt.
                    </p>
                  </div>
                </label>

                <label
                  onClick={() => setPaymentMethod("paymob")}
                  className={`p-4 rounded-2xl border transition-all cursor-pointer flex items-start gap-3.5 ${
                    paymentMethod === "paymob"
                      ? "border-emerald-500 bg-emerald-500/10 dark:bg-emerald-500/10 shadow-sm"
                      : "border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 hover:border-zinc-300"
                  }`}
                >
                  <input
                    type="radio"
                    name="payment"
                    checked={paymentMethod === "paymob"}
                    onChange={() => setPaymentMethod("paymob")}
                    className="mt-1 accent-emerald-500"
                  />
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-1.5">
                        <ShieldCheck className="size-3.5 text-blue-500" />
                        Paymob Gateway (Cards, Meeza & Mobile Wallets)
                      </span>
                      <span className="rounded-full bg-blue-500/20 px-2 py-0.5 text-[10px] font-bold text-blue-400">
                        Paymob
                      </span>
                    </div>
                    <p className="text-[11px] text-zinc-500 mt-1 leading-relaxed">
                      Redirects to 3D-Secure hosted checkout window for Vodafone Cash, InstaPay, Meeza, and debit/credit cards.
                    </p>
                  </div>
                </label>
              </div>
            </section>

            <button
              type="submit"
              disabled={isSubmitting || items.length === 0}
              className="h-12 w-full flex items-center justify-center gap-2 rounded-xl bg-zinc-900 dark:bg-zinc-100 text-zinc-50 dark:text-zinc-900 font-semibold text-xs shadow-lg hover:scale-[1.01] active:scale-[0.99] transition-all disabled:opacity-50 cursor-pointer"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span>Processing Order...</span>
                </>
              ) : (
                <>
                  <Lock className="w-3.5 h-3.5 opacity-60" />
                  <span>
                    Confirm & Complete Order — EGP {grandTotal.toLocaleString()}
                  </span>
                </>
              )}
            </button>
          </form>
        )}
      </div>

      <div className="w-full lg:w-[420px] bg-zinc-100/70 dark:bg-zinc-900/40 border-t lg:border-t-0 lg:border-l border-zinc-200 dark:border-white/[0.06] p-6 sm:p-10 flex flex-col gap-6">
        <h2 className="text-sm font-bold uppercase tracking-wider text-zinc-900 dark:text-zinc-100">
          Order Summary ({items.length} items)
        </h2>

        <div className="flex flex-col gap-3.5 divide-y divide-zinc-200 dark:divide-white/[0.05]">
          {items.map((item) => (
            <div key={`${item.id}-${item.variantId || 'base'}`} className="pt-3.5 first:pt-0 flex items-center justify-between gap-3">
              <div className="flex items-center gap-3 min-w-0">
                <div className="relative size-14 rounded-xl overflow-hidden bg-white dark:bg-zinc-950 border border-zinc-200 dark:border-zinc-800 shrink-0">
                  {item.image ? (
                    <Image src={item.image} alt={item.title} fill className="object-cover" />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-[10px] text-zinc-500">No Img</div>
                  )}
                  <span className="absolute bottom-1 right-1 rounded-full bg-zinc-900 text-white font-mono text-[9px] size-4 flex items-center justify-center font-bold">
                    {item.quantity}
                  </span>
                </div>
                <div className="min-w-0">
                  <p className="text-xs font-semibold text-zinc-900 dark:text-zinc-100 truncate">{item.title}</p>
                  <p className="text-[10px] text-zinc-500 font-mono">
                    Qty: {item.quantity} × EGP {Number(item.price).toLocaleString()}
                  </p>
                </div>
              </div>
              <div className="text-right shrink-0">
                <span className="text-xs font-mono font-bold text-zinc-900 dark:text-zinc-100">
                  EGP {(item.price * item.quantity).toLocaleString()}
                </span>
              </div>
            </div>
          ))}
        </div>

        <div className="flex flex-col gap-2 pt-4 border-t border-zinc-200 dark:border-white/[0.08] text-xs">
          <div className="flex justify-between text-zinc-500">
            <span>Subtotal</span>
            <span className="font-mono text-zinc-800 dark:text-zinc-200">EGP {subtotal.toLocaleString()}</span>
          </div>
          <div className="flex justify-between text-zinc-500">
            <span>Cairo Courier Delivery</span>
            <span className="font-mono text-zinc-800 dark:text-zinc-200">
              {shippingFee === 0 ? "FREE" : `EGP ${shippingFee}`}
            </span>
          </div>
          <div className="flex justify-between text-sm font-bold text-zinc-900 dark:text-zinc-50 pt-2 border-t border-zinc-200 dark:border-white/[0.08]">
            <span>Total to Pay</span>
            <span className="font-mono text-emerald-500 text-base">EGP {grandTotal.toLocaleString()}</span>
          </div>
        </div>

        <div className="rounded-2xl bg-white dark:bg-zinc-950/80 border border-zinc-200 dark:border-white/[0.06] p-4 text-[11px] text-zinc-500 flex flex-col gap-2 leading-relaxed">
          <div className="flex items-center gap-2 text-zinc-700 dark:text-zinc-300 font-medium">
            <ShieldCheck className="size-4 text-emerald-500" />
            <span>Authenticated Goods Guaranteed</span>
          </div>
          <p>
            Your transaction is protected by end-to-end encryption. In case of sizing mismatches, our courier will facilitate a complimentary exchange.
          </p>
        </div>
      </div>

    </div>
  )
}