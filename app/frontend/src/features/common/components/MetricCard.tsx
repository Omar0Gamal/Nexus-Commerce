import { type LucideIcon, ArrowUpRight, ArrowDownRight } from "lucide-react"
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
  const isTrendUp = trendUp !== undefined ? trendUp : trend.startsWith("+")

  return (
    <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02] transition-colors hover:bg-white/[0.04]">
      <div className="relative rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] p-5 flex flex-col justify-between overflow-hidden group min-h-[140px]">
        <div className="flex flex-row items-center justify-between pb-2 z-10">
          <div className="text-sm font-medium text-zinc-400">
            {title}
          </div>
          <div className={`flex size-9 items-center justify-center rounded-xl bg-zinc-950 border border-zinc-800 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] transition-transform duration-300 group-hover:scale-110`}>
            <Icon className="size-4 text-zinc-100" />
          </div>
        </div>
        
        <div className="flex-1 flex flex-col justify-end z-10">
          <div className="text-3xl font-semibold tabular-nums tracking-tight text-zinc-50">
            {value}
          </div>
          
          <div className="flex items-center text-xs text-zinc-500 mt-1.5 gap-1">
            <span className="flex items-center gap-0.5 font-medium text-zinc-300">
              {isTrendUp ? <ArrowUpRight className="size-3 text-zinc-400" /> : <ArrowDownRight className="size-3 text-zinc-400" />}
              {trend}
            </span>
            {description && <span>{description}</span>}
          </div>
        </div>
        
        {sparklineData && (
          <div className="absolute bottom-0 left-0 right-0 h-[40px] opacity-20 group-hover:opacity-40 transition-opacity pointer-events-none">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={sparklineData}>
                <Line 
                  type="monotone" 
                  dataKey="v" 
                  stroke="#fafafa" 
                  strokeWidth={2} 
                  dot={false} 
                  isAnimationActive={true}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>
    </div>
  )
}
