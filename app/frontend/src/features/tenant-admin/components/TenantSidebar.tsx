"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  LayoutDashboard,
  Package,
  ShoppingCart,
  Users,
  BarChart3,
  Tag,
  Settings,
  ShieldCheck,
  ChevronsUpDown,
  Boxes,
  RotateCcw,
  Sparkles,
  Megaphone,
  Star,
  Search,
  Webhook,
} from "lucide-react"
import { useAuthStore } from "@/store/useAuthStore"

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
  permission?: string
}

const storeNav: NavItem[] = [
  { title: "Overview", href: "/admin", icon: LayoutDashboard },
  { title: "Orders", href: "/admin/orders", icon: ShoppingCart, badge: "12", permission: "orders.read" },
  { title: "Products", href: "/admin/catalog", icon: Package, permission: "products.read" },
  { title: "Inventory", href: "/admin/inventory", icon: Boxes, permission: "inventory.read" },
  { title: "Customers", href: "/admin/customers", icon: Users, permission: "customers.read" },
  { title: "Analytics", href: "/admin/analytics", icon: BarChart3, permission: "analytics.read" },
  { title: "Discounts", href: "/admin/discounts", icon: Tag, permission: "marketing.read" },
  { title: "Returns", href: "/admin/returns", icon: RotateCcw, permission: "returns.read" },
  { title: "Promotions", href: "/admin/promotions", icon: Sparkles, permission: "marketing.read" },
  { title: "Marketing", href: "/admin/marketing", icon: Megaphone, permission: "marketing.read" },
  { title: "Reviews", href: "/admin/reviews", icon: Star, permission: "products.read" },
  { title: "SEO", href: "/admin/seo", icon: Search, permission: "seo.read" },
  { title: "Webhooks", href: "/admin/webhooks", icon: Webhook, permission: "settings.read" },
]

const settingsNav: NavItem[] = [
  { title: "Team", href: "/admin/team", icon: ShieldCheck },
  { title: "Settings", href: "/admin/settings", icon: Settings },
]

