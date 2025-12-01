"use client"

import Link from "next/link"
import { ArrowRight, Play } from "lucide-react"
import { useEffect, useState } from "react"

const rotatingWords = ["e-commerce", "storefront", "business", "brand"]

function RotatingWord() {
  const [index, setIndex] = useState(0)

  useEffect(() => {
    const interval = setInterval(() => {
      setIndex((i) => (i + 1) % rotatingWords.length)
    }, 2800)
    return () => clearInterval(interval)
  }, [])

  return (
    <span className="relative inline-block overflow-hidden h-[1.15em] align-bottom">
      {rotatingWords.map((word, i) => (
        <span
          key={word}
          className="absolute left-0 right-0 transition-all duration-500 ease-out"
          style={{
            transform: i === index ? "translateY(0)" : i === (index - 1 + rotatingWords.length) % rotatingWords.length ? "translateY(-110%)" : "translateY(110%)",
            opacity: i === index ? 1 : 0,
            background: "linear-gradient(135deg, oklch(0.7 0.2 285), oklch(0.8 0.15 200))",
            WebkitBackgroundClip: "text",
            WebkitTextFillColor: "transparent",
            backgroundClip: "text",
          }}
        >
          {word}
        </span>
      ))}
    </span>
  )
}

function DashboardMockup() {
  return (
    <div className="relative w-full max-w-2xl mx-auto animate-float">
      {/* Browser chrome */}
      <div className="rounded-2xl border border-white/10 bg-black/60 backdrop-blur-xl shadow-2xl shadow-primary/10 overflow-hidden">
        {/* Title bar */}
        <div className="flex items-center gap-2 px-4 py-3 border-b border-white/5 bg-white/[0.02]">
          <div className="flex gap-1.5">
            <div className="size-3 rounded-full bg-red-500/60" />
            <div className="size-3 rounded-full bg-yellow-500/60" />
            <div className="size-3 rounded-full bg-green-500/60" />
          </div>
          <div className="flex-1 flex justify-center">
            <div className="px-4 py-1 rounded-md bg-white/5 text-xs text-zinc-500 font-mono">
              dashboard.nexus.com
            </div>
          </div>
        </div>

        {/* Dashboard content */}
        <div className="p-5 space-y-4">
          {/* Stats row */}
          <div className="grid grid-cols-3 gap-3">
            {[
              { label: "Revenue", value: "EGP 284,500", change: "+24.5%", color: "text-emerald-400" },
              { label: "Orders", value: "1,847", change: "+12.8%", color: "text-emerald-400" },
              { label: "Customers", value: "3,621", change: "+8.2%", color: "text-emerald-400" },
            ].map((stat) => (
              <div key={stat.label} className="rounded-xl bg-white/[0.03] border border-white/5 p-3">
                <div className="text-[10px] text-zinc-500 uppercase tracking-wider">{stat.label}</div>
                <div className="text-lg font-bold text-white mt-1">{stat.value}</div>
                <div className={`text-[10px] font-medium ${stat.color} mt-0.5`}>{stat.change}</div>
              </div>
            ))}
          </div>

          {/* Chart area */}
          <div className="rounded-xl bg-white/[0.03] border border-white/5 p-4">
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs text-zinc-400 font-medium">Revenue Analytics</span>
              <span className="text-[10px] text-zinc-600">Last 7 days</span>
            </div>
            {/* Fake chart bars */}
            <div className="flex items-end gap-1.5 h-20">
              {[40, 65, 45, 80, 55, 90, 70].map((h, i) => (
                <div
                  key={i}
                  className="flex-1 rounded-t-sm bg-gradient-to-t from-primary/80 to-primary/30 transition-all duration-700"
                  style={{
                    height: `${h}%`,
                    animationDelay: `${i * 100}ms`,
                  }}
                />
              ))}
            </div>
            <div className="flex justify-between mt-2">
              {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map((d) => (
                <span key={d} className="text-[8px] text-zinc-600 flex-1 text-center">{d}</span>
              ))}
            </div>
          </div>

          {/* Recent orders */}
          <div className="rounded-xl bg-white/[0.03] border border-white/5 p-3">
            <div className="text-xs text-zinc-400 font-medium mb-2">Recent Orders</div>
            {[
              { id: "#4821", customer: "Ahmed M.", amount: "EGP 1,250", status: "Completed" },
              { id: "#4820", customer: "Sara K.", amount: "EGP 890", status: "Processing" },
              { id: "#4819", customer: "Omar H.", amount: "EGP 2,100", status: "Completed" },
            ].map((order) => (
              <div key={order.id} className="flex items-center justify-between py-1.5 border-b border-white/5 last:border-0">
                <div className="flex items-center gap-2">
                  <span className="text-[10px] font-mono text-zinc-500">{order.id}</span>
                  <span className="text-xs text-zinc-300">{order.customer}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-medium text-white">{order.amount}</span>
                  <span className={`text-[9px] px-1.5 py-0.5 rounded-full ${order.status === "Completed" ? "bg-emerald-500/10 text-emerald-400" : "bg-amber-500/10 text-amber-400"}`}>
                    {order.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Floating notification card */}
      <div className="absolute -right-4 top-24 animate-float-delayed z-10">
        <div className="rounded-xl border border-white/10 bg-black/70 backdrop-blur-xl p-3 shadow-xl shadow-black/30 max-w-[180px]">
          <div className="flex items-center gap-2">
            <div className="size-7 rounded-full bg-emerald-500/20 flex items-center justify-center">
              <span className="text-emerald-400 text-xs">✓</span>
            </div>
            <div>
              <div className="text-[10px] font-medium text-white">New Order!</div>
              <div className="text-[9px] text-zinc-500">EGP 1,250 · Just now</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export function Hero() {
  return (
    <section className="relative w-full min-h-screen flex items-center overflow-hidden bg-black pt-20">
      {/* Aurora gradient blobs */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="aurora-blob absolute -top-[40%] -left-[20%] w-[70%] h-[70%] rounded-full bg-[radial-gradient(circle,oklch(0.7_0.2_285/0.25),transparent_60%)] blur-[40px]" />
        <div className="aurora-blob-reverse absolute -bottom-[30%] -right-[20%] w-[60%] h-[60%] rounded-full bg-[radial-gradient(circle,oklch(0.8_0.15_200/0.15),transparent_60%)] blur-[40px]" />
        <div className="aurora-blob-slow absolute top-[20%] right-[10%] w-[40%] h-[40%] rounded-full bg-[radial-gradient(circle,oklch(0.7_0.18_320/0.1),transparent_60%)] blur-[40px]" />
      </div>

      {/* Grid pattern overlay */}
      <div className="grid-pattern absolute inset-0 pointer-events-none opacity-60" />

      {/* Top edge gradient line */}
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-primary/50 to-transparent" />

      <div className="container relative mx-auto px-4 md:px-6 z-10 max-w-7xl">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 lg:gap-12 items-center">

          {/* Text Content */}
          <div className="flex flex-col items-center lg:items-start text-center lg:text-left gap-6 animate-fade-in-up">
            {/* Badge */}
            <div className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/5 px-4 py-1.5 text-sm font-medium text-primary backdrop-blur-sm badge-shimmer"
              style={{ backgroundImage: "linear-gradient(90deg, transparent, oklch(0.7 0.2 285 / 0.1), transparent)" }}
            >
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-primary" />
              </span>
              Now in public beta
            </div>

            <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-extrabold tracking-tight leading-[1.08]">
              <span className="bg-gradient-to-b from-white to-zinc-400 bg-clip-text text-transparent">
                Scale your{" "}
              </span>
              <br className="hidden sm:block" />
              <RotatingWord />
              <br className="hidden sm:block" />
              <span className="bg-gradient-to-b from-white to-zinc-400 bg-clip-text text-transparent">
                effortlessly
              </span>
            </h1>

            <p className="max-w-lg text-zinc-400 text-base sm:text-lg leading-relaxed">
              The enterprise-grade commerce engine with bank-grade security, sub-50ms performance, and everything you need to build, manage, and scale.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 mt-2 w-full sm:w-auto">
              <Link
                href="/superadmin/login"
                className="group relative inline-flex h-13 items-center justify-center overflow-hidden rounded-full bg-white px-8 text-base font-semibold text-black transition-all duration-300 hover:scale-105 shadow-[0_0_30px_rgba(255,255,255,0.15)] hover:shadow-[0_0_50px_rgba(255,255,255,0.25)] btn-shimmer"
              >
                Start free trial
                <ArrowRight className="ml-2 size-4 transition-transform duration-300 group-hover:translate-x-1" />
              </Link>
              <Link
                href="#features"
                className="group inline-flex h-13 items-center justify-center rounded-full border border-white/10 bg-white/5 backdrop-blur-sm px-8 text-base font-medium text-zinc-300 transition-all duration-300 hover:bg-white/10 hover:text-white hover:border-white/20"
              >
                <Play className="mr-2 size-4 text-primary" />
                See how it works
              </Link>
            </div>

            {/* Social proof mini */}
            <div className="flex items-center gap-3 mt-4">
              <div className="flex -space-x-2">
                {[
                  "bg-gradient-to-br from-violet-500 to-purple-600",
                  "bg-gradient-to-br from-blue-500 to-cyan-500",
                  "bg-gradient-to-br from-emerald-500 to-teal-500",
                  "bg-gradient-to-br from-orange-500 to-amber-500",
                ].map((bg, i) => (
                  <div key={i} className={`size-8 rounded-full ${bg} border-2 border-black flex items-center justify-center text-[10px] font-bold text-white`}>
                    {["A", "S", "M", "K"][i]}
                  </div>
                ))}
              </div>
              <div className="text-sm text-zinc-400">
                <span className="text-white font-semibold">500+</span> businesses growing with Nexus
              </div>
            </div>
          </div>

          {/* Dashboard Mockup */}
          <div className="relative" style={{ animationDelay: "200ms" }}>
            <DashboardMockup />
          </div>
        </div>
      </div>

      {/* Bottom fade */}
      <div className="absolute bottom-0 left-0 right-0 h-32 bg-gradient-to-t from-black to-transparent pointer-events-none" />
    </section>
  )
}
