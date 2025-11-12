import { ExternalLink, Store, ChevronRight } from "lucide-react"
import { Card, CardHeader, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import type { Shop } from "@/lib/shops"
import { StatusBadge } from "@/components/common/status-badge"

interface ShopCardProps {
  shop: Shop
  onSelect: (shop: Shop) => void
}

export function ShopCard({ shop, onSelect }: ShopCardProps) {
  return (
    <Card className="group relative overflow-hidden card-elevated border border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl hover:-translate-y-1 hover:ring-1 hover:ring-border/50 dark:hover:ring-border/30 transition-all duration-300 flex flex-col justify-between h-[280px] cursor-pointer"
      onClick={() => onSelect(shop)}
    >
      {/* Noise and glow effects */}
      <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-10 mix-blend-overlay pointer-events-none" />
      <div className="absolute top-0 right-0 w-32 h-32 bg-primary/5 rounded-full blur-2xl -mr-10 -mt-10 group-hover:bg-primary/10 transition-colors duration-500 pointer-events-none" />
      <div className="absolute left-0 top-0 bottom-0 w-1 bg-primary rounded-l-lg opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
      
      <CardHeader className="pb-4 relative z-10">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary ring-1 ring-primary/20 shadow-inner group-hover:scale-105 transition-transform duration-300">
              <Store className="size-5" />
            </div>
            <div className="flex flex-col">
              <span className="font-bold text-foreground tracking-tight group-hover:text-primary transition-colors text-base">{shop.name}</span>
              <span className="font-mono text-xs text-muted-foreground flex items-center gap-1.5 mt-0.5">
                {shop.subdomain}.saas.app
                <ExternalLink className="size-3 cursor-pointer hover:text-primary transition-colors" onClick={(e) => {
                  e.stopPropagation();
                  window.open(`https://${shop.subdomain}.saas.app`, "_blank");
                }} />
              </span>
            </div>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="pb-0 relative z-10 flex-1">
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-2">
            <StatusBadge status={shop.status} />
            <Badge variant="outline" className="font-semibold text-[11px] shadow-sm bg-primary/5 text-primary border-primary/20">
              {shop.plan_name} Plan
            </Badge>
          </div>
          
          <div className="grid grid-cols-2 gap-2 text-sm mt-1 p-3 bg-muted/30 rounded-lg border border-border/50">
            <div className="flex flex-col gap-1">
              <span className="text-[10px] uppercase font-bold tracking-wider text-muted-foreground">ID</span>
              <span className="font-mono text-[11px] text-foreground/80">{shop.id.slice(0, 8)}</span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-[10px] uppercase font-bold tracking-wider text-muted-foreground">Created</span>
              <span className="font-medium text-[11px] text-foreground/80">{shop.created_at ? new Date(shop.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) : 'Unknown'}</span>
            </div>
          </div>
        </div>
      </CardContent>
      
      {/* Footer quick actions */}
      <div className="mt-5 p-2 bg-muted/40 border-t border-border/40 flex items-center justify-between relative z-10 group-hover:bg-muted/60 transition-colors">
        <Button 
          variant="ghost" 
          size="sm" 
          className="text-xs h-8 text-muted-foreground hover:text-foreground"
          onClick={(e) => {
            e.stopPropagation();
            window.open(`https://${shop.subdomain}.saas.app`, "_blank");
          }}
        >
          <ExternalLink className="size-3.5 mr-1.5" />
          Visit Store
        </Button>
        <Button 
          variant="secondary" 
          size="sm" 
          className="text-xs h-8 bg-primary/10 text-primary hover:bg-primary/20"
          onClick={(e) => {
            e.stopPropagation();
            onSelect(shop);
          }}
        >
          View Details
          <ChevronRight className="size-3.5 ml-1" />
        </Button>
      </div>
    </Card>
  )
}
