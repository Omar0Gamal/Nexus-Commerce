"use client"

import * as React from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { Star, MessageSquare, Loader2, Search, Check, X } from "lucide-react"
import { useAuthStore } from "@/store/useAuthStore"
import { Card } from "@/components/ui/card"
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

export default function ReviewsPage() {
  const { token, activeShop } = useAuthStore()
  const queryClient = useQueryClient()
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || ""
  
  const [statusFilter, setStatusFilter] = React.useState("pending")

  const { data: reviewsData, isLoading } = useQuery({
    queryKey: ["reviews", activeShop?.id, statusFilter],
    queryFn: async () => {
      const url = new URL(`${apiUrl}/api/v1/reviews/staff`)
      if (statusFilter !== "all") {
        url.searchParams.append("status", statusFilter)
      }
      
      const res = await fetch(url.toString(), {
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Shop-ID": activeShop?.id || "",
        },
      })
      if (!res.ok) throw new Error("Failed to fetch reviews")
      const reviewsResponse = await res.json()
      return reviewsResponse.data || []
    },
    enabled: !!token && !!activeShop?.id,
  })

  const updateStatusMutation = useMutation({
    mutationFn: async ({ id, status }: { id: string, status: string }) => {
      const res = await fetch(`${apiUrl}/api/v1/reviews/staff/${id}`, {
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Shop-ID": activeShop?.id || "",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ status }),
      })
      if (!res.ok) throw new Error("Failed to update review status")
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reviews"] })
    },
  })

  return (
    <div className="flex flex-col gap-8 p-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-950 dark:text-zinc-50">Review Moderation</h1>
          <p className="text-zinc-500 mt-2">Approve or reject customer product reviews before they appear on the storefront.</p>
        </div>
      </div>

      <div className="flex gap-2 border-b border-zinc-200 dark:border-zinc-800 pb-px">
        {["pending", "approved", "rejected", "all"].map(tab => (
          <button
            key={tab}
            onClick={() => setStatusFilter(tab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              statusFilter === tab 
                ? "border-zinc-900 text-zinc-900 dark:border-zinc-100 dark:text-zinc-100" 
                : "border-transparent text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300"
            }`}
          >
            <span className="capitalize">{tab}</span>
          </button>
        ))}
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800 shadow-sm rounded-xl overflow-hidden">
        <div className="p-4 border-b border-zinc-200 dark:border-zinc-800 flex items-center gap-4 bg-zinc-50/50 dark:bg-zinc-900/50">
          <div className="relative w-72">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
            <Input placeholder="Search reviews..." className="pl-9 bg-white dark:bg-zinc-950" />
          </div>
        </div>

        {isLoading ? (
          <div className="p-12 flex justify-center">
            <Loader2 className="w-8 h-8 animate-spin text-zinc-400" />
          </div>
        ) : !reviewsData || reviewsData.length === 0 ? (
          <div className="p-24 flex flex-col items-center justify-center text-center">
            <MessageSquare className="w-16 h-16 text-zinc-200 dark:text-zinc-800 mb-6 stroke-[1]" />
            <h3 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">No {statusFilter !== 'all' ? statusFilter : ''} reviews</h3>
            <p className="text-sm text-zinc-500 max-w-sm mt-2">When customers leave reviews for your products, they will appear here.</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent border-zinc-200 dark:border-zinc-800">
                <TableHead>Customer</TableHead>
                <TableHead>Product</TableHead>
                <TableHead>Rating</TableHead>
                <TableHead className="w-1/2">Review</TableHead>
                <TableHead>Date</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {reviewsData.map((review: any) => (
                <TableRow key={review.id} className="border-zinc-200 dark:border-zinc-800 hover:bg-zinc-50 dark:hover:bg-zinc-900/50 transition-colors">
                  <TableCell className="font-medium">{review.customer_name || review.customer_email || "Anonymous"}</TableCell>
                  <TableCell className="text-zinc-500 max-w-[150px] truncate" title={review.product_id}>{review.product_id.substring(0, 8)}</TableCell>
                  <TableCell>
                    <div className="flex items-center text-amber-500">
                      {Array.from({ length: 5 }).map((_, i) => (
                        <Star key={i} className={`w-3.5 h-3.5 ${i < review.rating ? 'fill-current' : 'text-zinc-200 dark:text-zinc-800'}`} />
                      ))}
                    </div>
                  </TableCell>
                  <TableCell>
                    <p className="font-medium text-sm text-zinc-900 dark:text-zinc-100">{review.title}</p>
                    <p className="text-sm text-zinc-500 line-clamp-2 mt-0.5">{review.content}</p>
                  </TableCell>
                  <TableCell className="text-zinc-500 text-sm whitespace-nowrap">
                    {new Date(review.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell className="text-right whitespace-nowrap">
                    {review.status === "pending" ? (
                      <div className="flex justify-end gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => updateStatusMutation.mutate({ id: review.id, status: "rejected" })}
                          disabled={updateStatusMutation.isPending}
                          className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950/50"
                        >
                          <X className="w-4 h-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => updateStatusMutation.mutate({ id: review.id, status: "approved" })}
                          disabled={updateStatusMutation.isPending}
                          className="h-8 w-8 p-0 text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-950/50"
                        >
                          <Check className="w-4 h-4" />
                        </Button>
                      </div>
                    ) : (
                      <Badge variant="outline" className={`capitalize ${
                        review.status === 'approved' 
                          ? 'text-emerald-600 border-emerald-200 dark:border-emerald-900/50' 
                          : 'text-red-600 border-red-200 dark:border-red-900/50'
                      }`}>
                        {review.status}
                      </Badge>
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
