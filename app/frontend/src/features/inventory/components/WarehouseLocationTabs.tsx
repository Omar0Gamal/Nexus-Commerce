'use client';

import * as React from 'react';
import { Warehouse as WarehouseIcon, Plus, MapPin, Check } from 'lucide-react';
import { Warehouse } from '../types';
import { getWarehouseAddress } from '../api/inventoryApi';

interface WarehouseLocationTabsProps {
  warehouses: Warehouse[];
  activeWarehouseId: string;
  onSelectWarehouse: (warehouseId: string) => void;
  onOpenAddWarehouse: () => void;
}

export function WarehouseLocationTabs({
  warehouses,
  activeWarehouseId,
  onSelectWarehouse,
  onOpenAddWarehouse,
}: WarehouseLocationTabsProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-1 border-b border-white/[0.06]">
      <div className="flex items-center gap-2 overflow-x-auto pb-2 sm:pb-0 scrollbar-none">
        <button
          onClick={() => onSelectWarehouse('all')}
          className={`flex items-center gap-2 px-4 py-2 rounded-full text-xs font-medium transition-all whitespace-nowrap ${
            activeWarehouseId === 'all'
              ? 'bg-zinc-100 text-zinc-900 shadow-sm shadow-black/20 font-semibold'
              : 'bg-zinc-900/80 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/80 border border-zinc-800/70'
          }`}
        >
          <WarehouseIcon className="w-3.5 h-3.5" />
          <span>All Warehouses</span>
          <span
            className={`px-1.5 py-0.5 rounded-full text-[10px] font-mono ${
              activeWarehouseId === 'all' ? 'bg-zinc-900/10 text-zinc-900' : 'bg-zinc-800 text-zinc-400'
            }`}
          >
            {warehouses.length}
          </span>
        </button>

        {warehouses.map((wh) => {
          const isActive = activeWarehouseId === wh.id;
          const address = getWarehouseAddress(wh.address);

          return (
            <button
              key={wh.id}
              onClick={() => onSelectWarehouse(wh.id)}
              title={address ? `${wh.name} • ${address}` : wh.name}
              className={`group flex items-center gap-2 px-4 py-2 rounded-full text-xs font-medium transition-all whitespace-nowrap ${
                isActive
                  ? 'bg-zinc-100 text-zinc-900 shadow-sm shadow-black/20 font-semibold'
                  : 'bg-zinc-900/80 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/80 border border-zinc-800/70'
              }`}
            >
              <MapPin
                className={`w-3.5 h-3.5 ${
                  isActive ? 'text-zinc-900' : 'text-zinc-500 group-hover:text-zinc-300'
                }`}
              />
              <span>{wh.name}</span>
              {wh.is_default && (
                <span
                  className={`px-1.5 py-0.2 rounded-full text-[9px] uppercase tracking-wider font-semibold ${
                    isActive ? 'bg-zinc-900/15 text-zinc-900' : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20'
                  }`}
                >
                  Default
                </span>
              )}
            </button>
          );
        })}
      </div>

      <button
        onClick={onOpenAddWarehouse}
        className="group flex items-center gap-1.5 px-3.5 py-1.5 rounded-full bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-zinc-50 hover:bg-zinc-800 transition-all self-start sm:self-auto shrink-0 shadow-sm active:scale-95"
      >
        <Plus className="w-3.5 h-3.5 text-zinc-400 group-hover:text-zinc-100" />
        <span>Add Location</span>
      </button>
    </div>
  );
}
