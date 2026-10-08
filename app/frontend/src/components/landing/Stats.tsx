"use client"

import { useReveal } from "@/hooks/useReveal"
import { useEffect, useRef, useState } from "react"

function AnimatedNumber({
  target,
  suffix = "",
  prefix = "",
}: {
  target: number
  suffix?: string
  prefix?: string
}) {
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
      const eased = 1 - Math.pow(1 - progress, 3)
      setCount(Math.floor(eased * target))
      if (progress < 1) requestAnimationFrame(step)
    }
    requestAnimationFrame(step)
  }, [started, target])

  return (
    <span ref={ref}>
      {prefix}
      {count.toLocaleString()}
      {suffix}
    </span>
  )
}

const stats = [
  {
    value: 500,
    suffix: "+",
    label: "Active stores",
    description: "Businesses running on Nexus Commerce",
    accent: true,
  },
  {
    value: 2,
    suffix: "M+",
    label: "Orders processed",
    description: "Transactions handled without a single outage",
    accent: false,
  },
  {
    value: 99,
    suffix: ".9%",
    label: "Uptime",
    description: "SLA-backed reliability, monitored 24/7",
    accent: false,
  },
  {
    value: 50,
    suffix: "ms",
    prefix: "<",
    label: "Avg. response",
    description: "Edge-deployed globally for instant load times",
    accent: true,
  },
]

export function Stats() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      className="w-full relative overflow-hidden"
    >

      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-primary/25 to-transparent z-0" />
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-border/30 to-transparent z-0" />

      <div className="container relative mx-auto px-4 md:px-6 max-w-7xl z-10 py-16 md:py-24 border-y border-border/20">
        <p
          className={`text-center text-sm md:text-base font-medium text-muted-foreground mb-16 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] ${
            visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"
          }`}
        >
          Platform by the numbers
        </p>

        <div
          className={`grid grid-cols-2 lg:grid-cols-4 gap-8 lg:gap-4 transition-all duration-700 delay-100 ease-[cubic-bezier(0.32,0.72,0,1)] ${
            visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-10"
          }`}
        >
          {stats.map((stat, i) => (
            <div
              key={stat.label}
              className="group flex flex-col items-center text-center relative"
              style={{ transitionDelay: `${i * 80}ms` }}
            >
              {i > 0 && (
                <div className="absolute left-0 top-4 bottom-4 w-px bg-border/25 hidden lg:block" />
              )}

              <div
                className={`text-[4rem] sm:text-[5rem] md:text-[7rem] font-medium tracking-tighter leading-none mb-3 tabular-nums
                  ${stat.accent ? "text-foreground" : "text-foreground/80"}`}
              >
                <AnimatedNumber target={stat.value} suffix={stat.suffix} prefix={stat.prefix} />
              </div>

              <div className="text-sm font-semibold text-foreground/80 uppercase tracking-widest mb-1.5">
                {stat.label}
              </div>

              <div className="text-xs text-muted-foreground/50 max-w-[20ch] leading-relaxed">
                {stat.description}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
