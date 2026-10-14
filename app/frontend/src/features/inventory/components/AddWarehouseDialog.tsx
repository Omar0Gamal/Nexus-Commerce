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
import { Warehouse as WarehouseIcon, Loader2, AlertCircle } from 'lucide-react';
import { CreateWarehousePayload, Warehouse } from '../types';

interface AddWarehouseDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onWarehouseCreated: (warehouse: Warehouse) => void;
  onSubmitCreate: (payload: CreateWarehousePayload) => Promise<Warehouse>;
}

export function AddWarehouseDialog({
  isOpen,
  onClose,
  onWarehouseCreated,
  onSubmitCreate,
}: AddWarehouseDialogProps) {
  const [name, setName] = React.useState('');
  const [address, setAddress] = React.useState('');
  const [isDefault, setIsDefault] = React.useState(false);
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [errorMsg, setErrorMsg] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (isOpen) {
      setName('');
      setAddress('');
      setIsDefault(false);
      setErrorMsg(null);
    }
  }, [isOpen]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setErrorMsg('Facility / Warehouse name is required.');
      return;
    }

    setIsSubmitting(true);
    setErrorMsg(null);

    try {
      const created = await onSubmitCreate({
        name: name.trim(),
        address: address.trim() || undefined,
        is_default: isDefault,
      });

      onWarehouseCreated(created);
      onClose();
    } catch (err: any) {
      console.error('Create warehouse error:', err);
      const apiMsg = err?.response?.data?.error || err?.message || 'Failed to create warehouse.';
      setErrorMsg(apiMsg);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && !isSubmitting && onClose()}>
      <DialogContent className="max-w-md bg-zinc-950 border border-zinc-800 text-zinc-100 p-0 overflow-hidden shadow-2xl rounded-3xl">
        <div className="p-6 flex flex-col gap-5">
          <DialogHeader className="gap-1.5">
            <div className="flex items-center gap-2">
              <span className="rounded-full px-2.5 py-0.5 text-[10px] uppercase tracking-wider font-semibold bg-zinc-900 text-zinc-400 border border-zinc-800">
                Logistics Setup
              </span>
            </div>
            <DialogTitle className="text-xl font-semibold tracking-tight text-zinc-50 flex items-center gap-2">
              <WarehouseIcon className="w-5 h-5 text-zinc-400" />
              Add Fulfillment Center
            </DialogTitle>
            <DialogDescription className="text-zinc-400 text-xs">
              Configure a physical warehouse, retail shop backroom, or 3PL distribution hub.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <Label className="text-xs font-medium text-zinc-300">Facility / Warehouse Name *</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Giza Express Hub, Alex Depot"
                required
                className="h-10 bg-zinc-900 border-zinc-800 text-sm text-zinc-100 rounded-xl focus-visible:ring-zinc-600"
              />
            </div>

            <div className="flex flex-col gap-2">
              <Label className="text-xs font-medium text-zinc-300">Street Address & Logistics Details</Label>
              <Input
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                placeholder="e.g. Building 12, Industrial Zone 4, Cairo"
                className="h-10 bg-zinc-900 border-zinc-800 text-sm text-zinc-100 rounded-xl focus-visible:ring-zinc-600"
              />
            </div>

            <div className="flex items-center justify-between p-3 rounded-2xl bg-zinc-900/60 border border-zinc-800/80">
              <div className="flex flex-col">
                <span className="text-xs font-medium text-zinc-200">Default Shipping Origin</span>
                <span className="text-[11px] text-zinc-400">
                  New orders will automatically reserve stock from this facility first.
                </span>
              </div>
              <input
                type="checkbox"
                checked={isDefault}
                onChange={(e) => setIsDefault(e.target.checked)}
                className="w-4 h-4 rounded bg-zinc-900 border-zinc-700 text-zinc-100 focus:ring-0 focus:ring-offset-0 cursor-pointer"
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
                disabled={isSubmitting || !name.trim()}
                className="group relative h-10 rounded-full bg-zinc-100 text-zinc-900 hover:bg-white text-xs font-semibold px-6 shadow-sm active:scale-95 transition-all"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="w-3.5 h-3.5 animate-spin mr-2" />
                    Adding Location...
                  </>
                ) : (
                  <>
                    <span>Create Warehouse</span>
                    <span className="ml-2 flex h-5 w-5 items-center justify-center rounded-full bg-black/10 transition-transform group-hover:scale-110">
                      +
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
