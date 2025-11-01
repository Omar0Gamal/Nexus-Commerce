"use client"

import type { ReactNode } from "react"
import { usePathname } from "next/navigation"
import { AppSidebar } from "./app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Clock } from "lucide-react"

type DashboardShellProps = {
  title: string
  description: string
  children: ReactNode
}

function formatTime() {
  return new Date().toLocaleTimeString("en-US", {
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  })
}

export function DashboardShell({ title, description, children }: DashboardShellProps) {
  const pathname = usePathname()

  const segments = pathname.split("/").filter(Boolean)
  const currentSegment = segments.length > 1 ? segments[segments.length - 1] : null

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset className="bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-blue-50/40 via-background to-background dark:from-blue-900/10 dark:via-background dark:to-background">
        <header className="relative overflow-hidden border-b border-border/40 dark:border-border/30">
          <div className="dot-grid absolute inset-0 pointer-events-none" />

          <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-primary/80 via-primary/40 to-transparent" />

          <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.03] via-transparent to-transparent dark:from-primary/[0.06]" />

          <div className="relative flex flex-col gap-3 p-4 sm:p-6 lg:py-6 mx-auto w-full max-w-7xl">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <SidebarTrigger className="-ml-1 text-muted-foreground hover:text-foreground" />
                <Separator orientation="vertical" className="h-4" />
                <Breadcrumb>
                  <BreadcrumbList>
                    <BreadcrumbItem>
                      <BreadcrumbLink href="/superadmin" className="text-xs text-muted-foreground hover:text-primary transition-colors">
                        Dashboard
                      </BreadcrumbLink>
                    </BreadcrumbItem>
                    {currentSegment && (
                      <>
                        <BreadcrumbSeparator />
                        <BreadcrumbItem>
                          <BreadcrumbPage className="text-xs font-medium capitalize">
                            {currentSegment}
                          </BreadcrumbPage>
                        </BreadcrumbItem>
                      </>
                    )}
                  </BreadcrumbList>
                </Breadcrumb>
              </div>

              <div className="hidden sm:flex items-center gap-1.5 text-[11px] text-muted-foreground/70 bg-muted/50 dark:bg-muted/30 px-2.5 py-1 rounded-full border border-border/30 backdrop-blur-sm">
                <Clock className="size-3" />
                <span>{formatTime()}</span>
              </div>
            </div>

            <div className="flex flex-col gap-1 mt-1">
              <h1 className="text-2xl font-bold tracking-tight bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text">
                {title}
              </h1>
              <p className="text-sm text-muted-foreground/80 max-w-xl">
                {description}
              </p>
            </div>
          </div>
        </header>

        <main className="flex flex-1 flex-col gap-4 p-4 sm:p-6 mx-auto w-full max-w-7xl">
          {children}
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
