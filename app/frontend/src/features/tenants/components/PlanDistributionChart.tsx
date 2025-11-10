"use client"

import { PieChart as PieChartIcon } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts"

const COLORS = ['#8b5cf6', '#3b82f6', '#10b981', '#f59e0b']

interface PlanDistributionChartProps {
  totalShops: number
  planData: { name: string; value: number }[]
  isLoading?: boolean
}

export function PlanDistributionChart({ totalShops, planData, isLoading }: PlanDistributionChartProps) {
  return (
    <Card className="col-span-1 lg:col-span-1 card-elevated border border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl flex flex-col relative overflow-hidden group">
      <div className="absolute top-0 right-0 w-40 h-40 bg-primary/5 rounded-full blur-3xl -mr-10 -mt-10 group-hover:bg-primary/10 transition-colors duration-500" />
      <CardHeader className="pb-0 pt-6 px-6 z-10">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-base font-semibold">Plan Distribution</CardTitle>
            <CardDescription className="text-xs mt-1">Tenant subscription tiers</CardDescription>
          </div>
          <div className="flex size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground ring-1 ring-border/50 group-hover:text-primary group-hover:ring-primary/30 transition-colors">
            <PieChartIcon className="size-4" />
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col items-center justify-center relative min-h-[220px] px-6 pb-6 z-10">
        {isLoading ? (
          <div className="flex h-full w-full items-center justify-center">
            <Skeleton className="h-36 w-36 rounded-full" />
          </div>
        ) : (
          <>
            <div className="h-[180px] w-full mt-4 relative">
              {/* Inner Label */}
              <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                <span className="text-4xl font-bold text-foreground">{totalShops}</span>
                <span className="text-[10px] uppercase font-bold tracking-wider text-muted-foreground">Total</span>
              </div>
              
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={planData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={5}
                    dataKey="value"
                    stroke="none"
                    cornerRadius={4}
                  >
                    {planData.map((entry, index) => (
                      <Cell 
                        key={`cell-${index}`} 
                        fill={COLORS[index % COLORS.length]} 
                        className="hover:opacity-80 transition-opacity outline-none cursor-pointer"
                        style={{ filter: `drop-shadow(0px 4px 6px ${COLORS[index % COLORS.length]}40)` }}
                      />
                    ))}
                  </Pie>
                  <Tooltip 
                    cursor={false}
                    contentStyle={{ 
                      borderRadius: '12px', 
                      border: '1px solid hsl(var(--border)/0.5)', 
                      backgroundColor: 'hsl(var(--card)/0.9)',
                      backdropFilter: 'blur(8px)',
                      fontSize: '13px',
                      fontWeight: 600,
                      boxShadow: '0 10px 25px -5px rgba(0,0,0,0.1)'
                    }} 
                    itemStyle={{ color: 'hsl(var(--foreground))' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
            
            {/* Custom Legend */}
            <div className="flex flex-wrap justify-center gap-x-5 gap-y-3 mt-4 w-full">
              {planData.map((entry, index) => (
                <div key={entry.name} className="flex items-center gap-2 text-xs font-semibold text-foreground/80 group/legend cursor-default">
                  <div className="relative flex items-center justify-center">
                    <div className="absolute inset-0 rounded-full blur-sm opacity-50 group-hover/legend:opacity-100 transition-opacity" style={{ backgroundColor: COLORS[index % COLORS.length] }} />
                    <div className="relative w-3 h-3 rounded-full border-2 border-background" style={{ backgroundColor: COLORS[index % COLORS.length] }} />
                  </div>
                  {entry.name}
                  <span className="text-muted-foreground ml-1">({entry.value})</span>
                </div>
              ))}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
