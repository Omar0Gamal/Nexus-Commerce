import { Store, ShieldCheck, TrendingUp, AlertTriangle } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

interface TenantsHeroProps {
  totalShops: number
  activeCount: number
  proPlusCount: number
  needsAttention: number
  isLoading?: boolean
}

export function TenantsHero({ totalShops, activeCount, proPlusCount, needsAttention, isLoading }: TenantsHeroProps) {
  return (
    <div className="lg:col-span-2 grid grid-cols-1 sm:grid-cols-3 gap-4 stagger-children">
      {/* Main Hero Card */}
      <Card className="sm:col-span-2 bg-gradient-to-br from-primary to-blue-600 text-primary-foreground relative overflow-hidden border-0 shadow-lg group">
         {/* Animated background noise & glow */}
         <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 mix-blend-overlay pointer-events-none" />
         <div className="absolute -right-20 -top-20 w-64 h-64 bg-white/10 blur-3xl rounded-full pointer-events-none group-hover:bg-white/20 transition-all duration-700" />
         
         <CardContent className="p-8 flex flex-col justify-between h-full relative z-10">
            <div>
               <h2 className="text-primary-foreground/90 font-medium text-lg mb-1 flex items-center gap-2">
                 <Store className="size-5" />
                 Total Platform Tenants
               </h2>
               <div className="flex items-baseline gap-3 mt-4">
                  {isLoading ? (
                    <Skeleton className="h-16 w-24 bg-primary-foreground/20" />
                  ) : (
                    <>
                      <span className="text-7xl font-bold tracking-tighter">{totalShops}</span>
                      <span className="text-primary-foreground/80 text-sm font-medium">registered shops</span>
                    </>
                  )}
               </div>
            </div>
            
            <div className="mt-10 grid grid-cols-2 gap-4">
               <div className="flex items-center gap-3 bg-black/10 rounded-xl p-3 backdrop-blur-sm border border-white/10">
                 <div className="flex size-10 rounded-full bg-emerald-500/20 text-emerald-300 items-center justify-center">
                   <ShieldCheck className="size-5" />
                 </div>
                 <div className="flex flex-col">
                   <span className="text-xl font-bold">{activeCount}</span>
                   <span className="text-[11px] text-primary-foreground/80 uppercase tracking-wider font-semibold">Active Now</span>
                 </div>
               </div>
               
               <div className="flex items-center gap-3 bg-black/10 rounded-xl p-3 backdrop-blur-sm border border-white/10">
                 <div className="flex size-10 rounded-full bg-blue-400/20 text-blue-200 items-center justify-center">
                   <TrendingUp className="size-5" />
                 </div>
                 <div className="flex flex-col">
                   <span className="text-xl font-bold">{proPlusCount}</span>
                   <span className="text-[11px] text-primary-foreground/80 uppercase tracking-wider font-semibold">Pro & Enterprise</span>
                 </div>
               </div>
            </div>
         </CardContent>
      </Card>

      {/* Needs Attention Card (Side stacked) */}
      <div className="flex flex-col gap-4">
        <Card className="flex-1 card-elevated border border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl hover:-translate-y-1 transition-all duration-300 relative overflow-hidden group flex flex-col justify-between">
          <div className="absolute left-0 top-0 bottom-0 w-1 bg-amber-500 rounded-l-lg opacity-80" />
          <div className="absolute inset-0 bg-gradient-to-br from-amber-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
          
          <CardHeader className="pb-2 pt-6 px-6 relative z-10">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center justify-between">
              Needs Attention
              <div className="flex size-8 items-center justify-center rounded-lg bg-amber-500/10 text-amber-600 ring-1 ring-amber-500/20 group-hover:scale-110 transition-transform">
                <AlertTriangle className="size-4" />
              </div>
            </CardTitle>
          </CardHeader>
          <CardContent className="px-6 pb-6 mt-auto relative z-10">
            {isLoading ? (
              <Skeleton className="h-10 w-16" />
            ) : (
              <div className="flex flex-col gap-1">
                <div className="text-4xl font-bold tracking-tight text-foreground">{needsAttention}</div>
                <p className="text-xs text-muted-foreground font-medium">Tenants past due or suspended</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
