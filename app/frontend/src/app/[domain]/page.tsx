import { env } from "@/env"
import Link from "next/link"
import Image from "next/image"
import type { Metadata } from "next"
import { ShieldCheck, Truck, CreditCard, ArrowRight, Sparkles, RefreshCw, Star } from "lucide-react"

export async function generateMetadata({ params }: { params: Promise<{ domain: string }> }): Promise<Metadata> {
  const { domain: rawDomain } = await params
  const domain = decodeURIComponent(rawDomain)
  const shop = await getShopContext(domain)

  if (!shop) {
    return { title: "Kicks & Co. Cairo | Luxury Sneaker Boutique" }
  }

  return {
    title: `${shop.name} | The Egyptian Sneaker Authority`,
    description: shop.storefront_config?.hero?.subheadline || "Authenticated Grails, Exclusive Releases & Iconic Streetwear Delivered Across Greater Cairo.",
    openGraph: {
      title: `${shop.name} | The Egyptian Sneaker Authority`,
      description: shop.storefront_config?.hero?.subheadline || "Authenticated Grails, Exclusive Releases & Iconic Streetwear Delivered Across Greater Cairo.",
      images: [
        shop.storefront_config?.hero?.bg_image ||
        "https://images.unsplash.com/photo-1552346154-21d32810aba3?auto=format&fit=crop&w=1600&q=80"
      ],
    },
  }
}

async function getShopContext(domain: string) {
  try {
    const shopResponse = await fetch(`${env.NEXT_PUBLIC_API_URL}/api/v1/storefront/shop/${domain}`, {
      next: { revalidate: 60 }
    })
    if (shopResponse.ok) {
      const shopPayload = await shopResponse.json()
      return shopPayload.data || shopPayload
    }
  } catch (err) {
    console.warn("Failed to fetch shop context")
  }
  return null
}

async function getProducts(shopId: string) {
  try {
    const productsResponse = await fetch(`${env.NEXT_PUBLIC_API_URL}/api/v1/catalog/products?per_page=12`, {
      headers: { "X-Shop-ID": shopId },
      next: { revalidate: 60 }
    })
    if (productsResponse.ok) {
      const productsPayload = await productsResponse.json()
      return productsPayload.data || []
    }
  } catch (err) {
    console.warn("Failed to fetch products")
  }
  return []
}

