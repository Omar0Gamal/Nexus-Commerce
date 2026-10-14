"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import { Search, AlertTriangle, Info, CheckCircle2, Link as LinkIcon, Plus } from "lucide-react"
import { useAuthStore } from "@/store/useAuthStore"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"

export default function SEOPage() {
  const { token, activeShop } = useAuthStore()
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || ""

  const { data: audit, isLoading } = useQuery({
    queryKey: ["seo-audit", activeShop?.id],
    queryFn: async () => {
      const res = await fetch(`${apiUrl}/api/v1/seo/audit`, {
        headers: { Authorization: `Bearer ${token}`, "X-Shop-ID": activeShop?.id || "" }
      })
      if (!res.ok) throw new Error("Failed to fetch")
      return (await res.json()).data
    },
    enabled: !!token && !!activeShop?.id,
  })

  const score = audit?.score || 85;

  return (
    <div className="flex flex-col gap-8 p-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-zinc-950 dark:text-zinc-50">SEO & Redirects</h1>
        <p className="text-zinc-500 mt-2">Manage your search engine visibility and URL redirects.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="col-span-1 border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm flex flex-col items-center justify-center p-8">
          <div className="relative w-40 h-40 flex items-center justify-center rounded-full border-[12px] border-zinc-100 dark:border-zinc-900">
            <svg className="absolute inset-0 w-full h-full -rotate-90" viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="44" fill="none" stroke="currentColor" strokeWidth="12" className="text-emerald-500" strokeDasharray={`${(score / 100) * 276} 276`} strokeLinecap="round" />
            </svg>
            <div className="text-center">
              <span className="text-4xl font-bold tracking-tight">{score}</span>
              <span className="block text-xs text-zinc-500 uppercase tracking-widest font-semibold mt-1">Score</span>
            </div>
          </div>
        </Card>

        <Card className="col-span-2 border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm">
          <CardHeader>
            <CardTitle>Recommendations</CardTitle>
            <CardDescription>Actionable steps to improve your SEO score.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-start gap-3 p-3 rounded-lg bg-zinc-50 dark:bg-zinc-900/50 border border-zinc-200 dark:border-zinc-800">
              <AlertTriangle className="w-5 h-5 text-amber-500 mt-0.5" />
              <div>
                <p className="font-medium text-sm">Missing Meta Descriptions</p>
                <p className="text-sm text-zinc-500">3 products are missing meta descriptions. This affects click-through rates.</p>
              </div>
            </div>
            <div className="flex items-start gap-3 p-3 rounded-lg bg-zinc-50 dark:bg-zinc-900/50 border border-zinc-200 dark:border-zinc-800">
              <Info className="w-5 h-5 text-blue-500 mt-0.5" />
              <div>
                <p className="font-medium text-sm">Image Alt Tags</p>
                <p className="text-sm text-zinc-500">12 images have generic or missing alt text. Describe your images for accessibility and SEO.</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm">
        <div className="p-6 border-b border-zinc-200 dark:border-zinc-800 flex justify-between items-center">
          <div>
            <CardTitle>URL Redirects</CardTitle>
            <CardDescription className="mt-1">Manage 301 and 302 redirects for your storefront.</CardDescription>
          </div>
          <Button className="rounded-xl h-10 px-4 font-medium bg-zinc-900 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200">
            <Plus className="w-4 h-4 mr-2" />
            Add Redirect
          </Button>
        </div>
        <CardContent className="p-0">
          <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
            <div className="p-4 flex items-center justify-between hover:bg-zinc-50 dark:hover:bg-zinc-900/50">
              <div className="flex items-center gap-3">
                <LinkIcon className="w-4 h-4 text-zinc-400" />
                <code className="text-sm bg-zinc-100 dark:bg-zinc-900 px-2 py-1 rounded">/old-collection</code>
                <span className="text-zinc-400">→</span>
                <code className="text-sm bg-zinc-100 dark:bg-zinc-900 px-2 py-1 rounded">/collections/summer</code>
              </div>
              <span className="text-xs text-zinc-500 bg-zinc-100 dark:bg-zinc-900 px-2 py-1 rounded">301 Permanent</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
