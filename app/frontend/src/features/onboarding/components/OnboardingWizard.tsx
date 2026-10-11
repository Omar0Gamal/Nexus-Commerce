"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2 } from "lucide-react"
import { usePostApiV1AuthMerchantRegister } from "@/api/endpoints/auth/auth";

export function OnboardingWizard() {
  const router = useRouter()
  const { mutate: registerMerchant, isPending } = usePostApiV1AuthMerchantRegister()
  
  const [formData, setFormData] = useState({
    firstName: "",
    lastName: "",
    email: "",
    password: "",
    storeName: "",
    subdomain: "",
  })
  
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    
    registerMerchant({
      data: {
        full_name: `${formData.firstName} ${formData.lastName}`,
        email: formData.email,
        password: formData.password,
        shop_name: formData.storeName,
        subdomain: formData.subdomain,
      } as any
    }, {
      onSuccess: () => {
        router.push("/admin/login?registered=true")
      },
      onError: (err: any) => {
        setError(err.response?.data?.error || "An error occurred during registration.")
      }
    })
  }

  return (
    <div className="min-h-screen w-full flex flex-col lg:flex-row bg-zinc-950 text-zinc-50 font-sans selection:bg-zinc-100/30 selection:text-zinc-50">
      <div className="hidden lg:flex relative w-1/2 flex-col justify-between p-12 bg-zinc-900 overflow-hidden">
        <div className="absolute inset-0 z-0">
          <img 
            src="https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=2564&auto=format&fit=crop" 
            alt="Abstract architecture" 
            className="w-full h-full object-cover opacity-30 mix-blend-luminosity"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-zinc-950 via-zinc-950/80 to-transparent" />
        </div>
        
        <div className="relative z-10">
          <div className="flex items-center gap-2 font-bold text-xl tracking-tight">
            <div className="w-6 h-6 bg-zinc-100 rounded-sm" />
            Nexus
          </div>
        </div>
        
        <div className="relative z-10 max-w-lg">
          <h1 className="text-5xl font-medium tracking-tight leading-[1.1] mb-6">
            Build your commerce empire.
          </h1>
          <p className="text-zinc-400 text-lg leading-relaxed">
            Create your account to launch a globally distributed, high-performance storefront in seconds.
          </p>
        </div>
      </div>

      <div className="flex-1 flex flex-col justify-center px-6 sm:px-12 lg:px-24 py-12">
        <div className="w-full max-w-md mx-auto">
          <div className="mb-10">
            <h2 className="text-2xl font-semibold tracking-tight">Create your store</h2>
            <p className="text-zinc-400 mt-2 text-sm">Enter your details to provision your new environment.</p>
          </div>

          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            {error && (
              <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
                {error}
              </div>
            )}
            
            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-2">
                <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">First Name</label>
                <input 
                  required
                  type="text" 
                  value={formData.firstName}
                  onChange={(e) => setFormData(p => ({ ...p, firstName: e.target.value }))}
                  className="h-12 w-full rounded-xl bg-zinc-900 border border-zinc-800 px-4 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all"
                  placeholder="Omar"
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Last Name</label>
                <input 
                  required
                  type="text" 
                  value={formData.lastName}
                  onChange={(e) => setFormData(p => ({ ...p, lastName: e.target.value }))}
                  className="h-12 w-full rounded-xl bg-zinc-900 border border-zinc-800 px-4 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all"
                  placeholder="Gamal"
                />
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Email Address</label>
              <input 
                required
                type="email" 
                value={formData.email}
                onChange={(e) => setFormData(p => ({ ...p, email: e.target.value }))}
                className="h-12 w-full rounded-xl bg-zinc-900 border border-zinc-800 px-4 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all"
                placeholder="omar@example.com"
              />
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Password</label>
              <input 
                required
                type="password" 
                value={formData.password}
                onChange={(e) => setFormData(p => ({ ...p, password: e.target.value }))}
                className="h-12 w-full rounded-xl bg-zinc-900 border border-zinc-800 px-4 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all"
                placeholder="••••••••"
              />
            </div>

            <div className="w-full h-px bg-zinc-800 my-2" />

            <div className="flex flex-col gap-2">
              <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Store Name</label>
              <input 
                required
                type="text" 
                value={formData.storeName}
                onChange={(e) => setFormData(p => ({ ...p, storeName: e.target.value }))}
                className="h-12 w-full rounded-xl bg-zinc-900 border border-zinc-800 px-4 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all"
                placeholder="Acme Co"
              />
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Store Subdomain</label>
              <div className="relative flex items-center">
                <input 
                  required
                  type="text" 
                  value={formData.subdomain}
                  onChange={(e) => setFormData(p => ({ ...p, subdomain: e.target.value }))}
                  className="h-12 w-full rounded-xl bg-zinc-900 border border-zinc-800 pl-4 pr-32 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all"
                  placeholder="acme"
                />
                <span className="absolute right-4 text-sm text-zinc-500 pointer-events-none">.localhost:3000</span>
              </div>
            </div>

            <button 
              type="submit" 
              disabled={isPending}
              className="mt-4 flex h-12 w-full items-center justify-center rounded-xl bg-zinc-100 text-sm font-semibold text-zinc-900 transition-all hover:bg-white active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
            >
              {isPending ? <Loader2 className="w-5 h-5 animate-spin" /> : "Launch Store"}
            </button>
          </form>
        </div>
      </div>
      
    </div>
  )
}
