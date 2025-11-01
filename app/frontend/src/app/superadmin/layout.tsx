import { TooltipProvider } from '@/components/ui/tooltip'

export const metadata = {
  title: 'Tenant Control | SaaS Superadmin',
  description: 'Monitor and administer all tenant shops across the platform.',
}

export default function SuperadminLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <TooltipProvider delay={0}>
      {children}
    </TooltipProvider>
  )
}
