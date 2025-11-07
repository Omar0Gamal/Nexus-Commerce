import { Activity, ChevronRight, LucideIcon } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"

export interface ActivityItem {
  time: string
  text: string
  icon: LucideIcon
  color: string
  lightColor: string
  textColor: string
}

interface RecentActivityListProps {
  activities: ActivityItem[]
}

export function RecentActivityList({ activities }: RecentActivityListProps) {
  return (
    <Card className="col-span-2 card-elevated border-border/40 dark:border-border/20 bg-card/90 dark:bg-card/50 backdrop-blur-xl">
      <CardHeader className="pb-3 pt-5 px-5">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-sm">Recent Activity</CardTitle>
            <CardDescription className="text-xs">Latest tenant and platform events</CardDescription>
          </div>
          <div className="flex size-7 items-center justify-center rounded-lg bg-muted/50 text-muted-foreground">
            <Activity className="size-3.5" />
          </div>
        </div>
      </CardHeader>
      <CardContent className="px-5 pb-5">
        <div className="flex flex-col gap-3">
          {activities.map((activity, i) => (
            <div 
              key={i} 
              className="group/item cursor-pointer relative flex items-center justify-between p-3 rounded-xl border border-transparent bg-muted/20 hover:bg-muted/40 hover:border-border/40 hover:shadow-sm transition-all duration-300"
            >
              <div className="flex items-center gap-4">
                <div className="relative z-10 flex-shrink-0">
                  {/* Ambient glow behind icon */}
                  <div className={`absolute inset-0 rounded-xl ${activity.lightColor} blur-md opacity-0 group-hover/item:opacity-70 transition-opacity duration-300`} />
                  <div className={`relative flex size-10 items-center justify-center rounded-xl ${activity.lightColor} ring-1 ring-background group-hover/item:scale-105 group-hover/item:shadow-sm transition-all duration-300`}>
                    <activity.icon className={`size-4 ${activity.textColor}`} />
                  </div>
                </div>

                <div className="flex flex-col gap-1 min-w-0">
                  <p className="text-sm font-semibold leading-snug text-foreground/90 group-hover/item:text-foreground transition-colors">
                    {activity.text}
                  </p>
                  <span className="text-[11px] font-medium text-muted-foreground/60 uppercase tracking-wider">
                    {activity.time}
                  </span>
                </div>
              </div>
              <div className="opacity-0 -translate-x-3 group-hover/item:opacity-100 group-hover/item:translate-x-0 transition-all duration-300 text-muted-foreground pr-2">
                <ChevronRight className="size-4" />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
