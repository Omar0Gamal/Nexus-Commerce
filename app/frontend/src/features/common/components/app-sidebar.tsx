"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  Store,
  BarChart3,
  CreditCard,
  ScrollText,
  Activity,
  ShieldCheck,
  ChevronsUpDown,
  Moon,
  Sun,
  LayoutDashboard,
} from "lucide-react"

import { useTheme } from "next-themes"

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

type NavItem = {
  title: string
  href: string
  icon: React.ComponentType<{ className?: string }>
  badge?: string
}

const platformNav: NavItem[] = [
  { title: "Overview", href: "/superadmin", icon: LayoutDashboard },
  { title: "Tenants", href: "/superadmin/tenants", icon: Store },
  { title: "Analytics", href: "/superadmin/analytics", icon: BarChart3 },
  { title: "Billing", href: "/superadmin/billing", icon: CreditCard },
  { title: "Audit Log", href: "#", icon: ScrollText, badge: "3" },
]

const systemNav: NavItem[] = [
  { title: "Infrastructure", href: "/superadmin/infrastructure", icon: Activity },
]

export function AppSidebar() {
  const pathname = usePathname()
  const { setTheme, theme } = useTheme()

  const renderNav = (items: NavItem[]) =>
    items.map((item) => {
      const isActive = item.href !== "#" && pathname === item.href
      return (
        <SidebarMenuItem key={item.title} className="group-data-[collapsible=icon]:flex group-data-[collapsible=icon]:justify-center">
          <SidebarMenuButton
            isActive={isActive}
            tooltip={item.title}
            className={`relative transition-all duration-200 rounded-lg ${
              isActive
                ? "bg-primary/10 text-primary font-semibold shadow-sm shadow-primary/10 dark:bg-primary/15 dark:text-primary dark:shadow-primary/20"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/70"
            }`}
            render={<Link href={item.href} />}
          >
            {/* Active indicator bar — hidden in icon mode */}
            {isActive && (
              <span className="absolute left-0 top-1.5 bottom-1.5 w-[3px] rounded-r-full bg-primary group-data-[collapsible=icon]:hidden" />
            )}
            <item.icon className={isActive ? "text-primary" : ""} />
            <span className="group-data-[collapsible=icon]:hidden">{item.title}</span>
          </SidebarMenuButton>
          {item.badge ? (
            <SidebarMenuBadge className="bg-primary/10 text-primary text-[10px] font-semibold border border-primary/20 group-data-[collapsible=icon]:hidden">
              {item.badge}
            </SidebarMenuBadge>
          ) : null}
        </SidebarMenuItem>
      )
    })

  return (
    <Sidebar
      collapsible="icon"
      className="border-r border-border/40 dark:border-border/30 bg-gradient-to-b from-background via-background to-muted/20 dark:from-background dark:via-background dark:to-muted/10"
    >
      <SidebarHeader className="border-b border-border/40 dark:border-border/30 pb-4 pt-4">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              className="hover:bg-transparent cursor-default px-1"
            >
              {/* Logo with glow ring */}
              <div className="relative flex items-center justify-center">
                <div className="absolute inset-0 rounded-lg bg-primary/20 blur-md animate-pulse" />
                <div className="relative flex size-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-primary/80 shadow-lg shadow-primary/25">
                  <ShieldCheck className="size-4.5 text-primary-foreground" />
                </div>
              </div>
              <div className="flex flex-col gap-0.5 leading-none ml-2 group-data-[collapsible=icon]:hidden">
                <span className="font-bold tracking-tight text-foreground text-sm">
                  Nexus Commerce
                </span>
                <span className="text-[9px] text-primary uppercase tracking-[0.2em] font-semibold">
                  Admin
                </span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent className="pt-4 gap-6">
        <SidebarGroup className="px-0">
          {/* Label shown in expanded mode */}
          <SidebarGroupLabel className="text-[10px] uppercase tracking-[0.15em] text-muted-foreground/60 mb-2 font-semibold px-4 group-data-[collapsible=icon]:hidden">
            Platform
          </SidebarGroupLabel>
          {/* Thin divider shown in icon mode as group separator */}
          <div className="hidden group-data-[collapsible=icon]:flex justify-center mb-3">
            <div className="w-5 h-px bg-border/60" />
          </div>
          <SidebarGroupContent>
            <SidebarMenu className="gap-1 px-2 group-data-[collapsible=icon]:px-1 group-data-[collapsible=icon]:gap-1">
              {renderNav(platformNav)}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup className="px-0">
          {/* Label shown in expanded mode */}
          <SidebarGroupLabel className="text-[10px] uppercase tracking-[0.15em] text-muted-foreground/60 mb-2 font-semibold px-4 group-data-[collapsible=icon]:hidden">
            System
          </SidebarGroupLabel>
          {/* Thin divider shown in icon mode as group separator */}
          <div className="hidden group-data-[collapsible=icon]:flex justify-center mb-3">
            <div className="w-5 h-px bg-border/60" />
          </div>
          <SidebarGroupContent>
            <SidebarMenu className="gap-1 px-2 group-data-[collapsible=icon]:px-1 group-data-[collapsible=icon]:gap-1">
              {renderNav(systemNav)}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t border-border/40 dark:border-border/30 pt-4 pb-4 px-2">
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger render={
                <SidebarMenuButton size="lg" className="hover:bg-muted/70 data-[state=open]:bg-muted/70 transition-colors rounded-xl">
                  <Avatar className="size-8 rounded-lg ring-2 ring-primary/10 ring-offset-1 ring-offset-background">
                    <AvatarFallback className="rounded-lg bg-gradient-to-br from-primary/20 to-primary/10 text-primary font-bold text-sm">
                      N
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex flex-col gap-1 leading-none ml-2 group-data-[collapsible=icon]:hidden">
                    <span className="font-semibold text-sm">Ava Reyes</span>
                    <span className="text-[10px] uppercase tracking-wider text-muted-foreground/70">Platform admin</span>
                  </div>
                  <ChevronsUpDown className="ml-auto size-4 text-muted-foreground/50 group-data-[collapsible=icon]:hidden" />
                </SidebarMenuButton>
              } />
              <DropdownMenuContent align="end" className="w-56 mt-2 rounded-xl border border-border/50 dark:border-border/30 shadow-lg">
                <DropdownMenuGroup>
                  <DropdownMenuLabel className="font-normal px-2 py-2">
                    <div className="flex flex-col gap-1">
                      <p className="text-sm font-semibold leading-none">Ava Reyes</p>
                      <p className="text-[10px] uppercase tracking-wider leading-none text-muted-foreground">
                        Platform admin
                      </p>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => setTheme(theme === "dark" ? "light" : "dark")} className="cursor-pointer">
                    {theme === "dark" ? <Sun className="mr-2 size-4" /> : <Moon className="mr-2 size-4" />}
                    Toggle Theme
                  </DropdownMenuItem>
                </DropdownMenuGroup>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuItem 
                    className="cursor-pointer text-destructive focus:bg-destructive/10 focus:text-destructive"
                    onClick={() => {
                      if (typeof window !== 'undefined') {
                        localStorage.removeItem('admin_access_token');
                        window.location.href = '/superadmin/login';
                      }
                    }}
                  >
                    Log out
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  )
}
