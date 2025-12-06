"use client"

import { useReveal } from "@/hooks/useReveal"
import { Star } from "lucide-react"

const testimonials = [
  {
    quote: "Nexus Commerce transformed our online presence. We went from 100 orders a month to 2,000+ in just three months. The platform is incredibly intuitive.",
    name: "Ahmed Mostafa",
    role: "CEO, TechStart Egypt",
    initials: "AM",
    gradient: "from-violet-500 to-purple-600",
  },
  {
    quote: "The sub-50ms response time is real. Our customers noticed the speed difference immediately and our conversion rate jumped by 34%.",
    name: "Sara El-Kady",
    role: "CTO, SwiftPay",
    initials: "SK",
    gradient: "from-blue-500 to-cyan-500",
  },
  {
    quote: "Best e-commerce platform we've used. The analytics dashboard alone is worth the subscription. We can see exactly what drives our revenue.",
    name: "Mohamed Hassan",
    role: "Founder, CraftHub",
    initials: "MH",
    gradient: "from-emerald-500 to-teal-500",
  },
  {
    quote: "We migrated from a competitor in under 48 hours. The import tools are seamless and the support team was incredibly responsive.",
    name: "Layla Nour",
    role: "Operations Lead, NovaTrade",
    initials: "LN",
    gradient: "from-orange-500 to-amber-500",
  },
  {
    quote: "Security was our biggest concern. Nexus gives us bank-grade protection with automatic SSL and encryption. Our customers trust us more.",
    name: "Karim Adel",
    role: "Head of Security, DataFlow",
    initials: "KA",
    gradient: "from-rose-500 to-pink-500",
  },
  {
    quote: "The AI-powered product descriptions saved our content team 20 hours a week. It's like having an extra team member who never sleeps.",
    name: "Nadia Ibrahim",
    role: "Marketing Director, PixelForge",
    initials: "NI",
    gradient: "from-indigo-500 to-violet-500",
  },
]

export function Testimonials() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      id="testimonials"
      className="w-full py-24 md:py-32 bg-black relative overflow-hidden"
    >
      {/* Background orbs */}
      <div className="absolute top-20 left-1/4 w-96 h-96 bg-primary/5 rounded-full blur-[150px] pointer-events-none" />
      <div className="absolute bottom-20 right-1/4 w-80 h-80 bg-purple-500/5 rounded-full blur-[150px] pointer-events-none" />

      <div className="container relative mx-auto px-4 md:px-6 max-w-7xl">
        {/* Section header */}
        <div className={`flex flex-col items-center text-center mb-16 transition-all duration-700 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}>
          <div className="inline-flex items-center rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-medium text-zinc-400 uppercase tracking-widest mb-5">
            Testimonials
          </div>
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold tracking-tight text-white mb-4">
            Loved by businesses
          </h2>
          <p className="max-w-xl text-zinc-500 md:text-lg leading-relaxed">
            Don&apos;t just take our word for it. Hear from the businesses that trust Nexus.
          </p>
        </div>

        {/* Testimonial grid */}
        <div className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 transition-all duration-700 delay-200 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}>
          {testimonials.map((t, i) => (
            <div
              key={t.name}
              className="group relative rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 transition-all duration-500 hover:border-white/10 hover:bg-white/[0.04] hover:-translate-y-1"
              style={{ animationDelay: `${i * 80}ms` }}
            >
              {/* Stars */}
              <div className="flex gap-1 mb-4">
                {[...Array(5)].map((_, j) => (
                  <Star key={j} className="size-3.5 fill-amber-400 text-amber-400" />
                ))}
              </div>

              <blockquote className="text-sm text-zinc-400 leading-relaxed mb-6">
                &ldquo;{t.quote}&rdquo;
              </blockquote>

              <div className="flex items-center gap-3 mt-auto">
                <div className={`size-9 rounded-full bg-gradient-to-br ${t.gradient} flex items-center justify-center text-xs font-bold text-white shrink-0`}>
                  {t.initials}
                </div>
                <div>
                  <div className="text-sm font-medium text-white">{t.name}</div>
                  <div className="text-xs text-zinc-500">{t.role}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
