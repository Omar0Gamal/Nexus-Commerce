"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { usePostApiV1AuthAdminLogin } from "@/api/endpoints/auth/auth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Terminal } from "lucide-react"
import { Alert, AlertDescription } from "@/components/ui/alert"

export default function SuperadminLoginPage() {
  const router = useRouter()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")

  const { mutate: login, isPending } = usePostApiV1AuthAdminLogin({
    mutation: {
      onSuccess: (data) => {
        if (data.data?.access_token) {
          localStorage.setItem("admin_access_token", data.data.access_token)
          router.push("/superadmin")
        } else {
          setError("Invalid response from server.")
        }
      },
      onError: (err: any) => {
        setError(err.response?.data?.error || "Invalid email or password.")
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
    <div className="flex min-h-screen items-center justify-center bg-muted/40 px-4 py-12 sm:px-6 lg:px-8">
      <Card className="w-full max-w-sm">
        <CardHeader className="space-y-1">
          <div className="flex justify-center mb-4">
            <div className="flex size-10 items-center justify-center rounded-lg bg-primary">
              <Terminal className="size-6 text-primary-foreground" />
            </div>
          </div>
          <CardTitle className="text-2xl text-center">Superadmin Login</CardTitle>
          <CardDescription className="text-center">
            Enter your credentials to access the platform.
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="grid gap-4">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            <div className="grid gap-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="admin@nexus.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={isPending}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={isPending}
              />
            </div>
          </CardContent>
          <CardFooter>
            <Button className="w-full" type="submit" disabled={isPending}>
              {isPending ? "Signing in..." : "Sign in"}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}
