import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer
} from "recharts"

interface SubscriptionChurnChartProps {
  churnHistory: { month: string; new: number; churned: number }[]
}

export function SubscriptionChurnChart({ churnHistory }: SubscriptionChurnChartProps) {
  return (
    <Card className="col-span-1 card-elevated border-border/40 dark:border-border/30 bg-card relative overflow-hidden">
      {/* Decorative top accent */}
      <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-violet-500 via-rose-500 to-rose-400" />
      <CardHeader className="pt-5">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-base">Subscription Churn</CardTitle>
            <CardDescription className="mt-1">Lost vs new subscriptions.</CardDescription>
          </div>
          {/* Legend */}
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1.5">
              <span className="size-2 rounded-full bg-violet-500" />
              New
            </span>
            <span className="flex items-center gap-1.5">
              <span className="size-2 rounded-full bg-rose-500" />
              Churned
            </span>
          </div>
        </div>
      </CardHeader>
      <CardContent className="h-[280px]">
        {churnHistory.length === 0 ? (
           <div className="flex h-full items-center justify-center text-sm text-muted-foreground bg-muted/10 rounded-lg border border-dashed">
             No churn data available.
           </div>
        ) : (
           <ResponsiveContainer width="100%" height="100%">
             <BarChart data={churnHistory} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
               <defs>
                 <linearGradient id="colorNew" x1="0" y1="0" x2="0" y2="1">
                   <stop offset="5%" stopColor="#8b5cf6" stopOpacity={1}/>
                   <stop offset="95%" stopColor="#3b82f6" stopOpacity={1}/>
                 </linearGradient>
                 <linearGradient id="colorChurn" x1="0" y1="0" x2="0" y2="1">
                   <stop offset="5%" stopColor="#f43f5e" stopOpacity={1}/>
                   <stop offset="95%" stopColor="#9f1239" stopOpacity={1}/>
                 </linearGradient>
               </defs>
               <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--muted))" />
               <XAxis dataKey="month" fontSize={12} tickLine={false} axisLine={false} />
               <YAxis fontSize={12} tickLine={false} axisLine={false} />
               <Tooltip 
                 contentStyle={{ borderRadius: '10px', border: '1px solid hsl(var(--border))', backgroundColor: 'hsl(var(--card))', boxShadow: '0 4px 12px rgb(0 0 0 / 0.08)' }}
                 cursor={{ fill: 'hsl(var(--muted))', opacity: 0.2 }}
               />
               <Bar dataKey="new" name="New Subscriptions" fill="url(#colorNew)" radius={[6, 6, 0, 0]} />
               <Bar dataKey="churned" name="Churned" fill="url(#colorChurn)" radius={[6, 6, 0, 0]} />
             </BarChart>
           </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
