'use client';

import * as React from 'react';
import { Boxes, Warehouse as WarehouseIcon, AlertTriangle, ShieldCheck, ArrowUpRight, CheckCircle2 } from 'lucide-react';
import { Warehouse, InventoryItemRow } from '../types';

interface WarehouseMetricsProps {
  items: InventoryItemRow[];
  warehouses: Warehouse[];
  activeWarehouseId: string;
  onFilterLowStock: () => void;
  isLowStockFiltered: boolean;
}

export function WarehouseMetrics({
  items,
  warehouses,
  activeWarehouseId,
  onFilterLowStock,
  isLowStockFiltered,
}: WarehouseMetricsProps) {
  const totalOnHand = items.reduce((sum, item) => {
    if (activeWarehouseId === 'all') {
      return sum + item.totalOnHand;
    }
    return sum + (item.levels[activeWarehouseId] || 0);
  }, 0);

  const totalReserved = items.reduce((sum, item) => {
    if (activeWarehouseId === 'all') {
      return sum + item.totalReserved;
    }
    return sum + (item.reservedLevels[activeWarehouseId] || 0);
  }, 0);

  const totalAvailable = Math.max(0, totalOnHand - totalReserved);

  const lowStockCount = items.filter((item) => item.isLowStock || item.isOutOfStock).length;
  const outOfStockCount = items.filter((item) => item.isOutOfStock).length;

  const activeWarehouseName =
    activeWarehouseId === 'all'
      ? 'All Locations'
      : warehouses.find((w) => w.id === activeWarehouseId)?.name || 'Selected Location';

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02] transition-colors hover:bg-white/[0.04]">
        <div className="relative rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] p-5 flex flex-col justify-between min-h-[140px] group overflow-hidden">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-medium tracking-wide uppercase text-zinc-400">Total On Hand</span>
            <div className="flex size-9 items-center justify-center rounded-xl bg-zinc-950 border border-zinc-800 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] transition-transform duration-300 group-hover:scale-110">
              <Boxes className="size-4 text-zinc-100" />
            </div>
          </div>
          <div>
            <div className="text-3xl font-semibold tabular-nums tracking-tight text-zinc-50">
              {totalOnHand.toLocaleString()}
            </div>
            <div className="flex items-center text-xs text-zinc-500 mt-1.5 gap-1.5">
              <span className="flex items-center gap-0.5 font-medium text-emerald-400">
                <ArrowUpRight className="size-3" />
                Live
              </span>
              <span>across {activeWarehouseName}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02] transition-colors hover:bg-white/[0.04]">
        <div className="relative rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] p-5 flex flex-col justify-between min-h-[140px] group overflow-hidden">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-medium tracking-wide uppercase text-zinc-400">Available Stock</span>
            <div className="flex size-9 items-center justify-center rounded-xl bg-zinc-950 border border-zinc-800 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] transition-transform duration-300 group-hover:scale-110">
              <CheckCircle2 className="size-4 text-emerald-400" />
            </div>
          </div>
          <div>
            <div className="text-3xl font-semibold tabular-nums tracking-tight text-zinc-50">
              {totalAvailable.toLocaleString()}
            </div>
            <div className="flex items-center text-xs text-zinc-500 mt-1.5 gap-1.5">
              <span className="font-medium text-zinc-300">{totalReserved} reserved</span>
              <span>in pending orders</span>
            </div>
          </div>
        </div>
      </div>

      <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02] transition-colors hover:bg-white/[0.04]">
        <div className="relative rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] p-5 flex flex-col justify-between min-h-[140px] group overflow-hidden">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-medium tracking-wide uppercase text-zinc-400">Active Facilities</span>
            <div className="flex size-9 items-center justify-center rounded-xl bg-zinc-950 border border-zinc-800 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] transition-transform duration-300 group-hover:scale-110">
              <WarehouseIcon className="size-4 text-zinc-100" />
            </div>
          </div>
          <div>
            <div className="text-3xl font-semibold tabular-nums tracking-tight text-zinc-50">
              {warehouses.length}
            </div>
            <div className="flex items-center text-xs text-zinc-500 mt-1.5 gap-1.5">
              <span className="font-medium text-zinc-300">
                {warehouses.filter((w) => w.is_default).length > 0 ? 'Primary Default Set' : 'Custom Routing'}
              </span>
              <span>locations active</span>
            </div>
          </div>
        </div>
      </div>

      <div
        onClick={onFilterLowStock}
        className={`p-1.5 rounded-3xl border transition-all cursor-pointer ${
          isLowStockFiltered
            ? 'border-amber-500/40 bg-amber-500/[0.08] ring-1 ring-amber-500/30'
            : 'border-white/[0.05] bg-white/[0.02] hover:bg-white/[0.04]'
        }`}
      >
        <div className="relative rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] p-5 flex flex-col justify-between min-h-[140px] group overflow-hidden">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-medium tracking-wide uppercase text-zinc-400">Stock Alerts</span>
            <div
              className={`flex size-9 items-center justify-center rounded-xl border transition-transform duration-300 group-hover:scale-110 ${
                lowStockCount > 0
                  ? 'bg-amber-500/10 border-amber-500/20 text-amber-400'
                  : 'bg-zinc-950 border-zinc-800 text-zinc-400'
              }`}
            >
              <AlertTriangle className="size-4" />
            </div>
          </div>
          <div>
            <div className="flex items-baseline gap-2">
              <div
                className={`text-3xl font-semibold tabular-nums tracking-tight ${
                  lowStockCount > 0 ? 'text-amber-400' : 'text-zinc-50'
                }`}
              >
                {lowStockCount}
              </div>
              {outOfStockCount > 0 && (
                <span className="text-xs font-medium text-rose-400">({outOfStockCount} zero stock)</span>
              )}
            </div>
            <div className="flex items-center text-xs text-zinc-500 mt-1.5 gap-1.5">
              <span
                className={`font-medium ${
                  isLowStockFiltered ? 'text-amber-300 underline' : 'text-zinc-400 group-hover:text-zinc-200'
                }`}
              >
                {isLowStockFiltered ? 'Clear filter' : 'Filter low stock items'}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
