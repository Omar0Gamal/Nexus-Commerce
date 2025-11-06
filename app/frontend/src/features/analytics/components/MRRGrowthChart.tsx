import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer
} from "recharts"

interface MRRGrowthChartProps {
  mrrHistory: { month: string; mrr: number }[]
}

export function MRRGrowthChart({ mrrHistory }: MRRGrowthChartProps) {
  return (
    <Card className="col-span-1 card-elevated border-border/40 dark:border-border/30 bg-card relative overflow-hidden">
      {/* Decorative top accent */}
      <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-violet-500 via-blue-500 to-cyan-500" />
      <CardHeader className="pt-5">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-base">MRR Growth</CardTitle>
            <CardDescription className="mt-1">Monthly recurring revenue over the past 6 months.</CardDescription>
          </div>
          {/* Legend */}
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <span className="size-2 rounded-full bg-gradient-to-r from-violet-500 to-cyan-500" />
            MRR
          </div>
        </div>
      </CardHeader>
      <CardContent className="h-[280px]">
        {mrrHistory.length === 0 ? (
           <div className="flex h-full items-center justify-center text-sm text-muted-foreground bg-muted/10 rounded-lg border border-dashed">
             No historical data available.
           </div>
        ) : (
           <ResponsiveContainer width="100%" height="100%">
             <AreaChart data={mrrHistory} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
               <defs>
                 <linearGradient id="colorMrr" x1="0" y1="0" x2="1" y2="0">
                   <stop offset="5%" stopColor="#8b5cf6" stopOpacity={1}/>
                   <stop offset="95%" stopColor="#06b6d4" stopOpacity={1}/>
                 </linearGradient>
                 <linearGradient id="fillMrr" x1="0" y1="0" x2="0" y2="1">
                   <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.15}/>
                   <stop offset="95%" stopColor="#06b6d4" stopOpacity={0.02}/>
                 </linearGradient>
               </defs>
               <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
               <XAxis dataKey="month" fontSize={12} tickLine={false} axisLine={false} />
               <YAxis 
                 fontSize={12} 
                 tickLine={false} 
                 axisLine={false} 
                 tickFormatter={(value) => `$${value / 1000}k`}
               />
               <Tooltip 
                 contentStyle={{ borderRadius: '10px', border: '1px solid hsl(var(--border))', backgroundColor: 'hsl(var(--card))', boxShadow: '0 4px 12px rgb(0 0 0 / 0.08)' }}
                 formatter={(value: any) => [`$${Number(value).toLocaleString()}`, 'MRR']}
               />
               <Area
                 type="monotone"
                 dataKey="mrr"
                 stroke="url(#colorMrr)"
                 strokeWidth={3}
                 fill="url(#fillMrr)"
                 dot={{ r: 4, fill: "#06b6d4", strokeWidth: 0 }}
                 activeDot={{ r: 6, fill: "#8b5cf6", stroke: "#fff", strokeWidth: 2 }}
               />
             </AreaChart>
           </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
