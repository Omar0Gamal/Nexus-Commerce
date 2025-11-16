import { useGetApiV1PlatformAdminHealth } from "@/api/endpoints/admin/admin"
import type { HealthData } from "../types"

export function useHealthData() {
  const { data: healthEnvelope, isLoading, isError } = useGetApiV1PlatformAdminHealth()
  const healthData = (healthEnvelope?.data as unknown as HealthData) || {}

  return {
    healthData,
    isLoading,
    isError,
  }
}
