import Link from "next/link"
import { Trophy, ChevronRight } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button, buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export interface TopTenantItem {
  name: string
  revenue: string
  plan: string
}

interface TopTenantsListProps {
  tenants: TopTenantItem[]
}

export function TopTenantsList({ tenants }: TopTenantsListProps) {
  return (
    <Card className="col-span-1 card-elevated border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl flex flex-col">
      <CardHeader className="pb-3 pt-5 px-5">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-sm">Top Tenants</CardTitle>
            <CardDescription className="text-xs">Highest MRR contributors</CardDescription>
          </div>
          <div className="flex size-7 items-center justify-center rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 ring-1 ring-amber-500/20">
            <Trophy className="size-3.5" />
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col px-5 pb-5">
        <div className="flex flex-col gap-3">
          {tenants.map((tenant, idx) => (
            <div key={tenant.name} className="flex items-center justify-between group">
              <div className="flex items-center gap-3">
                <div className="flex size-7 items-center justify-center rounded-full bg-muted text-xs font-bold text-muted-foreground">
                  {idx + 1}
                </div>
                <div className="flex flex-col">
                  <span className="text-sm font-semibold text-foreground">{tenant.name}</span>
                  <span className="text-xs text-muted-foreground">{tenant.plan}</span>
                </div>
              </div>
              <span className="text-sm font-bold tracking-tight">{tenant.revenue}</span>
            </div>
          ))}
        </div>

        <div className="mt-auto pt-5">
          <Link 
            href="/superadmin/analytics" 
            className={cn(buttonVariants({ variant: "outline" }), "w-full text-[11px] h-8 shadow-sm hover:bg-muted/70 flex items-center justify-center")}
          >
            View Full Analytics
            <ChevronRight className="ml-1 size-3" />
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
