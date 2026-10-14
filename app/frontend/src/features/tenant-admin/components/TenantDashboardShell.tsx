"use client"

import * as React from "react"
import { usePathname } from "next/navigation"
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { useBreadcrumbStore } from "@/store/breadcrumb"
import { TenantSidebar } from "./TenantSidebar"

interface TenantDashboardShellProps {
  children: React.ReactNode
}

export function TenantDashboardShell({ children }: TenantDashboardShellProps) {
  const pathname = usePathname()
  const titleOverrides = useBreadcrumbStore((state) => state.titleOverrides)
  
  const paths = pathname.split('/').filter(Boolean)
  
  return (
    <SidebarProvider>
      <TenantSidebar />
      <SidebarInset className="bg-zinc-950 min-h-[100dvh] flex flex-col">
        <header className="sticky top-0 z-50 flex h-16 shrink-0 items-center gap-4 bg-zinc-950/80 backdrop-blur-xl px-6 lg:px-12 border-b border-zinc-800/50">
          <SidebarTrigger className="-ml-2 text-zinc-400 hover:text-zinc-50 transition-colors" />
          <div className="h-4 w-px bg-zinc-800" />
          <Breadcrumb>
            <BreadcrumbList>
              {paths.map((path, index) => {
                const isLast = index === paths.length - 1
                const isUUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(path)
                const override = titleOverrides[path.toLowerCase()]
                let title = override
                if (!title) {
                  if (isUUID) {
                    const prev = index > 0 ? paths[index - 1].toLowerCase() : ""
                    if (prev === "orders") {
                      title = `Order #${path.slice(0, 8)}`
                    } else if (prev === "products" || prev === "catalog") {
                      title = "Product Details"
                    } else {
                      title = path.slice(0, 8)
                    }
                  } else {
                    title = path.charAt(0).toUpperCase() + path.slice(1)
                  }
                }
                
                return (
                  <React.Fragment key={path}>
                    <BreadcrumbItem className="hidden md:block">
                      {isLast ? (
                        <BreadcrumbPage className="font-medium text-zinc-100 tracking-tight">
                          {title}
                        </BreadcrumbPage>
                      ) : (
                        <BreadcrumbLink 
                          href={`/${paths.slice(0, index + 1).join('/')}`}
                          className="text-zinc-500 transition-colors hover:text-zinc-200"
                        >
                          {title}
                        </BreadcrumbLink>
                      )}
                    </BreadcrumbItem>
                    {!isLast && (
                      <BreadcrumbSeparator className="hidden md:block text-zinc-700" />
                    )}
                  </React.Fragment>
                )
              })}
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        <main className="flex-1 p-6 md:p-8 lg:p-12 xl:p-16">
          <div className="mx-auto w-full max-w-7xl animate-in fade-in slide-in-from-bottom-8 duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]">
            {children}
          </div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
