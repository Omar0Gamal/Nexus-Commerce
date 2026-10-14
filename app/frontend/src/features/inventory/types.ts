export interface Warehouse {
  id: string;
  shop_id: string;
  name: string;
  address?: string | { String?: string; Valid?: boolean } | null;
  is_default: boolean;
  is_active: boolean;
  created_at: string;
}

export interface InventoryLevel {
  id: string;
  shop_id: string;
  variant_id: string;
  warehouse_id: string;
  quantity_on_hand: number;
  quantity_reserved: number;
  quantity_incoming: number;
  low_stock_threshold: number;
  updated_at: string;
}

export type MovementType =
  | 'restock'
  | 'adjustment'
  | 'transfer'
  | 'return'
  | 'sale'
  | 'pre_order_reserve';

export interface StockMovement {
  id: string;
  shop_id: string;
  variant_id: string;
  warehouse_id?: string | { Bytes?: string; Valid?: boolean } | null;
  movement_type: MovementType;
  quantity_delta: number;
  reference_id?: string | { Bytes?: string; Valid?: boolean } | null;
  reference_type?: string | { String?: string; Valid?: boolean } | null;
  note?: string | { String?: string; Valid?: boolean } | null;
  created_by?: string | { String?: string; Valid?: boolean } | null;
  created_at: string;
}

export interface LowStockVariant {
  variant_id: string;
  quantity_on_hand: number;
  low_stock_threshold: number;
  warehouse_id: string;
}

export interface ReorderRule {
  id: string;
  shop_id: string;
  variant_id: string;
  reorder_point: number;
  reorder_qty: number;
  supplier_email?: string | { String?: string; Valid?: boolean } | null;
  is_active: boolean;
  last_triggered_at?: string | null;
}

export interface AdjustStockPayload {
  variant_id: string;
  warehouse_id: string;
  delta: number;
  movement_type: MovementType;
  reference_id?: string;
  reference_type?: string;
  note?: string;
}

export interface CreateWarehousePayload {
  name: string;
  address?: string;
  is_default?: boolean;
}

export interface InventoryItemRow {
  productId: string;
  variantId: string;
  productTitle: string;
  variantTitle: string;
  sku: string;
  price: string;
  imageUrl?: string;
  status: string;
  levels: Record<string, number>; // warehouseId -> quantityOnHand
  reservedLevels: Record<string, number>; // warehouseId -> quantityReserved
  totalOnHand: number;
  totalReserved: number;
  totalAvailable: number;
  lowStockThreshold: number;
  isLowStock: boolean;
  isOutOfStock: boolean;
}
