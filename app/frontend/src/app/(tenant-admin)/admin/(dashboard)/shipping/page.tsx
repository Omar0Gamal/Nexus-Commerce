"use client"

import * as React from "react"
import { Globe, Plus, MoreHorizontal } from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

export default function ShippingPage() {
  return (
    <div className="flex flex-col gap-8 p-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-950 dark:text-zinc-50">Shipping Zones</h1>
          <p className="text-zinc-500 mt-2">Manage where you ship and how much it costs.</p>
        </div>
        <Button className="rounded-xl h-10 px-4 font-medium bg-zinc-900 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200">
          <Plus className="w-4 h-4 mr-2" />
          Add Zone
        </Button>
      </div>

      <div className="grid gap-6">
        <Card className="border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm overflow-hidden">
          <div className="bg-zinc-50 dark:bg-zinc-900/50 p-4 border-b border-zinc-200 dark:border-zinc-800 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Globe className="w-5 h-5 text-zinc-400" />
              <div>
                <h3 className="font-semibold text-zinc-900 dark:text-zinc-100">Domestic (Egypt)</h3>
                <p className="text-sm text-zinc-500">1 region</p>
              </div>
            </div>
            <Button variant="ghost" size="icon">
              <MoreHorizontal className="w-4 h-4" />
            </Button>
          </div>
          <CardContent className="p-0">
            <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
              <div className="p-4 flex items-center justify-between">
                <div>
                  <p className="font-medium">Standard Shipping</p>
                  <p className="text-sm text-zinc-500">0 kg - 5 kg</p>
                </div>
                <div className="flex items-center gap-4">
                  <span className="font-mono">EGP 50.00</span>
                </div>
              </div>
              <div className="p-4 flex items-center justify-between">
                <div>
                  <p className="font-medium">Heavy Goods</p>
                  <p className="text-sm text-zinc-500">5.01 kg - 20 kg</p>
                </div>
                <div className="flex items-center gap-4">
                  <span className="font-mono">EGP 150.00</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm overflow-hidden">
          <div className="bg-zinc-50 dark:bg-zinc-900/50 p-4 border-b border-zinc-200 dark:border-zinc-800 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Globe className="w-5 h-5 text-zinc-400" />
              <div>
                <h3 className="font-semibold text-zinc-900 dark:text-zinc-100">Rest of World</h3>
                <p className="text-sm text-zinc-500">195 regions</p>
              </div>
            </div>
            <Button variant="ghost" size="icon">
              <MoreHorizontal className="w-4 h-4" />
            </Button>
          </div>
          <CardContent className="p-0">
            <div className="p-4 flex items-center justify-between">
              <div>
                <p className="font-medium">International Express</p>
                <p className="text-sm text-zinc-500">0 kg - 2 kg</p>
              </div>
              <div className="flex items-center gap-4">
                <span className="font-mono">EGP 400.00</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
