"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"

import { useAuthStore } from "@/store/useAuthStore"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { ArrowRight, Storefront } from "@phosphor-icons/react"
import { usePostApiV1AuthStaffLogin } from "@/api/endpoints/auth/auth";

export default function TenantLoginPage() {
  const router = useRouter()
  const { setAuth } = useAuthStore()
  
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleDemoLogin = async (demoEmail = "owner@coolshoes.com", demoPass = "password123") => {
    setEmail(demoEmail)
    setPassword(demoPass)
    setError("")
    setIsSubmitting(true)

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"
      const res = await fetch(`${apiUrl}/api/v1/auth/staff/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: demoEmail, password: demoPass }),
      })

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}))
        throw new Error(errData.error || errData.message || "Invalid credentials")
      }

      const authResponse = await res.json()
      const authData = authResponse.data || authResponse
      let token = authData.access_token
      let shop = authData.shop
      let user = authData.user

      try {
        const meRes = await fetch(`${apiUrl}/api/v1/auth/staff/me`, {
          headers: { Authorization: `Bearer ${token}` },
        })
        if (meRes.ok) {
          const meData = await meRes.json()
          const meInfo = meData.data || meData
          user = meInfo.user || user
          const shops = meInfo.shops || []
          if (shops.length > 0) {
            shop = shops[0]
            const selectRes = await fetch(`${apiUrl}/api/v1/auth/staff/select-shop`, {
              method: "POST",
              headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/json",
              },
              body: JSON.stringify({ shop_id: shop.id }),
            })
            if (selectRes.ok) {
              const selectShopResponse = await selectRes.json()
              token = selectShopResponse.data?.access_token || selectShopResponse.access_token || token
            }
          }
        }
      } catch (meErr) {
        console.warn("Could not fetch staff shops:", meErr)
      }

      localStorage.setItem("admin_access_token", token)
      document.cookie = `admin_access_token=${token}; path=/; SameSite=Lax`
      if (shop?.id) {
        localStorage.setItem("nexus_shop_id", shop.id)
        localStorage.setItem("nexus_shop_name", shop.name)
      }
      setAuth(token, user, shop, shop)
      router.push("/admin")
    } catch (err: any) {
      setError(err.message || "Failed to sign in.")
    } finally {
      setIsSubmitting(false)
    }
  }

  const fillCredentials = (userEmail: string) => {
    setEmail(userEmail)
    setPassword("password123")
    setError("")
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    if (!email || !password) {
      setError("Please fill out all fields.")
      return
    }

    setIsSubmitting(true)
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"
      const res = await fetch(`${apiUrl}/api/v1/auth/staff/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      })

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}))
        throw new Error(errData.error || errData.message || "Invalid credentials")
      }

      const authResponse = await res.json()
      const authData = authResponse.data || authResponse
      let token = authData.access_token

      let shop = authData.shop
      let user = authData.user

      try {
        const meRes = await fetch(`${apiUrl}/api/v1/auth/staff/me`, {
          headers: { Authorization: `Bearer ${token}` },
        })
        if (meRes.ok) {
          const meData = await meRes.json()
          const meInfo = meData.data || meData
          user = meInfo.user || user
          const shops = meInfo.shops || []
          if (shops.length > 0) {
            shop = shops[0]
            const selectRes = await fetch(`${apiUrl}/api/v1/auth/staff/select-shop`, {
              method: "POST",
              headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/json",
              },
              body: JSON.stringify({ shop_id: shop.id }),
            })
            if (selectRes.ok) {
              const selectShopResponse = await selectRes.json()
              token = selectShopResponse.data?.access_token || selectShopResponse.access_token || token
            }
          }
        }
      } catch (meErr) {
        console.warn("Could not fetch staff shops:", meErr)
      }

      localStorage.setItem("admin_access_token", token)
      document.cookie = `admin_access_token=${token}; path=/; SameSite=Lax`
      if (shop?.id) {
        localStorage.setItem("nexus_shop_id", shop.id)
        localStorage.setItem("nexus_shop_name", shop.name)
      }
      setAuth(token, user, shop, shop)
      router.push("/admin")
    } catch (err: any) {
      setError(err.message || "Failed to sign in.")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center px-4 py-12 overflow-hidden bg-zinc-950 selection:bg-zinc-100 selection:text-zinc-900">
      
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-zinc-900/40 via-zinc-950 to-zinc-950 pointer-events-none" />

      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />

      <div className="relative w-full max-w-sm z-10 animate-in fade-in slide-in-from-bottom-8 duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]">
        <div className="absolute inset-0 rounded-[2.25rem] bg-white/[0.02] border border-white/[0.05]" />

        <div className="relative rounded-[calc(2.25rem-0.375rem)] m-[0.375rem] bg-zinc-900/80 backdrop-blur-xl border border-white/[0.05] shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] px-8 py-10">

          <div className="flex flex-col items-center mb-8">
            <div className="relative mb-5">
              <div className="relative flex size-12 items-center justify-center rounded-[1rem] bg-zinc-950 border border-zinc-800 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)]">
                <Storefront weight="bold" className="size-5 text-zinc-100" />
              </div>
            </div>
            <h1 className="text-2xl font-semibold tracking-tight text-zinc-50">
              Tenant Login
            </h1>
            <p className="text-sm text-zinc-500 mt-1.5 text-center">
              Welcome back to your shop.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            {error && (
              <Alert variant="destructive" className="rounded-xl bg-red-500/10 border-red-500/20 text-red-400">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <div className="flex flex-col gap-2">
              <Label htmlFor="email" className="text-sm font-medium text-zinc-400">
                Email
              </Label>
              <Input
                id="email"
                type="email"
                placeholder="owner@coolshoes.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={isSubmitting}
                className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 placeholder:text-zinc-600 transition-all"
              />
            </div>

            <div className="flex flex-col gap-2">
              <Label htmlFor="password" className="text-sm font-medium text-zinc-400">
                Password
              </Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={isSubmitting}
                className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 transition-all"
              />
            </div>

            <div className="flex flex-col gap-1.5 pt-1">
              <span className="text-[11px] font-medium uppercase tracking-wider text-zinc-500">Quick Test Accounts</span>
              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() => fillCredentials("owner@coolshoes.com")}
                  className="rounded-lg bg-zinc-800/60 border border-zinc-700/50 px-2.5 py-1 text-xs text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors"
                >
                  Cool Shoes
                </button>
                <button
                  type="button"
                  onClick={() => fillCredentials("owner@techstore.com")}
                  className="rounded-lg bg-zinc-800/60 border border-zinc-700/50 px-2.5 py-1 text-xs text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors"
                >
                  Tech Store
                </button>
                <button
                  type="button"
                  onClick={() => handleDemoLogin("owner@coolshoes.com", "password123")}
                  disabled={isSubmitting}
                  className="rounded-lg bg-emerald-950/40 border border-emerald-800/40 px-2.5 py-1 text-xs text-emerald-300 hover:bg-emerald-900/60 transition-colors disabled:opacity-50"
                >
                  Demo Store (1-Click)
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="group mt-2 flex items-center justify-between w-full rounded-full bg-zinc-100 pl-7 pr-2 py-2 text-sm font-semibold text-zinc-900 transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98] hover:bg-white disabled:opacity-60 disabled:cursor-not-allowed"
            >
              <span>{isSubmitting ? "Signing in..." : "Sign in"}</span>
              <div className="w-9 h-9 rounded-full bg-zinc-950/10 flex items-center justify-center transition-transform duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] group-hover:translate-x-1 group-hover:-translate-y-[1px] group-hover:scale-105 group-disabled:translate-x-0">
                <ArrowRight weight="bold" className="size-4 text-zinc-900" />
              </div>
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
