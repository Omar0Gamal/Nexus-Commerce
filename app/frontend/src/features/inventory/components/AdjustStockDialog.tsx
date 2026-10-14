'use client';

import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import {
  Boxes,
  Plus,
  Minus,
  Equal,
  ArrowRight,
  Warehouse as WarehouseIcon,
  Loader2,
  AlertCircle,
  CheckCircle2,
} from 'lucide-react';
import { Warehouse, InventoryItemRow, MovementType, AdjustStockPayload } from '../types';

interface AdjustStockDialogProps {
  isOpen: boolean;
  onClose: () => void;
  selectedItem: InventoryItemRow | null;
  warehouses: Warehouse[];
  defaultWarehouseId?: string;
  onAdjustSuccess: () => void;
  onAdjustSubmit: (payload: AdjustStockPayload) => Promise<void>;
}

export function AdjustStockDialog({
  isOpen,
  onClose,
  selectedItem,
  warehouses,
  defaultWarehouseId,
  onAdjustSuccess,
  onAdjustSubmit,
}: AdjustStockDialogProps) {
  const [selectedWarehouseId, setSelectedWarehouseId] = React.useState<string>('');
  const [adjustmentMode, setAdjustmentMode] = React.useState<'add' | 'remove' | 'exact'>('add');
  const [quantityInput, setQuantityInput] = React.useState<number>(1);
  const [movementType, setMovementType] = React.useState<MovementType>('restock');
  const [note, setNote] = React.useState<string>('');
  const [referenceId, setReferenceId] = React.useState<string>('');
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [errorMsg, setErrorMsg] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (isOpen) {
      if (defaultWarehouseId && defaultWarehouseId !== 'all') {
        setSelectedWarehouseId(defaultWarehouseId);
      } else if (warehouses.length > 0) {
        const def = warehouses.find((w) => w.is_default) || warehouses[0];
        setSelectedWarehouseId(def.id);
      }
      setAdjustmentMode('add');
      setQuantityInput(1);
      setMovementType('restock');
      setNote('');
      setReferenceId('');
      setErrorMsg(null);
    }
  }, [isOpen, defaultWarehouseId, warehouses]);

  const currentQuantity =
    selectedItem && selectedWarehouseId ? selectedItem.levels[selectedWarehouseId] ?? 0 : 0;

  let delta = 0;
  let newQuantity = currentQuantity;

  if (adjustmentMode === 'add') {
    delta = Math.max(0, quantityInput || 0);
    newQuantity = currentQuantity + delta;
  } else if (adjustmentMode === 'remove') {
    delta = -Math.max(0, quantityInput || 0);
    newQuantity = currentQuantity + delta;
  } else if (adjustmentMode === 'exact') {
    const target = Math.max(0, quantityInput || 0);
    delta = target - currentQuantity;
    newQuantity = target;
  }

  const handleModeChange = (mode: 'add' | 'remove' | 'exact') => {
    setAdjustmentMode(mode);
    setErrorMsg(null);
    if (mode === 'add') {
      setMovementType('restock');
      if (quantityInput <= 0) setQuantityInput(1);
    } else if (mode === 'remove') {
      setMovementType('adjustment');
      if (quantityInput <= 0) setQuantityInput(1);
    } else if (mode === 'exact') {
      setMovementType('adjustment');
      setQuantityInput(currentQuantity);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedItem) return;

    if (!selectedWarehouseId) {
      setErrorMsg('Please select a target warehouse.');
      return;
    }

    if (delta === 0) {
      setErrorMsg('Adjustment delta cannot be 0.');
      return;
    }

    if (newQuantity < 0) {
      setErrorMsg(
        `Cannot reduce stock below 0. Current stock is ${currentQuantity}, attempted reduction is ${Math.abs(delta)}.`
      );
      return;
    }

    setIsSubmitting(true);
    setErrorMsg(null);

    try {
      await onAdjustSubmit({
        variant_id: selectedItem.variantId,
        warehouse_id: selectedWarehouseId,
        delta,
        movement_type: movementType,
        note: note.trim() || undefined,
        reference_id: referenceId.trim() || undefined,
      });

      onAdjustSuccess();
      onClose();
    } catch (err: any) {
      console.error('Stock adjustment error:', err);
      const apiMsg = err?.response?.data?.error || err?.message || 'Failed to adjust stock. Please try again.';
      setErrorMsg(apiMsg);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!selectedItem) return null;

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && !isSubmitting && onClose()}>
      <DialogContent className="max-w-lg bg-zinc-950 border border-zinc-800 text-zinc-100 p-0 overflow-hidden shadow-2xl rounded-3xl">
        <div className="p-6 flex flex-col gap-6">
          <DialogHeader className="gap-1.5">
            <div className="flex items-center gap-2">
              <span className="rounded-full px-2.5 py-0.5 text-[10px] uppercase tracking-wider font-semibold bg-zinc-900 text-zinc-400 border border-zinc-800">
                Inventory Action
              </span>
            </div>
            <DialogTitle className="text-xl font-semibold tracking-tight text-zinc-50">
              Adjust Stock Level
            </DialogTitle>
            <DialogDescription className="text-zinc-400 text-xs">
              Record a stock receipt, cycle count correction, transfer, or manual write-off.
            </DialogDescription>
          </DialogHeader>

          <div className="p-3.5 rounded-2xl bg-zinc-900/70 border border-zinc-800/80 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-zinc-950 border border-zinc-800 flex items-center justify-center shrink-0">
                {selectedItem.imageUrl ? (
                  <img
                    src={selectedItem.imageUrl}
                    alt={selectedItem.productTitle}
                    className="w-full h-full object-cover rounded-xl"
                  />
                ) : (
                  <Boxes className="w-5 h-5 text-zinc-400" />
                )}
              </div>
              <div>
                <h4 className="text-sm font-medium text-zinc-100 line-clamp-1">{selectedItem.productTitle}</h4>
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
          </div>

          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            <div className="flex flex-col gap-2">
              <Label className="text-xs font-medium text-zinc-300">Target Warehouse / Location</Label>
              <select
                value={selectedWarehouseId}
                onChange={(e) => setSelectedWarehouseId(e.target.value)}
                className="h-10 px-3.5 rounded-xl bg-zinc-900 border border-zinc-800 text-sm font-medium text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-600 transition-shadow"
              >
                {warehouses.map((wh) => (
                  <option key={wh.id} value={wh.id}>
                    {wh.name} {wh.is_default ? '(Primary Default)' : ''}
                  </option>
                ))}
              </select>
            </div>

            <div className="flex flex-col gap-2">
              <Label className="text-xs font-medium text-zinc-300">Adjustment Type</Label>
              <div className="grid grid-cols-3 gap-2 p-1 rounded-2xl bg-zinc-900 border border-zinc-800/80">
                <button
                  type="button"
                  onClick={() => handleModeChange('add')}
                  className={`flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-semibold transition-all ${
                    adjustmentMode === 'add'
                      ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 shadow-sm'
                      : 'text-zinc-400 hover:text-zinc-200'
                  }`}
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>Receive / Add</span>
                </button>
                <button
                  type="button"
                  onClick={() => handleModeChange('remove')}
                  className={`flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-semibold transition-all ${
                    adjustmentMode === 'remove'
                      ? 'bg-rose-500/15 text-rose-400 border border-rose-500/30 shadow-sm'
                      : 'text-zinc-400 hover:text-zinc-200'
                  }`}
                >
                  <Minus className="w-3.5 h-3.5" />
                  <span>Remove / Loss</span>
                </button>
                <button
                  type="button"
                  onClick={() => handleModeChange('exact')}
                  className={`flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-semibold transition-all ${
                    adjustmentMode === 'exact'
                      ? 'bg-blue-500/15 text-blue-400 border border-blue-500/30 shadow-sm'
                      : 'text-zinc-400 hover:text-zinc-200'
                  }`}
                >
                  <Equal className="w-3.5 h-3.5" />
                  <span>Set Exact Total</span>
                </button>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between">
                <Label className="text-xs font-medium text-zinc-300">
                  {adjustmentMode === 'exact' ? 'New Total Quantity' : 'Quantity Delta'}
                </Label>
                {adjustmentMode !== 'exact' && (
                  <div className="flex items-center gap-1">
                    {[5, 10, 25, 50].map((step) => (
                      <button
                        key={step}
                        type="button"
                        onClick={() => setQuantityInput(step)}
                        className="px-2 py-0.5 rounded-lg text-[10px] font-mono bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 transition-colors"
                      >
                        +{step}
                      </button>
                    ))}
                  </div>
                )}
              </div>
              <Input
                type="number"
                min={adjustmentMode === 'exact' ? 0 : 1}
                value={quantityInput}
                onChange={(e) => setQuantityInput(parseInt(e.target.value, 10) || 0)}
                placeholder="Enter quantity"
                className="h-10 bg-zinc-900 border-zinc-800 text-zinc-100 rounded-xl tabular-nums text-base focus-visible:ring-zinc-600"
              />
            </div>

            <div className="p-3.5 rounded-2xl bg-zinc-900/60 border border-zinc-800/80 flex items-center justify-between text-xs">
              <div className="flex flex-col">
                <span className="text-zinc-400">Current</span>
                <span className="text-sm font-semibold tabular-nums text-zinc-200">{currentQuantity} units</span>
              </div>
              <div className="flex items-center gap-2">
                <span
                  className={`font-mono font-semibold ${
                    delta > 0 ? 'text-emerald-400' : delta < 0 ? 'text-rose-400' : 'text-zinc-400'
                  }`}
                >
                  {delta > 0 ? `+${delta}` : delta}
                </span>
                <ArrowRight className="w-3.5 h-3.5 text-zinc-500" />
              </div>
              <div className="flex flex-col items-end">
                <span className="text-zinc-400">New On Hand</span>
                <span
                  className={`text-sm font-semibold tabular-nums ${
                    newQuantity < 0
                      ? 'text-rose-400'
                      : newQuantity <= selectedItem.lowStockThreshold
                      ? 'text-amber-400'
                      : 'text-zinc-50'
                  }`}
                >
                  {newQuantity} units
                </span>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="flex flex-col gap-2">
                <Label className="text-xs font-medium text-zinc-300">Movement Reason</Label>
                <select
                  value={movementType}
                  onChange={(e) => setMovementType(e.target.value as MovementType)}
                  className="h-10 px-3 rounded-xl bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-600"
                >
                  <option value="restock">Restock / Shipment Inbound</option>
                  <option value="adjustment">Cycle Count / Audit Correction</option>
                  <option value="transfer">Inter-Warehouse Transfer</option>
                  <option value="return">Customer Return</option>
                  <option value="sale">Manual / Offline Sale</option>
                  <option value="pre_order_reserve">Pre-order Reserve</option>
                </select>
              </div>

              <div className="flex flex-col gap-2">
                <Label className="text-xs font-medium text-zinc-300">Reference / PO # (Optional)</Label>
                <Input
                  value={referenceId}
                  onChange={(e) => setReferenceId(e.target.value)}
                  placeholder="e.g. PO-8491, RMA-22"
                  className="h-10 bg-zinc-900 border-zinc-800 text-xs text-zinc-200 rounded-xl"
                />
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <Label className="text-xs font-medium text-zinc-300">Internal Audit Note (Optional)</Label>
              <Input
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="e.g. Verified during Q3 warehouse cycle count"
                className="h-10 bg-zinc-900 border-zinc-800 text-xs text-zinc-200 rounded-xl"
              />
            </div>

            {errorMsg && (
              <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs flex items-center gap-2">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{errorMsg}</span>
              </div>
            )}

            <DialogFooter className="gap-2 sm:gap-0 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={onClose}
                disabled={isSubmitting}
                className="h-10 rounded-full border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-800 text-xs"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={isSubmitting || newQuantity < 0 || delta === 0}
                className="group relative h-10 rounded-full bg-zinc-100 text-zinc-900 hover:bg-white text-xs font-semibold px-6 shadow-sm active:scale-95 transition-all"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="w-3.5 h-3.5 animate-spin mr-2" />
                    Updating Inventory...
                  </>
                ) : (
                  <>
                    <span>Confirm Adjustment</span>
                    <span className="ml-2 flex h-5 w-5 items-center justify-center rounded-full bg-black/10 transition-transform group-hover:scale-110">
                      ✓
                    </span>
                  </>
                )}
              </Button>
            </DialogFooter>
          </form>
        </div>
      </DialogContent>
    </Dialog>
  );
}
