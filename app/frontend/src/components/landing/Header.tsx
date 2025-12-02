"use client"

import Link from "next/link"
import { useEffect, useState } from "react"
import { Menu, X } from "lucide-react"

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
        className={`fixed top-0 left-0 right-0 z-50 transition-all duration-500 ${scrolled
            ? "top-4 mx-auto max-w-4xl left-4 right-4 md:left-auto md:right-auto rounded-full border border-white/10 bg-black/70 backdrop-blur-xl shadow-2xl shadow-black/40"
            : "bg-transparent"
          }`}
      >
        <div className={`flex items-center justify-between px-6 ${scrolled ? "h-14" : "h-20"} transition-all duration-500`}>
          <Link href="/" className="flex items-center gap-2.5 group">
            <div className="relative size-8 rounded-lg bg-gradient-to-br from-primary to-purple-400 flex items-center justify-center shadow-lg shadow-primary/25 transition-transform duration-300 group-hover:scale-110 group-hover:rotate-6">
              <span className="text-sm font-black text-white">N</span>
            </div>
            <span className="font-bold tracking-tight text-lg text-white">
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
                className="relative px-4 py-2 text-sm font-medium text-zinc-400 hover:text-white transition-colors duration-300 group"
              >
                {item.label}
                <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-0 h-0.5 bg-gradient-to-r from-primary to-purple-400 rounded-full transition-all duration-300 group-hover:w-2/3" />
              </Link>
            ))}
          </nav>

          <div className="flex items-center gap-3">
            <Link
              href="/superadmin/login"
              className="hidden md:inline-flex relative overflow-hidden items-center justify-center px-5 py-2 text-sm font-medium rounded-full bg-white text-black hover:bg-zinc-200 transition-all duration-300 shadow-lg shadow-white/10 hover:shadow-white/20 btn-shimmer"
            >
              Get Started
            </Link>
            <button
              onClick={() => setMobileOpen(!mobileOpen)}
              className="md:hidden p-2 text-zinc-400 hover:text-white transition-colors"
              aria-label="Toggle menu"
            >
              {mobileOpen ? <X className="size-5" /> : <Menu className="size-5" />}
            </button>
          </div>
        </div>
      </header>

      {/* Mobile menu overlay */}
      {mobileOpen && (
        <div className="fixed inset-0 z-40 bg-black/95 backdrop-blur-xl flex flex-col items-center justify-center gap-8 md:hidden">
          {[
            { label: "Features", href: "#features" },
            { label: "Pricing", href: "#pricing" },
            { label: "Testimonials", href: "#testimonials" },
          ].map((item) => (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => setMobileOpen(false)}
              className="text-2xl font-medium text-zinc-300 hover:text-white transition-colors"
            >
              {item.label}
            </Link>
          ))}
          <Link
            href="/superadmin/login"
            onClick={() => setMobileOpen(false)}
            className="mt-4 px-8 py-3 bg-white text-black rounded-full font-medium text-lg"
          >
            Get Started
          </Link>
        </div>
      )}
    </>
  )
}
