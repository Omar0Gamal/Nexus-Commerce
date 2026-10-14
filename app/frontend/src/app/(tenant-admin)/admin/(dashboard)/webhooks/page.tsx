"use client"

import * as React from "react"
import { Webhook, Plus, Trash2, ShieldAlert } from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

export default function WebhooksPage() {
  return (
    <div className="flex flex-col gap-8 p-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-950 dark:text-zinc-50">Webhooks</h1>
          <p className="text-zinc-500 mt-2">Subscribe to real-time events for your applications.</p>
        </div>
        <Button className="rounded-xl h-10 px-4 font-medium bg-zinc-900 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200">
          <Plus className="w-4 h-4 mr-2" />
          Add Webhook
        </Button>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm">
        <CardContent className="p-0">
          <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
            <div className="p-6 flex flex-col gap-4">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-medium text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
                    Order Sync
                    <Badge variant="outline" className="text-emerald-600 border-emerald-600 bg-emerald-50 dark:bg-emerald-950 dark:text-emerald-400">Active</Badge>
                  </h3>
                  <code className="text-sm text-zinc-500 mt-1 block">https://api.erp-system.example.com/webhooks/nexus</code>
                </div>
                <Button variant="ghost" size="icon" className="text-zinc-400 hover:text-red-500">
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
              <div>
                <p className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-2">Events</p>
                <div className="flex gap-2">
                  <Badge variant="secondary" className="font-mono text-xs">order.created</Badge>
                  <Badge variant="secondary" className="font-mono text-xs">order.paid</Badge>
                </div>
              </div>
              <div className="flex items-center gap-2 text-xs text-zinc-500 bg-zinc-50 dark:bg-zinc-900/50 p-2 rounded-md border border-zinc-200 dark:border-zinc-800">
                <ShieldAlert className="w-3 h-3" />
                <span>Secret: whsec_••••••••••••••••••••••••</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
