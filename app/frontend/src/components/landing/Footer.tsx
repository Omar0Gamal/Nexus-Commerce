import Link from "next/link"

export function Footer() {
  return (
    <footer className="w-full bg-black relative overflow-hidden">
      {/* Top gradient border */}
      <div className="h-px bg-gradient-to-r from-transparent via-primary/30 to-transparent" />

      <div className="container mx-auto px-4 md:px-6 max-w-7xl py-16">
        <div className="grid gap-12 md:grid-cols-2 lg:grid-cols-12">
          {/* Brand column */}
          <div className="lg:col-span-4 flex flex-col gap-4">
            <Link href="/" className="flex items-center gap-2.5 group w-fit">
              <div className="size-8 rounded-lg bg-gradient-to-br from-primary to-purple-400 flex items-center justify-center shadow-lg shadow-primary/20 transition-transform duration-300 group-hover:scale-110 group-hover:rotate-6">
                <span className="text-sm font-black text-white">N</span>
              </div>
              <span className="font-bold tracking-tight text-lg text-white">
                Nexus Commerce
              </span>
            </Link>
            <p className="text-sm text-zinc-500 max-w-xs leading-relaxed">
              The modern e-commerce platform for businesses of all sizes. Built for speed, scale, and conversion.
            </p>

            {/* Newsletter */}
            <div className="mt-2">
              <p className="text-xs font-medium text-zinc-400 mb-2">Stay in the loop</p>
              <div className="flex gap-2">
                <input
                  type="email"
                  placeholder="your@email.com"
                  className="flex-1 px-4 py-2 text-sm bg-white/5 border border-white/10 rounded-lg text-white placeholder:text-zinc-600 focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
                />
                <button className="px-4 py-2 text-sm font-medium bg-white text-black rounded-lg hover:bg-zinc-200 transition-colors shrink-0">
                  Subscribe
                </button>
              </div>
            </div>
          </div>

          {/* Links columns */}
          <div className="lg:col-span-8 grid grid-cols-2 sm:grid-cols-3 gap-8">
            <div className="flex flex-col gap-3">
              <h3 className="text-xs font-semibold text-zinc-300 uppercase tracking-wider mb-1">Platform</h3>
              {["Features", "Pricing", "Integrations", "API Docs", "Changelog"].map((link) => (
                <Link key={link} href="#" className="text-sm text-zinc-500 hover:text-white transition-colors duration-200 w-fit">
                  {link}
                </Link>
              ))}
            </div>
            <div className="flex flex-col gap-3">
              <h3 className="text-xs font-semibold text-zinc-300 uppercase tracking-wider mb-1">Company</h3>
              {["About Us", "Careers", "Blog", "Contact", "Partners"].map((link) => (
                <Link key={link} href="#" className="text-sm text-zinc-500 hover:text-white transition-colors duration-200 w-fit">
                  {link}
                </Link>
              ))}
            </div>
            <div className="flex flex-col gap-3">
              <h3 className="text-xs font-semibold text-zinc-300 uppercase tracking-wider mb-1">Legal</h3>
              {["Privacy Policy", "Terms of Service", "Cookie Policy", "GDPR", "Security"].map((link) => (
                <Link key={link} href="#" className="text-sm text-zinc-500 hover:text-white transition-colors duration-200 w-fit">
                  {link}
                </Link>
              ))}
            </div>
          </div>
        </div>

        {/* Bottom bar */}
        <div className="mt-16 pt-8 border-t border-white/5 flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-xs text-zinc-600">
            © 2026 Nexus Commerce. All rights reserved.
          </p>
          <div className="flex items-center gap-5">
            {[
              { name: "Twitter", icon: "𝕏" },
              { name: "GitHub", icon: "⌘" },
              { name: "Discord", icon: "◆" },
            ].map((social) => (
              <Link
                key={social.name}
                href="#"
                className="size-8 rounded-full bg-white/5 border border-white/[0.06] flex items-center justify-center text-xs text-zinc-500 hover:text-white hover:bg-white/10 hover:border-white/15 transition-all duration-300"
                aria-label={social.name}
              >
                {social.icon}
              </Link>
            ))}
          </div>
        </div>
      </div>
    </footer>
  )
}
