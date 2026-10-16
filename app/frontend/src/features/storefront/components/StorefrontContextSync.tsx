"use client"

import * as React from "react"
import { useCartStore } from "@/store/cart"

interface StorefrontContextSyncProps {
  shopId?: string;
  shopName?: string;
}

export function StorefrontContextSync({ shopId, shopName }: StorefrontContextSyncProps) {
  const fetchCart = useCartStore((state) => state.fetchCart);

  React.useEffect(() => {
    if (typeof window !== "undefined") {
      if (shopId) {
        localStorage.setItem("nexus_shop_id", shopId);
      }
      if (shopName) {
        localStorage.setItem("nexus_shop_name", shopName);
      }
      fetchCart();
    }
  }, [shopId, shopName, fetchCart]);

  return null;
}
