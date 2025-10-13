import { type LucideIcon, ArrowUpRight, ArrowDownRight } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { LineChart, Line, ResponsiveContainer } from "recharts"
import { accentColors, type AccentColor } from "../constants/colors"

export interface MetricCardProps {
  title: string
  value: string | number
  icon: LucideIcon
  trend: string
  trendUp?: boolean
  description?: string
  accent: typeof accentColors[AccentColor]
  sparklineData?: { v: number }[]
}

export function MetricCard({ 
  title, 
  value, 
  icon: Icon, 
  trend, 
  trendUp,
  description, 
  accent, 
  sparklineData 
}: MetricCardProps) {
  // If trendUp is explicit (Analytics Dashboard), use it. 
  // Otherwise, fallback to checking if trend starts with '+' (Overview Dashboard).
  const isTrendUp = trendUp !== undefined ? trendUp : trend.startsWith("+")

  // For Analytics dashboard which has a description, trend color depends on trendUp
  const trendColorClass = description 
    ? (isTrendUp ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400')
    : accent.trend;

  return (
    <Card className="group relative overflow-hidden card-elevated border-border/40 dark:border-border/30 bg-card hover:-translate-y-0.5 hover:ring-1 hover:ring-border/50 dark:hover:ring-border/30 transition-all duration-300 flex flex-col">
      {/* Left accent bar */}
      <div className={`absolute left-0 top-0 bottom-0 w-[3px] ${accent.bar} rounded-l-lg opacity-70 group-hover:opacity-100 transition-opacity z-10`} />

      <CardHeader className="flex flex-row items-center justify-between pb-2 pl-5 pt-5 pr-5 z-10">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        <div className={`flex size-9 items-center justify-center rounded-lg ${accent.iconBg} ring-1 ${accent.ring} transition-transform duration-300 group-hover:scale-110`}>
          <Icon className={`size-4 ${accent.iconText}`} />
        </div>
      </CardHeader>
      <CardContent className="pl-5 pr-5 pb-3 flex-1 flex flex-col justify-between z-10">
        <div>
          <div className="text-3xl font-bold tabular-nums tracking-tight text-foreground">
            {value}
          </div>
          
          {description ? (
            <p className="flex items-center text-xs text-muted-foreground mt-1.5 gap-1">
              <span className={`flex items-center gap-0.5 font-semibold ${trendColorClass}`}>
                {isTrendUp ? <ArrowUpRight className="size-3" /> : <ArrowDownRight className="size-3" />}
                {trend}
              </span>
              {description}
            </p>
          ) : (
            <p className={`flex items-center gap-1 text-xs font-medium mt-1 ${trendColorClass}`}>
              <ArrowUpRight className="size-3.5" />
              {trend}
            </p>
          )}
        </div>
        
        {/* Optional sparkline chart */}
        {sparklineData && (
          <div className="h-[28px] w-full mt-2 opacity-60 group-hover:opacity-100 transition-opacity">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={sparklineData}>
                <Line 
                  type="monotone" 
                  dataKey="v" 
                  stroke={accent.stroke} 
                  strokeWidth={2} 
                  dot={false} 
                  isAnimationActive={true}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
