import { Header } from "@/components/landing/Header"
import { Hero } from "@/components/landing/Hero"
import { TrustedBy } from "@/components/landing/TrustedBy"
import { Features } from "@/components/landing/Features"
import { Stats } from "@/components/landing/Stats"
import { Pricing } from "@/components/landing/Pricing"
import { Testimonials } from "@/components/landing/Testimonials"
import { CTA } from "@/components/landing/CTA"
import { Footer } from "@/components/landing/Footer"
import { GlobalGrid } from "@/components/landing/GlobalGrid"

export default function Home() {
  return (
    <div className="flex min-h-[100dvh] flex-col font-sans text-foreground selection:bg-primary/30 selection:text-primary relative z-0 overflow-x-hidden">
      <GlobalGrid />
      <Header />
      <main className="flex-1">
        <Hero />
        <TrustedBy />
        <Features />
        <Stats />
        <Pricing />
        <Testimonials />
        <CTA />
      </main>
      <Footer />
    </div>
  )
}
