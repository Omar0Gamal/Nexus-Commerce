"use client"

import { Database, Zap, Bot, Activity, type LucideIcon } from "lucide-react"

import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"
import { useHealthData } from "../hooks/useHealthData"
import type { HealthState } from "../types"

type ServiceHealth = {
  name: string
  icon: LucideIcon
  state: HealthState
  metricLabel: string
  metricValue: string
  detail: string
}

const stateConfig: Record<
  HealthState,
  { label: string; dot: string; text: string; ring: string; bar: string; iconBg: string; glow: string }
> = {
  operational: {
    label: "Operational",
    dot: "bg-emerald-500",
    text: "text-emerald-600 dark:text-emerald-400",
    ring: "ring-emerald-500/20",
    bar: "bg-emerald-500",
    iconBg: "bg-emerald-500/10 dark:bg-emerald-500/15",
    glow: "shadow-emerald-500/10",
  },
  degraded: {
    label: "Degraded",
    dot: "bg-amber-500",
    text: "text-amber-600 dark:text-amber-400",
    ring: "ring-amber-500/20",
    bar: "bg-amber-500",
    iconBg: "bg-amber-500/10 dark:bg-amber-500/15",
    glow: "shadow-amber-500/10",
  },
  down: {
    label: "Outage",
    dot: "bg-red-500",
    text: "text-red-600 dark:text-red-400",
    ring: "ring-red-500/20",
    bar: "bg-red-500",
    iconBg: "bg-red-500/10 dark:bg-red-500/15",
    glow: "shadow-red-500/10",
  },
  unknown: {
    label: "Unknown",
    dot: "bg-muted-foreground",
    text: "text-muted-foreground",
    ring: "ring-muted-foreground/20",
    bar: "bg-muted-foreground",
    iconBg: "bg-muted/50",
    glow: "",
  },
}

export function InfraHealthCards() {
  const { healthData, isLoading } = useHealthData()

  if (isLoading) {
    return (
      <section className="grid grid-cols-1 gap-4 md:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i} className="overflow-hidden card-elevated border-border/40">
            <div className="h-[2px] w-full bg-muted" />
            <CardHeader className="flex flex-row items-center justify-between gap-2 pb-2 pt-5">
              <div className="flex items-center gap-2.5">
                <Skeleton className="size-10 rounded-xl" />
                <Skeleton className="h-4 w-28" />
              </div>
              <Skeleton className="h-5 w-20 rounded-full" />
            </CardHeader>
            <CardContent className="flex flex-col gap-1.5">
              <div className="flex items-baseline justify-between">
                <Skeleton className="h-3 w-20" />
                <Skeleton className="h-6 w-16" />
              </div>
              <Skeleton className="mt-1 h-3 w-3/4" />
            </CardContent>
          </Card>
        ))}
      </section>
    )
  }

  if (!healthData || Object.keys(healthData).length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center border rounded-xl border-dashed bg-muted/10">
        <div className="flex size-12 items-center justify-center rounded-full bg-muted mb-4">
          <Activity className="size-6 text-muted-foreground" />
        </div>
        <h3 className="text-lg font-medium text-foreground">No telemetry data</h3>
        <p className="text-sm text-muted-foreground mt-2 max-w-sm">
          Infrastructure health metrics are currently unavailable. Ensure telemetry agents are active.
        </p>
      </div>
    )
  }

  const dynamicServices: ServiceHealth[] = [
    {
      name: "Database Health",
      icon: Database,
      state: (healthData.database?.state as HealthState) || "unknown",
      metricLabel: "Primary latency",
      metricValue: healthData.database?.latency || "N/A",
      detail: healthData.database?.detail || "No data reported",
    },
    {
      name: "Redis Health",
      icon: Zap,
      state: (healthData.redis?.state as HealthState) || "unknown",
      metricLabel: "Cache hit rate",
      metricValue: healthData.redis?.hit_rate || "N/A",
      detail: healthData.redis?.detail || "No data reported",
    },
    {
      name: "AI Engine Status",
      icon: Bot,
      state: (healthData.ai?.state as HealthState) || "unknown",
      metricLabel: "Inference queue",
      metricValue: healthData.ai?.queue || "N/A",
      detail: healthData.ai?.detail || "No data reported",
    },
  ]

  return (
    <section
      aria-label="Infrastructure health"
      className="grid grid-cols-1 gap-4 md:grid-cols-3 stagger-children"
    >
      {dynamicServices.map((service) => {
        const config = stateConfig[service.state]
        return (
          <Card
            key={service.name}
            className={cn(
              "group relative overflow-hidden card-elevated border-border/40 dark:border-border/30 bg-card hover:-translate-y-0.5 transition-all duration-300",
              config.glow
            )}
          >
            {/* Top accent bar */}
            <div className={cn("absolute top-0 left-0 right-0 h-[2px]", config.bar)} aria-hidden="true" />

            <CardHeader className="flex flex-row items-center justify-between gap-2 pb-2 pt-5">
              <div className="flex items-center gap-3">
                <div className={cn(
                  "flex size-10 items-center justify-center rounded-xl ring-1 transition-transform duration-300 group-hover:scale-110",
                  config.iconBg,
                  config.ring,
                )}>
                  <service.icon className={cn("size-5", config.text)} aria-hidden="true" />
                </div>
                <h3 className="text-sm font-semibold">{service.name}</h3>
              </div>

              {/* Status badge */}
              <div className={cn("flex items-center gap-2 text-xs font-semibold px-2.5 py-1 rounded-full", config.text, config.iconBg)}>
                <span className="relative flex size-2">
                  {service.state !== "down" ? (
                    <span
                      className={cn(
                        "absolute inline-flex size-full animate-ping rounded-full opacity-60",
                        config.dot,
                      )}
                    />
                  ) : null}
                  <span className={cn("relative inline-flex size-2 rounded-full", config.dot)} />
                </span>
                {config.label}
              </div>
            </CardHeader>

            <CardContent className="flex flex-col gap-1.5">
              <div className="flex items-baseline justify-between">
                <span className="text-xs text-muted-foreground/70">{service.metricLabel}</span>
                <span className="text-xl font-bold tabular-nums tracking-tight">{service.metricValue}</span>
              </div>
              <p className="text-xs text-muted-foreground/60 text-pretty leading-relaxed">{service.detail}</p>
            </CardContent>
          </Card>
        )
      })}
    </section>
  )
}
