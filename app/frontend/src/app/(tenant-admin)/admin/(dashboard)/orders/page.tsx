import { OrdersTable } from "@/features/orders/components/OrdersTable"

export default function AdminOrdersPage() {
  return (
    <div className="min-h-screen bg-zinc-950 p-8 lg:p-12">
      <div className="max-w-7xl mx-auto flex flex-col gap-12 animate-in fade-in slide-in-from-bottom-8 duration-700 ease-[cubic-bezier(0.32,0.72,0,1)]">
        <div className="flex flex-col gap-3">
          <span className="rounded-full px-3 py-1 text-[10px] uppercase tracking-[0.2em] font-medium text-zinc-400 border border-zinc-800 w-max">
            Fulfillment
          </span>
          <h1 className="text-4xl font-semibold tracking-tight text-zinc-50">Orders</h1>
          <p className="text-zinc-400 text-lg">
            Manage and fulfill your customer orders.
          </p>
        </div>
        
        <OrdersTable />
      </div>
    </div>
  )
}
