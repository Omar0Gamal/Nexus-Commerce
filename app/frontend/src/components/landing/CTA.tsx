"use client"

import Link from "next/link"
import { ArrowRight } from "@phosphor-icons/react"
import { useReveal } from "@/hooks/useReveal"

export function CTA() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      className="w-full relative overflow-hidden"
    >


      <div className="container relative mx-auto px-4 md:px-6 max-w-4xl z-10 py-32 md:py-48">
        <div className={`flex flex-col items-center text-center transition-all duration-1000 ease-[cubic-bezier(0.32,0.72,0,1)] ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}>
          


          <h2 className="text-5xl md:text-7xl lg:text-[6.5rem] font-medium tracking-tighter mb-8 leading-[1.05] text-foreground text-wrap-balance">
            The infrastructure for <br className="hidden md:block" />
            <span className="text-muted-foreground">modern commerce.</span>
          </h2>

          <p className="text-muted-foreground/60 text-sm font-medium tracking-wide mb-12 uppercase">
            No credit card required · 14-day free trial · Cancel anytime
          </p>

          <div className="flex flex-col sm:flex-row gap-4 items-center">
            <Link
              href="/superadmin/login"
              className="group flex items-center gap-4 rounded-full bg-primary pl-8 pr-2 py-2.5 text-base font-medium text-primary-foreground transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98] shadow-[0_0_40px_-10px_oklch(0.6_0.28_260_/_0.5)] hover:shadow-[0_0_60px_-10px_oklch(0.6_0.28_260_/_0.7)] relative overflow-hidden"
            >
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-[200%] group-hover:translate-x-[200%] transition-transform duration-1000 ease-in-out" />
              <span>Start free trial</span>
              <div className="w-10 h-10 rounded-full bg-primary-foreground/20 flex items-center justify-center transition-transform duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] group-hover:translate-x-1 group-hover:-translate-y-[1px] group-hover:scale-105">
                <ArrowRight weight="bold" className="size-4.5 text-primary-foreground" />
              </div>
            </Link>
          </div>
        </div>
      </div>
    </section>
  )
}
