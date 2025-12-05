"use client"

import { useReveal } from "@/hooks/useReveal"

const companies = [
  "TechStart", "AlphaRetail", "GrowthCo", "SwiftPay", "CloudNine",
  "DataFlow", "PulseMedia", "CraftHub", "NovaTrade", "PixelForge",
  "VeloCity", "ZenMarket", "BrightPath", "CoreStack", "FluxPoint",
]

export function TrustedBy() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      className={`relative w-full py-16 bg-black border-y border-white/5 overflow-hidden transition-all duration-700 ${visible ? "opacity-100" : "opacity-0"}`}
    >
      <div className="container mx-auto px-4 md:px-6 max-w-7xl mb-8">
        <p className="text-center text-sm font-medium text-zinc-500 uppercase tracking-widest">
          Trusted by <span className="text-zinc-300">500+</span> businesses worldwide
        </p>
      </div>

      {/* Marquee container */}
      <div className="relative">
        {/* Edge fades */}
        <div className="absolute left-0 top-0 bottom-0 w-32 bg-gradient-to-r from-black to-transparent z-10 pointer-events-none" />
        <div className="absolute right-0 top-0 bottom-0 w-32 bg-gradient-to-l from-black to-transparent z-10 pointer-events-none" />

        <div className="flex overflow-hidden">
          <div className="animate-marquee flex shrink-0 items-center gap-12 pr-12">
            {[...companies, ...companies].map((name, i) => (
              <div
                key={`${name}-${i}`}
                className="flex items-center gap-2 text-zinc-600 hover:text-zinc-300 transition-colors duration-300 select-none cursor-default"
              >
                <div className="size-6 rounded-md bg-zinc-800/80 flex items-center justify-center text-[10px] font-bold text-zinc-500">
                  {name.charAt(0)}
                </div>
                <span className="text-sm font-medium whitespace-nowrap">{name}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
