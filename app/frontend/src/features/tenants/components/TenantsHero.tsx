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
      <Card className="sm:col-span-2 relative overflow-hidden border border-border/40 bg-card/80 backdrop-blur-xl shadow-sm group">
        <div className="absolute left-0 top-0 bottom-0 w-[3px] bg-primary rounded-l-lg opacity-70 group-hover:opacity-100 transition-opacity" />

        <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.04] via-transparent to-transparent pointer-events-none" />

        <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-[0.03] mix-blend-overlay pointer-events-none" />

        <div className="absolute -right-16 -top-16 w-48 h-48 bg-primary/10 blur-3xl rounded-full pointer-events-none group-hover:bg-primary/15 transition-all duration-700" />

        <CardContent className="p-8 flex flex-col justify-between h-full relative z-10">
          <div>
            <h2 className="text-muted-foreground font-medium text-sm mb-1 flex items-center gap-2 uppercase tracking-wider">
              <Store className="size-4 text-primary/70" />
              Total Platform Tenants
            </h2>
            <div className="flex items-baseline gap-3 mt-4">
              {isLoading ? (
                <Skeleton className="h-16 w-24" />
              ) : (
                <>
                  <span className="text-7xl font-bold tracking-tighter text-foreground">{totalShops}</span>
                  <span className="text-muted-foreground text-sm font-medium">registered shops</span>
                </>
              )}
            </div>
          </div>

          <div className="mt-10 grid grid-cols-2 gap-4">
            <div className="flex items-center gap-3 bg-muted/40 rounded-xl p-3 border border-border/40 group-hover:border-primary/20 transition-colors">
              <div className="flex size-10 rounded-full bg-emerald-500/10 text-emerald-400 items-center justify-center border border-emerald-500/20">
                <ShieldCheck className="size-5" />
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-bold text-foreground">{activeCount}</span>
                <span className="text-[11px] text-muted-foreground uppercase tracking-wider font-semibold">Active Now</span>
              </div>
            </div>

            <div className="flex items-center gap-3 bg-muted/40 rounded-xl p-3 border border-border/40 group-hover:border-primary/20 transition-colors">
              <div className="flex size-10 rounded-full bg-primary/10 text-primary items-center justify-center border border-primary/20">
                <TrendingUp className="size-5" />
              </div>
              <div className="flex flex-col">
                <span className="text-xl font-bold text-foreground">{proPlusCount}</span>
                <span className="text-[11px] text-muted-foreground uppercase tracking-wider font-semibold">Pro & Enterprise</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="flex flex-col gap-4">
        <Card className="flex-1 card-elevated border border-border/40 bg-card/80 backdrop-blur-xl hover:-translate-y-1 transition-all duration-300 relative overflow-hidden group flex flex-col justify-between">
          <div className="absolute left-0 top-0 bottom-0 w-[3px] bg-amber-500 rounded-l-lg opacity-80" />
          <div className="absolute inset-0 bg-gradient-to-br from-amber-500/[0.04] to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />

          <CardHeader className="pb-2 pt-6 px-6 relative z-10">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center justify-between">
              Needs Attention
              <div className="flex size-8 items-center justify-center rounded-lg bg-amber-500/10 text-amber-500 ring-1 ring-amber-500/20 group-hover:scale-110 transition-transform">
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
