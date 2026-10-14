"use client"

import * as React from "react"
import { useQueryState } from 'nuqs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Package, MoreHorizontal, Edit, Trash, Search, Filter } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useGetApiV1CatalogProducts } from "@/api/endpoints/catalog/catalog";

interface ProductData {
  id: string
  title: string
  status: string
  price: number
  currency: string
  inventory_quantity: number
  image_url?: string
}

export function ProductsTable() {
  const [search, setSearch] = useQueryState('search', { defaultValue: '' })
  const [statusFilter, setStatusFilter] = useQueryState('status', { defaultValue: 'all' })

  const { data: responseData, isLoading, isError } = useGetApiV1CatalogProducts()
  
  const catalogPayload = responseData as any
  const products: ProductData[] = catalogPayload?.data || []

  const filteredProducts = products.filter(p => {
    const matchesSearch = p.title.toLowerCase().includes(search.toLowerCase())
    const matchesStatus = statusFilter === 'all' || p.status.toLowerCase() === statusFilter
    return matchesSearch && matchesStatus
  })

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-zinc-500" />
            <Input 
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search products..."
              className="pl-9 bg-zinc-900 border-zinc-800 text-zinc-100 placeholder:text-zinc-600 focus-visible:ring-zinc-700 w-[300px] rounded-xl"
            />
          </div>
          <select 
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="h-9 px-3 rounded-xl bg-zinc-900 border border-zinc-800 text-sm font-medium text-zinc-300 focus:outline-none focus:ring-2 focus:ring-zinc-700 transition-shadow"
          >
            <option value="all">All Statuses</option>
            <option value="active">Active</option>
            <option value="draft">Draft</option>
          </select>
        </div>
        <button className="group relative flex h-9 items-center justify-center rounded-full bg-zinc-100 px-4 text-sm font-medium text-zinc-900 hover:bg-zinc-200 transition-all active:scale-[0.98]">
          <span>Add Product</span>
          <span className="ml-2 flex h-5 w-5 items-center justify-center rounded-full bg-white/10 transition-transform group-hover:translate-x-0.5 group-hover:scale-105">
            +
          </span>
        </button>
      </div>

      <div className="p-1.5 rounded-[2rem] border border-white/[0.05] bg-white/[0.02]">
        <div className="rounded-[1.625rem] bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] overflow-hidden">
          
          {isLoading ? (
            <Table>
              <TableHeader className="border-zinc-800/50 bg-zinc-900/50">
                <TableRow className="border-zinc-800/50 hover:bg-transparent">
                  <TableHead className="text-zinc-500">Product</TableHead>
                  <TableHead className="text-zinc-500">Status</TableHead>
                  <TableHead className="text-zinc-500">Inventory</TableHead>
                  <TableHead className="text-zinc-500">Price</TableHead>
                  <TableHead className="w-[80px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {Array.from({ length: 5 }).map((_, i) => (
                  <TableRow key={i} className="border-zinc-800/30">
                    <TableCell><Skeleton className="h-6 w-[200px] bg-zinc-800" /></TableCell>
                    <TableCell><Skeleton className="h-6 w-[80px] bg-zinc-800" /></TableCell>
                    <TableCell><Skeleton className="h-6 w-[60px] bg-zinc-800" /></TableCell>
                    <TableCell><Skeleton className="h-6 w-[80px] bg-zinc-800" /></TableCell>
                    <TableCell><Skeleton className="h-8 w-8 rounded-md bg-zinc-800" /></TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : isError ? (
            <div className="p-16 text-center text-red-400">
              <p className="font-medium tracking-tight">Failed to load products.</p>
              <p className="text-sm mt-1 text-red-500/70 font-mono">Check network connection</p>
            </div>
          ) : filteredProducts.length === 0 ? (
            <div className="p-16 text-center flex flex-col items-center justify-center">
              <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-zinc-800/50 border border-zinc-700/50 mb-4 shadow-inner">
                <Package className="h-8 w-8 text-zinc-500" />
              </div>
              <h3 className="text-lg font-medium text-zinc-100 tracking-tight">No products found</h3>
              <p className="text-sm mt-1 text-zinc-500 max-w-xs mx-auto">
                {search ? `No results match "${search}". Try clearing filters.` : "Get started by adding your first product."}
              </p>
              {search && (
                <button onClick={() => setSearch('')} className="mt-6 text-sm text-zinc-400 hover:text-zinc-100 underline decoration-zinc-700 underline-offset-4">
                  Clear search
                </button>
              )}
            </div>
          ) : (
            <Table>
              <TableHeader className="border-zinc-800/50 bg-zinc-900/50">
                <TableRow className="border-zinc-800/50 hover:bg-transparent">
                  <TableHead className="text-zinc-500 font-medium tracking-tight">Product</TableHead>
                  <TableHead className="text-zinc-500 font-medium tracking-tight">Status</TableHead>
                  <TableHead className="text-zinc-500 font-medium tracking-tight">Inventory</TableHead>
                  <TableHead className="text-zinc-500 font-medium tracking-tight">Price</TableHead>
                  <TableHead className="w-[80px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredProducts.map((product) => (
                  <TableRow key={product.id} className="border-zinc-800/30 hover:bg-zinc-800/30 transition-colors group">
                    <TableCell className="font-medium text-zinc-100">
                      <div className="flex items-center gap-4">
                        {product.image_url ? (
                          <img 
                            src={product.image_url} 
                            alt={product.title} 
                            className="h-10 w-10 rounded-xl object-cover ring-1 ring-white/10"
                          />
                        ) : (
                          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-zinc-800 ring-1 ring-white/5">
                            <Package className="h-4 w-4 text-zinc-500" />
                          </div>
                        )}
                        <span className="tracking-tight">{product.title}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge 
                        variant="secondary"
                        className={`font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-md ${
                          product.status?.toLowerCase() === 'active' 
                            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' 
                            : 'bg-zinc-800 text-zinc-400 border-zinc-700'
                        }`}
                      >
                        {product.status || 'Draft'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className={`font-mono text-sm ${product.inventory_quantity <= 5 ? "text-amber-400" : "text-zinc-400"}`}>
                        {product.inventory_quantity ?? 0}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="font-mono text-sm text-zinc-300">
                        {new Intl.NumberFormat('en-US', {
                          style: 'currency',
                          currency: product.currency || 'USD'
                        }).format(product.price || 0)}
                      </span>
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger className="inline-flex items-center justify-center rounded-md h-8 w-8 p-0 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-zinc-700">
                          <MoreHorizontal className="h-4 w-4 text-zinc-400" />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-[160px] rounded-xl border-zinc-700 bg-zinc-800">
                          <DropdownMenuItem className="cursor-pointer text-zinc-300 focus:bg-zinc-700 focus:text-zinc-50">
                            <Edit className="mr-2 h-4 w-4" /> Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem className="cursor-pointer text-red-400 focus:bg-red-400/10 focus:text-red-400">
                            <Trash className="mr-2 h-4 w-4" /> Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </div>
    </div>
  )
}
