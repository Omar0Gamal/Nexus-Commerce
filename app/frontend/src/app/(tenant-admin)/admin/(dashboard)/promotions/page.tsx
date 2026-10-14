"use client"

import * as React from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { Sparkles, Loader2, Plus, Calendar, Tag } from "lucide-react"
import { useAuthStore } from "@/store/useAuthStore"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"

export default function PromotionsPage() {
  const { token, activeShop } = useAuthStore()
  const queryClient = useQueryClient()
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || ""

  const [isCreateOpen, setIsCreateOpen] = React.useState(false)
  const [newName, setNewName] = React.useState("")
  const [newType, setNewType] = React.useState("flash_sale")
  const [newValue, setNewValue] = React.useState("")

  const { data: promotions, isLoading } = useQuery({
    queryKey: ["promotions", activeShop?.id],
    queryFn: async () => {
      const res = await fetch(`${apiUrl}/api/v1/promotions`, {
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Shop-ID": activeShop?.id || "",
        },
      })
      if (!res.ok) throw new Error("Failed to fetch promotions")
      const promotionsResponse = await res.json()
      return promotionsResponse.data || []
    },
    enabled: !!token && !!activeShop?.id,
  })

  const createMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch(`${apiUrl}/api/v1/promotions`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Shop-ID": activeShop?.id || "",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: newName,
          promotion_type: newType,
          discount_value: parseFloat(newValue),
          start_date: new Date().toISOString(),
          end_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
          is_active: true,
        }),
      })
      if (!res.ok) throw new Error("Failed to create promotion")
      return res.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["promotions"] })
      setIsCreateOpen(false)
      setNewName("")
      setNewValue("")
    },
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    createMutation.mutate()
  }

  return (
    <div className="flex flex-col gap-8 p-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-950 dark:text-zinc-50">Promotions & Flash Sales</h1>
          <p className="text-zinc-500 mt-2">Create and manage discounts, flash sales, and special offers.</p>
        </div>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger className="inline-flex items-center justify-center rounded-xl h-10 px-4 text-sm font-medium bg-zinc-900 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200 transition-colors">
            <Plus className="w-4 h-4 mr-2" />
            Create Promotion
          </DialogTrigger>
          <DialogContent className="sm:max-w-md bg-white dark:bg-zinc-950 border border-zinc-200 dark:border-zinc-800 rounded-2xl shadow-xl">
            <DialogHeader>
              <DialogTitle>Create New Promotion</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4 pt-4">
              <div>
                <label className="text-sm font-medium mb-1.5 block">Promotion Name</label>
                <Input
                  required
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="e.g. Summer Flash Sale"
                  className="bg-white dark:bg-zinc-900"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium mb-1.5 block">Type</label>
                  <select
                    value={newType}
                    onChange={(e) => setNewType(e.target.value)}
                    className="w-full h-10 rounded-md border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-3 text-sm"
                  >
                    <option value="flash_sale">Flash Sale</option>
                    <option value="volume_discount">Volume Discount</option>
                  </select>
                </div>
                <div>
                  <label className="text-sm font-medium mb-1.5 block">Discount Value (%)</label>
                  <Input
                    type="number"
                    required
                    min="1"
                    max="100"
                    value={newValue}
                    onChange={(e) => setNewValue(e.target.value)}
                    placeholder="20"
                    className="bg-white dark:bg-zinc-900"
                  />
                </div>
              </div>
              <Button
                type="submit"
                disabled={createMutation.isPending}
                className="w-full h-11 rounded-xl font-medium mt-4 bg-zinc-900 text-white hover:bg-zinc-800 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
              >
                {createMutation.isPending ? "Creating..." : "Create Promotion"}
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800 shadow-sm rounded-xl overflow-hidden">
        {isLoading ? (
          <div className="p-12 flex justify-center">
            <Loader2 className="w-8 h-8 animate-spin text-zinc-400" />
          </div>
        ) : !promotions || promotions.length === 0 ? (
          <div className="p-24 flex flex-col items-center justify-center text-center">
            <Sparkles className="w-16 h-16 text-zinc-200 dark:text-zinc-800 mb-6 stroke-[1]" />
            <h3 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">No active promotions</h3>
            <p className="text-sm text-zinc-500 max-w-sm mt-2">Create a flash sale or volume discount to boost your sales.</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent border-zinc-200 dark:border-zinc-800">
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Value</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>End Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {promotions.map((promo: any) => (
                <TableRow key={promo.id} className="border-zinc-200 dark:border-zinc-800 hover:bg-zinc-50 dark:hover:bg-zinc-900/50 transition-colors">
                  <TableCell className="font-medium">{promo.name}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2 text-zinc-500">
                      {promo.promotion_type === 'flash_sale' ? <Tag className="w-3.5 h-3.5" /> : <Sparkles className="w-3.5 h-3.5" />}
                      <span className="capitalize">{promo.promotion_type.replace('_', ' ')}</span>
                    </div>
                  </TableCell>
                  <TableCell className="font-mono">{promo.discount_value}%</TableCell>
                  <TableCell>
                    {promo.is_active ? (
                      <Badge variant="outline" className="text-emerald-600 border-emerald-600 bg-emerald-50 dark:bg-emerald-950 dark:text-emerald-400">Active</Badge>
                    ) : (
                      <Badge variant="outline" className="text-zinc-500 border-zinc-300">Inactive</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-zinc-500">
                    <div className="flex items-center gap-2">
                      <Calendar className="w-3.5 h-3.5" />
                      {new Date(promo.end_date).toLocaleDateString()}
                    </div>
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
