export type HealthState = "operational" | "degraded" | "down" | "unknown"

export interface HealthData {
  database?: { state: string; latency: string; detail: string };
  redis?: { state: string; hit_rate: string; detail: string };
  ai?: { state: string; queue: string; detail: string };
}
