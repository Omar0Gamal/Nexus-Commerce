"use client"

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useGetApiV1AnalyticsCohorts } from "@/api/endpoints/analytics/analytics"

export function CohortRetention() {
  const { data: cohortsResponse, isLoading, isError } = useGetApiV1AnalyticsCohorts()

  const rawCohorts = cohortsResponse?.data?.cohorts || []
  const rawMatrix = cohortsResponse?.data?.matrix || []

  const getColor = (percentage: number | null | undefined) => {
    if (percentage === null || percentage === undefined) return "bg-transparent"
    if (percentage >= 80) return "bg-emerald-500/25 text-emerald-400 font-medium"
    if (percentage >= 60) return "bg-emerald-500/15 text-emerald-300"
    if (percentage >= 40) return "bg-primary/40 text-foreground"
    if (percentage >= 20) return "bg-primary/25 text-foreground/80"
    if (percentage > 0) return "bg-primary/10 text-muted-foreground"
    return "bg-muted/20 text-muted-foreground/60"
  }

  const maxOffsets = Math.max(
    5,
    rawMatrix.reduce((max, row) => Math.max(max, row.length), 0)
  )
  const displayedOffsets = Math.min(maxOffsets, 6)

  return (
    <Card className="card-elevated border-border/40 overflow-hidden">
      <CardHeader>
        <CardTitle className="text-base font-medium">Cohort Retention</CardTitle>
        <CardDescription>User retention rate across periodic cohorts</CardDescription>
      </CardHeader>
      <CardContent className="overflow-x-auto">
        {isLoading ? (
          <div className="space-y-3 py-2">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        ) : isError ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            Failed to load cohort analytics from server.
          </div>
        ) : rawCohorts.length === 0 ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            No cohort retention data available yet. Data accumulates as customers return.
          </div>
        ) : (
          <table className="w-full text-sm text-left">
            <thead className="text-muted-foreground border-b border-border/50">
              <tr>
                <th className="font-medium pb-3 pr-4">Cohort</th>
                {Array.from({ length: displayedOffsets }).map((_, idx) => (
                  <th key={idx} className="font-medium pb-3 px-2 text-center">
                    {`Period ${idx}`}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rawCohorts.map((cohortName, i) => {
                const row = rawMatrix[i] || []
                return (
                  <tr key={cohortName} className="border-b border-border/20 last:border-0 hover:bg-muted/10 transition-colors">
                    <td className="py-3 pr-4 whitespace-nowrap font-medium text-foreground">{cohortName}</td>
                    {Array.from({ length: displayedOffsets }).map((_, j) => {
                      const retentionRate = row[j]
                      return (
                        <td key={j} className="p-1">
                          {retentionRate !== undefined && retentionRate !== null ? (
                            <div className={`flex items-center justify-center py-2 rounded-xl text-xs ${getColor(retentionRate)}`}>
                              {Math.round(retentionRate)}%
                            </div>
                          ) : (
                            <div className="flex items-center justify-center py-2 text-xs text-muted-foreground/30">
                              -
                            </div>
                          )}
                        </td>
                      )
                    })}
                  </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </CardContent>
    </Card>
  )
}
