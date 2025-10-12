export type ShopPlan = "Free" | "Starter" | "Pro" | "Enterprise"
export type ShopStatus = "active" | "trialing" | "past_due" | "suspended"

export interface Shop {
  id: string
  name: string
  subdomain: string
  plan_name: ShopPlan
  status: ShopStatus
  created_at?: string
}

export const shops: Shop[] = [
  {
    id: "8f14e45f-ceea-467d-9a1e-3c2b8a0d1f01",
    name: "Northwind Coffee",
    subdomain: "northwind",
    plan_name: "Pro",
    status: "active",
  },
  {
    id: "2c1a7b3e-9d5f-4a12-b8e0-6f4d2a9c7e10",
    name: "Atlas Outdoors",
    subdomain: "atlas-outdoors",
    plan_name: "Enterprise",
    status: "active",
  },
  {
    id: "5e9c0d21-8b74-49f3-a2c6-1d7e4f8b0a33",
    name: "Bloom & Petal",
    subdomain: "bloompetal",
    plan_name: "Starter",
    status: "trialing",
  },
  {
    id: "a3d8f612-4c9e-4d70-91b2-7e5a2c1f9d04",
    name: "Pixel Forge Studio",
    subdomain: "pixelforge",
    plan_name: "Pro",
    status: "active",
  },
  {
    id: "b7e2c948-1f30-4a6d-8c15-9d3e6a7b2f88",
    name: "Harbor Freight Co",
    subdomain: "harborfreight",
    plan_name: "Enterprise",
    status: "past_due",
  },
  {
    id: "d4f1a093-6e28-4b57-a9d3-2c8f5e0b1a76",
    name: "Sunrise Bakery",
    subdomain: "sunrise",
    plan_name: "Free",
    status: "active",
  },
  {
    id: "c0a5b731-2d84-4e19-b6f7-3a1e9c8d4f22",
    name: "Vertex Analytics",
    subdomain: "vertex",
    plan_name: "Pro",
    status: "active",
  },
  {
    id: "e6b3d270-9a41-4c83-8e05-1f7d2a6b9c50",
    name: "Meridian Health",
    subdomain: "meridian-health",
    plan_name: "Enterprise",
    status: "active",
  },
  {
    id: "f1c8e402-3b76-4d95-a0e2-8c4f1d7a6b93",
    name: "Cobalt Fitness",
    subdomain: "cobalt",
    plan_name: "Starter",
    status: "suspended",
  },
  {
    id: "0d9a2f68-7c15-4e30-b8a4-6f2e9d1c5a07",
    name: "Willow & Wren",
    subdomain: "willowwren",
    plan_name: "Pro",
    status: "active",
  },
  {
    id: "3a7f1b90-5d62-4c48-9e13-2b8a6f4d0c15",
    name: "Ironclad Security",
    subdomain: "ironclad",
    plan_name: "Enterprise",
    status: "active",
  },
  {
    id: "9c4e0a83-1b57-4f26-a7d9-5e3c8b2f6a41",
    name: "Lumen Media",
    subdomain: "lumen",
    plan_name: "Starter",
    status: "trialing",
  },
]
