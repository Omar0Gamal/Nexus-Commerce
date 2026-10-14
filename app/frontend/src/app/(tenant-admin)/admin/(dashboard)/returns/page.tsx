"use client"

import * as React from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
  RotateCcw,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Loader2,
  Search,
} from "lucide-react"
import { useAuthStore } from "@/store/useAuthStore"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export default function ReturnsPage() {
  const { token, activeShop } = useAuthStore()
  const queryClient = useQueryClient()
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || ""

  const { data: returnsData, isLoading } = useQuery({
    queryKey: ["returns", activeShop?.id],
    queryFn: async () => {
      const res = await fetch(`${apiUrl}/api/v1/returns`, {
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Shop-ID": activeShop?.id || "",
        },
      })
      if (!res.ok) throw new Error("Failed to fetch returns")
      const returnsResponse = await res.json()
      return returnsResponse.data || []
    },
    enabled: !!token && !!activeShop?.id,
  })

  const refundMutation = useMutation({
    mutationFn: async (returnId: string) => {
      const res = await fetch(`${apiUrl}/api/v1/returns/${returnId}/refund`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Shop-ID": activeShop?.id || "",
        },
      })
      if (!res.ok) throw new Error("Failed to process refund")
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["returns"] })
    },
  })

  const formatStatus = (status: string) => {
    switch (status.toLowerCase()) {
      case "pending":
        return <Badge variant="outline" className="text-yellow-600 border-yellow-600 bg-yellow-50 dark:bg-yellow-950 dark:text-yellow-400">Pending</Badge>
      case "approved":
        return <Badge variant="outline" className="text-emerald-600 border-emerald-600 bg-emerald-50 dark:bg-emerald-950 dark:text-emerald-400">Approved</Badge>
      case "rejected":
        return <Badge variant="outline" className="text-red-600 border-red-600 bg-red-50 dark:bg-red-950 dark:text-red-400">Rejected</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  return (
    <div className="flex flex-col gap-8 p-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-950 dark:text-zinc-50">Returns & RMAs</h1>
          <p className="text-zinc-500 mt-2">Manage customer return requests and process refunds.</p>
        </div>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800 shadow-sm rounded-xl overflow-hidden">
        <div className="p-4 border-b border-zinc-200 dark:border-zinc-800 flex items-center gap-4 bg-zinc-50/50 dark:bg-zinc-900/50">
          <div className="relative w-72">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
            <Input placeholder="Search returns..." className="pl-9 bg-white dark:bg-zinc-950" />
          </div>
        </div>

        {isLoading ? (
          <div className="p-12 flex justify-center">
            <Loader2 className="w-8 h-8 animate-spin text-zinc-400" />
          </div>
        ) : !returnsData || returnsData.length === 0 ? (
          <div className="p-24 flex flex-col items-center justify-center text-center">
            <RotateCcw className="w-16 h-16 text-zinc-200 dark:text-zinc-800 mb-6 stroke-[1]" />
            <h3 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">No return requests</h3>
            <p className="text-sm text-zinc-500 max-w-sm mt-2">When customers request a return or refund, they will appear here for you to process.</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent border-zinc-200 dark:border-zinc-800">
                <TableHead>Return ID</TableHead>
                <TableHead>Order #</TableHead>
                <TableHead>Customer</TableHead>
                <TableHead>Reason</TableHead>
                <TableHead>Date</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {returnsData.map((ret: any) => (
                <TableRow key={ret.id} className="border-zinc-200 dark:border-zinc-800 hover:bg-zinc-50 dark:hover:bg-zinc-900/50 transition-colors">
                  <TableCell className="font-mono text-xs">{ret.id.substring(0, 8)}</TableCell>
                  <TableCell className="font-mono text-xs text-zinc-500">{ret.order_id?.substring(0, 8)}</TableCell>
                  <TableCell>{ret.customer_email || "Customer"}</TableCell>
                  <TableCell className="max-w-[200px] truncate">{ret.reason}</TableCell>
                  <TableCell className="text-zinc-500">
                    {new Date(ret.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell>{formatStatus(ret.status)}</TableCell>
                  <TableCell className="text-right">
                    {ret.status === "pending" && (
                      <Button
                        size="sm"
                        variant="default"
                        onClick={() => refundMutation.mutate(ret.id)}
                        disabled={refundMutation.isPending}
                        className="h-8 text-xs font-medium rounded-lg"
                      >
                        {refundMutation.isPending ? "Processing..." : "Approve Refund"}
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </Card>
    </div>
  )
}
