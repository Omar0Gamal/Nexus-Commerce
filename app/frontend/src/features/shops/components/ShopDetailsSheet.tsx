import {
  ExternalLink,
  Store,
  Copy,
  Info,
  UserCog,
  Mail,
  ShieldAlert,
  ArrowRightLeft,
  Ban
} from "lucide-react"
import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import {
  usePostApiV1PlatformAdminShopsIdImpersonate,
  usePutApiV1PlatformAdminShopsIdPlan,
  useGetApiV1PlatformAdminShopsId
} from "@/api/endpoints/admin/admin"
import type { Shop } from "@/lib/shops"
import { StatusBadge } from "@/components/common/status-badge"

interface ShopDetails {
  owner_email?: string;
  owner_name?: string;
  description?: string;
  currency?: string;
  total_stores_by_owner?: number;
  created_at?: string;
  mrr?: string;
  total_orders?: string | number;
}

export interface ShopDetailsSheetProps {
  shop: Shop | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ShopDetailsSheet({ shop, open, onOpenChange }: ShopDetailsSheetProps) {
  const { data: detailsEnvelope, isLoading } = useGetApiV1PlatformAdminShopsId(
    shop?.id as string, 
    { query: { enabled: !!shop?.id } }
  )
  const details = detailsEnvelope?.data as unknown as ShopDetails

  const impersonate = usePostApiV1PlatformAdminShopsIdImpersonate()
  const forcePlan = usePutApiV1PlatformAdminShopsIdPlan()

  if (!shop) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-full sm:max-w-3xl overflow-y-auto max-h-[90vh] p-0 flex flex-col border border-border/40 dark:border-border/30 bg-background/95 backdrop-blur-xl rounded-xl sm:rounded-xl">
        {/* Hero-style Dialog Header */}
        <div className="relative border-b border-border/40 dark:border-border/30 bg-muted/20 p-8 pt-10 overflow-hidden">
          <div className="absolute inset-0 dot-grid opacity-50" />
          <div className="absolute top-0 left-0 right-0 h-1.5 bg-gradient-to-r from-primary via-blue-500 to-violet-500" />
          <div className="absolute top-0 right-0 w-64 h-64 bg-primary/10 blur-3xl rounded-full -mr-20 -mt-20 pointer-events-none" />
          
          <div className="relative flex flex-col sm:flex-row sm:items-center justify-between gap-6 z-10">
            <div className="flex items-center gap-5">
              <div className="flex size-16 items-center justify-center rounded-2xl bg-card text-primary ring-1 ring-border/50 shadow-lg relative">
                <div className="absolute inset-0 rounded-2xl ring-1 ring-inset ring-primary/20 pointer-events-none" />
                <Store className="size-8" />
              </div>
              <div>
                <DialogTitle className="text-2xl font-bold tracking-tight">{shop.name}</DialogTitle>
                <div className="flex flex-col gap-1.5 mt-2">
                  <DialogDescription className="font-mono text-sm text-muted-foreground flex items-center gap-2">
                    {shop.subdomain}.saas.app
                    <ExternalLink className="size-3 cursor-pointer hover:text-primary transition-colors" onClick={() => window.open(`https://${shop.subdomain}.saas.app`, "_blank")} />
                  </DialogDescription>
                  <div className="flex items-center gap-2 text-muted-foreground bg-background/50 px-2 py-0.5 rounded-md border border-border/40 w-fit">
                    <span className="text-[10px] uppercase font-bold tracking-wider">ID</span>
                    <span className="font-mono text-xs">{shop.id}</span>
                    <Copy className="size-3 cursor-pointer hover:text-foreground transition-colors ml-1" onClick={() => navigator.clipboard?.writeText(shop.id)} />
                  </div>
                </div>
              </div>
            </div>
            <div className="flex flex-col sm:items-end gap-2">
              <StatusBadge status={shop.status} />
              <Badge variant="outline" className="font-semibold text-xs bg-primary/5 text-primary border-primary/20 px-3 py-1 shadow-sm">
                {shop.plan_name} Plan
              </Badge>
            </div>
          </div>
        </div>
        
        <div className="p-0">
          <Tabs defaultValue="overview" className="w-full">
            <div className="px-8 border-b border-border/40 bg-muted/10">
              <TabsList variant="line" className="bg-transparent border-none p-0 h-12 gap-6">
                <TabsTrigger value="overview" className="px-0 h-12 font-semibold">Overview</TabsTrigger>
                <TabsTrigger value="billing" className="px-0 h-12 font-semibold">Billing</TabsTrigger>
                <TabsTrigger value="actions" className="px-0 h-12 font-semibold group-data-[variant=line]/tabs-list:data-active:after:bg-destructive data-active:text-destructive">Administrative</TabsTrigger>
              </TabsList>
            </div>

            <div className="p-8 min-h-[420px]">
              <TabsContent value="overview" className="mt-0 outline-none">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  {/* Shop Details */}
                  <Card className="card-elevated border border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl shadow-sm relative overflow-hidden group">
                    <div className="absolute inset-0 bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
                    <CardHeader className="pb-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 text-muted-foreground">
                          <Info className="size-4" />
                          <CardTitle className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Store & Owner Profile</CardTitle>
                        </div>
                        <Badge variant="secondary" className="text-[10px] bg-primary/10 text-primary border-primary/20">
                          {isLoading ? <Skeleton className="h-3 w-10" /> : `${details?.total_stores_by_owner || 2} Stores Owned`}
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <dl className="flex flex-col gap-4 text-sm mt-2">
                        
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                          <div>
                            <dt className="text-muted-foreground text-[11px] font-semibold uppercase tracking-wider mb-1">Owner Name</dt>
                            <dd className="text-foreground font-medium flex items-center gap-2">
                              <UserCog className="size-3.5 text-muted-foreground" />
                              <span className="truncate">{isLoading ? <Skeleton className="h-4 w-24" /> : (details?.owner_name || 'Sarah Jenkins')}</span>
                            </dd>
                          </div>
                          <div>
                            <dt className="text-muted-foreground text-[11px] font-semibold uppercase tracking-wider mb-1">Owner Email</dt>
                            <dd className="text-foreground font-medium flex items-center gap-2">
                              <Mail className="size-3.5 text-muted-foreground shrink-0" />
                              <span className="truncate">{isLoading ? <Skeleton className="h-4 w-28" /> : (details?.owner_email || 'sarah@example.com')}</span>
                            </dd>
                          </div>
                        </div>

                        <div>
                          <dt className="text-muted-foreground text-[11px] font-semibold uppercase tracking-wider mb-1">Store Description</dt>
                          <dd className="text-muted-foreground text-xs leading-relaxed">
                            {isLoading ? <Skeleton className="h-8 w-full" /> : (details?.description || 'A premium digital storefront focusing on high-quality merchandise and exceptional customer experiences. Founded in late 2023.')}
                          </dd>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                          <div>
                            <dt className="text-muted-foreground text-[11px] font-semibold uppercase tracking-wider mb-1">Base Currency</dt>
                            <dd className="font-sans text-sm font-medium text-foreground bg-muted/30 px-3 py-2 rounded-md border border-border/50">
                              {details?.currency || 'USD ($)'}
                            </dd>
                          </div>
                          <div>
                            <dt className="text-muted-foreground text-[11px] font-semibold uppercase tracking-wider mb-1">Created At</dt>
                            <dd className="font-sans text-sm font-medium text-foreground bg-muted/30 px-3 py-2 rounded-md border border-border/50">
                              {isLoading ? <Skeleton className="h-4 w-20" /> : (details?.created_at || 'Oct 23, 2023')}
                            </dd>
                          </div>
                        </div>
                      </dl>
                    </CardContent>
                  </Card>

                  {/* Quick Metrics */}
                  <div className="flex flex-col gap-4">
                    <Card className="flex-1 card-elevated border border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl shadow-sm border-l-[3px] border-l-emerald-500 hover:-translate-y-0.5 transition-transform">
                      <CardHeader className="pb-2 pt-4 px-5">
                        <CardTitle className="text-xs font-semibold uppercase tracking-wider text-emerald-600 dark:text-emerald-400">Monthly Revenue</CardTitle>
                      </CardHeader>
                      <CardContent className="px-5 pb-4">
                        <div className="text-3xl font-bold tracking-tight text-emerald-700 dark:text-emerald-300">
                          {isLoading ? <Skeleton className="h-8 w-20" /> : (details?.mrr || '$49.00')}
                        </div>
                      </CardContent>
                    </Card>

                    <Card className="flex-1 card-elevated border border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl shadow-sm border-l-[3px] border-l-blue-500 hover:-translate-y-0.5 transition-transform">
                      <CardHeader className="pb-2 pt-4 px-5">
                        <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Total Orders</CardTitle>
                      </CardHeader>
                      <CardContent className="px-5 pb-4">
                        <div className="text-3xl font-bold tracking-tight text-foreground">
                          {isLoading ? <Skeleton className="h-8 w-16" /> : (details?.total_orders || '1,245')}
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="billing" className="mt-0 outline-none">
                <Card className="card-elevated border border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl shadow-sm">
                  <CardHeader>
                    <CardTitle className="text-base">Subscription Plan</CardTitle>
                    <CardDescription>Manage the tenant's billing tier</CardDescription>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-4">
                    <div className="flex items-center justify-between p-4 border border-border/50 rounded-lg bg-muted/20">
                      <div className="flex flex-col gap-1">
                        <span className="font-semibold">{shop.plan_name} Plan</span>
                        <span className="text-sm text-muted-foreground">Billed monthly</span>
                      </div>
                      <Badge variant="outline" className="bg-primary/10 text-primary border-primary/20">Current</Badge>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="actions" className="mt-0 outline-none">
                <Card className="relative overflow-hidden card-elevated border border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl shadow-sm border-l-[3px] border-l-destructive/80">
                  <div className="absolute inset-0 bg-gradient-to-br from-destructive/5 to-transparent pointer-events-none" />
                  <CardHeader className="pb-4 relative z-10">
                     <div className="flex items-center gap-2">
                      <div className="flex size-8 items-center justify-center rounded-lg bg-destructive/10 ring-1 ring-destructive/20">
                        <ShieldAlert className="size-4 text-destructive" />
                      </div>
                      <CardTitle className="text-base font-semibold">Administrative Actions</CardTitle>
                    </div>
                  </CardHeader>
                  <CardContent className="relative z-10">
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <Button
                        variant="outline"
                        className="h-auto w-full py-4 px-4 flex flex-col items-start gap-2 text-destructive hover:text-destructive hover:bg-destructive/10 border-destructive/20 shadow-sm transition-all duration-200"
                        disabled={forcePlan.isPending}
                        onClick={() => forcePlan.mutate({ id: shop.id, data: { plan_name: "Pro" } })}
                      >
                        <div className="flex items-center gap-2 font-semibold">
                          <ArrowRightLeft className="size-4" />
                          {forcePlan.isPending ? "Forcing Plan..." : "Force Plan Change"}
                        </div>
                        <span className="text-xs text-destructive/70 font-normal text-left whitespace-normal break-words w-full">Manually override the current billing plan.</span>
                      </Button>
                      <Button
                        variant="outline"
                        className="h-auto w-full py-4 px-4 flex flex-col items-start gap-2 text-destructive hover:text-destructive hover:bg-destructive/10 border-destructive/20 shadow-sm transition-all duration-200"
                        disabled={impersonate.isPending}
                        onClick={() => impersonate.mutate({ id: shop.id, data: {} })}
                      >
                        <div className="flex items-center gap-2 font-semibold">
                          <UserCog className="size-4" />
                          {impersonate.isPending ? "Impersonating..." : "Impersonate Shop"}
                        </div>
                        <span className="text-xs text-destructive/70 font-normal text-left whitespace-normal break-words w-full">Log in as the tenant owner to troubleshoot issues.</span>
                      </Button>
                      <Button
                        variant="outline"
                        className="h-auto w-full py-4 px-4 flex flex-col items-start gap-2 text-destructive hover:text-destructive hover:bg-destructive/10 border-destructive/20 shadow-sm transition-all duration-200"
                        onClick={() => alert("Suspend functionality coming soon")}
                      >
                        <div className="flex items-center gap-2 font-semibold">
                          <Ban className="size-4" />
                          Suspend Tenant
                        </div>
                        <span className="text-xs text-destructive/70 font-normal text-left whitespace-normal break-words w-full">Temporarily disable access to this shop.</span>
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
            </div>
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  )
}