export default async function StorefrontPage({ params }: { params: Promise<{ domain: string }> }) {
  const { domain: rawDomain } = await params
  const domain = decodeURIComponent(rawDomain)
  const shop = await getShopContext(domain)

  const shopName = shop?.name || "Kicks & Co. Cairo"
  const shopId = shop?.shop_id || "11111111-aaaa-bbbb-cccc-000000000001"
  const config = shop?.storefront_config || {}

  const products = await getProducts(shopId)

  const heroImage = config.hero?.bg_image || "https://images.unsplash.com/photo-1552346154-21d32810aba3?auto=format&fit=crop&w=1600&q=80"
  const heroHeadline = config.hero?.headline || "The Egyptian Sneaker Authority"
  const heroSubheadline = config.hero?.subheadline || "Authenticated Grails, Exclusive Releases & Iconic Streetwear Delivered Across Greater Cairo & Alexandria."

  return (
    <div className="flex flex-col gap-16 md:gap-24 animate-in fade-in duration-700">
      <section className="relative flex flex-col items-center justify-center text-center max-w-7xl mx-auto w-full min-h-[540px] md:min-h-[600px] overflow-hidden rounded-[2.5rem] p-8 md:p-16 border border-zinc-200/80 dark:border-white/10 shadow-2xl">
        <div className="absolute inset-0 z-0">
          <Image
            src={heroImage}
            alt={heroHeadline}
            fill
            priority
            className="object-cover object-center brightness-[0.4] scale-105 transition-transform duration-1000"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-zinc-950 via-zinc-950/60 to-transparent" />
          <div className="absolute inset-0 bg-gradient-to-r from-zinc-950/80 via-transparent to-zinc-950/80" />
        </div>

        <div className="relative z-10 flex flex-col items-center max-w-3xl mx-auto">
          <div className="inline-flex items-center gap-2 rounded-full bg-white/10 backdrop-blur-md border border-white/20 px-3.5 py-1 text-xs font-semibold text-zinc-100 mb-6 tracking-wider uppercase">
            <Sparkles className="size-3.5 text-amber-400" />
            Autumn / Winter 2026 Drop
          </div>

          <h1 className="text-4xl sm:text-6xl md:text-7xl font-bold tracking-tight text-white mb-6 leading-[1.1] drop-shadow-md">
            {heroHeadline}
          </h1>

          <p className="text-base sm:text-lg md:text-xl text-zinc-300 mb-10 max-w-2xl leading-relaxed font-light">
            {heroSubheadline}
          </p>

          <div className="flex flex-col sm:flex-row items-center gap-4 w-full sm:w-auto">
            <a
              href="#catalog"
              className="w-full sm:w-auto flex items-center justify-center gap-2.5 h-12 px-8 rounded-full bg-zinc-100 hover:bg-white text-zinc-950 font-semibold text-sm transition-all hover:scale-105 active:scale-95 shadow-lg shadow-black/40"
            >
              <span>Explore Collection</span>
              <ArrowRight className="size-4" />
            </a>
            <div className="flex items-center gap-2 text-xs text-zinc-300 font-mono">
              <span className="size-2 rounded-full bg-emerald-400 animate-pulse" />
              <span>Same-Day Cairo Courier Active</span>
            </div>
          </div>
        </div>
      </section>

      <section className="max-w-7xl mx-auto w-full px-4">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 p-6 rounded-3xl bg-zinc-100 dark:bg-zinc-900/60 border border-zinc-200 dark:border-white/[0.06] backdrop-blur-sm">
          <div className="flex items-center gap-3.5">
            <div className="size-10 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400 shrink-0">
              <ShieldCheck className="size-5" />
            </div>
            <div className="flex flex-col">
              <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100">100% Authenticated</span>
              <span className="text-[11px] text-zinc-500">Every pair physical verified</span>
            </div>
          </div>

          <div className="flex items-center gap-3.5">
            <div className="size-10 rounded-2xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400 shrink-0">
              <Truck className="size-5" />
            </div>
            <div className="flex flex-col">
              <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100">Same-Day Dispatch</span>
              <span className="text-[11px] text-zinc-500">Cairo, Giza & Alex express</span>
            </div>
          </div>

          <div className="flex items-center gap-3.5">
            <div className="size-10 rounded-2xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400 shrink-0">
              <CreditCard className="size-5" />
            </div>
            <div className="flex flex-col">
              <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100">Flexible Payment</span>
              <span className="text-[11px] text-zinc-500">Cash on Delivery & Paymob</span>
            </div>
          </div>

          <div className="flex items-center gap-3.5">
            <div className="size-10 rounded-2xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400 shrink-0">
              <RefreshCw className="size-5" />
            </div>
            <div className="flex flex-col">
              <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100">14-Day Exchanges</span>
              <span className="text-[11px] text-zinc-500">Guaranteed size swap policy</span>
            </div>
          </div>
        </div>
      </section>

      <section id="catalog" className="max-w-7xl mx-auto w-full px-4 scroll-mt-24">
        <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4 mb-8">
          <div>
            <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-widest text-emerald-500 mb-1">
              <Star className="size-3.5 fill-current" /> Curated Inventory
            </div>
            <h2 className="text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
              Featured Releases
            </h2>
            <p className="text-sm text-zinc-500 mt-1">
              Rare grails, timeless Jordan silhouettes, and luxury street staples.
            </p>
          </div>

          <div className="flex items-center gap-2 overflow-x-auto pb-1 text-xs">
            <span className="rounded-full bg-zinc-900 dark:bg-zinc-100 text-zinc-100 dark:text-zinc-900 font-semibold px-3 py-1.5 cursor-pointer">
              All Shoes
            </span>
            <span className="rounded-full bg-zinc-200/60 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-300 font-medium px-3 py-1.5 cursor-pointer hover:bg-zinc-300 dark:hover:bg-zinc-700 transition-colors">
              Air Jordan
            </span>
            <span className="rounded-full bg-zinc-200/60 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-300 font-medium px-3 py-1.5 cursor-pointer hover:bg-zinc-300 dark:hover:bg-zinc-700 transition-colors">
              Yeezy
            </span>
            <span className="rounded-full bg-zinc-200/60 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-300 font-medium px-3 py-1.5 cursor-pointer hover:bg-zinc-300 dark:hover:bg-zinc-700 transition-colors">
              New Balance
            </span>
          </div>
        </div>

        {products.length === 0 ? (
          <div className="text-center py-20 text-zinc-500 border border-dashed border-zinc-200 dark:border-white/10 rounded-3xl">
            No products found in the catalog.
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 md:gap-8">
            {products.map((product: any) => {
              const primaryImg = product.images?.find((img: any) => img.is_primary)?.url || product.images?.[0]?.url || product.image_url
              const secondaryImg = product.images?.[1]?.url || primaryImg
              const price = Number(product.price || 0)
              const compareAt = Number(product.compare_at_price || 0)
              const hasDiscount = compareAt > price
              const stock = product.stock_quantity ?? 0

              return (
                <Link
                  href={`/product/${product.id}`}
                  key={product.id}
                  className="group relative flex flex-col rounded-3xl border border-zinc-200 dark:border-white/[0.08] bg-white dark:bg-zinc-900/40 hover:border-zinc-300 dark:hover:border-white/20 transition-all duration-300 overflow-hidden hover:shadow-xl hover:shadow-black/10"
                >
                  <div className="relative aspect-[4/3] w-full bg-zinc-100 dark:bg-zinc-950 overflow-hidden">
                    {primaryImg ? (
                      <>
                        <Image
                          src={primaryImg}
                          alt={product.title}
                          fill
                          className="object-cover object-center transition-all duration-700 group-hover:scale-105 group-hover:opacity-0"
                        />
                        <Image
                          src={secondaryImg}
                          alt={product.title}
                          fill
                          className="object-cover object-center opacity-0 transition-all duration-700 group-hover:scale-105 group-hover:opacity-100"
                        />
                      </>
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-xs text-zinc-500">
                        No Image Available
                      </div>
                    )}

                    <div className="absolute top-3 left-3 flex items-center gap-1.5">
                      {hasDiscount && (
                        <span className="rounded-full bg-red-500 px-2.5 py-0.5 text-[10px] font-bold text-white shadow-sm">
                          Save {Math.round(((compareAt - price) / compareAt) * 100)}%
                        </span>
                      )}
                      {stock > 0 && stock <= 4 && (
                        <span className="rounded-full bg-amber-500/90 backdrop-blur-sm px-2.5 py-0.5 text-[10px] font-bold text-zinc-950 shadow-sm">
                          Only {stock} Left
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="p-5 flex flex-col justify-between flex-1 gap-4">
                    <div>
                      <span className="text-[11px] font-mono uppercase tracking-wider text-zinc-400">
                        {product.sku || "Exclusive Drop"}
                      </span>
                      <h3 className="font-semibold text-zinc-900 dark:text-zinc-100 text-sm mt-0.5 line-clamp-1 group-hover:text-emerald-500 transition-colors">
                        {product.title}
                      </h3>
                      <p className="text-xs text-zinc-500 dark:text-zinc-400 mt-1 line-clamp-2 leading-relaxed">
                        {product.description || "Crafted with premium materials and signature design."}
                      </p>
                    </div>

                    <div className="flex items-center justify-between pt-3 border-t border-zinc-100 dark:border-white/[0.05]">
                      <div>
                        <div className="flex items-baseline gap-2">
                          <span className="text-base font-bold font-mono text-zinc-900 dark:text-zinc-50">
                            EGP {price.toLocaleString()}
                          </span>
                          {hasDiscount && (
                            <span className="text-xs font-mono text-zinc-400 line-through">
                              EGP {compareAt.toLocaleString()}
                            </span>
                          )}
                        </div>
                        <span className="text-[10px] text-zinc-400 block mt-0.5">
                          Taxes included • Free Cairo shipping
                        </span>
                      </div>

                      <div className="size-9 rounded-full bg-zinc-900 dark:bg-zinc-100 text-zinc-100 dark:text-zinc-900 flex items-center justify-center group-hover:scale-110 transition-transform">
                        <ArrowRight className="size-4" />
                      </div>
                    </div>
                  </div>
                </Link>
              )
            })}
          </div>
        )}
      </section>

      <section className="max-w-7xl mx-auto w-full px-4">
        <div className="rounded-[2.5rem] bg-zinc-900 text-zinc-100 p-8 md:p-16 border border-zinc-800 flex flex-col md:flex-row items-center justify-between gap-8">
          <div className="max-w-xl flex flex-col gap-3">
            <span className="text-xs font-mono uppercase tracking-widest text-emerald-400">
              The Cairo Flagship Experience
            </span>
            <h2 className="text-2xl md:text-4xl font-bold tracking-tight">
              Direct from Collector Vaults to Your Doorstep.
            </h2>
            <p className="text-sm text-zinc-400 leading-relaxed">
              Every sneaker passing through {shopName} undergoes rigorous multi-point verification: stitching tolerance, ultraviolet tag fluorescence, and manufacturer scent profiling. We take the anxiety out of buying rare footwear in Egypt.
            </p>
          </div>
          <div className="flex flex-col gap-3 shrink-0">
            <div className="flex items-center gap-3 bg-zinc-950/80 rounded-2xl p-4 border border-zinc-800">
              <span className="text-2xl font-bold font-mono text-emerald-400">100%</span>
              <div className="text-xs text-zinc-400">
                <span className="font-semibold text-zinc-200 block">Authenticity Guarantee</span>
                Or double your money back.
              </div>
            </div>
            <div className="flex items-center gap-3 bg-zinc-950/80 rounded-2xl p-4 border border-zinc-800">
              <span className="text-2xl font-bold font-mono text-blue-400">24h</span>
              <div className="text-xs text-zinc-400">
                <span className="font-semibold text-zinc-200 block">Cairo Rapid Dispatch</span>
                Delivered across all Cairo zones.
              </div>
            </div>
          </div>
        </div>
      </section>

    </div>
  )
}
