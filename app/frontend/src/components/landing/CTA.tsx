"use client"

import Link from "next/link"
import { ArrowRight } from "lucide-react"
import { useReveal } from "@/hooks/useReveal"

export function CTA() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      className="w-full py-24 md:py-32 bg-black relative overflow-hidden"
    >
      {/* Aurora background */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="aurora-blob absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[80%] h-[60%] rounded-full bg-[radial-gradient(circle,oklch(0.7_0.2_285/0.15),transparent_60%)] blur-[60px]" />
        <div className="aurora-blob-reverse absolute top-1/2 left-1/3 -translate-x-1/2 -translate-y-1/2 w-[40%] h-[50%] rounded-full bg-[radial-gradient(circle,oklch(0.8_0.15_200/0.1),transparent_60%)] blur-[60px]" />
      </div>

      <div className="container relative mx-auto px-4 md:px-6 max-w-4xl">
        <div className={`flex flex-col items-center text-center transition-all duration-700 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}>
          <h2 className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-6">
            <span className="bg-gradient-to-b from-white to-zinc-400 bg-clip-text text-transparent">
              Ready to transform
            </span>
            <br />
            <span className="text-gradient-primary">your commerce?</span>
          </h2>

          <p className="max-w-lg text-zinc-500 text-base sm:text-lg leading-relaxed mb-10">
            Join 500+ businesses already scaling with Nexus. Start your free 14-day trial today — no credit card required.
          </p>

          <div className="flex flex-col sm:flex-row gap-4">
            <Link
              href="/superadmin/login"
              className="group relative inline-flex h-14 items-center justify-center overflow-hidden rounded-full bg-white px-10 text-base font-semibold text-black transition-all duration-300 hover:scale-105 shadow-[0_0_40px_rgba(255,255,255,0.15)] hover:shadow-[0_0_60px_rgba(255,255,255,0.25)] btn-shimmer"
            >
              Start your free trial
              <ArrowRight className="ml-2 size-4 transition-transform duration-300 group-hover:translate-x-1" />
            </Link>
            <Link
              href="#pricing"
              className="inline-flex h-14 items-center justify-center rounded-full border border-white/10 bg-white/5 backdrop-blur-sm px-10 text-base font-medium text-zinc-300 transition-all duration-300 hover:bg-white/10 hover:text-white hover:border-white/20"
            >
              View pricing
            </Link>
          </div>
        </div>
      </div>
    </section>
  )
}
