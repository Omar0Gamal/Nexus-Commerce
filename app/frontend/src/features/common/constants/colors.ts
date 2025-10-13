export const accentColors = {
  emerald: {
    bar: "bg-emerald-500",
    iconBg: "bg-emerald-500/10 dark:bg-emerald-500/15",
    iconText: "text-emerald-600 dark:text-emerald-400",
    trend: "text-emerald-600 dark:text-emerald-400",
    ring: "ring-emerald-500/20",
    stroke: "#10b981",
  },
  blue: {
    bar: "bg-blue-500",
    iconBg: "bg-blue-500/10 dark:bg-blue-500/15",
    iconText: "text-blue-600 dark:text-blue-400",
    trend: "text-blue-600 dark:text-blue-400",
    ring: "ring-blue-500/20",
    stroke: "#3b82f6",
  },
  violet: {
    bar: "bg-violet-500",
    iconBg: "bg-violet-500/10 dark:bg-violet-500/15",
    iconText: "text-violet-600 dark:text-violet-400",
    trend: "text-violet-600 dark:text-violet-400",
    ring: "ring-violet-500/20",
    stroke: "#8b5cf6",
  },
  amber: {
    bar: "bg-amber-500",
    iconBg: "bg-amber-500/10 dark:bg-amber-500/15",
    iconText: "text-amber-600 dark:text-amber-400",
    trend: "text-amber-600 dark:text-amber-400",
    ring: "ring-amber-500/20",
    stroke: "#f59e0b",
  },
} as const;

export type AccentColor = keyof typeof accentColors;
