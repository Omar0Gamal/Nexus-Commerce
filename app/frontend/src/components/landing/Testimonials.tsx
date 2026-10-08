"use client"

import { useReveal } from "@/hooks/useReveal"
import { Star } from "@phosphor-icons/react"

const testimonials = [
  {
    quote: "Nexus Commerce transformed our online presence. We went from 100 orders a month to 2,000+ in just three months. The platform is incredibly intuitive.",
    name: "Ahmed Mostafa",
    role: "CEO, TechStart Egypt",
    initials: "AM",
  },
  {
    quote: "The sub-50ms response time is real. Our customers noticed the speed difference immediately and our conversion rate jumped by 34%.",
    name: "Sara El-Kady",
    role: "CTO, SwiftPay",
    initials: "SK",
  },
  {
    quote: "Best e-commerce platform we've used. The analytics dashboard alone is worth the subscription. We can see exactly what drives our revenue.",
    name: "Mohamed Hassan",
    role: "Founder, CraftHub",
    initials: "MH",
  },
  {
    quote: "We migrated from a competitor in under 48 hours. The import tools are seamless and the support team was incredibly responsive.",
    name: "Layla Nour",
    role: "Operations Lead, NovaTrade",
    initials: "LN",
  },
  {
    quote: "Security was our biggest concern. Nexus gives us bank-grade protection with automatic SSL and encryption. Our customers trust us more.",
    name: "Karim Adel",
    role: "Head of Security, DataFlow",
    initials: "KA",
  },
  {
    quote: "The AI-powered product descriptions saved our content team 20 hours a week. It's like having an extra team member who never sleeps.",
    name: "Nadia Ibrahim",
    role: "Marketing Director, PixelForge",
    initials: "NI",
  },
]

export function Testimonials() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      id="testimonials"
      className="w-full py-28 md:py-40 relative overflow-hidden"
    >
      <div className="absolute inset-0 pointer-events-none opacity-[0.025] z-0" style={{ backgroundImage: "url('https://grainy-gradients.vercel.app/noise.svg')" }} />

      <div className="container relative mx-auto px-4 md:px-6 max-w-7xl z-10">
        <div className={`flex flex-col items-center text-center mb-20 transition-all duration-1000 ease-[cubic-bezier(0.32,0.72,0,1)] ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}>
          <div className="inline-flex items-center gap-2 rounded-full px-3 py-1 mb-6 text-[10px] uppercase tracking-[0.2em] font-semibold border border-primary/20 bg-primary/10 text-primary backdrop-blur-md">
            Testimonials
          </div>
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-medium tracking-tighter text-foreground mb-5 text-wrap-balance">
            Loved by fast-growing businesses.
          </h2>
          <p className="max-w-[54ch] text-muted-foreground md:text-lg leading-relaxed">
            Don&apos;t just take our word for it. Hear from the operators and engineers scaling their commerce on Nexus.
          </p>
        </div>

        <div className={`relative w-full overflow-hidden transition-all duration-1000 delay-150 ease-[cubic-bezier(0.32,0.72,0,1)] ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-16"}`}>
          <div className="absolute inset-0 z-20 pointer-events-none" style={{ maskImage: 'linear-gradient(to right, transparent, black 5%, black 95%, transparent)', WebkitMaskImage: 'linear-gradient(to right, transparent, black 5%, black 95%, transparent)' }}></div>
          
          <div className="flex w-max animate-marquee hover:[animation-play-state:paused] gap-6 pb-10">
            {[...testimonials, ...testimonials].map((t, i) => (
              <div
                key={`${t.name}-${i}`}
                className="group relative flex flex-col shrink-0 w-[85vw] sm:w-[380px] md:w-[420px] transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] hover:-translate-y-1"
              >
              <div className="absolute inset-0 rounded-[1.5rem] bg-card/60 backdrop-blur-md border border-border/40 transition-colors duration-500 group-hover:border-primary/30" />

               <div
                className="absolute inset-0 rounded-[1.5rem] opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none"
                style={{
                  background: "radial-gradient(circle 200px at 50% 0%, oklch(0.6 0.28 260 / 0.04), transparent 70%)",
                }}
              />

              <div className="relative flex-1 flex flex-col justify-between p-7 md:p-8">

                <div className="flex gap-1 mb-6">
                  {[...Array(5)].map((_, j) => (
                    <Star key={j} weight="fill" className="size-3.5 text-primary/80" />
                  ))}
                </div>

                <blockquote className="text-[0.95rem] text-muted-foreground leading-relaxed mb-8 line-clamp-3">
                  &ldquo;{t.quote}&rdquo;
                </blockquote>

                <div className="flex items-center gap-3.5 mt-auto">
                  <div className="size-10 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center text-xs font-semibold text-primary shrink-0">
                    {t.initials}
                  </div>
                  <div>
                    <div className="text-sm font-medium text-foreground">{t.name}</div>
                    <div className="text-xs text-muted-foreground/80 mt-0.5">{t.role}</div>
                  </div>
                </div>
              </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
