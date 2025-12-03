"use client"

import { Layers, Zap, Globe, Shield, Activity, Users } from "lucide-react"
import { useReveal } from "@/hooks/useReveal"
import { useCallback, useRef } from "react"

const features = [
  {
    title: "Built for Scale",
    description: "Enterprise-grade infrastructure built for high traffic and reliability. Handle millions of requests without breaking a sweat.",
    icon: Layers,
    gradient: "from-primary/20 via-primary/5 to-transparent",
    iconBg: "bg-primary/10 text-primary",
    span: "md:col-span-2 md:row-span-2",
  },
  {
    title: "Custom Domains",
    description: "Connect your own domain instantly and build your unique brand identity online.",
    icon: Globe,
    gradient: "from-blue-500/10 to-transparent",
    iconBg: "bg-blue-500/10 text-blue-400",
    span: "md:col-span-1",
  },
  {
    title: "Blazing Performance",
    description: "Sub-50ms response times globally. Your customers never wait.",
    icon: Zap,
    gradient: "from-amber-500/10 to-transparent",
    iconBg: "bg-amber-500/10 text-amber-400",
    span: "md:col-span-1",
  },
  {
    title: "Bank-Grade Security",
    description: "Automatic SSL, data encryption, and complete customer data protection built into every layer.",
    icon: Shield,
    gradient: "from-emerald-500/10 to-transparent",
    iconBg: "bg-emerald-500/10 text-emerald-400",
    span: "md:col-span-1 md:row-span-2",
  },
  {
    title: "Advanced Analytics",
    description: "Real-time insights into performance, revenue growth, and customer behavior.",
    icon: Activity,
    gradient: "from-purple-500/10 to-transparent",
    iconBg: "bg-purple-500/10 text-purple-400",
    span: "md:col-span-1",
  },
  {
    title: "AI-Powered Operations",
    description: "Automate catalog management, customer support, and SEO with dedicated AI tools.",
    icon: Users,
    gradient: "from-rose-500/10 to-transparent",
    iconBg: "bg-rose-500/10 text-rose-400",
    span: "md:col-span-1",
  },
]

function FeatureCard({ feature, index }: { feature: typeof features[0]; index: number }) {
  const cardRef = useRef<HTMLDivElement>(null)

  const handleMouseMove = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    const el = cardRef.current
    if (!el) return
    const rect = el.getBoundingClientRect()
    el.style.setProperty("--mouse-x", `${e.clientX - rect.left}px`)
    el.style.setProperty("--mouse-y", `${e.clientY - rect.top}px`)
  }, [])

  return (
    <div
      ref={cardRef}
      onMouseMove={handleMouseMove}
      className={`spotlight-card group relative rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 md:p-8 transition-all duration-500 hover:border-white/15 hover:-translate-y-1 ${feature.span} flex flex-col justify-between min-h-[200px]`}
      style={{ animationDelay: `${index * 100}ms` }}
    >
      {/* Background gradient */}
      <div className={`absolute inset-0 bg-gradient-to-br ${feature.gradient} rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none`} />

      <div className="relative z-10">
        <div className={`mb-5 inline-flex h-12 w-12 items-center justify-center rounded-xl ${feature.iconBg} transition-all duration-500 group-hover:scale-110 group-hover:shadow-lg`}>
          <feature.icon className="size-6" />
        </div>
      </div>

      <div className="mt-auto relative z-10">
        <h3 className="mb-2 text-xl font-bold tracking-tight text-white group-hover:text-gradient-primary transition-colors duration-300">
          {feature.title}
        </h3>
        <p className="text-sm text-zinc-500 leading-relaxed group-hover:text-zinc-400 transition-colors duration-300">
          {feature.description}
        </p>
      </div>
    </div>
  )
}

export function Features() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      id="features"
      className="w-full py-24 md:py-32 bg-black relative overflow-hidden"
    >
      {/* Subtle background orbs */}
      <div className="absolute top-40 -left-40 w-96 h-96 bg-primary/10 rounded-full blur-[150px] pointer-events-none" />
      <div className="absolute bottom-40 -right-40 w-80 h-80 bg-purple-500/8 rounded-full blur-[150px] pointer-events-none" />

      <div className="container relative mx-auto px-4 md:px-6 max-w-7xl">
        {/* Section header */}
        <div className={`flex flex-col items-center text-center mb-16 transition-all duration-700 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}>
          <div className="inline-flex items-center rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-medium text-zinc-400 uppercase tracking-widest mb-5">
            Everything you need
          </div>
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-bold tracking-tight text-white mb-4">
            Powering modern commerce
          </h2>
          <p className="max-w-xl text-zinc-500 md:text-lg leading-relaxed">
            The complete infrastructure to run a world-class online store.
          </p>
        </div>

        {/* Bento Grid */}
        <div className={`grid grid-cols-1 md:grid-cols-4 gap-3 md:gap-4 auto-rows-[220px] transition-all duration-700 delay-200 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}>
          {features.map((feature, i) => (
            <FeatureCard key={feature.title} feature={feature} index={i} />
          ))}
        </div>
      </div>
    </section>
  )
}
