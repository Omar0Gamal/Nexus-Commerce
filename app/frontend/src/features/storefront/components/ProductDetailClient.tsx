"use client"

import * as React from "react"
import Image from "next/image"
import Link from "next/link"
import {
  ShieldCheck,
  Truck,
  RefreshCw,
  ShoppingBag,
  Check,
  ChevronRight,
  Minus,
  Plus,
  Star,
  Sparkles,
  ArrowLeft
} from "lucide-react"
import { useCartStore } from "@/store/cart"
import { WishlistButton } from "./WishlistButton"

interface ProductDetailClientProps {
  product: any
  shop: any
}

export function ProductDetailClient({ product, shop }: ProductDetailClientProps) {
  const addItem = useCartStore((state) => state.addItem)

  const imageList: string[] = React.useMemo(() => {
    if (product.images && Array.isArray(product.images) && product.images.length > 0) {
      return product.images.map((im: any) => im.url).filter(Boolean)
    }
    if (product.image_url) return [product.image_url]
    return ["https://images.unsplash.com/photo-1552346154-21d32810aba3?auto=format&fit=crop&w=800&q=80"]
  }, [product])

  const [activeImage, setActiveImage] = React.useState<string>(imageList[0] || "")

  const variants: any[] = React.useMemo(() => {
    if (product.variants && Array.isArray(product.variants) && product.variants.length > 0) {
      return product.variants
    }
    return [
      { id: "v1", title: "US 8.5", sku: `${product.sku || 'SKU'}-85`, price: product.price, inventory_quantity: 4 },
      { id: "v2", title: "US 9", sku: `${product.sku || 'SKU'}-90`, price: product.price, inventory_quantity: 6 },
      { id: "v3", title: "US 9.5", sku: `${product.sku || 'SKU'}-95`, price: product.price, inventory_quantity: 5 },
      { id: "v4", title: "US 10", sku: `${product.sku || 'SKU'}-100`, price: product.price, inventory_quantity: 8 },
      { id: "v5", title: "US 10.5", sku: `${product.sku || 'SKU'}-105`, price: product.price, inventory_quantity: 3 },
      { id: "v6", title: "US 11", sku: `${product.sku || 'SKU'}-110`, price: product.price, inventory_quantity: 2 },
    ]
  }, [product])

  const [selectedVariant, setSelectedVariant] = React.useState<any>(variants[0] || null)
  const [quantity, setQuantity] = React.useState(1)
  const [isAdding, setIsAdding] = React.useState(false)
  const [justAdded, setJustAdded] = React.useState(false)

  const activePrice = Number(selectedVariant?.price ?? product.price ?? 0)
  const compareAtPrice = Number(product.compare_at_price ?? 0)
  const hasDiscount = compareAtPrice > activePrice

  const stock = Number(
    selectedVariant?.inventory_quantity ??
    selectedVariant?.stock_quantity ??
    product.stock_quantity ??
    product.inventory_count ??
    8
  )

  const handleAddToCart = async () => {
    setIsAdding(true)
    try {
      await addItem({
        id: product.id,
        productId: product.id,
        variantId: selectedVariant?.id,
        title: selectedVariant?.title
          ? `${product.title} (${selectedVariant.title})`
          : product.title,
        price: activePrice,
        quantity: quantity,
        image: activeImage || imageList[0] || "",
      })
      setJustAdded(true)
      setTimeout(() => setJustAdded(false), 2000)
    } finally {
      setIsAdding(false)
    }
  }

  return (
    <div className="flex flex-col gap-10">
      <nav className="flex items-center gap-2 text-xs font-medium text-zinc-500">
        <Link href="/" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors flex items-center gap-1">
          <ArrowLeft className="size-3.5" />
          <span>Home</span>
        </Link>
        <ChevronRight className="size-3 text-zinc-400" />
        <Link href="/#catalog" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
          Sneakers
        </Link>
        <ChevronRight className="size-3 text-zinc-400" />
        <span className="text-zinc-900 dark:text-zinc-200 truncate max-w-xs">{product.title}</span>
      </nav>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16 items-start">
        <div className="lg:col-span-7 flex flex-col gap-4">
          <div className="relative aspect-[4/3] w-full rounded-3xl overflow-hidden bg-zinc-100 dark:bg-zinc-900/60 border border-zinc-200 dark:border-white/[0.08] shadow-lg">
            {activeImage ? (
              <Image
                src={activeImage}
                alt={product.title}
                fill
                priority
                className="object-cover object-center transition-all duration-500"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-zinc-500 text-sm">
                No Photo Available
              </div>
            )}

            <div className="absolute top-4 left-4 inline-flex items-center gap-1.5 rounded-full bg-zinc-950/80 backdrop-blur-md border border-white/10 px-3 py-1 text-xs font-medium text-zinc-200">
              <ShieldCheck className="size-3.5 text-emerald-400" />
              <span>Physically Authenticated</span>
            </div>
          </div>

          {imageList.length > 1 && (
            <div className="flex items-center gap-3 overflow-x-auto pb-2">
              {imageList.map((img, idx) => (
                <button
                  key={idx}
                  type="button"
                  onClick={() => setActiveImage(img)}
                  className={`relative size-20 rounded-2xl overflow-hidden border-2 transition-all shrink-0 cursor-pointer ${
                    activeImage === img
                      ? "border-emerald-500 shadow-md scale-105"
                      : "border-zinc-200 dark:border-zinc-800 opacity-60 hover:opacity-100"
                  }`}
                >
                  <Image src={img} alt={`View ${idx + 1}`} fill className="object-cover" />
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="lg:col-span-5 flex flex-col gap-6">
          <div className="flex flex-col gap-2">
            <span className="text-xs font-mono font-semibold uppercase tracking-widest text-emerald-500">
              {product.sku || "Exclusive Drop"}
            </span>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50 leading-tight">
              {product.title}
            </h1>
            
            <div className="flex items-baseline gap-3 mt-1">
              <span className="text-2xl font-bold font-mono text-zinc-900 dark:text-zinc-50">
                EGP {activePrice.toLocaleString()}
              </span>
              {hasDiscount && (
                <span className="text-sm font-mono text-zinc-400 line-through">
                  EGP {compareAtPrice.toLocaleString()}
                </span>
              )}
              {hasDiscount && (
                <span className="rounded-full bg-red-500/10 text-red-500 border border-red-500/20 px-2 py-0.5 text-xs font-semibold">
                  Save {Math.round(((compareAtPrice - activePrice) / compareAtPrice) * 100)}%
                </span>
              )}
            </div>
            <span className="text-[11px] text-zinc-500">
              All import customs, VAT & taxes included. No hidden fees at checkout.
            </span>
          </div>

          <div className="h-px bg-zinc-200 dark:bg-zinc-800" />

          <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold uppercase tracking-wider text-zinc-700 dark:text-zinc-300">
                Select Size: <strong className="text-zinc-900 dark:text-zinc-100">{selectedVariant?.title || "Standard"}</strong>
              </span>
              <span className="text-xs text-zinc-500 font-mono">True to Size (TTS)</span>
            </div>

            <div className="grid grid-cols-3 sm:grid-cols-4 gap-2">
              {variants.map((v) => {
                const isSelected = selectedVariant?.id === v.id
                const vStock = v.inventory_quantity ?? v.stock_quantity ?? 5
                const isOutOfStock = vStock === 0

                return (
                  <button
                    key={v.id}
                    type="button"
                    disabled={isOutOfStock}
                    onClick={() => setSelectedVariant(v)}
                    className={`h-11 rounded-xl border text-xs font-medium transition-all flex flex-col items-center justify-center cursor-pointer ${
                      isSelected
                        ? "border-emerald-500 bg-emerald-500/10 text-emerald-400 shadow-sm"
                        : isOutOfStock
                        ? "border-zinc-200 dark:border-zinc-800/60 opacity-30 cursor-not-allowed line-through text-zinc-400"
                        : "border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 text-zinc-800 dark:text-zinc-200 hover:border-zinc-400 dark:hover:border-zinc-700"
                    }`}
                  >
                    <span>{v.title}</span>
                  </button>
                )
              })}
            </div>
          </div>

          <div className="flex items-center gap-2 text-xs">
            {stock > 4 ? (
              <div className="flex items-center gap-2 text-emerald-600 dark:text-emerald-400 font-medium">
                <span className="size-2 rounded-full bg-emerald-500 animate-pulse" />
                <span>In Stock — Dispatched within 24h across Greater Cairo</span>
              </div>
            ) : stock > 0 ? (
              <div className="flex items-center gap-2 text-amber-600 dark:text-amber-400 font-medium">
                <span className="size-2 rounded-full bg-amber-500 animate-pulse" />
                <span>Low Stock: Only {stock} pairs remaining in Cairo vault</span>
              </div>
            ) : (
              <div className="flex items-center gap-2 text-red-500 font-medium">
                <span>Currently Sold Out</span>
              </div>
            )}
          </div>

          <div className="flex items-center gap-3">
            <div className="flex items-center border border-zinc-200 dark:border-zinc-800 rounded-2xl bg-zinc-100 dark:bg-zinc-900 p-1">
              <button
                type="button"
                onClick={() => setQuantity(q => Math.max(1, q - 1))}
                className="size-9 rounded-xl flex items-center justify-center text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors"
              >
                <Minus className="size-3.5" />
              </button>
              <span className="w-9 text-center text-xs font-mono font-bold text-zinc-900 dark:text-zinc-100">
                {quantity}
              </span>
              <button
                type="button"
                onClick={() => setQuantity(q => q + 1)}
                className="size-9 rounded-xl flex items-center justify-center text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors"
              >
                <Plus className="size-3.5" />
              </button>
            </div>

            <button
              type="button"
              onClick={handleAddToCart}
              disabled={isAdding || stock === 0}
              className="flex-1 h-12 rounded-2xl bg-zinc-900 hover:bg-zinc-800 dark:bg-zinc-100 dark:hover:bg-white text-zinc-50 dark:text-zinc-950 font-semibold text-sm transition-all hover:scale-[1.01] active:scale-[0.99] flex items-center justify-center gap-2 shadow-lg disabled:opacity-50 cursor-pointer"
            >
              {justAdded ? (
                <>
                  <Check className="size-4 text-emerald-500" />
                  <span>Added to Shopping Bag!</span>
                </>
              ) : (
                <>
                  <ShoppingBag className="size-4 opacity-70" />
                  <span>Add to Bag — EGP {(activePrice * quantity).toLocaleString()}</span>
                </>
              )}
            </button>

            <WishlistButton productId={product.id} shopId={shop?.shop_id} />
          </div>

          <div className="flex flex-col gap-2 pt-2">
            <span className="text-xs font-semibold uppercase tracking-wider text-zinc-500">
              Collector Details & Craftsmanship
            </span>
            <p className="text-xs text-zinc-600 dark:text-zinc-400 leading-relaxed font-sans">
              {product.description || "The definitive release for modern collectors, manufactured with premium full-grain leather, archival color blocking, and authenticated construction."}
            </p>
          </div>

          <div className="rounded-2xl border border-zinc-200 dark:border-white/[0.06] bg-zinc-50 dark:bg-zinc-900/40 divide-y divide-zinc-200 dark:divide-white/[0.06] overflow-hidden text-xs">
            <div className="p-3.5 flex items-start gap-3">
              <ShieldCheck className="size-4 text-emerald-500 shrink-0 mt-0.5" />
              <div className="flex flex-col gap-0.5">
                <span className="font-semibold text-zinc-900 dark:text-zinc-100">100% Authenticity Verified</span>
                <span className="text-zinc-500">Includes NFC verification tag and serialized invoice.</span>
              </div>
            </div>
            <div className="p-3.5 flex items-start gap-3">
              <Truck className="size-4 text-blue-500 shrink-0 mt-0.5" />
              <div className="flex flex-col gap-0.5">
                <span className="font-semibold text-zinc-900 dark:text-zinc-100">Express Cairo Delivery</span>
                <span className="text-zinc-500">Same-day courier across Cairo & Giza. 48h to Alexandria.</span>
              </div>
            </div>
            <div className="p-3.5 flex items-start gap-3">
              <RefreshCw className="size-4 text-purple-500 shrink-0 mt-0.5" />
              <div className="flex flex-col gap-0.5">
                <span className="font-semibold text-zinc-900 dark:text-zinc-100">14-Day Size Exchange</span>
                <span className="text-zinc-500">Need a different size? We swap it at your door with zero courier hassle.</span>
              </div>
            </div>
          </div>

        </div>

      </div>
    </div>
  )
}
