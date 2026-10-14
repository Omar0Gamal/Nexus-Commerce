"use client"

import * as React from "react"
import { CheckCircle2, CreditCard, Download, ExternalLink, Zap, Star, Shield, ArrowUpRight, MoreHorizontal } from "lucide-react"

import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table } from "@/components/ui/table"
import { useGetApiV1BillingInvoices } from "@/api/endpoints/billing/billing";

export default function BillingPage() {
  const { data: apiInvoices, isError } = useGetApiV1BillingInvoices();
  const invoicesToUse = (apiInvoices?.data && !isError) ? apiInvoices.data.map((inv: any) => ({
    id: inv.id,
    date: new Date(inv.created_at * 1000).toLocaleDateString(),
    amount: `$${(inv.amount_due / 100).toFixed(2)}`,
    status: inv.status,
    pdf: inv.hosted_invoice_url || "#"
  })) : [];

  const handlePortalRedirect = () => {
    
  }

  return (
    <div className="flex-1 space-y-8 p-8 md:p-10 max-w-6xl mx-auto w-full">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-4 sm:space-y-0">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-zinc-950 dark:text-zinc-50">
            Billing & Subscriptions
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-400 mt-2">
            Manage your subscription plan, payment methods, and billing history.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline" className="gap-2" onClick={handlePortalRedirect}>
            <ExternalLink className="h-4 w-4" />
            Billing Portal
          </Button>
        </div>
      </div>

      <Tabs defaultValue="overview" className="space-y-6">
        <TabsList className="bg-zinc-100/50 dark:bg-zinc-900/50 border border-zinc-200 dark:border-zinc-800">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="invoices">Invoices</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-8 outline-none">
          <section>
            <Card className="p-6 md:p-8 bg-white dark:bg-zinc-950 border-zinc-200 dark:border-zinc-800 rounded-2xl shadow-sm relative overflow-hidden">
              <div className="absolute top-0 right-0 -mt-4 -mr-4 w-32 h-32 bg-indigo-500/10 blur-3xl rounded-full pointer-events-none" />
              
              <div className="flex flex-col md:flex-row justify-between md:items-start gap-6 relative z-10">
                <div className="space-y-4">
                  <div className="flex items-center gap-3">
                    <h2 className="text-xl font-medium text-zinc-900 dark:text-zinc-100">Professional Plan</h2>
                    <Badge variant="default" className="bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 dark:hover:bg-indigo-500/20 border-indigo-200 dark:border-indigo-500/20">Current</Badge>
                  </div>
                  <p className="text-sm text-zinc-600 dark:text-zinc-400 max-w-md leading-relaxed">
                    You are currently on the Professional Plan. Your next billing date is <span className="text-zinc-900 dark:text-zinc-100 font-medium">September 1, 2026</span> for <span className="text-zinc-900 dark:text-zinc-100 font-medium">$99.00</span>.
                  </p>
                </div>
                <div className="flex flex-col sm:flex-row items-center gap-3">
                  <Button variant="outline" className="w-full sm:w-auto">Cancel Subscription</Button>
                  <Button className="w-full sm:w-auto bg-zinc-900 dark:bg-zinc-100 text-zinc-50 dark:text-zinc-900">Upgrade Plan</Button>
                </div>
              </div>
              
              <Separator className="my-8" />
              
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
                <div>
                  <p className="text-sm font-medium text-zinc-500 dark:text-zinc-400 mb-1">Products Limit</p>
                  <p className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">254 <span className="text-lg text-zinc-400 font-normal">/ 1,000</span></p>
                </div>
                <div>
                  <p className="text-sm font-medium text-zinc-500 dark:text-zinc-400 mb-1">Monthly Orders</p>
                  <p className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">1,248 <span className="text-lg text-zinc-400 font-normal">/ 5,000</span></p>
                </div>
                <div>
                  <p className="text-sm font-medium text-zinc-500 dark:text-zinc-400 mb-1">Staff Accounts</p>
                  <p className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">3 <span className="text-lg text-zinc-400 font-normal">/ 5</span></p>
                </div>
              </div>
            </Card>
          </section>

          <section className="space-y-4">
            <h3 className="text-lg font-medium text-zinc-900 dark:text-zinc-100">Payment Method</h3>
            <Card className="p-0 overflow-hidden border-zinc-200 dark:border-zinc-800 rounded-xl shadow-sm">
              <div className="p-6 flex flex-col sm:flex-row sm:items-center justify-between gap-6 bg-zinc-50/50 dark:bg-zinc-900/20">
                <div className="flex items-start gap-4">
                  <div className="bg-white dark:bg-zinc-900 p-2 rounded-xl border border-zinc-200 dark:border-zinc-800 flex items-center justify-center shrink-0 h-12 w-16">
                    <svg className="w-8 h-8 text-zinc-700 dark:text-zinc-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <rect x="2" y="5" width="20" height="14" rx="2" />
                      <line x1="2" y1="10" x2="22" y2="10" />
                    </svg>
                  </div>
                  <div>
                    <p className="font-medium text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
                      Visa ending in 4242
                      <Badge variant="outline" className="text-[10px] uppercase tracking-wider py-0 px-1.5 font-semibold text-zinc-500 dark:text-zinc-400">Default</Badge>
                    </p>
                    <p className="text-sm text-zinc-500 dark:text-zinc-400 mt-1">Expires 12/2028</p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <Button variant="ghost" size="sm">Edit</Button>
                  <Button variant="outline" size="sm">Add new card</Button>
                </div>
              </div>
            </Card>
          </section>

          <section className="space-y-4 pt-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-medium text-zinc-900 dark:text-zinc-100">Available Plans</h3>
                <p className="text-sm text-zinc-500 dark:text-zinc-400">Find a plan that scales with your business.</p>
              </div>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <Card className="p-6 border-zinc-200 dark:border-zinc-800 rounded-xl hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors bg-white dark:bg-zinc-950/50 flex flex-col">
                <div className="mb-4">
                  <h4 className="font-medium text-zinc-900 dark:text-zinc-100">Basic</h4>
                  <div className="mt-2 flex items-baseline gap-1">
                    <span className="text-3xl font-semibold text-zinc-900 dark:text-zinc-50">$29</span>
                    <span className="text-sm text-zinc-500 dark:text-zinc-400">/mo</span>
                  </div>
                </div>
                <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-6">Perfect for new merchants just getting started.</p>
                <ul className="space-y-3 mb-8 flex-1">
                  {["Up to 100 products", "500 monthly orders", "Basic analytics", "Email support"].map(feature => (
                    <li key={feature} className="flex items-center gap-3 text-sm text-zinc-600 dark:text-zinc-400">
                      <CheckCircle2 className="h-4 w-4 text-zinc-400 dark:text-zinc-500 shrink-0" />
                      {feature}
                    </li>
                  ))}
                </ul>
                <Button variant="outline" className="w-full">Downgrade</Button>
              </Card>

              <Card className="p-6 border-indigo-500/30 dark:border-indigo-500/30 rounded-xl bg-indigo-50/30 dark:bg-indigo-950/20 ring-1 ring-indigo-500/10 flex flex-col relative overflow-hidden">
                <div className="absolute top-0 right-0 p-3">
                  <Badge className="bg-indigo-500 hover:bg-indigo-600 text-zinc-50 border-transparent">Current</Badge>
                </div>
                <div className="mb-4">
                  <h4 className="font-medium text-zinc-900 dark:text-zinc-100 flex items-center gap-2">Professional <Zap className="h-4 w-4 text-indigo-500" /></h4>
                  <div className="mt-2 flex items-baseline gap-1">
                    <span className="text-3xl font-semibold text-zinc-900 dark:text-zinc-50">$99</span>
                    <span className="text-sm text-zinc-500 dark:text-zinc-400">/mo</span>
                  </div>
                </div>
                <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-6">For growing businesses that need more power.</p>
                <ul className="space-y-3 mb-8 flex-1">
                  {["Up to 1,000 products", "5,000 monthly orders", "Advanced analytics", "Priority support", "Custom domain"].map(feature => (
                    <li key={feature} className="flex items-center gap-3 text-sm text-zinc-600 dark:text-zinc-300">
                      <CheckCircle2 className="h-4 w-4 text-indigo-500 shrink-0" />
                      {feature}
                    </li>
                  ))}
                </ul>
                <Button className="w-full bg-indigo-500 hover:bg-indigo-600 text-zinc-50" disabled>Current Plan</Button>
              </Card>

              <Card className="p-6 border-zinc-200 dark:border-zinc-800 rounded-xl hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors bg-zinc-950 dark:bg-zinc-900 flex flex-col group relative overflow-hidden">
                <div className="absolute inset-0 bg-gradient-to-br from-zinc-800/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
                <div className="mb-4 relative z-10">
                  <h4 className="font-medium text-zinc-100 dark:text-zinc-100 flex items-center gap-2">Premium <Shield className="h-4 w-4 text-zinc-400" /></h4>
                  <div className="mt-2 flex items-baseline gap-1">
                    <span className="text-3xl font-semibold text-zinc-50 dark:text-zinc-50">$299</span>
                    <span className="text-sm text-zinc-400 dark:text-zinc-400">/mo</span>
                  </div>
                </div>
                <p className="text-sm text-zinc-400 dark:text-zinc-400 mb-6 relative z-10">Unlimited scale and enterprise-grade features.</p>
                <ul className="space-y-3 mb-8 flex-1 relative z-10">
                  {["Unlimited products", "Unlimited orders", "Custom reporting", "24/7 dedicated support", "API access", "White-labeling"].map(feature => (
                    <li key={feature} className="flex items-center gap-3 text-sm text-zinc-300 dark:text-zinc-300">
                      <CheckCircle2 className="h-4 w-4 text-zinc-500 shrink-0" />
                      {feature}
                    </li>
                  ))}
                </ul>
                <Button className="w-full bg-white text-zinc-950 hover:bg-zinc-200 relative z-10">Upgrade to Premium</Button>
              </Card>
            </div>
          </section>
        </TabsContent>

        <TabsContent value="invoices" className="outline-none">
          <Card className="border-zinc-200 dark:border-zinc-800 shadow-sm rounded-xl overflow-hidden">
            <div className="p-6 border-b border-zinc-200 dark:border-zinc-800 flex items-center justify-between bg-white dark:bg-zinc-950">
              <div>
                <h3 className="text-lg font-medium text-zinc-900 dark:text-zinc-100">Billing History</h3>
                <p className="text-sm text-zinc-500 dark:text-zinc-400 mt-1">View and download your past invoices.</p>
              </div>
              <Button variant="outline" size="sm" className="gap-2 hidden sm:flex">
                <Download className="h-4 w-4" />
                Download All
              </Button>
            </div>
            
            <div className="bg-white dark:bg-zinc-950">
              <table className="w-full text-sm text-left">
                <thead className="bg-zinc-50 dark:bg-zinc-900/50 border-b border-zinc-200 dark:border-zinc-800 text-zinc-500 dark:text-zinc-400">
                  <tr>
                    <th className="px-6 py-4 font-medium">Invoice</th>
                    <th className="px-6 py-4 font-medium">Date</th>
                    <th className="px-6 py-4 font-medium">Amount</th>
                    <th className="px-6 py-4 font-medium">Status</th>
                    <th className="px-6 py-4 font-medium text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-200 dark:divide-zinc-800">
                  {invoicesToUse.map((inv: any) => (
                    <tr key={inv.id} className="hover:bg-zinc-50/50 dark:hover:bg-zinc-900/20 transition-colors">
                      <td className="px-6 py-4 font-medium text-zinc-900 dark:text-zinc-100">{inv.id}</td>
                      <td className="px-6 py-4 text-zinc-600 dark:text-zinc-400">{inv.date}</td>
                      <td className="px-6 py-4 text-zinc-600 dark:text-zinc-400">{inv.amount}</td>
                      <td className="px-6 py-4">
                        <Badge variant="outline" className="bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20 font-medium">
                          {inv.status}
                        </Badge>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <Button variant="ghost" size="sm" className="gap-2 text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100">
                          <Download className="h-4 w-4" />
                          <span className="hidden sm:inline">PDF</span>
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
