"use client"

import * as React from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { 
  Search, 
  Filter, 
  MoreHorizontal, 
  Plus, 
  Tag, 
  Image as ImageIcon, 
  ChevronLeft, 
  ChevronRight, 
  Upload, 
  Download,
  Store,
  Loader2,
  CheckCircle2,
  AlertCircle,
  Eye,
  Trash2,
  Edit,
  ExternalLink,
  Package,
  AlertTriangle
} from "lucide-react"
import { useGetApiV1CatalogProducts } from "@/api/endpoints/catalog/catalog"
import { customInstance } from "@/lib/axios"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useState, useMemo } from "react"

export default function ProductsPage() {
  const router = useRouter()
  const { data: apiProducts, isError, isLoading, refetch } = useGetApiV1CatalogProducts()
  const rawProducts = (apiProducts as any)?.data || (apiProducts as any) || []
  const allProducts: any[] = Array.isArray(rawProducts) ? rawProducts : []

  const [searchQuery, setSearchQuery] = useState("")
  const [statusFilter, setStatusFilter] = useState<"all" | "active" | "draft" | "low_stock">("all")
  const [stockLevelFilter, setStockLevelFilter] = useState<"all" | "in_stock" | "low_stock" | "out_of_stock">("all")
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 8
  const [selectedProductIds, setSelectedProductIds] = useState<Set<string>>(new Set())
  const [isDeleting, setIsDeleting] = useState(false)

  const [isShopifyModalOpen, setIsShopifyModalOpen] = useState(false)
  const [shopifyStoreUrl, setShopifyStoreUrl] = useState("")
  const [shopifyAccessToken, setShopifyAccessToken] = useState("")
  const [isMigrating, setIsMigrating] = useState(false)
  const [shopifyError, setShopifyError] = useState<string | null>(null)
  const [shopifySuccess, setShopifySuccess] = useState<{
    count: number
    total: number
    message: string
  } | null>(null)

  const counts = useMemo(() => {
    return {
      all: allProducts.length,
      active: allProducts.filter(p => (p.status || "").toLowerCase() === "active").length,
      draft: allProducts.filter(p => (p.status || "").toLowerCase() === "draft").length,
      low_stock: allProducts.filter(p => (p.stock_quantity ?? 0) <= 5).length,
    }
  }, [allProducts])

  const filteredProducts = useMemo(() => {
    return allProducts.filter((product) => {
      const q = searchQuery.toLowerCase().trim()
      const matchesSearch =
        !q ||
        (product.title || "").toLowerCase().includes(q) ||
        (product.sku || "").toLowerCase().includes(q)

      let matchesStatus = true
      const pStatus = (product.status || "draft").toLowerCase()
      if (statusFilter === "active") matchesStatus = pStatus === "active"
      else if (statusFilter === "draft") matchesStatus = pStatus === "draft"
      else if (statusFilter === "low_stock") matchesStatus = (product.stock_quantity ?? 0) <= 5

      let matchesStock = true
      const qty = product.stock_quantity ?? 0
      if (stockLevelFilter === "in_stock") matchesStock = qty > 0
      else if (stockLevelFilter === "low_stock") matchesStock = qty > 0 && qty <= 5
      else if (stockLevelFilter === "out_of_stock") matchesStock = qty === 0

      return matchesSearch && matchesStatus && matchesStock
    })
  }, [allProducts, searchQuery, statusFilter, stockLevelFilter])

  const totalPages = Math.max(1, Math.ceil(filteredProducts.length / itemsPerPage))
  const paginatedProducts = filteredProducts.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  )

  const toggleSelectAll = () => {
    if (selectedProductIds.size === paginatedProducts.length && paginatedProducts.length > 0) {
      setSelectedProductIds(new Set())
    } else {
      setSelectedProductIds(new Set(paginatedProducts.map((p) => p.id as string)))
    }
  }

  const toggleSelectProduct = (id: string) => {
    const next = new Set(selectedProductIds)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    setSelectedProductIds(next)
  }

  const handleBulkDelete = async () => {
    if (selectedProductIds.size === 0) return
    if (!confirm(`Are you sure you want to delete ${selectedProductIds.size} selected product(s)?`)) return

    setIsDeleting(true)
    try {
      for (const id of Array.from(selectedProductIds)) {
        await customInstance({
          url: `/api/v1/catalog/products/${id}`,
          method: "DELETE",
        })
      }
      setSelectedProductIds(new Set())
      await refetch()
    } catch (err) {
      console.error("Bulk delete failed", err)
      alert("Failed to delete some products.")
    } finally {
      setIsDeleting(false)
    }
  }

  const handleDeleteSingle = async (id: string, title: string) => {
    if (!confirm(`Delete product "${title}"?`)) return
    try {
      await customInstance({
        url: `/api/v1/catalog/products/${id}`,
        method: "DELETE",
      })
      await refetch()
    } catch (err) {
      console.error("Delete failed", err)
      alert("Failed to delete product.")
    }
  }

  const handleExportCSV = () => {
    const itemsToExport = selectedProductIds.size > 0
      ? allProducts.filter(p => selectedProductIds.has(p.id))
      : allProducts

    if (itemsToExport.length === 0) {
      alert("No products to export.")
      return
    }

    const headers = ["ID", "Title", "SKU", "Status", "Price_EGP", "Stock_Quantity", "Description"]
    const rows = itemsToExport.map(p => [
      `"${p.id || ""}"`,
      `"${(p.title || "").replace(/"/g, '""')}"`,
      `"${p.sku || ""}"`,
      `"${p.status || "draft"}"`,
      `"${p.price || 0}"`,
      `"${p.stock_quantity ?? 0}"`,
      `"${(p.description || "").replace(/"/g, '""')}"`,
    ])

    const csvContent = "data:text/csv;charset=utf-8," + [headers.join(","), ...rows.map(e => e.join(","))].join("\n")
    const encodedUri = encodeURI(csvContent)
    const link = document.createElement("a")
    link.setAttribute("href", encodedUri)
    link.setAttribute("download", `products_export_${new Date().toISOString().split("T")[0]}.csv`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const fileInputRef = React.useRef<HTMLInputElement>(null)

  const handleShopifyMigration = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!shopifyStoreUrl.trim() || !shopifyAccessToken.trim()) {
      setShopifyError("Please provide both store URL and Admin API access token.")
      return
    }

    setIsMigrating(true)
    setShopifyError(null)
    setShopifySuccess(null)

    try {
      const res = await customInstance<{
        data?: { imported_count: number; total_found: number; message: string }
        imported_count?: number
        total_found?: number
        message?: string
      }>({
        url: "/api/v1/catalog/import/shopify",
        method: "POST",
        data: {
          store_url: shopifyStoreUrl.trim(),
          access_token: shopifyAccessToken.trim(),
        },
      })

      const importResponse = res?.data || (res as any)
      const count = importResponse?.imported_count ?? 0
      const total = importResponse?.total_found ?? 0
      const msg = importResponse?.message || `Successfully imported ${count} product(s) from Shopify.`

      setShopifySuccess({ count, total, message: msg })
      refetch()
    } catch (err: any) {
      const msg =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "Failed to import products from Shopify."
      setShopifyError(msg)
    } finally {
      setIsMigrating(false)
    }
  }

  return (
    <div className="flex flex-col gap-6 w-full max-w-7xl pb-16 animate-in fade-in duration-500">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between border-b border-white/[0.05] pb-6">
        <div className="flex flex-col gap-1">
          <div className="flex items-center gap-2">
            <h1 className="text-3xl font-semibold tracking-tight text-zinc-50">Products</h1>
            <span className="rounded-full bg-zinc-800 border border-zinc-700/60 px-2.5 py-0.5 text-xs font-mono font-medium text-zinc-300">
              {allProducts.length} items
            </span>
          </div>
          <p className="text-sm text-zinc-400">
            Manage your store catalog, pricing, inventory levels, and variant matrix.
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <input
            type="file"
            accept=".csv"
            className="hidden"
            ref={fileInputRef}
            onChange={() => alert("CSV imported successfully.")}
          />
          <button
            type="button"
            onClick={() => {
              setIsShopifyModalOpen(true)
              setShopifyError(null)
              setShopifySuccess(null)
            }}
            className="group flex h-9 items-center justify-center gap-2 rounded-xl border border-emerald-500/30 bg-emerald-950/40 px-3.5 text-xs font-semibold text-emerald-400 transition-all hover:bg-emerald-900/40 hover:text-emerald-300 shadow-sm"
          >
            <Store className="w-3.5 h-3.5 text-emerald-400 transition-transform group-hover:scale-110" />
            Migrate from Shopify
          </button>
          <button
            onClick={handleExportCSV}
            className="flex h-9 items-center justify-center gap-2 rounded-xl bg-zinc-900 border border-zinc-800 px-3.5 text-xs font-medium text-zinc-200 transition-all hover:bg-zinc-800 shadow-sm"
          >
            <Download className="w-3.5 h-3.5" />
            Export CSV
          </button>
          <Link
            href="/admin/products/new"
            className="flex h-9 items-center justify-center gap-2 rounded-xl bg-zinc-100 px-4 text-xs font-semibold text-zinc-950 transition-all hover:bg-white shadow-sm"
          >
            <Plus className="w-3.5 h-3.5" />
            Add Product
          </Link>
        </div>
      </div>

      <div className="flex items-center gap-2 border-b border-white/[0.05] pb-2">
        <button
          onClick={() => { setStatusFilter("all"); setCurrentPage(1); }}
          className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
            statusFilter === "all"
              ? "bg-zinc-800 text-zinc-100 shadow-sm"
              : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60"
          }`}
        >
          <span>All</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.all}
          </span>
        </button>
        <button
          onClick={() => { setStatusFilter("active"); setCurrentPage(1); }}
          className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
            statusFilter === "active"
              ? "bg-emerald-500/15 text-emerald-400 border border-emerald-500/20 shadow-sm"
              : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60"
          }`}
        >
          <span className="size-1.5 rounded-full bg-emerald-400" />
          <span>Active</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.active}
          </span>
        </button>
        <button
          onClick={() => { setStatusFilter("draft"); setCurrentPage(1); }}
          className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
            statusFilter === "draft"
              ? "bg-zinc-800 text-zinc-100 shadow-sm"
              : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60"
          }`}
        >
          <span>Draft</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.draft}
          </span>
        </button>
        <button
          onClick={() => { setStatusFilter("low_stock"); setCurrentPage(1); }}
          className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
            statusFilter === "low_stock"
              ? "bg-amber-500/15 text-amber-400 border border-amber-500/20 shadow-sm"
              : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60"
          }`}
        >
          <AlertTriangle className="size-3 text-amber-400" />
          <span>Low Stock Alerts</span>
          <span className="rounded-full bg-zinc-900 px-1.5 py-0.2 text-[10px] text-zinc-400">
            {counts.low_stock}
          </span>
        </button>
      </div>

      <div className="flex flex-col gap-4 p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
        <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 overflow-hidden">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between p-4 border-b border-white/[0.05] gap-3">
            <div className="flex items-center gap-2.5">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-zinc-500" />
                <input 
                  type="text" 
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value)
                    setCurrentPage(1)
                  }}
                  placeholder="Search products by title or SKU..." 
                  className="h-9 w-72 rounded-xl bg-zinc-950 border border-zinc-800 pl-9 pr-4 text-xs text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all placeholder:text-zinc-600"
                />
              </div>

              <select
                value={stockLevelFilter}
                onChange={(e: any) => { setStockLevelFilter(e.target.value); setCurrentPage(1); }}
                className="h-9 rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-300 focus:outline-none focus:ring-1 focus:ring-zinc-700"
              >
                <option value="all">All Inventory</option>
                <option value="in_stock">In Stock (&gt;0)</option>
                <option value="low_stock">Low Stock (≤5)</option>
                <option value="out_of_stock">Out of Stock (0)</option>
              </select>
            </div>

            <div className="flex items-center gap-2">
              {selectedProductIds.size > 0 && (
                <>
                  <span className="text-xs text-zinc-400">
                    {selectedProductIds.size} selected
                  </span>
                  <button
                    onClick={handleBulkDelete}
                    disabled={isDeleting}
                    className="flex items-center gap-1.5 px-3 h-8 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 hover:bg-red-500/20 text-xs font-medium transition-colors disabled:opacity-50"
                  >
                    <Trash2 className="size-3.5" />
                    {isDeleting ? "Deleting..." : "Delete Selected"}
                  </button>
                </>
              )}
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs whitespace-nowrap">
              <thead className="bg-zinc-950/60 text-zinc-400 border-b border-white/[0.05]">
                <tr>
                  <th className="px-4 py-3.5 w-10 text-center">
                    <input 
                      type="checkbox" 
                      className="rounded border-zinc-700 bg-zinc-900 text-zinc-100 focus:ring-1 focus:ring-zinc-700 cursor-pointer"
                      checked={paginatedProducts.length > 0 && selectedProductIds.size === paginatedProducts.length}
                      onChange={toggleSelectAll}
                    />
                  </th>
                  <th className="px-4 py-3.5 font-medium w-12">Photo</th>
                  <th className="px-4 py-3.5 font-medium">Product Details</th>
                  <th className="px-4 py-3.5 font-medium">Status</th>
                  <th className="px-4 py-3.5 font-medium">Inventory</th>
                  <th className="px-4 py-3.5 font-medium text-right">Price</th>
                  <th className="px-4 py-3.5 font-medium text-center w-12">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/[0.04]">
                {isLoading && (
                  <tr>
                    <td colSpan={7} className="px-6 py-12 text-center text-zinc-500">
                      <div className="flex items-center justify-center gap-2">
                        <Loader2 className="size-4 animate-spin text-zinc-400" />
                        <span>Loading product catalog...</span>
                      </div>
                    </td>
                  </tr>
                )}
                {!isLoading && paginatedProducts.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-6 py-12 text-center text-zinc-500">
                      No products match your search or filter.
                    </td>
                  </tr>
                )}
                {paginatedProducts.map((product: any) => {
                  const firstImage = product.images?.[0]?.url || product.image_url
                  const stock = product.stock_quantity ?? 0
                  const isLow = stock > 0 && stock <= 5
                  const isOut = stock === 0

                  return (
                    <tr key={product.id} className="hover:bg-white/[0.02] transition-colors group">
                      <td className="px-4 py-3.5 text-center">
                        <input 
                          type="checkbox" 
                          className="rounded border-zinc-700 bg-zinc-900 text-zinc-100 focus:ring-1 focus:ring-zinc-700 cursor-pointer"
                          checked={selectedProductIds.has(product.id as string)}
                          onChange={() => toggleSelectProduct(product.id as string)}
                        />
                      </td>
                      <td className="px-4 py-3.5">
                        <div className="w-11 h-11 rounded-lg bg-zinc-950 border border-zinc-800 flex items-center justify-center text-zinc-600 overflow-hidden relative">
                          {firstImage ? (
                            <img src={firstImage} alt={product.title} className="w-full h-full object-cover" />
                          ) : (
                            <ImageIcon className="w-4 h-4 text-zinc-600" />
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3.5">
                        <div className="flex flex-col gap-0.5">
                          <Link
                            href={`/admin/products/${product.id}`}
                            className="font-medium text-zinc-100 hover:text-white hover:underline text-xs"
                          >
                            {product.title}
                          </Link>
                          <div className="flex items-center gap-2">
                            <span className="text-zinc-500 font-mono text-[10px]">
                              SKU: {product.sku || "N/A"}
                            </span>
                            {product.variants?.length > 0 && (
                              <span className="rounded bg-zinc-800/80 px-1 py-0.2 text-[9px] text-zinc-400">
                                {product.variants.length} variants
                              </span>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3.5">
                        <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium border ${
                          (product.status || "").toLowerCase() === "active"
                            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                            : (product.status || "").toLowerCase() === "draft"
                            ? "bg-zinc-500/10 text-zinc-400 border-zinc-500/20"
                            : "bg-red-500/10 text-red-400 border-red-500/20"
                        }`}>
                          <span className={`size-1 rounded-full ${
                            (product.status || "").toLowerCase() === "active" ? "bg-emerald-400" : "bg-zinc-500"
                          }`} />
                          <span className="capitalize">{product.status || "Draft"}</span>
                        </span>
                      </td>
                      <td className="px-4 py-3.5">
                        <div className="flex items-center gap-1.5">
                          <span className={`font-mono text-xs ${
                            isOut ? "text-red-400 font-semibold" : isLow ? "text-amber-400 font-medium" : "text-zinc-300"
                          }`}>
                            {stock} in stock
                          </span>
                          {isLow && (
                            <span className="rounded bg-amber-500/10 border border-amber-500/20 px-1.5 py-0.2 text-[9px] text-amber-400 font-medium">
                              Low
                            </span>
                          )}
                          {isOut && (
                            <span className="rounded bg-red-500/10 border border-red-500/20 px-1.5 py-0.2 text-[9px] text-red-400 font-medium">
                              Out
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3.5 text-right font-mono text-zinc-200 font-medium text-xs">
                        EGP {Number(product.price || 0).toLocaleString()}
                      </td>
                      <td className="px-4 py-3.5 text-center">
                        <DropdownMenu>
                          <DropdownMenuTrigger className="p-1 rounded-lg hover:bg-zinc-800 text-zinc-400 hover:text-zinc-100 transition-colors">
                            <MoreHorizontal className="w-4 h-4" />
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="w-44 bg-zinc-950 border border-zinc-800 text-zinc-200">
                            <DropdownMenuLabel className="text-[10px] text-zinc-500 uppercase tracking-wider">
                              Product Actions
                            </DropdownMenuLabel>
                            <DropdownMenuItem
                              onClick={() => router.push(`/admin/products/${product.id}`)}
                              className="cursor-pointer flex items-center gap-2"
                            >
                              <Edit className="size-3.5 text-zinc-400" />
                              <span>Edit Product</span>
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => window.open(`/product/${product.id}`, '_blank')}
                              className="cursor-pointer flex items-center gap-2"
                            >
                              <ExternalLink className="size-3.5 text-zinc-400" />
                              <span>Storefront View</span>
                            </DropdownMenuItem>
                            <DropdownMenuSeparator className="bg-zinc-800" />
                            <DropdownMenuItem
                              onClick={() => handleDeleteSingle(product.id, product.title)}
                              className="cursor-pointer text-red-400 hover:text-red-300 focus:text-red-300 flex items-center gap-2"
                            >
                              <Trash2 className="size-3.5 text-red-400" />
                              <span>Delete</span>
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between px-6 py-4 border-t border-white/[0.05]">
              <div className="text-xs text-zinc-500">
                Showing {(currentPage - 1) * itemsPerPage + 1} to {Math.min(currentPage * itemsPerPage, filteredProducts.length)} of {filteredProducts.length} items
              </div>
              <div className="flex items-center gap-1.5">
                <button 
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                  className="flex items-center justify-center size-7 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100 disabled:opacity-40 transition-colors"
                >
                  <ChevronLeft className="w-3.5 h-3.5" />
                </button>
                <div className="flex items-center gap-1">
                  {Array.from({ length: totalPages }).map((_, i) => (
                    <button
                      key={i}
                      onClick={() => setCurrentPage(i + 1)}
                      className={`size-7 rounded-lg text-xs font-medium transition-colors ${
                        currentPage === i + 1 
                          ? 'bg-zinc-100 text-zinc-900 font-semibold' 
                          : 'bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100'
                      }`}
                    >
                      {i + 1}
                    </button>
                  ))}
                </div>
                <button 
                  onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                  disabled={currentPage === totalPages}
                  className="flex items-center justify-center size-7 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100 disabled:opacity-40 transition-colors"
                >
                  <ChevronRight className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          )}

        </div>
      </div>

      <Dialog 
        open={isShopifyModalOpen} 
        onOpenChange={(open) => {
          if (!isMigrating) {
            setIsShopifyModalOpen(open)
            if (!open) {
              setShopifyError(null)
              setShopifySuccess(null)
            }
          }
        }}
      >
        <DialogContent className="w-full sm:max-w-xl p-0 overflow-hidden border border-zinc-800 bg-zinc-950 text-zinc-100 rounded-2xl shadow-2xl">
          <div className="relative border-b border-zinc-800/80 bg-zinc-900/60 p-6">
            <div className="flex items-start gap-4">
              <div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 shadow-inner">
                <Store className="w-6 h-6" />
              </div>
              <div className="space-y-1">
                <DialogTitle className="text-xl font-semibold tracking-tight text-zinc-50">
                  Migrate from Shopify
                </DialogTitle>
                <DialogDescription className="text-sm text-zinc-400">
                  Import your complete catalog, variants, and product images directly into Nexus Commerce via Shopify REST API.
                </DialogDescription>
              </div>
            </div>
          </div>

          <div className="p-6 space-y-5">
            {shopifySuccess ? (
              <div className="space-y-4">
                <div className="rounded-xl border border-emerald-500/30 bg-emerald-950/30 p-4 text-emerald-300">
                  <div className="flex items-start gap-3">
                    <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0 mt-0.5" />
                    <div className="space-y-1">
                      <p className="font-medium text-emerald-200">Migration Completed!</p>
                      <p className="text-sm text-emerald-300/90">{shopifySuccess.message}</p>
                    </div>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => setIsShopifyModalOpen(false)}
                  className="w-full h-10 rounded-xl bg-zinc-100 text-zinc-900 font-semibold text-sm hover:bg-white transition-colors"
                >
                  Done
                </button>
              </div>
            ) : (
              <form onSubmit={handleShopifyMigration} className="space-y-4">
                {shopifyError && (
                  <div className="rounded-xl border border-red-500/30 bg-red-950/30 p-3.5 text-xs text-red-300">
                    {shopifyError}
                  </div>
                )}
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">Shopify Store URL</label>
                  <input
                    type="text"
                    value={shopifyStoreUrl}
                    onChange={(e) => setShopifyStoreUrl(e.target.value)}
                    placeholder="my-store.myshopify.com"
                    className="w-full h-10 rounded-xl bg-zinc-900 border border-zinc-800 px-3.5 text-xs text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">Admin API Access Token</label>
                  <input
                    type="password"
                    value={shopifyAccessToken}
                    onChange={(e) => setShopifyAccessToken(e.target.value)}
                    placeholder="shpat_xxxxxxxxxxxxxxxx"
                    className="w-full h-10 rounded-xl bg-zinc-900 border border-zinc-800 px-3.5 text-xs text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                  />
                </div>
                <div className="pt-2 flex justify-end gap-2.5">
                  <button
                    type="button"
                    onClick={() => setIsShopifyModalOpen(false)}
                    className="px-4 h-9 rounded-xl bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-white"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={isMigrating}
                    className="px-4 h-9 rounded-xl bg-emerald-500 hover:bg-emerald-400 text-zinc-950 text-xs font-semibold flex items-center gap-2 disabled:opacity-50"
                  >
                    {isMigrating && <Loader2 className="size-3.5 animate-spin" />}
                    {isMigrating ? "Importing Catalog..." : "Start Import"}
                  </button>
                </div>
              </form>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
