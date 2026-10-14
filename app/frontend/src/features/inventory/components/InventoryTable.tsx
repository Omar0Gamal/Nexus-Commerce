'use client';

import * as React from 'react';
import {
  Boxes,
  Search,
  SlidersHorizontal,
  History,
  Plus,
  Minus,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Warehouse as WarehouseIcon,
  RefreshCw,
  ArrowUpDown,
  ExternalLink,
} from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Warehouse, InventoryItemRow } from '../types';

interface InventoryTableProps {
  items: InventoryItemRow[];
  warehouses: Warehouse[];
  activeWarehouseId: string;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  statusFilter: 'all' | 'in_stock' | 'low_stock' | 'out_of_stock';
  onStatusFilterChange: (status: 'all' | 'in_stock' | 'low_stock' | 'out_of_stock') => void;
  onOpenAdjust: (item: InventoryItemRow) => void;
  onOpenHistory: (item: InventoryItemRow) => void;
  isLoading: boolean;
  onRefresh: () => void;
}

export function InventoryTable({
  items,
  warehouses,
  activeWarehouseId,
  searchQuery,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  onOpenAdjust,
  onOpenHistory,
  isLoading,
  onRefresh,
}: InventoryTableProps) {
  const filteredItems = React.useMemo(() => {
    return items.filter((item) => {
      const matchesSearch =
        searchQuery === '' ||
        item.productTitle.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.variantTitle.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.sku.toLowerCase().includes(searchQuery.toLowerCase());

      if (!matchesSearch) return false;

      if (statusFilter === 'in_stock') {
        return !item.isLowStock && !item.isOutOfStock;
      }
      if (statusFilter === 'low_stock') {
        return item.isLowStock;
      }
      if (statusFilter === 'out_of_stock') {
        return item.isOutOfStock;
      }

      return true;
    });
  }, [items, searchQuery, statusFilter]);

  const selectedWarehouse =
    activeWarehouseId === 'all' ? null : warehouses.find((w) => w.id === activeWarehouseId);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
          <Input
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Search variant name, product, or SKU..."
            className="pl-9 pr-4 h-10 bg-zinc-900/80 border-zinc-800 text-sm text-zinc-100 rounded-full focus-visible:ring-zinc-600 transition-all placeholder:text-zinc-500"
          />
          {searchQuery && (
            <button
              onClick={() => onSearchChange('')}
              className="absolute right-3.5 top-1/2 -translate-y-1/2 text-xs text-zinc-500 hover:text-zinc-300"
            >
              Clear
            </button>
          )}
        </div>

        <div className="flex items-center gap-2">
          <select
            value={statusFilter}
            onChange={(e) => onStatusFilterChange(e.target.value as any)}
            className="h-10 px-3.5 rounded-full bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 focus:outline-none focus:ring-1 focus:ring-zinc-600 cursor-pointer"
          >
            <option value="all">All Inventory States</option>
            <option value="in_stock">Healthy (In Stock)</option>
            <option value="low_stock">Low Stock Warnings</option>
            <option value="out_of_stock">Out of Stock (Zero)</option>
          </select>

          <button
            onClick={onRefresh}
            title="Refresh inventory levels"
            disabled={isLoading}
            className="h-10 w-10 flex items-center justify-center rounded-full bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-all active:scale-95"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin text-zinc-200' : ''}`} />
          </button>
        </div>
      </div>

      <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
        <div className="rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-zinc-300">
              <thead className="bg-zinc-950/40 text-[11px] uppercase tracking-wider text-zinc-400 border-b border-white/[0.05]">
                <tr>
                  <th className="py-3.5 px-4 font-semibold">Product & Variant</th>
                  <th className="py-3.5 px-3 font-semibold">SKU / Code</th>
                  <th className="py-3.5 px-3 font-semibold">Stock Health</th>
                  {activeWarehouseId === 'all' ? (
                    <>
                      <th className="py-3.5 px-3 font-semibold">Distribution Breakdown</th>
                      <th className="py-3.5 px-3 font-semibold text-right">Total Available</th>
                      <th className="py-3.5 px-3 font-semibold text-right">Total On Hand</th>
                    </>
                  ) : (
                    <>
                      <th className="py-3.5 px-3 font-semibold text-right">On Hand</th>
                      <th className="py-3.5 px-3 font-semibold text-right">Reserved</th>
                      <th className="py-3.5 px-3 font-semibold text-right">Available</th>
                    </>
                  )}
                  <th className="py-3.5 px-4 font-semibold text-right">Actions</th>
                </tr>
              </thead>

              <tbody className="divide-y divide-white/[0.03]">
                {isLoading ? (
                  Array.from({ length: 5 }).map((_, idx) => (
                    <tr key={idx} className="animate-pulse">
                      <td className="py-4 px-4">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 rounded-xl bg-zinc-800" />
                          <div className="flex flex-col gap-1.5">
                            <div className="h-4 w-32 bg-zinc-800 rounded" />
                            <div className="h-3 w-20 bg-zinc-800/60 rounded" />
                          </div>
                        </div>
                      </td>
                      <td className="py-4 px-3">
                        <div className="h-3 w-16 bg-zinc-800 rounded" />
                      </td>
                      <td className="py-4 px-3">
                        <div className="h-5 w-20 bg-zinc-800 rounded-full" />
                      </td>
                      <td className="py-4 px-3">
                        <div className="h-4 w-28 bg-zinc-800 rounded" />
                      </td>
                      <td className="py-4 px-3 text-right">
                        <div className="h-4 w-12 bg-zinc-800 rounded ml-auto" />
                      </td>
                      <td className="py-4 px-3 text-right">
                        <div className="h-4 w-12 bg-zinc-800 rounded ml-auto" />
                      </td>
                      <td className="py-4 px-4 text-right">
                        <div className="h-8 w-20 bg-zinc-800 rounded-full ml-auto" />
                      </td>
                    </tr>
                  ))
                ) : filteredItems.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="py-16 text-center">
                      <div className="flex flex-col items-center justify-center gap-3 max-w-sm mx-auto">
                        <div className="w-12 h-12 rounded-2xl bg-zinc-950 border border-zinc-800 flex items-center justify-center text-zinc-500">
                          <Boxes className="w-6 h-6" />
                        </div>
                        <div className="flex flex-col gap-1">
                          <h4 className="text-sm font-semibold text-zinc-200">No inventory matches found</h4>
                          <p className="text-xs text-zinc-500">
                            {searchQuery || statusFilter !== 'all'
                              ? 'Try adjusting your search criteria or clearing active filters.'
                              : 'No products or variants have been tracked in inventory yet.'}
                          </p>
                        </div>
                        {(searchQuery || statusFilter !== 'all') && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              onSearchChange('');
                              onStatusFilterChange('all');
                            }}
                            className="rounded-full text-xs border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-800"
                          >
                            Reset Filters
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ) : (
                  filteredItems.map((item) => {
                    const singleOnHand =
                      activeWarehouseId === 'all' ? item.totalOnHand : item.levels[activeWarehouseId] ?? 0;
                    const singleReserved =
                      activeWarehouseId === 'all'
                        ? item.totalReserved
                        : item.reservedLevels[activeWarehouseId] ?? 0;
                    const singleAvailable = Math.max(0, singleOnHand - singleReserved);

                    return (
                      <tr
                        key={item.variantId}
                        className="group hover:bg-white/[0.02] transition-colors"
                      >
                        <td className="py-3.5 px-4">
                          <div className="flex items-center gap-3">
                            <div className="w-10 h-10 rounded-xl bg-zinc-950 border border-zinc-800/80 flex items-center justify-center shrink-0 overflow-hidden">
                              {item.imageUrl ? (
                                <img
                                  src={item.imageUrl}
                                  alt={item.productTitle}
                                  className="w-full h-full object-cover"
                                />
                              ) : (
                                <Boxes className="w-5 h-5 text-zinc-500" />
                              )}
                            </div>
                            <div className="flex flex-col min-w-0">
                              <span className="font-medium text-zinc-100 text-sm line-clamp-1 group-hover:text-white transition-colors">
                                {item.productTitle}
                              </span>
                              <div className="flex items-center gap-1.5 text-xs text-zinc-400">
                                <span className="font-mono text-zinc-300">{item.variantTitle}</span>
                                {item.price && (
                                  <>
                                    <span>•</span>
                                    <span className="text-zinc-400">{item.price}</span>
                                  </>
                                )}
                              </div>
                            </div>
                          </div>
                        </td>

                        <td className="py-3.5 px-3">
                          <span className="font-mono text-xs text-zinc-400">
                            {item.sku || '—'}
                          </span>
                        </td>

                        <td className="py-3.5 px-3">
                          {item.isOutOfStock ? (
                            <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-rose-500/15 text-rose-400 border border-rose-500/20">
                              <span className="w-1.5 h-1.5 rounded-full bg-rose-400 animate-pulse" />
                              Out of Stock
                            </span>
                          ) : item.isLowStock ? (
                            <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-amber-500/15 text-amber-400 border border-amber-500/20">
                              <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
                              Low Stock ({singleOnHand} left)
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/20">
                              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                              In Stock
                            </span>
                          )}
                        </td>

                        {activeWarehouseId === 'all' ? (
                          <>
                            <td className="py-3.5 px-3">
                              <div className="flex items-center flex-wrap gap-1.5 max-w-xs">
                                {warehouses.map((wh) => {
                                  const count = item.levels[wh.id] ?? 0;
                                  return (
                                    <span
                                      key={wh.id}
                                      title={`${wh.name}: ${count} on hand`}
                                      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-mono border ${
                                        count === 0
                                          ? 'bg-zinc-950/60 text-zinc-600 border-zinc-800/40'
                                          : count <= item.lowStockThreshold
                                          ? 'bg-amber-500/10 text-amber-300 border-amber-500/20'
                                          : 'bg-zinc-900 text-zinc-300 border-zinc-800'
                                      }`}
                                    >
                                      <span className="text-[10px] text-zinc-500 truncate max-w-[65px]">
                                        {wh.name.split(' ')[0]}
                                      </span>
                                      <span className="font-semibold">{count}</span>
                                    </span>
                                  );
                                })}
                              </div>
                            </td>

                            <td className="py-3.5 px-3 text-right">
                              <span className="font-mono text-sm font-semibold tabular-nums text-zinc-100">
                                {item.totalAvailable}
                              </span>
                            </td>

                            <td className="py-3.5 px-3 text-right">
                              <div className="flex flex-col items-end">
                                <span className="font-mono text-sm font-semibold tabular-nums text-zinc-200">
                                  {item.totalOnHand}
                                </span>
                                {item.totalReserved > 0 && (
                                  <span className="text-[10px] text-zinc-500 font-mono">
                                    {item.totalReserved} reserved
                                  </span>
                                )}
                              </div>
                            </td>
                          </>
                        ) : (
                          <>
                            <td className="py-3.5 px-3 text-right">
                              <span className="font-mono text-sm font-semibold tabular-nums text-zinc-100">
                                {singleOnHand}
                              </span>
                            </td>

                            <td className="py-3.5 px-3 text-right">
                              <span className="font-mono text-xs tabular-nums text-zinc-400">
                                {singleReserved}
                              </span>
                            </td>

                            <td className="py-3.5 px-3 text-right">
                              <span className="font-mono text-sm font-semibold tabular-nums text-emerald-400">
                                {singleAvailable}
                              </span>
                            </td>
                          </>
                        )}

                        <td className="py-3.5 px-4 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              onClick={() => onOpenAdjust(item)}
                              className="group/btn flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-zinc-100 hover:bg-white text-zinc-950 text-xs font-semibold shadow-sm active:scale-95 transition-all"
                            >
                              <span>Adjust</span>
                              <span className="w-4 h-4 rounded-full bg-black/10 flex items-center justify-center text-[10px] group-hover/btn:scale-110 transition-transform">
                                ±
                              </span>
                            </button>

                            <button
                              onClick={() => onOpenHistory(item)}
                              title="Inspect stock movement history"
                              className="p-1.5 rounded-full bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
                            >
                              <History className="w-3.5 h-3.5" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
