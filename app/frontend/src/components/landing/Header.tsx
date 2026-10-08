"use client"

import Link from "next/link"
import { useEffect, useState } from "react"
import { List, X } from "@phosphor-icons/react"

export function Header() {
  const [scrolled, setScrolled] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40)
    window.addEventListener("scroll", onScroll, { passive: true })
    return () => window.removeEventListener("scroll", onScroll)
  }, [])

  return (
    <>
      <header
        className={`fixed z-50 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] left-4 right-4 md:left-auto md:right-auto md:mx-auto md:inset-x-0 rounded-full border border-border/40 backdrop-blur-2xl ${
          scrolled
            ? "top-4 md:max-w-[820px] bg-background/70 shadow-[0_8px_32px_rgba(0,0,0,0.4)] ring-1 ring-white/[0.05]"
            : "top-6 md:max-w-[1000px] bg-background/40 shadow-sm"
        }`}
      >
        <div className="absolute inset-0 rounded-full pointer-events-none border border-white/[0.04] mask-image-[linear-gradient(to_bottom,black,transparent)]" />
        <div className={`flex items-center justify-between px-6 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] ${scrolled ? "h-14" : "h-20"}`}>
          <Link href="/" className="flex items-center gap-3 group">
            <div className="relative size-9 rounded-full bg-gradient-to-br from-primary/20 to-primary/5 border border-primary/20 flex items-center justify-center transition-all duration-500 group-hover:scale-105 group-hover:bg-primary/20 group-hover:border-primary/40 shadow-[inset_0_1px_1px_rgba(255,255,255,0.1)]">
              <span className="text-base font-bold bg-gradient-to-br from-white to-white/70 bg-clip-text text-transparent">N</span>
            </div>
            <span className="font-bold tracking-tight text-lg text-foreground">
              Nexus
            </span>
          </Link>

          <nav className="hidden md:flex items-center gap-1">
            {[
              { label: "Features", href: "#features" },
              { label: "Pricing", href: "#pricing" },
              { label: "Testimonials", href: "#testimonials" },
            ].map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="relative px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-all duration-300 group"
              >
                {item.label}
                <span className="absolute inset-x-4 -bottom-px h-px bg-gradient-to-r from-transparent via-primary/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              </Link>
            ))}
          </nav>

          <div className="flex items-center gap-3">
            <Link
              href="/superadmin/login"
              className="hidden md:inline-flex relative overflow-hidden items-center gap-3 justify-center pl-5 pr-2 py-1.5 text-sm font-medium rounded-full bg-primary text-primary-foreground hover:opacity-90 transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98] shadow-sm shadow-primary/25 group btn-shimmer"
            >
              <span>Start free trial</span>
              <div className="w-7 h-7 rounded-full bg-primary-foreground/20 flex items-center justify-center transition-transform duration-500 group-hover:translate-x-0.5 group-hover:-translate-y-px group-hover:scale-105">
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M2.5 6H9.5M9.5 6L6.5 3M9.5 6L6.5 9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
            </Link>
            <button
              onClick={() => setMobileOpen(!mobileOpen)}
              className="md:hidden p-2 text-muted-foreground hover:text-foreground transition-colors"
              aria-label="Toggle menu"
            >
              {mobileOpen ? <X className="size-5" weight="bold" /> : <List className="size-5" weight="bold" />}
            </button>
          </div>
        </div>
      </header>

      {mobileOpen && (
        <div className="fixed inset-0 z-40 bg-background/95 backdrop-blur-2xl flex flex-col items-center justify-center gap-8 md:hidden">
          {[
            { label: "Features", href: "#features" },
            { label: "Pricing", href: "#pricing" },
            { label: "Testimonials", href: "#testimonials" },
          ].map((item, i) => (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => setMobileOpen(false)}
              className="text-3xl font-medium text-foreground/70 hover:text-foreground transition-all duration-500 animate-fade-in-up"
              style={{ animationDelay: `${i * 100 + 100}ms` }}
            >
              {item.label}
            </Link>
          ))}
          <Link
            href="/superadmin/login"
            onClick={() => setMobileOpen(false)}
            className="mt-4 flex items-center gap-3 pl-7 pr-2 py-2 bg-primary text-primary-foreground rounded-full font-medium text-lg shadow-sm shadow-primary/25 group animate-fade-in-up"
            style={{ animationDelay: `${3 * 100 + 100}ms` }}
          >
            <span>Start free trial</span>
            <div className="w-9 h-9 rounded-full bg-primary-foreground/20 flex items-center justify-center transition-transform duration-500 group-hover:translate-x-0.5">
              <svg width="14" height="14" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M2.5 6H9.5M9.5 6L6.5 3M9.5 6L6.5 9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
          </Link>
        </div>
      )}
    </>
  )
}
