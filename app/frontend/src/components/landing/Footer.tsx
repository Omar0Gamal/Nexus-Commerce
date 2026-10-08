import Link from "next/link"
import { TwitterLogo, GithubLogo, DiscordLogo } from "@phosphor-icons/react/dist/ssr"

export function Footer() {
  return (
    <footer className="w-full relative overflow-hidden">
      <div className="h-px bg-gradient-to-r from-transparent via-primary/20 to-transparent" />

      <div className="container mx-auto px-4 md:px-6 max-w-7xl py-16">
        <div className="grid gap-12 md:grid-cols-2 lg:grid-cols-12">
          <div className="lg:col-span-4 flex flex-col gap-4">
            <Link href="/" className="flex items-center gap-3 group w-fit">
              <div className="relative size-9 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center transition-all duration-500 group-hover:scale-105 group-hover:bg-primary/20">
                <span className="text-base font-bold text-primary">N</span>
              </div>
              <span className="font-bold tracking-tight text-lg text-foreground">
                Nexus Commerce
              </span>
            </Link>
            <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
              The modern e-commerce platform for businesses of all sizes. Built for speed, scale, and conversion.
            </p>

            <div className="mt-2">
              <p className="text-xs font-medium text-muted-foreground/70 mb-2.5">Stay in the loop</p>
              <div className="flex gap-2">
                <input
                  type="email"
                  placeholder="your@email.com"
                  className="flex-1 px-4 py-2.5 text-sm bg-muted/20 border border-border/40 rounded-full text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:border-primary/50 focus:bg-muted/30 focus:shadow-[0_0_15px_-3px_oklch(0.6_0.28_260_/_0.3)] transition-all"
                />
                <button className="px-5 py-2 text-sm font-medium bg-primary text-primary-foreground rounded-full hover:opacity-90 transition-all active:scale-[0.98] shrink-0 shadow-sm shadow-primary/20">
                  Subscribe
                </button>
              </div>
            </div>
          </div>

          <div className="lg:col-span-8 grid grid-cols-2 sm:grid-cols-3 gap-8">
            <div className="flex flex-col gap-3">
              <h3 className="text-[10px] font-semibold text-foreground/50 uppercase tracking-[0.2em] mb-1">Platform</h3>
              {["Features", "Pricing", "Integrations", "API Docs", "Changelog"].map((link) => (
                <Link key={link} href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors duration-200 w-fit hover:translate-x-0.5 transform-gpu transition-transform">
                  {link}
                </Link>
              ))}
            </div>
            <div className="flex flex-col gap-3">
              <h3 className="text-[10px] font-semibold text-foreground/50 uppercase tracking-[0.2em] mb-1">Company</h3>
              {["About Us", "Careers", "Blog", "Contact", "Partners"].map((link) => (
                <Link key={link} href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors duration-200 w-fit">
                  {link}
                </Link>
              ))}
            </div>
            <div className="flex flex-col gap-3">
              <h3 className="text-[10px] font-semibold text-foreground/50 uppercase tracking-[0.2em] mb-1">Legal</h3>
              {["Privacy Policy", "Terms of Service", "Cookie Policy", "GDPR", "Security"].map((link) => (
                <Link key={link} href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors duration-200 w-fit">
                  {link}
                </Link>
              ))}
            </div>
          </div>
        </div>

        <div className="mt-16 pt-8 border-t border-border/30 flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-xs text-muted-foreground/50">
            © 2026 Nexus Commerce. All rights reserved.
          </p>
          <div className="flex items-center gap-2">
            {[
              { name: "Twitter", Icon: TwitterLogo },
              { name: "GitHub", Icon: GithubLogo },
              { name: "Discord", Icon: DiscordLogo },
            ].map(({ name, Icon }) => (
              <Link
                key={name}
                href="#"
                className="size-8 rounded-full bg-muted/50 border border-border/40 flex items-center justify-center text-muted-foreground/60 hover:text-foreground hover:bg-muted/80 hover:border-border transition-all duration-300"
                aria-label={name}
              >
                <Icon weight="fill" className="size-3.5" />
              </Link>
            ))}
          </div>
        </div>
      </div>
    </footer>
  )
}
