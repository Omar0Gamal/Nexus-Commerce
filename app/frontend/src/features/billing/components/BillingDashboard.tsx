"use client"

import { Receipt } from "lucide-react"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useBillingData } from "../hooks/useBillingData"
import { BillingHeroCards } from "./BillingHeroCards"
import { PayoutsTable } from "./PayoutsTable"

export function BillingDashboard() {
  const { billingData, isLoading } = useBillingData()

  if (isLoading) {
    return (
      <div className="flex flex-col gap-6">
        <section className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="card-elevated border-border/40">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="size-9 rounded-lg" />
              </CardHeader>
              <CardContent>
                <div className="flex flex-col gap-2 mt-1">
                  <Skeleton className="h-7 w-32" />
                  <Skeleton className="h-4 w-48" />
                </div>
              </CardContent>
            </Card>
          ))}
        </section>
        <section>
          <Card className="card-elevated border-border/40">
            <CardHeader className="flex flex-row items-center justify-between">
              <div className="flex flex-col gap-1.5">
                <Skeleton className="h-6 w-32" />
                <Skeleton className="h-4 w-64" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="rounded-lg border p-4">
                <div className="flex flex-col gap-4">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="flex items-center justify-between">
                      <Skeleton className="h-5 w-24" />
                      <Skeleton className="h-5 w-32" />
                      <Skeleton className="h-5 w-20" />
                      <Skeleton className="h-5 w-16 rounded-full" />
                      <Skeleton className="h-5 w-24" />
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    )
  }

  if (!billingData || Object.keys(billingData).length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-24 text-center border rounded-xl border-dashed bg-muted/10">
        <div className="flex size-12 items-center justify-center rounded-full bg-muted mb-4">
          <Receipt className="size-6 text-muted-foreground" />
        </div>
        <h3 className="text-lg font-medium text-foreground">No billing records found</h3>
        <p className="text-sm text-muted-foreground mt-2 max-w-sm">
          Platform billing data will appear here once transactions have been processed and fees collected.
        </p>
      </div>
    )
  }

  const payouts = billingData.recentPayouts || []

  return (
    <div className="flex flex-col gap-6">
      <BillingHeroCards billingData={billingData} />
      <PayoutsTable payouts={payouts} />
    </div>
  )
}
