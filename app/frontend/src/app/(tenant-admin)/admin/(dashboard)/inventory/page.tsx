'use client';

import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Boxes, Warehouse as WarehouseIcon, Plus, RefreshCw, CheckCircle2, SlidersHorizontal, Sparkles } from 'lucide-react';
import { useGetApiV1CatalogProducts } from '@/api/endpoints/catalog/catalog';
import { WarehouseMetrics } from '@/features/inventory/components/WarehouseMetrics';
import { WarehouseLocationTabs } from '@/features/inventory/components/WarehouseLocationTabs';
import { InventoryTable } from '@/features/inventory/components/InventoryTable';
import { AdjustStockDialog } from '@/features/inventory/components/AdjustStockDialog';
import { StockMovementsDrawer } from '@/features/inventory/components/StockMovementsDrawer';
import { AddWarehouseDialog } from '@/features/inventory/components/AddWarehouseDialog';
import {
  getWarehouses,
  createWarehouse,
  getInventoryLevels,
  adjustStock,
  getLowStockVariants,
} from '@/features/inventory/api/inventoryApi';
import {
  Warehouse,
  InventoryItemRow,
  AdjustStockPayload,
  CreateWarehousePayload,
  InventoryLevel,
} from '@/features/inventory/types';

export default function InventoryPage() {
  const queryClient = useQueryClient();

  const [activeWarehouseId, setActiveWarehouseId] = React.useState<string>('all');
  const [searchQuery, setSearchQuery] = React.useState<string>('');
  const [statusFilter, setStatusFilter] = React.useState<'all' | 'in_stock' | 'low_stock' | 'out_of_stock'>('all');
  const [isLowStockFiltered, setIsLowStockFiltered] = React.useState<boolean>(false);

  const [selectedAdjustItem, setSelectedAdjustItem] = React.useState<InventoryItemRow | null>(null);
  const [selectedHistoryItem, setSelectedHistoryItem] = React.useState<InventoryItemRow | null>(null);
  const [isAddWarehouseOpen, setIsAddWarehouseOpen] = React.useState<boolean>(false);

  const [bannerMessage, setBannerMessage] = React.useState<string | null>(null);

  const {
    data: warehouses = [],
    isLoading: isWarehousesLoading,
    refetch: refetchWarehouses,
  } = useQuery({
    queryKey: ['inventory', 'warehouses'],
    queryFn: getWarehouses,
  });

  const {
    data: catalogEnvelope,
    isLoading: isCatalogLoading,
    refetch: refetchCatalog,
  } = useGetApiV1CatalogProducts();

  const products = catalogEnvelope?.data || [];

  const {
    data: lowStockList = [],
    refetch: refetchLowStock,
  } = useQuery({
    queryKey: ['inventory', 'low-stock'],
    queryFn: getLowStockVariants,
  });

  const lowStockVariantIds = React.useMemo(() => {
    return new Set(lowStockList.map((ls) => ls.variant_id));
  }, [lowStockList]);

  const [itemRows, setItemRows] = React.useState<InventoryItemRow[]>([]);

  React.useEffect(() => {
    if (!products || products.length === 0) {
      const sampleWarehouses = warehouses.length > 0 ? warehouses : [
        { id: 'wh-cairo-main', shop_id: 'sample', name: 'Cairo Fulfillment Hub', is_default: true, is_active: true, created_at: '' },
        { id: 'wh-alex-sea', shop_id: 'sample', name: 'Alexandria Coastal Hub', is_default: false, is_active: true, created_at: '' },
        { id: 'wh-delta-express', shop_id: 'sample', name: 'Delta Express Center', is_default: false, is_active: true, created_at: '' },
      ];

      const sampleRows: InventoryItemRow[] = [
        {
          productId: 'prod-001',
          variantId: 'var-001-red-42',
          productTitle: 'Air Flow Velocity Sneaker',
          variantTitle: 'Size: 42 / Crimson Red',
          sku: 'AFV-42-RED',
          price: 'EGP 1,899.00',
          status: 'active',
          levels: {
            [sampleWarehouses[0]?.id || 'wh-cairo-main']: 45,
            [sampleWarehouses[1]?.id || 'wh-alex-sea']: 12,
            [sampleWarehouses[2]?.id || 'wh-delta-express']: 8,
          },
          reservedLevels: {
            [sampleWarehouses[0]?.id || 'wh-cairo-main']: 3,
            [sampleWarehouses[1]?.id || 'wh-alex-sea']: 0,
            [sampleWarehouses[2]?.id || 'wh-delta-express']: 1,
          },
          totalOnHand: 65,
          totalReserved: 4,
          totalAvailable: 61,
          lowStockThreshold: 15,
          isLowStock: false,
          isOutOfStock: false,
        },
        {
          productId: 'prod-001',
          variantId: 'var-001-blk-43',
          productTitle: 'Air Flow Velocity Sneaker',
          variantTitle: 'Size: 43 / Stealth Black',
          sku: 'AFV-43-BLK',
          price: 'EGP 1,899.00',
          status: 'active',
          levels: {
            [sampleWarehouses[0]?.id || 'wh-cairo-main']: 4,
            [sampleWarehouses[1]?.id || 'wh-alex-sea']: 2,
            [sampleWarehouses[2]?.id || 'wh-delta-express']: 0,
          },
          reservedLevels: {
            [sampleWarehouses[0]?.id || 'wh-cairo-main']: 2,
            [sampleWarehouses[1]?.id || 'wh-alex-sea']: 0,
            [sampleWarehouses[2]?.id || 'wh-delta-express']: 0,
          },
          totalOnHand: 6,
          totalReserved: 2,
          totalAvailable: 4,
          lowStockThreshold: 10,
          isLowStock: true,
          isOutOfStock: false,
        },
        {
          productId: 'prod-002',
          variantId: 'var-002-boots-41',
          productTitle: 'Tuscan Heritage Leather Boots',
          variantTitle: 'Size: 41 / Chestnut Brown',
          sku: 'THB-41-BRN',
          price: 'EGP 3,250.00',
          status: 'active',
          levels: {
            [sampleWarehouses[0]?.id || 'wh-cairo-main']: 0,
            [sampleWarehouses[1]?.id || 'wh-alex-sea']: 0,
            [sampleWarehouses[2]?.id || 'wh-delta-express']: 0,
          },
          reservedLevels: {
            [sampleWarehouses[0]?.id || 'wh-cairo-main']: 0,
            [sampleWarehouses[1]?.id || 'wh-alex-sea']: 0,
            [sampleWarehouses[2]?.id || 'wh-delta-express']: 0,
          },
          totalOnHand: 0,
          totalReserved: 0,
          totalAvailable: 0,
          lowStockThreshold: 5,
          isLowStock: false,
          isOutOfStock: true,
        },
        {
          productId: 'prod-003',
          variantId: 'var-003-linen-l',
          productTitle: 'Relaxed Tailored Linen Shirt',
          variantTitle: 'Size: L / Olive Green',
          sku: 'RTL-L-OLV',
          price: 'EGP 1,120.00',
          status: 'active',
          levels: {
            [sampleWarehouses[0]?.id || 'wh-cairo-main']: 28,
            [sampleWarehouses[1]?.id || 'wh-alex-sea']: 14,
            [sampleWarehouses[2]?.id || 'wh-delta-express']: 6,
          },
          reservedLevels: {
            [sampleWarehouses[0]?.id || 'wh-cairo-main']: 1,
            [sampleWarehouses[1]?.id || 'wh-alex-sea']: 0,
            [sampleWarehouses[2]?.id || 'wh-delta-express']: 0,
          },
          totalOnHand: 48,
          totalReserved: 1,
          totalAvailable: 47,
          lowStockThreshold: 8,
          isLowStock: false,
          isOutOfStock: false,
        },
      ];
      setItemRows(sampleRows);
      return;
    }

    const rows: InventoryItemRow[] = [];
    const defaultWh = warehouses.find((w) => w.is_default) || warehouses[0];

    products.forEach((product) => {
      const pImages = product.images || [];
      const primaryImage = pImages.find((img) => img.is_primary)?.url || pImages[0]?.url;

      if (product.variants && product.variants.length > 0) {
        product.variants.forEach((v) => {
          const varId = v.id || `var-${product.id}-${v.title}`;
          const onHand = v.stock_quantity ?? product.stock_quantity ?? 0;
          const threshold = product.low_stock_threshold ?? 5;
          const isLow = lowStockVariantIds.has(varId) || (onHand > 0 && onHand <= threshold);
          const isOut = onHand === 0;

          const levels: Record<string, number> = {};
          const reserved: Record<string, number> = {};

          if (warehouses.length > 0) {
            warehouses.forEach((wh, index) => {
              if (wh.is_default || index === 0) {
                levels[wh.id] = onHand;
                reserved[wh.id] = 0;
              } else {
                levels[wh.id] = 0;
                reserved[wh.id] = 0;
              }
            });
          }

          rows.push({
            productId: product.id || '',
            variantId: varId,
            productTitle: product.title || 'Product',
            variantTitle: v.title || 'Standard',
            sku: v.sku || product.sku || '',
            price: v.price || product.price || '',
            imageUrl: primaryImage,
            status: product.status || 'active',
            levels,
            reservedLevels: reserved,
            totalOnHand: onHand,
            totalReserved: 0,
            totalAvailable: onHand,
            lowStockThreshold: threshold,
            isLowStock: isLow,
            isOutOfStock: isOut,
          });
        });
      } else {
        const prodId = product.id || '';
        const onHand = product.stock_quantity ?? 0;
        const threshold = product.low_stock_threshold ?? 5;
        const isLow = lowStockVariantIds.has(prodId) || (onHand > 0 && onHand <= threshold);
        const isOut = onHand === 0;

        const levels: Record<string, number> = {};
        const reserved: Record<string, number> = {};

        if (warehouses.length > 0) {
          warehouses.forEach((wh, index) => {
            if (wh.is_default || index === 0) {
              levels[wh.id] = onHand;
            } else {
              levels[wh.id] = 0;
            }
            reserved[wh.id] = 0;
          });
        }

        rows.push({
          productId: prodId,
          variantId: prodId,
          productTitle: product.title || 'Product',
          variantTitle: 'Standard / Base',
          sku: product.sku || '',
          price: product.price || '',
          imageUrl: primaryImage,
          status: product.status || 'active',
          levels,
          reservedLevels: reserved,
          totalOnHand: onHand,
          totalReserved: 0,
          totalAvailable: onHand,
          lowStockThreshold: threshold,
          isLowStock: isLow,
          isOutOfStock: isOut,
        });
      }
    });

    setItemRows(rows);
  }, [products, warehouses, lowStockVariantIds]);

  const handleAdjustSubmit = async (payload: AdjustStockPayload) => {
    await adjustStock(payload);

    setItemRows((prevRows) =>
      prevRows.map((row) => {
        if (row.variantId === payload.variant_id) {
          const prevWarehouseQty = row.levels[payload.warehouse_id] ?? 0;
          const newWarehouseQty = Math.max(0, prevWarehouseQty + payload.delta);
          const updatedLevels = {
            ...row.levels,
            [payload.warehouse_id]: newWarehouseQty,
          };

          const newTotalOnHand = Object.values(updatedLevels).reduce((a, b) => a + b, 0);
          const newTotalAvailable = Math.max(0, newTotalOnHand - row.totalReserved);

          return {
            ...row,
            levels: updatedLevels,
            totalOnHand: newTotalOnHand,
            totalAvailable: newTotalAvailable,
            isLowStock: newTotalOnHand > 0 && newTotalOnHand <= row.lowStockThreshold,
            isOutOfStock: newTotalOnHand === 0,
          };
        }
        return row;
      })
    );

    const sign = payload.delta > 0 ? `+${payload.delta}` : `${payload.delta}`;
    setBannerMessage(`Successfully adjusted quantity (${sign} units). Changes saved.`);
    setTimeout(() => setBannerMessage(null), 5000);

    queryClient.invalidateQueries({ queryKey: ['inventory'] });
    queryClient.invalidateQueries({ queryKey: ['/api/v1/catalog/products'] });
  };

  const handleWarehouseCreated = (newWh: Warehouse) => {
    refetchWarehouses();
    setActiveWarehouseId(newWh.id);
    setBannerMessage(`Fulfillment facility "${newWh.name}" created successfully.`);
    setTimeout(() => setBannerMessage(null), 5000);
  };

  const handleRefresh = async () => {
    await Promise.all([refetchWarehouses(), refetchCatalog(), refetchLowStock()]);
    setBannerMessage('Inventory levels refreshed.');
    setTimeout(() => setBannerMessage(null), 3000);
  };

  const handleFilterLowStock = () => {
    if (isLowStockFiltered) {
      setIsLowStockFiltered(false);
      setStatusFilter('all');
    } else {
      setIsLowStockFiltered(true);
      setStatusFilter('low_stock');
    }
  };

  const isLoading = isWarehousesLoading || isCatalogLoading;

  return (
    <div className="flex flex-col gap-8 pb-16">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <span className="rounded-full px-3 py-1 text-[10px] uppercase tracking-[0.2em] font-medium text-zinc-400 border border-zinc-800 bg-zinc-900/60 w-max flex items-center gap-1.5 shadow-sm">
              <Sparkles className="w-3 h-3 text-emerald-400" />
              Multi-Warehouse Logistics
            </span>
          </div>

          <h1 className="text-3xl font-semibold tracking-tight text-zinc-50">
            Inventory & Distribution
          </h1>
          <p className="text-sm text-zinc-400 max-w-xl">
            Monitor real-time stock levels across regional warehouses, execute manual inventory
            adjustments, audit stock movement logs, and track fulfillment thresholds.
          </p>
        </div>

        <div className="flex items-center gap-3 self-start md:self-auto">
          <button
            onClick={() => {
              if (itemRows.length > 0) {
                setSelectedAdjustItem(itemRows[0]);
              }
            }}
            disabled={itemRows.length === 0}
            className="group flex h-10 items-center justify-center rounded-full bg-zinc-100 hover:bg-white px-5 text-xs font-semibold text-zinc-950 transition-all shadow-sm active:scale-95 disabled:opacity-50"
          >
            <span>Manual Adjustment</span>
            <span className="ml-2 flex h-5 w-5 items-center justify-center rounded-full bg-black/10 transition-transform group-hover:scale-110">
              ±
            </span>
          </button>

          <button
            onClick={() => setIsAddWarehouseOpen(true)}
            className="group flex h-10 items-center justify-center rounded-full bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 px-4 text-xs font-semibold text-zinc-200 transition-all shadow-sm active:scale-95"
          >
            <Plus className="w-3.5 h-3.5 mr-1.5 text-zinc-400 group-hover:text-zinc-100" />
            <span>New Facility</span>
          </button>
        </div>
      </div>

      {bannerMessage && (
        <div className="p-3 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs flex items-center justify-between animate-in fade-in slide-in-from-top-2 duration-300">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
            <span>{bannerMessage}</span>
          </div>
          <button
            onClick={() => setBannerMessage(null)}
            className="text-zinc-500 hover:text-zinc-300 text-xs px-2"
          >
            Dismiss
          </button>
        </div>
      )}

      <WarehouseMetrics
        items={itemRows}
        warehouses={warehouses}
        activeWarehouseId={activeWarehouseId}
        onFilterLowStock={handleFilterLowStock}
        isLowStockFiltered={isLowStockFiltered || statusFilter === 'low_stock'}
      />

      <WarehouseLocationTabs
        warehouses={warehouses}
        activeWarehouseId={activeWarehouseId}
        onSelectWarehouse={(whId) => setActiveWarehouseId(whId)}
        onOpenAddWarehouse={() => setIsAddWarehouseOpen(true)}
      />

      <InventoryTable
        items={itemRows}
        warehouses={warehouses}
        activeWarehouseId={activeWarehouseId}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        statusFilter={statusFilter}
        onStatusFilterChange={(status) => {
          setStatusFilter(status);
          setIsLowStockFiltered(status === 'low_stock');
        }}
        onOpenAdjust={(item) => setSelectedAdjustItem(item)}
        onOpenHistory={(item) => setSelectedHistoryItem(item)}
        isLoading={isLoading}
        onRefresh={handleRefresh}
      />

      <AdjustStockDialog
        isOpen={!!selectedAdjustItem}
        onClose={() => setSelectedAdjustItem(null)}
        selectedItem={selectedAdjustItem}
        warehouses={warehouses}
        defaultWarehouseId={activeWarehouseId}
        onAdjustSuccess={() => setSelectedAdjustItem(null)}
        onAdjustSubmit={handleAdjustSubmit}
      />

      <StockMovementsDrawer
        isOpen={!!selectedHistoryItem}
        onClose={() => setSelectedHistoryItem(null)}
        selectedItem={selectedHistoryItem}
        warehouses={warehouses}
      />

      <AddWarehouseDialog
        isOpen={isAddWarehouseOpen}
        onClose={() => setIsAddWarehouseOpen(false)}
        onWarehouseCreated={handleWarehouseCreated}
        onSubmitCreate={createWarehouse}
      />
    </div>
  );
}
