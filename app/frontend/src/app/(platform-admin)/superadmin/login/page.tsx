"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { ArrowRight, Terminal } from "@phosphor-icons/react"
import { usePostApiV1AuthAdminLogin } from "@/api/endpoints/auth/auth";

export default function SuperadminLoginPage() {
  const router = useRouter()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")

  const handleDemoLogin = () => {
    localStorage.setItem("admin_access_token", "fake_superadmin_token_123")
    document.cookie = "admin_access_token=fake_superadmin_token_123; path=/; SameSite=Lax"
    router.push("/superadmin")
  }

  const fillCredentials = (userEmail: string) => {
    setEmail(userEmail)
    setPassword("password123")
    setError("")
  }

  const { mutate: login, isPending } = usePostApiV1AuthAdminLogin({
    mutation: {
      onSuccess: (authResponse: any) => {
        if (authResponse.data?.access_token) {
          localStorage.setItem("admin_access_token", authResponse.data.access_token)
          document.cookie = `admin_access_token=${authResponse.data.access_token}; path=/; SameSite=Lax`
          router.push("/superadmin")
        } else {
          setError("Invalid response from server.")
        }
      },
      onError: (err: any) => {
        setError(err.response?.data?.error || err.response?.data?.message || "Invalid email or password.")
      },
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    if (!email || !password) {
      setError("Please fill out all fields.")
      return
    }
    login({ data: { email, password } })
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center px-4 py-12 overflow-hidden bg-background">
      <div className="dot-grid absolute inset-0 pointer-events-none opacity-40" />
      <div className="grid-pattern absolute inset-0 pointer-events-none opacity-[0.15]" />

      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <div className="w-[700px] h-[500px] rounded-full bg-primary/[0.07] blur-[140px]" />
      </div>

      <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-transparent via-primary/60 to-transparent" />

      <div className="relative w-full max-w-sm z-10">
        <div className="absolute inset-0 rounded-[2rem] bg-muted/10 border border-border/60 ring-1 ring-white/[0.04]" />

        <div className="relative rounded-[calc(2rem-0.25rem)] m-[0.25rem] bg-card/80 backdrop-blur-xl border border-border/30 shadow-[inset_0_1px_1px_rgba(255,255,255,0.06)] px-8 py-10">

          <div className="flex flex-col items-center mb-8">
            <div className="relative mb-5">
              <div className="absolute inset-0 rounded-2xl bg-primary/10 border border-primary/20 ring-1 ring-primary/10 blur-[1px]" />
              <div className="relative flex size-12 items-center justify-center rounded-[calc(1rem-0.1875rem)] bg-card border border-primary/25 shadow-[inset_0_1px_1px_rgba(255,255,255,0.1)]">
                <Terminal weight="bold" className="size-5 text-primary" />
              </div>
            </div>
            <h1 className="text-2xl font-semibold tracking-tight text-foreground">
              Superadmin Login
            </h1>
            <p className="text-sm text-muted-foreground/70 mt-1.5 text-center">
              Enter your credentials to access the platform.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            {error && (
              <Alert variant="destructive" className="rounded-xl">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <div className="flex flex-col gap-2">
              <Label htmlFor="email" className="text-sm font-medium text-foreground/80">
                Email
              </Label>
              <Input
                id="email"
                type="email"
                placeholder="admin@nexus.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={isPending}
                className="h-11 rounded-xl bg-muted/30 border-border/50 focus-visible:border-primary/50 focus-visible:ring-primary/20 placeholder:text-muted-foreground/40 transition-all"
              />
            </div>

            <div className="flex flex-col gap-2">
              <Label htmlFor="password" className="text-sm font-medium text-foreground/80">
                Password
              </Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={isPending}
                className="h-11 rounded-xl bg-muted/30 border-border/50 focus-visible:border-primary/50 focus-visible:ring-primary/20 transition-all"
              />
            </div>

            <div className="flex flex-col gap-1.5 pt-1">
              <span className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground/60">Quick Test Credentials</span>
              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() => fillCredentials("admin@nexus-commerce.com")}
                  className="rounded-lg bg-primary/10 border border-primary/20 px-2.5 py-1 text-xs text-primary hover:bg-primary/20 transition-colors"
                >
                  Platform Admin
                </button>
                <button
                  type="button"
                  onClick={handleDemoLogin}
                  className="rounded-lg bg-emerald-950/40 border border-emerald-800/40 px-2.5 py-1 text-xs text-emerald-300 hover:bg-emerald-900/60 transition-colors"
                >
                  Offline Demo
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isPending}
              className="group mt-2 flex items-center justify-between w-full rounded-full bg-primary pl-7 pr-2 py-2 text-sm font-medium text-primary-foreground transition-all duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] active:scale-[0.98] shadow-sm shadow-primary/25 hover:shadow-sm hover:shadow-primary/40 disabled:opacity-60 disabled:cursor-not-allowed"
            >
              <span>{isPending ? "Signing in..." : "Sign in"}</span>
              <div className="w-9 h-9 rounded-full bg-primary-foreground/20 flex items-center justify-center transition-transform duration-700 ease-[cubic-bezier(0.32,0.72,0,1)] group-hover:translate-x-1 group-hover:-translate-y-[1px] group-hover:scale-105 group-disabled:translate-x-0">
                <ArrowRight weight="bold" className="size-4 text-primary-foreground" />
              </div>
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
