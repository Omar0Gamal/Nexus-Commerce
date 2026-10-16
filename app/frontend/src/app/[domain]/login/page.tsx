"use client"

import * as React from "react"
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { usePostApiV1AuthCustomerRegister, usePostApiV1AuthCustomerLogin } from "@/api/endpoints/auth/auth";
import { useCustomerAuthStore } from "@/store/useCustomerAuthStore";
import { useRouter } from "next/navigation";
import { Alert, AlertDescription } from "@/components/ui/alert";

export default function LoginPage() {
  const router = useRouter()
  const { setCustomerAuth } = useCustomerAuthStore()
  
  const [loginEmail, setLoginEmail] = useState("")
  const [loginPassword, setLoginPassword] = useState("")
  const [loginError, setLoginError] = useState("")

  const [firstName, setFirstName] = useState("")
  const [lastName, setLastName] = useState("")
  const [regEmail, setRegEmail] = useState("")
  const [regPassword, setRegPassword] = useState("")
  const [regError, setRegError] = useState("")

  const registerMutation = usePostApiV1AuthCustomerRegister({
    mutation: {
      onSuccess: (registerResponse: any) => {
        const token = registerResponse.data.access_token;
        const user = registerResponse.data.user || registerResponse.data.customer;
        setCustomerAuth(token, user);
        router.push("/account");
      },
      onError: (err: any) => {
        setRegError(err?.response?.data?.message || "Registration failed");
      }
    }
  })

  const loginMutation = usePostApiV1AuthCustomerLogin({
    mutation: {
      onSuccess: (loginResponse: any) => {
        const token = loginResponse.data.access_token;
        const user = loginResponse.data.user || loginResponse.data.customer;
        setCustomerAuth(token, user);
        router.push("/account");
      },
      onError: (err: any) => {
        setLoginError(err?.response?.data?.message || "Login failed");
      }
    }
  })

  const handleRegister = (e: React.FormEvent) => {
    e.preventDefault()
    setRegError("")
    registerMutation.mutate({
      data: {
        first_name: firstName,
        last_name: lastName,
        email: regEmail,
        password: regPassword,
      }
    })
  }

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault()
    setLoginError("")
    loginMutation.mutate({
      data: {
        email: loginEmail,
        password: loginPassword,
      }
    })
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-zinc-950 p-4">
      <Card className="w-full max-w-md shadow-sm border-zinc-200 dark:border-zinc-800">
        <Tabs defaultValue="login" className="w-full">
          <CardHeader>
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="login">Login</TabsTrigger>
              <TabsTrigger value="register">Register</TabsTrigger>
            </TabsList>
            <CardTitle className="text-2xl font-semibold tracking-tight text-center">Welcome</CardTitle>
            <CardDescription className="text-center text-zinc-500">
              Sign in to your account or create a new one.
            </CardDescription>
          </CardHeader>
          
          <CardContent>
            <TabsContent value="login">
              <form onSubmit={handleLogin} className="space-y-4">
                {loginError && (
                  <Alert variant="destructive">
                    <AlertDescription>{loginError}</AlertDescription>
                  </Alert>
                )}
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input id="email" type="email" placeholder="m@example.com" required value={loginEmail} onChange={e => setLoginEmail(e.target.value)} className="bg-white dark:bg-zinc-900" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input id="password" type="password" required value={loginPassword} onChange={e => setLoginPassword(e.target.value)} className="bg-white dark:bg-zinc-900" />
                </div>
                <Button type="submit" disabled={loginMutation.isPending} className="w-full mt-2">
                  {loginMutation.isPending ? "Signing In..." : "Sign In"}
                </Button>
              </form>
            </TabsContent>
            
            <TabsContent value="register">
              <form onSubmit={handleRegister} className="space-y-4">
                {regError && (
                  <Alert variant="destructive">
                    <AlertDescription>{regError}</AlertDescription>
                  </Alert>
                )}
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="firstName">First name</Label>
                    <Input id="firstName" required value={firstName} onChange={e => setFirstName(e.target.value)} className="bg-white dark:bg-zinc-900" />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="lastName">Last name</Label>
                    <Input id="lastName" required value={lastName} onChange={e => setLastName(e.target.value)} className="bg-white dark:bg-zinc-900" />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="reg-email">Email</Label>
                  <Input id="reg-email" type="email" placeholder="m@example.com" required value={regEmail} onChange={e => setRegEmail(e.target.value)} className="bg-white dark:bg-zinc-900" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="reg-password">Password</Label>
                  <Input id="reg-password" type="password" required value={regPassword} onChange={e => setRegPassword(e.target.value)} className="bg-white dark:bg-zinc-900" />
                </div>
                <Button type="submit" disabled={registerMutation.isPending} className="w-full mt-2">
                  {registerMutation.isPending ? "Creating account..." : "Create Account"}
                </Button>
              </form>
            </TabsContent>
          </CardContent>
        </Tabs>
      </Card>
    </div>
  )
}
