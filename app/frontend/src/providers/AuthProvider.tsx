"use client"

import { useEffect, useState } from "react"
import { useRouter, usePathname } from "next/navigation"
import { useAuthStore } from "@/store/useAuthStore"
import { Storefront } from "@phosphor-icons/react"

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const { token, user, setAuth } = useAuthStore()
  const [isInitializing, setIsInitializing] = useState(true)

  useEffect(() => {
    async function initializeAuth() {
      const currentToken = token || localStorage.getItem("admin_access_token")
      
      if (!currentToken) {
        if (pathname.startsWith("/admin") && pathname !== "/admin/login") {
          router.replace("/admin/login")
        }
        setIsInitializing(false)
        return
      }

      if (!user) {
        try {
          localStorage.setItem("admin_access_token", currentToken)
          document.cookie = `admin_access_token=${currentToken}; path=/; SameSite=Lax`

          const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"
          const res = await fetch(`${apiUrl}/api/v1/auth/staff/me`, {
            headers: { Authorization: `Bearer ${currentToken}` },
          })

          if (!res.ok) {
            throw new Error("Invalid or expired session")
          }

          const json = await res.json()
          const staffProfile = json.data || json
          const userData = staffProfile.user
          const shops = staffProfile.shops || []
          const activeShop = shops[0] || null

          if (activeShop?.id) {
            localStorage.setItem("nexus_shop_id", activeShop.id)
            localStorage.setItem("nexus_shop_name", activeShop.name)
          }

          setAuth(currentToken, userData, activeShop, activeShop)
        } catch (error) {
          localStorage.removeItem("admin_access_token")
          document.cookie = "admin_access_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT"
          if (pathname.startsWith("/admin") && pathname !== "/admin/login") {
            router.replace("/admin/login")
          }
        }
      }

      setIsInitializing(false)
    }

    initializeAuth()
  }, [pathname, token, user, router, setAuth])

  if (isInitializing && pathname.startsWith("/admin") && pathname !== "/admin/login") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-4">
          <div className="relative flex size-12 animate-pulse items-center justify-center rounded-2xl bg-card border border-primary/20 shadow-sm">
            <Storefront weight="bold" className="size-5 text-primary" />
          </div>
          <p className="text-sm text-muted-foreground animate-pulse">
            Loading your workspace...
          </p>
        </div>
      </div>
    )
  }

  return <>{children}</>
}
