"use client"

import { Check, ArrowRight } from "@phosphor-icons/react"
import { useReveal } from "@/hooks/useReveal"
import Link from "next/link"
import { useGetApiV1Plans } from "@/api/endpoints/billing/billing";

export function Pricing() {
  const { data: plansEnvelope, isLoading } = useGetApiV1Plans()
  const plans = plansEnvelope?.data || []
  const [ref, visible] = useReveal<HTMLElement>()

  const formatPrice = (priceStr: string | undefined) => {
    if (!priceStr) return { amount: "0", currency: "EGP" }
    const price = parseInt(priceStr, 10)
    const formatted = new Intl.NumberFormat("en-US", {
      maximumFractionDigits: 0,
    }).format(price)
    return { amount: formatted, currency: "EGP" }
  }

  return (
    <section
      ref={ref}
      id="pricing"
      className="w-full py-32 md:py-40 relative overflow-hidden"
    >
      <div className="absolute inset-0 pointer-events-none opacity-[0.025] z-0" style={{ backgroundImage: "url('https://grainy-gradients.vercel.app/noise.svg')" }} />

      <div className="container relative mx-auto px-4 md:px-6 z-10 max-w-6xl">
        <div className={`flex flex-col items-center text-center mb-20 transition-all duration-1000 ease-[cubic-bezier(0.32,0.72,0,1)] ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}>
          <div className="inline-flex items-center gap-2 rounded-full px-3 py-1 mb-6 text-[10px] uppercase tracking-[0.2em] font-semibold border border-primary/20 bg-primary/10 text-primary backdrop-blur-md">
            Pricing
          </div>
          <h2 className="text-4xl sm:text-5xl md:text-6xl font-medium tracking-tighter text-foreground mb-6">
            Transparent scaling.
          </h2>
          <p className="max-w-[50ch] text-muted-foreground md:text-lg leading-relaxed">
            No hidden fees, no complex tiers. Start with a 14-day free trial on any plan and scale as your commerce grows.
          </p>
        </div>

        {isLoading ? (
          <div className="flex justify-center py-20">
            <div className="size-8 rounded-full border-2 border-muted border-t-primary animate-spin" />
          </div>
        ) : (
          <div className={`grid grid-cols-1 lg:grid-cols-3 gap-6 md:gap-8 items-stretch transition-all duration-1000 delay-150 ease-[cubic-bezier(0.32,0.72,0,1)] ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-16"}`}>
            {plans.map((plan: any, i: number) => {
              const isPro = plan.name?.toLowerCase() === "pro"

              return (
                <div
                  key={plan.id || i}
                  className={`group relative flex flex-col transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] ${
                    isPro ? "lg:scale-[1.04] z-10" : "hover:-translate-y-1 z-0"
                  }`}
                  style={{ animationDelay: `${i * 100}ms` }}
                >
                  <div className={`absolute inset-0 rounded-[2rem] border transition-all duration-500
                    ${isPro 
                      ? "animated-border" 
                      : "bg-muted/10 border-border/40 group-hover:border-primary/20"
                    }`} 
                  />
                  
                  <div
                    className="absolute inset-0 rounded-[2rem] opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none"
                    style={{
                      background: "radial-gradient(circle 250px at 50% 0%, oklch(0.6 0.28 260 / 0.05), transparent 70%)",
                    }}
                  />
                  
                  <div className={`relative flex-1 flex flex-col rounded-[calc(2rem-0.25rem)] m-1 p-8 md:p-10 shadow-[inset_0_1px_1px_rgba(255,255,255,0.02)] border transition-colors duration-500
                    ${isPro 
                      ? "bg-card/70 backdrop-blur-md border-primary/20" 
                      : "bg-card/70 backdrop-blur-md border-border/40 group-hover:bg-card/90"
                    }`}>
                    
                    {isPro && (
                      <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 rounded-full bg-primary px-4 py-1 text-[10px] font-bold text-primary-foreground uppercase tracking-[0.2em] shadow-[0_0_15px_-3px_oklch(0.6_0.28_260_/_0.5)]">
                        Most Popular
                      </div>
                    )}

                    <div className="mb-10 relative z-10">
                      <h3 className="text-2xl font-medium text-foreground tracking-tight">
                        {plan.name}
                      </h3>
                      <p className="text-sm text-muted-foreground mt-2 min-h-[40px]">
                        {isPro ? "The complete engine for growing businesses." : "Everything you need to start selling."}
                      </p>
                      
                      <div className="mt-8 flex flex-col gap-1">
                        <span className="text-[10px] font-bold text-muted-foreground/60 uppercase tracking-[0.2em]">
                          {formatPrice(plan.monthly_price).currency} / month
                        </span>
                        <div className="flex items-baseline gap-1.5">
                          <span className="text-5xl md:text-6xl font-medium tracking-tighter text-foreground tabular-nums">
                            {formatPrice(plan.monthly_price).amount}
                          </span>
                        </div>
                      </div>
                    </div>

                    <div className="h-px w-full bg-border/40 mb-8" />

                    <ul className="mb-10 flex flex-1 flex-col gap-4 text-sm relative z-10">
                      {[
                        { text: `${plan.max_products === -1 ? "Unlimited" : plan.max_products} Products`, show: true },
                        { text: `${plan.max_staff_accounts} Staff Accounts`, show: true },
                        { text: `${plan.transaction_fee_percent}% Transaction Fee`, show: true },
                        { text: "Custom Domain", show: plan.features?.custom_domain },
                        { text: "Advanced Analytics", show: plan.features?.advanced_analytics },
                        { text: "Priority Support", show: plan.features?.priority_support },
                        { text: "Full API Access", show: plan.features?.api_access },
                      ].filter(f => f.show).map((feature) => (
                        <li key={feature.text} className="flex items-center gap-3 text-muted-foreground/80">
                          <div className={`flex h-4 w-4 shrink-0 items-center justify-center rounded-full ${isPro ? "bg-primary/20 shadow-[0_0_10px_oklch(0.6_0.28_260_/_0.3)]" : "bg-muted border border-border/50"}`}>
                            <div className={`w-1.5 h-1.5 rounded-full ${isPro ? "bg-primary" : "bg-muted-foreground/60"}`} />
                          </div>
                          <span>{feature.text}</span>
                        </li>
                      ))}
                    </ul>

                    <Link
                      href="/superadmin/login"
                      className={`group/btn mt-auto flex items-center justify-between rounded-full pl-6 pr-2 py-2 text-sm font-medium transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98] relative z-10 ${
                        isPro
                          ? "bg-primary text-primary-foreground shadow-[0_0_30px_-5px_oklch(0.6_0.28_260_/_0.4)] hover:shadow-[0_0_40px_-5px_oklch(0.6_0.28_260_/_0.6)]"
                          : "bg-muted/40 text-foreground border border-border/50 hover:bg-muted/80 hover:border-border"
                      }`}
                    >
                      <span>Start with {plan.name}</span>
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center transition-transform duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] group-hover/btn:translate-x-1 group-hover/btn:-translate-y-[1px] group-hover/btn:scale-105 ${isPro ? "bg-primary-foreground/20" : "bg-background/80"}`}>
                        <ArrowRight weight="bold" className={`size-3.5 ${isPro ? "text-primary-foreground" : "text-foreground"}`} />
                      </div>
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