export function TenantSidebar() {
  const pathname = usePathname()
  const { user, activeShop, shopBrief, logout } = useAuthStore()

  const shopName = activeShop?.name || "My Store"
  const shopRole = "Store Owner"
  const shopInitials = shopName.substring(0, 2).toUpperCase()

  const userName = user?.full_name || user?.email || "Staff"
  const userRole = (user as any)?.roles?.[0] || "Admin"
  const userInitials = userName.substring(0, 2).toUpperCase()

  // Inline RBAC check to avoid hook-in-loop issues
  const isOwner = shopBrief?.is_owner ?? false
  const perms = shopBrief?.permissions ?? []
  const checkPerm = (perm?: string) => {
    if (!perm) return true
    if (isOwner || perms.includes("*")) return true
    if (perms.includes(perm)) return true
    const ns = perm.split(".")[0]
    return perms.includes(`${ns}.*`)
  }

  const visibleStoreNav = storeNav.filter((item) => checkPerm(item.permission))

  React.useEffect(() => {
    const activeEl = document.querySelector('[data-sidebar="menu-button"][data-active="true"]')
    if (activeEl) {
      activeEl.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
    }
  }, [pathname])

  const renderNav = (items: NavItem[]) =>
    items.map((item) => {
      const isActive = item.href !== "#" && pathname === item.href
      return (
        <SidebarMenuItem key={item.title} className="group-data-[collapsible=icon]:flex group-data-[collapsible=icon]:justify-center">
          <SidebarMenuButton
            isActive={isActive}
            tooltip={item.title}
            className={`relative transition-all duration-200 rounded-xl ${
              isActive
                ? "bg-zinc-800 text-zinc-50 shadow-sm shadow-black/20"
                : "text-zinc-400 hover:text-zinc-50 hover:bg-zinc-900"
            }`}
            render={<Link href={item.href} />}
          >
            {isActive && (
              <span className="absolute left-0 top-2 bottom-2 w-[3px] rounded-r-full bg-zinc-100 group-data-[collapsible=icon]:hidden" />
            )}
            <item.icon className={isActive ? "text-zinc-50" : ""} />
            <span className="group-data-[collapsible=icon]:hidden">{item.title}</span>
          </SidebarMenuButton>
          {item.badge ? (
            <SidebarMenuBadge className="bg-zinc-800 text-zinc-300 text-[10px] font-mono border border-zinc-700 group-data-[collapsible=icon]:hidden">
              {item.badge}
            </SidebarMenuBadge>
          ) : null}
        </SidebarMenuItem>
      )
    })

  return (
    <Sidebar
      collapsible="icon"
      className="border-r border-zinc-800/50 bg-zinc-950"
    >
      <SidebarHeader className="border-b border-zinc-800/50 pb-6 pt-6">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              className="hover:bg-zinc-900 cursor-default px-2 rounded-xl transition-colors"
            >
              <div className="relative flex items-center justify-center p-0.5 rounded-[12px] border border-white/10 bg-white/5">
                <div className="relative flex size-8 items-center justify-center rounded-[10px] bg-zinc-900 text-zinc-100 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)]">
                  <span className="text-sm font-semibold tracking-tight">{shopInitials}</span>
                </div>
              </div>
              <div className="flex flex-col gap-0.5 leading-none ml-3 group-data-[collapsible=icon]:hidden">
                <span className="font-semibold tracking-tight text-zinc-100 text-sm">
                  {shopName}
                </span>
                <span className="text-[10px] text-zinc-500 font-mono">
                  {shopRole}
                </span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent className="pt-6 pb-28 gap-6 overflow-y-auto overscroll-contain">
        <SidebarGroup className="px-0">
          <SidebarGroupLabel className="text-[10px] uppercase tracking-[0.2em] text-zinc-500 mb-3 font-semibold px-6 group-data-[collapsible=icon]:hidden">
            Store
          </SidebarGroupLabel>
          <div className="hidden group-data-[collapsible=icon]:flex justify-center mb-4">
            <div className="w-5 h-px bg-zinc-800" />
          </div>
          <SidebarGroupContent>
            <SidebarMenu className="gap-1.5 px-3 group-data-[collapsible=icon]:px-2 group-data-[collapsible=icon]:gap-2">
              {renderNav(visibleStoreNav)}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup className="px-0">
          <SidebarGroupLabel className="text-[10px] uppercase tracking-[0.2em] text-zinc-500 mb-3 font-semibold px-6 group-data-[collapsible=icon]:hidden">
            Configuration
          </SidebarGroupLabel>
          <div className="hidden group-data-[collapsible=icon]:flex justify-center mb-4">
            <div className="w-5 h-px bg-zinc-800" />
          </div>
          <SidebarGroupContent>
            <SidebarMenu className="gap-1.5 px-3 group-data-[collapsible=icon]:px-2 group-data-[collapsible=icon]:gap-2">
              {renderNav(settingsNav)}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="shrink-0 relative z-20 border-t border-zinc-800/50 bg-zinc-950 pt-5 pb-5 px-3 shadow-lg">
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger render={
                <SidebarMenuButton size="lg" className="hover:bg-zinc-900 data-[state=open]:bg-zinc-900 transition-colors rounded-xl p-2">
                  <Avatar className="size-8 rounded-lg ring-1 ring-white/10 ring-offset-2 ring-offset-zinc-950">
                    <AvatarFallback className="rounded-lg bg-zinc-800 text-zinc-300 font-mono text-xs">
                      {userInitials}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex flex-col gap-1 leading-none ml-3 group-data-[collapsible=icon]:hidden">
                    <span className="font-semibold text-zinc-100 text-sm tracking-tight">{userName}</span>
                    <span className="text-[10px] text-zinc-500 font-mono">{userRole}</span>
                  </div>
                  <ChevronsUpDown className="ml-auto size-4 text-zinc-500 group-data-[collapsible=icon]:hidden" />
                </SidebarMenuButton>
              } />
              <DropdownMenuContent align="end" className="w-56 mt-2 rounded-xl border border-zinc-800 bg-zinc-950 shadow-2xl">
                <DropdownMenuGroup>
                  <DropdownMenuLabel className="font-normal px-3 py-3">
                    <div className="flex flex-col gap-1.5">
                      <p className="text-sm font-semibold tracking-tight text-zinc-100">{userName}</p>
                      <p className="text-[10px] font-mono leading-none text-zinc-500">
                        {userRole}
                      </p>
                    </div>
                  </DropdownMenuLabel>
                </DropdownMenuGroup>
                <DropdownMenuSeparator className="bg-zinc-800/50" />
                <DropdownMenuGroup>
                  <DropdownMenuItem
                    className="cursor-pointer text-red-400 focus:bg-red-400/10 focus:text-red-400 font-medium tracking-tight m-1 rounded-lg"
                    onClick={() => {
                      logout()
                      if (typeof window !== 'undefined') {
                        window.location.href = '/admin/login';
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
