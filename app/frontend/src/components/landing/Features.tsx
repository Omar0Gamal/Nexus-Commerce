"use client"

import { useReveal } from "@/hooks/useReveal"
import { ShieldCheck, ChartLineUp, Robot, ArrowUpRight } from "@phosphor-icons/react"

export function Features() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      id="features"
      className="w-full py-28 md:py-40 relative overflow-hidden"
    >
      <div className="absolute inset-0 pointer-events-none opacity-[0.02] z-0" style={{ backgroundImage: "url('https://grainy-gradients.vercel.app/noise.svg')" }} />

      <div className="container relative mx-auto px-4 md:px-6 max-w-7xl z-10">
        <div
          className={`mb-16 md:mb-24 max-w-3xl transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] ${
            visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-10"
          }`}
        >
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-medium tracking-tighter text-foreground mb-6 text-wrap-balance">
            Everything you need to sell at scale.
          </h2>
          <p className="text-muted-foreground text-lg md:text-xl leading-relaxed max-w-2xl">
            A complete commerce platform — not a patchwork of integrations. Security, speed, and intelligence built in from the start.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-12 gap-4 md:gap-6">
          <div 
            className={`group relative overflow-hidden rounded-[2rem] border border-border/40 bg-card/60 backdrop-blur-md md:col-span-7 min-h-[300px] flex flex-col justify-end p-8 md:p-10 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] hover:border-primary/40 hover:bg-card/80 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}
          >
            <div className="absolute inset-0 bg-gradient-to-br from-primary/10 via-background to-background opacity-50 group-hover:opacity-100 transition-opacity duration-700" />
            <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-full bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-primary/20 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700 blur-2xl pointer-events-none" />
            
            <div className="relative z-10">
              <h3 className="text-3xl font-medium tracking-tight text-foreground mb-3">Built for scale</h3>
              <p className="text-muted-foreground leading-relaxed max-w-md">
                Enterprise-grade infrastructure that handles millions of requests. No performance cliffs, no surprise outages — just reliable uptime when you need it most.
              </p>
            </div>
          </div>

          <div 
            className={`group relative overflow-hidden rounded-[2rem] border border-border/40 bg-card/60 backdrop-blur-md md:col-span-5 min-h-[300px] p-8 md:p-10 transition-all duration-700 delay-100 ease-[cubic-bezier(0.32,0.72,0,1)] hover:border-border/80 hover:bg-card/80 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}
          >
            <div className="absolute inset-0 bg-gradient-to-b from-transparent to-black/5 dark:to-white/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
            <div className="relative z-10 h-full flex flex-col">
              <div className="inline-flex h-12 w-12 items-center justify-center rounded-full bg-muted/50 text-foreground mb-auto">
                <ShieldCheck weight="light" className="size-6" />
              </div>
              <div className="mt-8">
                <h3 className="text-xl font-medium tracking-tight text-foreground mb-2">Bank-grade security</h3>
                <p className="text-sm text-muted-foreground leading-relaxed">
                  Automatic SSL, encrypted data at rest and in transit, and built-in fraud signals at every transaction layer.
                </p>
              </div>
            </div>
          </div>

          <div 
            className={`group relative overflow-hidden rounded-[2rem] border border-border/40 bg-card/60 backdrop-blur-md md:col-span-5 min-h-[300px] p-8 md:p-10 transition-all duration-700 delay-200 ease-[cubic-bezier(0.32,0.72,0,1)] hover:border-border/80 hover:bg-card/80 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}
          >
            <div className="relative z-10 h-full flex flex-col justify-between">
              <div className="text-[5rem] font-medium tracking-tighter leading-none text-primary/80 group-hover:text-primary transition-colors duration-500">
                &lt;50<span className="text-4xl text-primary/50 group-hover:text-primary/70">ms</span>
              </div>
              <div>
                <h3 className="text-xl font-medium tracking-tight text-foreground mb-2">Global edge network</h3>
                <p className="text-sm text-muted-foreground leading-relaxed">
                  Edge-deployed APIs that respond before your customers blink.
                </p>
              </div>
            </div>
          </div>

          <div 
            className={`group relative overflow-hidden rounded-[2rem] border border-border/40 bg-card/60 backdrop-blur-md md:col-span-7 min-h-[300px] flex flex-col sm:flex-row items-center p-8 md:p-10 transition-all duration-700 delay-150 ease-[cubic-bezier(0.32,0.72,0,1)] hover:border-border/80 hover:bg-card/80 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}
          >
            <div className="relative z-10 sm:w-1/2 flex flex-col justify-center mb-8 sm:mb-0 pr-6">
              <div className="inline-flex h-12 w-12 items-center justify-center rounded-full bg-muted/50 text-foreground mb-6">
                <ChartLineUp weight="light" className="size-6" />
              </div>
              <h3 className="text-xl font-medium tracking-tight text-foreground mb-2">Real-time analytics</h3>
              <p className="text-sm text-muted-foreground leading-relaxed">
                Revenue trends, cohort retention, and funnel analysis — updated in real time so you can act on data the moment it matters.
              </p>
            </div>
            <div className="relative w-full sm:w-1/2 h-40 sm:h-full border border-border/30 rounded-xl overflow-hidden bg-muted/20">
              <div className="absolute inset-x-0 bottom-0 h-[60%] bg-gradient-to-t from-primary/20 to-transparent flex items-end px-4 pb-4 gap-2">
                {[40, 70, 45, 90, 65, 100].map((h, i) => (
                  <div key={i} className="w-full bg-primary/40 rounded-t-sm transition-all duration-700 group-hover:bg-primary/70" style={{ height: `${h}%`, transitionDelay: `${i * 50}ms` }} />
                ))}
              </div>
            </div>
          </div>

          <div 
            className={`group relative overflow-hidden rounded-[2rem] border border-border/40 bg-card/60 backdrop-blur-md md:col-span-12 p-8 md:p-10 flex flex-col md:flex-row items-start md:items-center justify-between transition-all duration-700 delay-300 ease-[cubic-bezier(0.32,0.72,0,1)] hover:border-primary/40 hover:bg-card/80 ${visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-12"}`}
          >
             <div className="absolute right-0 top-0 bottom-0 w-1/2 pointer-events-none">
              <div className="absolute inset-0 bg-gradient-to-l from-primary/[0.08] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700" />
            </div>

            <div className="flex flex-col md:flex-row items-start md:items-center gap-6 relative z-10 w-full">
              <div className="flex-shrink-0 inline-flex h-14 w-14 items-center justify-center rounded-2xl bg-muted/60 border border-border/50 text-muted-foreground group-hover:text-primary group-hover:border-primary/40 group-hover:bg-primary/10 transition-all duration-500">
                <Robot weight="light" className="size-7" />
              </div>
              
              <div className="flex-1">
                <h3 className="text-2xl font-medium tracking-tight text-foreground mb-2">AI-powered ops &amp; Custom domains</h3>
                <p className="text-base text-muted-foreground leading-relaxed max-w-[70ch]">
                  Automate catalog enrichment, support triage, and SEO metadata with dedicated AI tooling. Connect your domain and build a brand identity that's entirely yours — no subdomains, no watermarks.
                </p>
              </div>

              <div className="flex-shrink-0 mt-6 md:mt-0">
                <button className="flex items-center gap-2 rounded-full border border-border/60 bg-muted/30 pl-5 pr-4 py-2.5 text-sm font-medium text-foreground transition-all duration-500 hover:bg-muted/60 hover:border-border">
                  Explore features <ArrowUpRight className="size-3.5 opacity-70 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                </button>
              </div>
            </div>
          </div>

        </div>
      </div>
    </section>
  )
}
