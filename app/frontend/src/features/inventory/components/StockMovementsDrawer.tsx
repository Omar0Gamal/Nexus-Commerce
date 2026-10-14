'use client';

import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import {
  History,
  Boxes,
  ArrowDownRight,
  ArrowUpRight,
  RefreshCw,
  Truck,
  RotateCcw,
  ShoppingBag,
  Clock,
  User,
  Loader2,
} from 'lucide-react';
import { StockMovement, InventoryItemRow, Warehouse } from '../types';
import { getStockMovements } from '../api/inventoryApi';

interface StockMovementsDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  selectedItem: InventoryItemRow | null;
  warehouses: Warehouse[];
}

export function StockMovementsDrawer({
  isOpen,
  onClose,
  selectedItem,
  warehouses,
}: StockMovementsDrawerProps) {
  const [movements, setMovements] = React.useState<StockMovement[]>([]);
  const [isLoading, setIsLoading] = React.useState<boolean>(false);

  React.useEffect(() => {
    if (isOpen && selectedItem?.variantId) {
      setIsLoading(true);
      getStockMovements(selectedItem.variantId)
        .then((movementsList) => setMovements(movementsList))
        .catch((err) => console.error('Failed to load movements:', err))
        .finally(() => setIsLoading(false));
    }
  }, [isOpen, selectedItem?.variantId]);

  const getWarehouseName = (warehouseId?: any) => {
    if (!warehouseId) return 'Main Facility';
    const id = typeof warehouseId === 'string' ? warehouseId : warehouseId?.Bytes || '';
    const wh = warehouses.find((w) => w.id === id);
    return wh ? wh.name : 'Distribution Center';
  };

  const getMovementBadge = (type: string, delta: number) => {
    switch (type) {
      case 'restock':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/20">
            <ArrowUpRight className="w-3 h-3" />
            Restock
          </span>
        );
      case 'adjustment':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-blue-500/15 text-blue-400 border border-blue-500/20">
            <RefreshCw className="w-3 h-3" />
            Cycle Audit
          </span>
        );
      case 'transfer':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-purple-500/15 text-purple-400 border border-purple-500/20">
            <Truck className="w-3 h-3" />
            Transfer
          </span>
        );
      case 'return':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-cyan-500/15 text-cyan-400 border border-cyan-500/20">
            <RotateCcw className="w-3 h-3" />
            Customer Return
          </span>
        );
      case 'sale':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-amber-500/15 text-amber-400 border border-amber-500/20">
            <ShoppingBag className="w-3 h-3" />
            Order Fulfillment
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-zinc-800 text-zinc-300">
            Movement
          </span>
        );
    }
  };

  if (!selectedItem) return null;

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-xl bg-zinc-950 border border-zinc-800 text-zinc-100 p-0 overflow-hidden shadow-2xl rounded-3xl">
        <div className="p-6 flex flex-col gap-5">
          <DialogHeader className="gap-1.5">
            <div className="flex items-center gap-2">
              <span className="rounded-full px-2.5 py-0.5 text-[10px] uppercase tracking-wider font-semibold bg-zinc-900 text-zinc-400 border border-zinc-800">
                Audit Trail
              </span>
            </div>
            <DialogTitle className="text-xl font-semibold tracking-tight text-zinc-50 flex items-center gap-2">
              <History className="w-5 h-5 text-zinc-400" />
              Stock Movement History
            </DialogTitle>
            <DialogDescription className="text-zinc-400 text-xs">
              Complete chronological audit ledger of stock additions, removals, and relocations.
            </DialogDescription>
          </DialogHeader>

          <div className="p-3.5 rounded-2xl bg-zinc-900/60 border border-zinc-800/80 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-zinc-950 border border-zinc-800 flex items-center justify-center shrink-0">
                <Boxes className="w-5 h-5 text-zinc-400" />
              </div>
              <div>
                <h4 className="text-sm font-medium text-zinc-100">{selectedItem.productTitle}</h4>
                <div className="flex items-center gap-2 text-xs text-zinc-400">
                  <span className="font-mono text-zinc-300">{selectedItem.variantTitle}</span>
                  {selectedItem.sku && (
                    <>
                      <span>•</span>
                      <span className="font-mono text-zinc-500">SKU: {selectedItem.sku}</span>
                    </>
                  )}
                </div>
              </div>
            </div>
            <div className="text-right">
              <span className="text-[10px] uppercase tracking-wider text-zinc-500 font-mono">Current Total</span>
              <div className="text-sm font-semibold tabular-nums text-zinc-100">
                {selectedItem.totalOnHand} units
              </div>
            </div>
          </div>

          <div className="max-h-[380px] overflow-y-auto pr-1 flex flex-col gap-2.5">
            {isLoading ? (
              <div className="py-12 flex flex-col items-center justify-center gap-2 text-zinc-500 text-xs">
                <Loader2 className="w-5 h-5 animate-spin text-zinc-400" />
                <span>Loading movement records...</span>
              </div>
            ) : movements.length === 0 ? (
              <div className="py-12 text-center text-xs text-zinc-500 flex flex-col items-center justify-center gap-2">
                <History className="w-8 h-8 text-zinc-700" />
                <span>No stock movement transactions recorded yet.</span>
              </div>
            ) : (
              movements.map((mov) => {
                const noteText = typeof mov.note === 'string' ? mov.note : mov.note?.String || '';
                const createdBy =
                  typeof mov.created_by === 'string' ? mov.created_by : mov.created_by?.String || 'Staff';
                const isPositive = mov.quantity_delta > 0;

                return (
                  <div
                    key={mov.id}
                    className="p-3.5 rounded-2xl bg-zinc-900/40 border border-zinc-800/60 hover:bg-zinc-900/70 transition-colors flex flex-col gap-2"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        {getMovementBadge(mov.movement_type, mov.quantity_delta)}
                        <span className="text-xs font-medium text-zinc-300">
                          {getWarehouseName(mov.warehouse_id)}
                        </span>
                      </div>
                      <div
                        className={`text-sm font-mono font-semibold tabular-nums flex items-center gap-0.5 ${
                          isPositive ? 'text-emerald-400' : 'text-rose-400'
                        }`}
                      >
                        {isPositive ? `+${mov.quantity_delta}` : mov.quantity_delta}
                        <span className="text-[10px] text-zinc-500 font-normal">units</span>
                      </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-zinc-500 pt-1 border-t border-zinc-800/40">
                      <div className="flex items-center gap-1.5 line-clamp-1">
                        {noteText ? (
                          <span className="text-zinc-400 italic">“{noteText}”</span>
                        ) : (
                          <span className="text-zinc-600">Direct adjustment</span>
                        )}
                      </div>
                      <div className="flex items-center gap-3 shrink-0 text-[11px]">
                        <span className="flex items-center gap-1 text-zinc-400">
                          <User className="w-3 h-3 text-zinc-500" />
                          {createdBy}
                        </span>
                        <span className="flex items-center gap-1 text-zinc-500">
                          <Clock className="w-3 h-3 text-zinc-600" />
                          {new Date(mov.created_at).toLocaleDateString(undefined, {
                            month: 'short',
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </span>
                      </div>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
