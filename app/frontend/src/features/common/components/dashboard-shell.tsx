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
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import React from "react"

type DashboardShellProps = {
  title: string
  description: string
  children: ReactNode
}

export function DashboardShell({ title, description, children }: DashboardShellProps) {
  const pathname = usePathname()
  const paths = pathname.split("/").filter(Boolean)

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset className="bg-zinc-950 min-h-[100dvh] flex flex-col">
        <header className="sticky top-0 z-50 flex h-16 shrink-0 items-center gap-4 bg-zinc-950/80 backdrop-blur-xl px-6 lg:px-12 border-b border-zinc-800/50">
          <SidebarTrigger className="-ml-2 text-zinc-400 hover:text-zinc-50 transition-colors" />
          <div className="h-4 w-px bg-zinc-800" />
          <Breadcrumb>
            <BreadcrumbList>
              {paths.map((path, index) => {
                const isLast = index === paths.length - 1
                const pathTitle = path.charAt(0).toUpperCase() + path.slice(1)
                
                return (
                  <React.Fragment key={path}>
                    <BreadcrumbItem className="hidden md:block">
                      {isLast ? (
                        <BreadcrumbPage className="font-medium text-zinc-100 tracking-tight">
                          {pathTitle}
                        </BreadcrumbPage>
                      ) : (
                        <BreadcrumbLink 
                          href={`/${paths.slice(0, index + 1).join('/')}`}
                          className="text-zinc-500 transition-colors hover:text-zinc-200"
                        >
                          {pathTitle}
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
            
            <div className="flex flex-col gap-3 mb-12">
              <span className="rounded-full px-3 py-1 text-[10px] uppercase tracking-[0.2em] font-medium text-zinc-400 border border-zinc-800 w-max">
                Platform Admin
              </span>
              <h1 className="text-4xl font-semibold tracking-tight text-zinc-50">{title}</h1>
              <p className="text-zinc-400 text-lg">
                {description}
              </p>
            </div>

            {children}
            
          </div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
