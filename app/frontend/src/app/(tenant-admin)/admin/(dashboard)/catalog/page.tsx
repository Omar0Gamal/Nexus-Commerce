import { ProductsTable } from "@/features/catalog/components/ProductsTable"

export default function AdminCatalogPage() {
  return (
    <div className="flex flex-col gap-12">
      <div className="flex flex-col sm:flex-row gap-4 sm:items-center justify-between">
        <div className="flex flex-col gap-3">
          <span className="rounded-full px-3 py-1 text-[10px] uppercase tracking-[0.2em] font-medium text-zinc-400 border border-zinc-800 w-max">
            Inventory
          </span>
          <h1 className="text-4xl font-semibold tracking-tight text-zinc-50">Catalog</h1>
          <p className="text-zinc-400 text-lg">
            Manage your store's catalog, inventory, and collections.
          </p>
        </div>
      </div>
      
      <ProductsTable />
    </div>
  )
}
