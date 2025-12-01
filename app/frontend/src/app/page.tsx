import { Header } from "@/components/landing/Header"
import { Hero } from "@/components/landing/Hero"
import { TrustedBy } from "@/components/landing/TrustedBy"
import { Features } from "@/components/landing/Features"
import { Stats } from "@/components/landing/Stats"
import { Pricing } from "@/components/landing/Pricing"
import { Testimonials } from "@/components/landing/Testimonials"
import { CTA } from "@/components/landing/CTA"
import { Footer } from "@/components/landing/Footer"

export default function Home() {
  return (
    <div className="dark flex min-h-screen flex-col font-sans bg-black text-zinc-50 selection:bg-primary/30 selection:text-primary">
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
