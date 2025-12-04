"use client"

import { Check, ArrowRight } from "lucide-react"
import { useGetApiV1Plans } from "@/api/endpoints/billing/billing"
import { useReveal } from "@/hooks/useReveal"
import Link from "next/link"

export function Pricing() {
  const { data: plansEnvelope, isLoading } = useGetApiV1Plans()
  const plans = plansEnvelope?.data || []
  const [ref, visible] = useReveal<HTMLElement>()

  // Formatting currency in EGP
  const formatEGP = (priceStr: string | undefined) => {
    if (!priceStr) return "0 EGP"
    const price = parseInt(priceStr, 10)
    return new Intl.NumberFormat("en-EG", {
      style: "currency",
      currency: "EGP",
      maximumFractionDigits: 0
    }).format(price)
  }

  return (
    <section
      ref={ref}
      id="pricing"
      className="w-full py-24 md:py-32 relative overflow-hidden bg-black"
    >
      {/* Background */}
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_bottom,_var(--tw-gradient-stops))] from-primary/10 via-transparent to-transparent pointer-events-none" />
      <div className="grid-pattern absolute inset-0 pointer-events-none opacity-30" />

      <div className="container relative mx-auto px-4 md:px-6 z-10 max-w-6xl">
        {/* Section header */}
        <div className={`flex flex-col items-center text-center mb-16 transition-all duration-700 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}>
          <div className="inline-flex items-center rounded-full border border-primary/20 bg-primary/5 px-3 py-1 text-xs font-medium text-primary uppercase tracking-widest mb-5">
            Simple Pricing
          </div>
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold tracking-tight text-white mb-4">
            Choose the perfect plan
          </h2>
          <p className="max-w-lg text-zinc-500 md:text-lg leading-relaxed">
            Transparent pricing with no hidden fees. Start with a 14-day free trial on any plan.
          </p>
        </div>

        {isLoading ? (
          <div className="flex justify-center py-20">
            <div className="size-12 rounded-full border-4 border-primary/30 border-t-primary animate-spin" />
          </div>
        ) : (
          <div className={`grid grid-cols-1 lg:grid-cols-3 gap-6 items-stretch transition-all duration-700 delay-200 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}>
            {plans.map((plan, i) => {
              const isPro = plan.name?.toLowerCase() === "pro"

              return (
                <div
                  key={plan.id || i}
                  className={`relative flex flex-col rounded-2xl p-px transition-all duration-500 hover:-translate-y-2 group ${
                    isPro ? "animated-border pulse-glow lg:scale-[1.04]" : ""
                  }`}
                >
                  {/* Inner card */}
                  <div className={`flex-1 flex flex-col rounded-[calc(1rem-1px)] p-8 ${
                    isPro
                      ? "bg-black/90 backdrop-blur-xl"
                      : "border border-white/[0.06] bg-white/[0.02] backdrop-blur-md hover:border-white/10"
                  }`}>
                    {isPro && (
                      <div className="absolute -top-4 left-1/2 -translate-x-1/2 rounded-full bg-gradient-to-r from-primary to-purple-400 px-4 py-1 text-xs font-bold text-white shadow-lg shadow-primary/30 z-10">
                        Most Popular
                      </div>
                    )}

                    <div className="mb-8">
                      <h3 className={`text-xl font-bold ${isPro ? "text-gradient-primary" : "text-white"}`}>
                        {plan.name}
                      </h3>
                      <p className="text-sm text-zinc-500 mt-1.5">
                        {isPro ? "Perfect for growing businesses." : "Everything you need to start."}
                      </p>
                      <div className="mt-6 flex items-baseline">
                        <span className="text-4xl sm:text-5xl font-extrabold tracking-tight text-white">
                          {formatEGP(plan.monthly_price)}
                        </span>
                        <span className="ml-1.5 text-sm font-medium text-zinc-600">/month</span>
                      </div>
                    </div>

                    <ul className="mb-8 flex flex-1 flex-col gap-3.5 text-sm">
                      {[
                        { text: `${plan.max_products === -1 ? "Unlimited" : plan.max_products} Products`, show: true },
                        { text: `${plan.max_staff_accounts} Staff Accounts`, show: true },
                        { text: `${plan.transaction_fee_percent}% Transaction Fee`, show: true },
                        { text: "Custom Domain", show: plan.features?.custom_domain },
                        { text: "Advanced Analytics", show: plan.features?.advanced_analytics },
                        { text: "Priority Support", show: plan.features?.priority_support },
                        { text: "Full API Access", show: plan.features?.api_access },
                      ].filter(f => f.show).map((feature) => (
                        <li key={feature.text} className="flex items-center gap-3 text-zinc-400">
                          <div className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                            <Check className="size-3" strokeWidth={3} />
                          </div>
                          <span>{feature.text}</span>
                        </li>
                      ))}
                    </ul>

                    <Link
                      href="/superadmin/login"
                      className={`group/btn relative mt-auto flex h-12 items-center justify-center rounded-xl font-medium text-sm transition-all duration-300 overflow-hidden btn-shimmer ${
                        isPro
                          ? "bg-white text-black hover:bg-zinc-200 shadow-lg shadow-white/10 hover:shadow-white/20"
                          : "bg-white/5 text-zinc-300 hover:bg-white/10 hover:text-white border border-white/[0.06]"
                      }`}
                    >
                      Get Started
                      <ArrowRight className="ml-2 size-3.5 transition-transform group-hover/btn:translate-x-0.5" />
                    </Link>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </section>
  )
}
