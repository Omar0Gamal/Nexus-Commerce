"use client"

import { useReveal } from "@/hooks/useReveal"

const companies = [
  { name: "Shopify",   slug: "shopify" },
  { name: "Stripe",    slug: "stripe" },
  { name: "Vercel",    slug: "vercel" },
  { name: "Supabase",  slug: "supabase" },
  { name: "GitHub",    slug: "github" },
  { name: "AWS",       slug: "amazonwebservices" },
  { name: "Cloudflare",slug: "cloudflare" },
  { name: "Notion",    slug: "notion" },
  { name: "Linear",    slug: "linear" },
  { name: "Figma",     slug: "figma" },
  { name: "Slack",     slug: "slack" },
  { name: "Discord",   slug: "discord" },
]

function BrandLogo({ slug, name }: { slug: string; name: string }) {
  return (
    <div className="flex items-center gap-2 opacity-40 group-hover:opacity-100 transition-opacity duration-500">
      <span className="text-xl md:text-2xl font-bold tracking-tight text-foreground/80 lowercase select-none">
        {name}
      </span>
    </div>
  )
}

export function TrustedBy() {
  const [ref, visible] = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      className={`relative w-full py-16 bg-transparent overflow-hidden transition-all duration-1000 ease-[cubic-bezier(0.32,0.72,0,1)] ${visible ? "opacity-100" : "opacity-0"}`}
    >
      <div className="container mx-auto px-4 md:px-6 max-w-7xl mb-10">
        <h2 className="text-center text-xl md:text-2xl font-medium tracking-tight text-foreground/80">
          Built on infrastructure you already trust
        </h2>
      </div>

      <div className="relative">
        <div className="absolute inset-0 z-10 pointer-events-none" 
             style={{ 
               maskImage: 'linear-gradient(to right, transparent, black 15%, black 85%, transparent)',
               WebkitMaskImage: 'linear-gradient(to right, transparent, black 15%, black 85%, transparent)' 
             }}>
        </div>

        <div className="flex overflow-hidden relative">
          <div className="w-full" style={{ maskImage: 'linear-gradient(to right, transparent, black 15%, black 85%, transparent)', WebkitMaskImage: 'linear-gradient(to right, transparent, black 15%, black 85%, transparent)' }}>
             <div className="animate-marquee flex shrink-0 items-center gap-16 md:gap-24 pr-16 md:pr-24">
                {[...companies, ...companies].map((company, i) => (
                  <div
                    key={`${company.slug}-${i}`}
                    className="group flex items-center justify-center transition-all duration-500 cursor-default"
                  >
                    <BrandLogo slug={company.slug} name={company.name} />
                  </div>
                ))}
             </div>
          </div>
        </div>
      </div>
    </section>
  )
}
