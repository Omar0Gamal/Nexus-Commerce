"use client"

import * as React from "react"
import { Search, Filter, MoreHorizontal, Download } from "lucide-react"
import { useGetApiV1Customers } from "@/api/endpoints/customers/customers";


export default function CustomersPage() {
  const { data: customersResponse, isLoading } = useGetApiV1Customers()
  
  const formatLastActive = (dateStr?: string) => {
    if (!dateStr || dateStr === "Never") return "Never";
    if (dateStr.includes("-")) {
      const parts = dateStr.split("T")[0].split("-");
      if (parts.length === 3) {
        const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
        const mIdx = parseInt(parts[1], 10) - 1;
        const day = parseInt(parts[2], 10);
        if (mIdx >= 0 && mIdx < 12 && !isNaN(day)) {
          return `${months[mIdx]} ${day}`;
        }
      }
    }
    return dateStr;
  };

  let customersList: any[] = [];
  if (customersResponse?.data && (customersResponse.data as unknown as any[]).length > 0) {
    customersList = (customersResponse.data as unknown as any[]).map((c: any) => ({
      id: c.id,
      name: `${c.first_name} ${c.last_name || ''}`.trim(),
      email: c.email,
      orders: c.total_orders || 0,
      spent: new Intl.NumberFormat("en-EG", { style: "currency", currency: "EGP", maximumFractionDigits: 0 }).format(Number(c.total_spent || 0)),
      lastActive: formatLastActive(c.last_order_date),
    }))
  }

  return (
    <div className="flex flex-col gap-8 w-full max-w-6xl">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-3xl font-semibold tracking-tight text-zinc-50">Customers</h1>
          <p className="text-zinc-400">Manage your customer base and view purchase history.</p>
        </div>
        <button className="group flex h-10 items-center justify-center rounded-full bg-zinc-100 px-6 text-sm font-semibold text-zinc-900 transition-all hover:bg-white active:scale-95">
          Add Customer
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
                  placeholder="Search customers..." 
                  className="h-9 w-64 rounded-full bg-zinc-950 border border-zinc-800 pl-9 pr-4 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all placeholder:text-zinc-600"
                />
              </div>
              <button className="flex items-center justify-center w-9 h-9 rounded-full bg-zinc-950 border border-zinc-800 text-zinc-400 hover:text-zinc-100 transition-colors">
                <Filter className="w-4 h-4" />
              </button>
            </div>
            <button className="flex items-center gap-2 px-4 h-9 rounded-full bg-zinc-950 border border-zinc-800 text-zinc-400 hover:text-zinc-100 text-sm font-medium transition-colors">
              <Download className="w-4 h-4" />
              Export
            </button>
          </div>

          <table className="w-full text-left text-sm whitespace-nowrap">
            <thead className="bg-zinc-950/50 text-zinc-500">
              <tr>
                <th className="px-6 py-3 font-medium">Customer</th>
                <th className="px-6 py-3 font-medium">Orders</th>
                <th className="px-6 py-3 font-medium">Amount Spent</th>
                <th className="px-6 py-3 font-medium text-right">Last Active</th>
                <th className="px-6 py-3 font-medium text-center w-12"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.05]">
              {isLoading ? (
                [1,2,3].map(i => (
                  <tr key={i}>
                    <td colSpan={5} className="px-6 py-4">
                      <div className="h-12 bg-zinc-800/50 animate-pulse rounded-xl" />
                    </td>
                  </tr>
                ))
              ) : customersList.length === 0 ? (
                <tr><td colSpan={5} className="px-6 py-12 text-center text-zinc-500">No data found</td></tr>
              ) : (
                customersList.map((customer: any) => (
                  <tr key={customer.id} className="hover:bg-white/[0.02] transition-colors group">
                    <td className="px-6 py-4">
                      <div className="flex flex-col">
                        <span className="font-medium text-zinc-100">{customer.name}</span>
                        <span className="text-zinc-500 text-xs">{customer.email}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-zinc-300">{customer.orders}</td>
                    <td className="px-6 py-4">
                      <span className="inline-flex items-center rounded-full bg-emerald-500/10 px-2 py-1 text-xs font-medium text-emerald-400 border border-emerald-500/20">
                        {customer.spent}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right text-zinc-500">{customer.lastActive}</td>
                    <td className="px-6 py-4 text-center">
                      <button className="text-zinc-600 hover:text-zinc-100 transition-colors">
                        <MoreHorizontal className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}