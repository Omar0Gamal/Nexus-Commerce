"use client"

import { useReveal } from "@/hooks/useReveal"
import { useEffect, useRef, useState } from "react"

function AnimatedNumber({ target, suffix = "", prefix = "" }: { target: number; suffix?: string; prefix?: string }) {
  const [count, setCount] = useState(0)
  const [started, setStarted] = useState(false)
  const ref = useRef<HTMLSpanElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !started) {
          setStarted(true)
          observer.unobserve(el)
        }
      },
      { threshold: 0.5 }
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [started])

  useEffect(() => {
    if (!started) return
    const duration = 2000
    const startTime = performance.now()
    const step = (now: number) => {
      const elapsed = now - startTime
      const progress = Math.min(elapsed / duration, 1)
      // Ease-out cubic
      const eased = 1 - Math.pow(1 - progress, 3)
      setCount(Math.floor(eased * target))
      if (progress < 1) requestAnimationFrame(step)
    }
    requestAnimationFrame(step)
  }, [started, target])

  return (
    <span ref={ref} className="text-gradient-primary">
      {prefix}{count.toLocaleString()}{suffix}
    </span>
  )
}

const stats = [
  { value: 500, suffix: "+", label: "Active Stores", description: "Businesses running on Nexus" },
  { value: 2, suffix: "M+", label: "Orders Processed", description: "Transactions handled seamlessly" },
  { value: 99, suffix: ".9%", label: "Uptime", description: "Enterprise-grade reliability" },
  { value: 50, suffix: "ms", prefix: "<", label: "Response Time", description: "Lightning-fast globally" },
]

export function Stats() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      className="w-full py-24 bg-black relative overflow-hidden"
    >
      {/* Grid pattern */}
      <div className="grid-pattern absolute inset-0 pointer-events-none opacity-40" />

      <div className="container relative mx-auto px-4 md:px-6 max-w-7xl">
        <div className={`grid grid-cols-2 lg:grid-cols-4 gap-8 lg:gap-4 transition-all duration-700 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}>
          {stats.map((stat, i) => (
            <div
              key={stat.label}
              className="flex flex-col items-center text-center group"
              style={{ transitionDelay: `${i * 100}ms` }}
            >
              <div className="text-4xl sm:text-5xl md:text-6xl font-extrabold tracking-tight mb-2">
                <AnimatedNumber target={stat.value} suffix={stat.suffix} prefix={stat.prefix} />
              </div>
              <div className="text-sm font-semibold text-white mb-1">{stat.label}</div>
              <div className="text-xs text-zinc-500">{stat.description}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
