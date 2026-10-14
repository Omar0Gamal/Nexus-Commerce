"use client"

import * as React from "react"
import { Search, Filter, MoreHorizontal, Plus, Trash2, Power } from "lucide-react"
import { useGetApiV1Coupons, usePatchApiV1CouponsIdToggle, useDeleteApiV1CouponsId } from "@/api/endpoints/coupons/coupons";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

export default function DiscountsPage() {
  const { data: couponsResponse, isLoading } = useGetApiV1Coupons()
  const toggleMutation = usePatchApiV1CouponsIdToggle()
  const deleteMutation = useDeleteApiV1CouponsId()
  const queryClient = useQueryClient()

  const [searchQuery, setSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  let allDiscounts: any[] = [];
  if (couponsResponse?.data && Array.isArray(couponsResponse.data)) {
    allDiscounts = couponsResponse.data.map((c: any) => {
      const isFreeShip = c.code?.toUpperCase().includes("FREESHIP");
      const typeLabel = isFreeShip ? "Free Shipping" : c.type === "percentage" ? "Percentage" : "Fixed Amount";
      const discountNumericValue = parseFloat(c.value || c.discount_amount || c.discount_percentage || "0");
      let displayValue = "-";
      if (isFreeShip) {
        displayValue = "Free shipping";
      } else if (c.type === "percentage") {
        displayValue = `${discountNumericValue}% off`;
      } else {
        displayValue = `EGP ${discountNumericValue.toLocaleString()} off`;
      }
      return {
        id: c.id,
        code: c.code,
        type: typeLabel,
        value: displayValue,
        status: c.is_active ? "Active" : "Expired",
        usage: `${c.usage_count || 0} / ${c.usage_limit ?? '∞'}`,
        expires: c.expires_at ? new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', year: 'numeric' }).format(new Date(c.expires_at)) : "No expiry",
      };
    });
  }

  const filteredDiscounts = allDiscounts.filter(d => 
    d.code.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const totalPages = Math.ceil(filteredDiscounts.length / itemsPerPage);
  const paginatedDiscounts = filteredDiscounts.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );

  const handleToggle = async (id: string) => {
    try {
      await toggleMutation.mutateAsync({ id, data: {} });
      queryClient.invalidateQueries({ queryKey: ['coupons'] });
    } catch (e) {
      console.error(e);
      alert("Failed to toggle coupon status");
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this coupon?")) return;
    try {
      await deleteMutation.mutateAsync({ id });
      queryClient.invalidateQueries({ queryKey: ['coupons'] });
    } catch (e) {
      console.error(e);
      alert("Failed to delete coupon");
    }
  }

  return (
    <div className="flex flex-col gap-8 w-full max-w-6xl animate-in fade-in slide-in-from-bottom-4 duration-700">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-3xl font-semibold tracking-tight text-zinc-50">Discounts</h1>
          <p className="text-zinc-400">Create and manage promotional codes and automatic discounts.</p>
        </div>
        <button className="group flex h-10 items-center justify-center gap-2 rounded-full bg-zinc-100 px-6 text-sm font-semibold text-zinc-900 transition-all hover:bg-white active:scale-95 shadow-sm">
          <Plus className="w-4 h-4" />
          Create Discount
        </button>
      </div>

      <div className="flex flex-col gap-4 p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
        <div className="rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] overflow-hidden">
          
          <div className="flex items-center justify-between p-4 border-b border-white/[0.05]">
            <div className="flex items-center gap-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                <input 
                  type="text" 
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    setCurrentPage(1);
                  }}
                  placeholder="Search discounts..." 
                  className="h-9 w-64 rounded-full bg-zinc-950 border border-zinc-800 pl-9 pr-4 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all placeholder:text-zinc-600"
                />
              </div>
              <button className="flex items-center justify-center w-9 h-9 rounded-full bg-zinc-950 border border-zinc-800 text-zinc-400 hover:text-zinc-100 transition-colors">
                <Filter className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm whitespace-nowrap">
              <thead className="bg-zinc-950/50 text-zinc-500">
                <tr>
                  <th className="px-6 py-4 font-medium">Discount Code</th>
                  <th className="px-6 py-4 font-medium">Type</th>
                  <th className="px-6 py-4 font-medium">Value</th>
                  <th className="px-6 py-4 font-medium">Status</th>
                  <th className="px-6 py-4 font-medium">Usage Limit</th>
                  <th className="px-6 py-4 font-medium text-right">Expires</th>
                  <th className="px-6 py-4 font-medium text-center w-24">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/[0.05]">
                {isLoading ? (
                  [1,2,3].map(i => (
                    <tr key={i}>
                      <td colSpan={7} className="px-6 py-4">
                        <div className="h-12 bg-zinc-800/50 animate-pulse rounded-xl" />
                      </td>
                    </tr>
                  ))
                ) : paginatedDiscounts.length === 0 ? (
                  <tr><td colSpan={7} className="px-6 py-12 text-center text-zinc-500">No data found</td></tr>
                ) : (
                  paginatedDiscounts.map((discount: any) => (
                    <tr key={discount.id} className="hover:bg-white/[0.02] transition-colors group">
                      <td className="px-6 py-4">
                        <span className="inline-flex items-center rounded-xl bg-zinc-800 px-2 py-1 text-xs font-mono font-medium text-zinc-300">
                          {discount.code}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-zinc-300">{discount.type}</td>
                      <td className="px-6 py-4 font-medium text-emerald-400 font-mono">{discount.value}</td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium border ${
                          discount.status === 'Active' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 
                          discount.status === 'Scheduled' ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' :
                          'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'
                        }`}>
                          {discount.status}
                        </span>
                      </td>
                      <td className="px-6 py-4 font-mono text-zinc-300">{discount.usage}</td>
                      <td className="px-6 py-4 text-right text-zinc-400">{discount.expires}</td>
                      <td className="px-6 py-4">
                        <div className="flex items-center justify-center gap-2">
                          <button 
                            onClick={() => handleToggle(discount.id)}
                            className="p-2 rounded-full hover:bg-zinc-800 text-zinc-400 hover:text-zinc-100 transition-colors"
                            title="Toggle Status"
                          >
                            <Power className="w-4 h-4" />
                          </button>
                          <button 
                            onClick={() => handleDelete(discount.id)}
                            className="p-2 rounded-full hover:bg-red-950/50 text-zinc-400 hover:text-red-400 transition-colors"
                            title="Delete Coupon"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between px-6 py-4 border-t border-white/[0.05]">
              <div className="text-sm text-zinc-500">
                Showing {(currentPage - 1) * itemsPerPage + 1} to {Math.min(currentPage * itemsPerPage, filteredDiscounts.length)} of {filteredDiscounts.length} results
              </div>
              <div className="flex items-center gap-2">
                <button 
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                  className="px-3 py-1 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100 disabled:opacity-50 text-sm"
                >
                  Previous
                </button>
                <button 
                  onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                  disabled={currentPage === totalPages}
                  className="px-3 py-1 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100 disabled:opacity-50 text-sm"
                >
                  Next
                </button>
              </div>
            </div>
          )}

        </div>
      </div>
    </div>
  )
}