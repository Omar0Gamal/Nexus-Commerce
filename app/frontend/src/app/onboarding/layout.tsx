import { ReactNode } from "react"
import { TooltipProvider } from "@/components/ui/tooltip"

export const metadata = {
  title: 'Setup Your Store | Nexus Commerce',
  description: 'Onboard and configure your new tenant store.',
}

export default function OnboardingLayout({
  children,
}: Readonly<{
  children: ReactNode
}>) {
  return (
    <TooltipProvider delay={0}>
      <main className="min-h-screen w-full bg-zinc-950">
        {children}
      </main>
    </TooltipProvider>
  )
}
