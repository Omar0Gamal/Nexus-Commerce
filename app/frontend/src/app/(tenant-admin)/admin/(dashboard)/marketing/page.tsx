"use client"

import * as React from "react"
import { Megaphone, Users, Award } from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

export default function MarketingPage() {
  return (
    <div className="flex flex-col gap-8 p-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-950 dark:text-zinc-50">Loyalty & Referrals</h1>
          <p className="text-zinc-500 mt-2">Manage customer loyalty points, rewards tiers, and referral programs.</p>
        </div>
      </div>

      <Tabs defaultValue="loyalty" className="w-full">
        <TabsList className="mb-8">
          <TabsTrigger value="loyalty" className="px-6 rounded-lg gap-2">
            <Award className="w-4 h-4" />
            Loyalty Program
          </TabsTrigger>
          <TabsTrigger value="referral" className="px-6 rounded-lg gap-2">
            <Users className="w-4 h-4" />
            Referral Config
          </TabsTrigger>
        </TabsList>
        
        <TabsContent value="loyalty">
          <div className="grid gap-6 md:grid-cols-2">
            <Card className="border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm">
              <CardHeader>
                <CardTitle>Points Earning</CardTitle>
                <CardDescription>Configure how customers earn points on purchases.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <label className="text-sm font-medium mb-1.5 block">Points per EGP 1.00 spent</label>
                  <div className="flex gap-2">
                    <Input defaultValue="10" className="bg-white dark:bg-zinc-950" />
                    <Button variant="secondary">Save</Button>
                  </div>
                </div>
                <div className="pt-4 border-t border-zinc-200 dark:border-zinc-800">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium text-sm">Program Status</p>
                      <p className="text-xs text-zinc-500">Currently active and awarding points</p>
                    </div>
                    <Button variant="outline" className="text-red-600 hover:text-red-700">Disable</Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm">
              <CardHeader>
                <CardTitle>Reward Tiers</CardTitle>
                <CardDescription>Set up customer tiers based on points balance.</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-3 border border-zinc-200 dark:border-zinc-800 rounded-lg bg-zinc-50 dark:bg-zinc-900/50">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-amber-700/20 text-amber-700 flex items-center justify-center font-bold">B</div>
                      <div>
                        <p className="font-medium text-sm">Bronze</p>
                        <p className="text-xs text-zinc-500">0 - 999 points</p>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center justify-between p-3 border border-zinc-200 dark:border-zinc-800 rounded-lg bg-zinc-50 dark:bg-zinc-900/50">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-slate-400/20 text-slate-500 flex items-center justify-center font-bold">S</div>
                      <div>
                        <p className="font-medium text-sm">Silver</p>
                        <p className="text-xs text-zinc-500">1,000 - 4,999 points (1.2x multiplier)</p>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center justify-between p-3 border border-amber-200 dark:border-amber-900/50 rounded-lg bg-amber-50 dark:bg-amber-950/20">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-amber-500/20 text-amber-600 flex items-center justify-center font-bold">G</div>
                      <div>
                        <p className="font-medium text-sm text-amber-900 dark:text-amber-500">Gold</p>
                        <p className="text-xs text-amber-700 dark:text-amber-600/80">5,000+ points (1.5x multiplier + free shipping)</p>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="referral">
          <Card className="border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm max-w-2xl">
            <CardHeader>
              <CardTitle>Referral Program</CardTitle>
              <CardDescription>Reward customers for bringing in new buyers.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div>
                <label className="text-sm font-medium mb-1.5 block">Referrer Bonus (Points)</label>
                <div className="flex gap-2">
                  <Input defaultValue="500" className="bg-white dark:bg-zinc-950" />
                  <Button variant="secondary">Save</Button>
                </div>
                <p className="text-xs text-zinc-500 mt-2">Points awarded to the referrer when a new customer completes their first order.</p>
              </div>

              <div>
                <label className="text-sm font-medium mb-1.5 block">Referred Friend Discount (%)</label>
                <div className="flex gap-2">
                  <Input defaultValue="10" className="bg-white dark:bg-zinc-950" />
                  <Button variant="secondary">Save</Button>
                </div>
                <p className="text-xs text-zinc-500 mt-2">Discount automatically applied to the new customer's first order.</p>
              </div>

              <div className="p-4 bg-zinc-50 dark:bg-zinc-900/50 rounded-lg border border-zinc-200 dark:border-zinc-800">
                <p className="text-sm font-medium mb-2">Example Referral Link</p>
                <code className="text-xs text-zinc-500 block p-2 bg-zinc-100 dark:bg-zinc-950 rounded border border-zinc-200 dark:border-zinc-800">
                  https://yourstore.com/?ref=CUST_8F92A
                </code>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
