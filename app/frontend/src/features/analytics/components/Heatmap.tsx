"use client"

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useMemo } from "react"
import { cn } from "@/lib/utils"
import { useGetApiV1AnalyticsTrend } from "@/api/endpoints/analytics/analytics"

interface HeatmapProps {
  title?: string
  description?: string
  data?: { date: string; count: number }[]
}

export function Heatmap({ 
  title = "Activity Heatmap", 
  description = "Daily user actions over the last 30 days",
  data 
}: HeatmapProps) {
  const { data: trendData, isLoading } = useGetApiV1AnalyticsTrend({
    query: { enabled: !data }
  })

  const heatmapData = useMemo(() => {
    if (data) return data

    const pointsMap = new Map<string, number>()
    if (trendData?.data) {
      for (const p of trendData.data) {
        if (p.date) {
          pointsMap.set(p.date, Number(p.visits || p.unique || 0))
        }
      }
    }

    const activityTimeline: { date: string; count: number }[] = []
    const today = new Date()
    for (let i = 29; i >= 0; i--) {
      const d = new Date(today)
      d.setDate(d.getDate() - i)
      const dateStr = d.toISOString().split("T")[0]
      activityTimeline.push({
        date: dateStr,
        count: pointsMap.get(dateStr) ?? 0,
      })
    }
    return activityTimeline
  }, [data, trendData])

  const getColor = (count: number) => {
    if (count === 0) return "bg-muted/40 hover:bg-muted/60"
    if (count < 3) return "bg-emerald-500/30 hover:bg-emerald-500/40"
    if (count < 6) return "bg-emerald-500/60 hover:bg-emerald-500/70"
    if (count < 9) return "bg-emerald-500/80 hover:bg-emerald-500/90"
    return "bg-emerald-500 hover:bg-emerald-400"
  }

  return (
    <Card className="card-elevated border-border/40">
      <CardHeader>
        <CardTitle className="text-base font-medium">{title}</CardTitle>
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent>
        {isLoading && !data ? (
          <div className="flex flex-wrap gap-2 py-2">
            {Array.from({ length: 30 }).map((_, i) => (
              <Skeleton key={i} className="size-6 rounded-sm" />
            ))}
          </div>
        ) : (
          <div className="flex flex-wrap gap-2">
            {heatmapData.map((day) => (
              <div
                key={day.date}
                className={cn(
                  "size-6 rounded-sm cursor-pointer transition-colors border border-border/20",
                  getColor(day.count)
                )}
                title={`${day.date}: ${day.count} events`}
              />
            ))}
          </div>
        )}
        <div className="mt-4 flex items-center gap-2 text-xs text-muted-foreground">
          <span>Less</span>
          <div className="size-3.5 rounded-sm bg-muted/40 border border-border/20" />
          <div className="size-3.5 rounded-sm bg-emerald-500/30" />
          <div className="size-3.5 rounded-sm bg-emerald-500/60" />
          <div className="size-3.5 rounded-sm bg-emerald-500/80" />
          <div className="size-3.5 rounded-sm bg-emerald-500" />
          <span>More</span>
        </div>
      </CardContent>
    </Card>
  )
}
