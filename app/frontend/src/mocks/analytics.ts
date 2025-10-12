import { TrendingUp, CreditCard, Store } from "lucide-react"

// Mock sparkline data for 7 days
export const mockSparklineData = {
  revenue: [{ v: 40 }, { v: 42 }, { v: 41 }, { v: 43 }, { v: 43 }, { v: 44 }, { v: 45.2 }],
  tenants: [{ v: 3 }, { v: 3 }, { v: 3 }, { v: 4 }, { v: 4 }, { v: 5 }, { v: 5 }],
  subs: [{ v: 110 }, { v: 112 }, { v: 115 }, { v: 120 }, { v: 122 }, { v: 125 }, { v: 128 }],
  health: [{ v: 100 }, { v: 100 }, { v: 99.8 }, { v: 100 }, { v: 99.9 }, { v: 100 }, { v: 99.9 }],
}

export const mockActivities = [
  {
    time: "2 hours ago",
    text: "New tenant 'Globex' upgraded to Enterprise.",
    icon: TrendingUp,
    color: "bg-emerald-500",
    lightColor: "bg-emerald-500/10",
    textColor: "text-emerald-600 dark:text-emerald-400",
  },
  {
    time: "5 hours ago",
    text: "Subscription renewed for 'Acme Corp'.",
    icon: CreditCard,
    color: "bg-blue-500",
    lightColor: "bg-blue-500/10",
    textColor: "text-blue-600 dark:text-blue-400",
  },
  {
    time: "1 day ago",
    text: "New tenant 'Massive Dynamic' joined.",
    icon: Store,
    color: "bg-violet-500",
    lightColor: "bg-violet-500/10",
    textColor: "text-violet-600 dark:text-violet-400",
  },
]

export const mockTopTenants = [
  { name: "Acme Corp", revenue: "$12,400", plan: "Enterprise" },
  { name: "Globex", revenue: "$8,900", plan: "Enterprise" },
  { name: "Soylent", revenue: "$5,200", plan: "Pro" },
  { name: "Massive Dynamic", revenue: "$4,100", plan: "Pro" },
]
