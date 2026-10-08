"use client"

import Link from "next/link"
import { ArrowRight, Play } from "@phosphor-icons/react"
import { useEffect, useRef, useState } from "react"
import Image from "next/image"



function HeroAsset() {
  return (
    <div className="relative w-full max-w-2xl mx-auto z-10">
      <div className="p-2 rounded-[2rem] bg-muted/20 border border-border/60 ring-1 ring-white/[0.04] shadow-sm ">
        <div className="relative overflow-hidden rounded-[calc(2rem-0.5rem)] bg-card shadow-[inset_0_1px_1px_rgba(255,255,255,0.08)] aspect-[16/9]">
          <Image
            src="https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=2564&auto=format&fit=crop"
            alt="Nexus Commerce Engine"
            fill
            className="object-cover opacity-90 transition-transform duration-1000 ease-[cubic-bezier(0.32,0.72,0,1)] hover:scale-105"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent pointer-events-none" />
        </div>
      </div>

      <div className="absolute -bottom-4 -left-4 flex items-center gap-2.5 px-4 py-2.5 rounded-2xl border border-border/60 bg-card/90 backdrop-blur-xl shadow-sm ">
        <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
        <span className="text-xs font-medium text-foreground">500+ businesses scaling</span>
      </div>

      <div className="absolute -top-4 -right-4 flex items-center gap-2 px-3.5 py-2 rounded-2xl border border-primary/30 bg-primary/10 backdrop-blur-xl shadow-sm">
        <span className="text-[11px] font-semibold text-primary tracking-tight">&lt;50ms</span>
        <span className="text-[10px] text-primary/70">global</span>
      </div>
    </div>
  )
}

export function Hero() {
  return (
    <section className="relative w-full flex items-center overflow-hidden bg-transparent pt-28 pb-24 md:pb-32">
      <div className="absolute inset-0 pointer-events-none opacity-[0.025] z-0" style={{ backgroundImage: "url('https://grainy-gradients.vercel.app/noise.svg')" }} />

      <div className="container relative mx-auto px-4 md:px-6 z-10 max-w-7xl">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 lg:gap-12 items-center">
          <div className="flex flex-col items-start text-left gap-6 animate-fade-in-up">
            <div className="inline-flex items-center gap-2 rounded-full px-3 py-1 text-[10px] uppercase tracking-[0.2em] font-semibold border border-primary/25 bg-primary/8 text-primary backdrop-blur-md">
              <span className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
              Now in public beta
            </div>

            <h1 className="text-[4.5rem] md:text-[6rem] lg:text-[7.5rem] font-medium tracking-tighter leading-[1] text-foreground text-wrap-balance">
              Scale your <br />
              <span className="text-muted-foreground/60">commerce.</span>
            </h1>

            <p className="max-w-[42ch] text-muted-foreground/75 text-base sm:text-lg leading-relaxed">
              The premium commerce engine with bank-grade security, sub-50ms performance, and the structural depth you need to build, manage, and scale.
            </p>

            <div className="flex gap-3 mt-4 w-full sm:w-auto items-center">
              <Link
                href="/superadmin/login"
                className="group flex items-center gap-4 rounded-full bg-primary pl-6 pr-2 py-2 text-base font-medium text-primary-foreground transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98] shadow-[0_0_30px_-5px_oklch(0.6_0.28_260_/_0.4)] hover:shadow-[0_0_40px_-5px_oklch(0.6_0.28_260_/_0.6)] btn-shimmer overflow-hidden relative"
              >
                <span>Start free trial</span>
                <div className="w-8 h-8 rounded-full bg-primary-foreground/20 flex items-center justify-center transition-transform duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] group-hover:translate-x-1 group-hover:-translate-y-[1px] group-hover:scale-105">
                  <ArrowRight weight="bold" className="size-4 text-primary-foreground" />
                </div>
              </Link>
            </div>
          </div>

          <div className="relative animate-fade-in-up" style={{ animationDelay: "200ms" }}>
            <HeroAsset />
          </div>
        </div>
      </div>
    </section>
  )
}
