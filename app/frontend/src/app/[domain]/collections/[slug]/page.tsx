import React from 'react';
import Link from 'next/link';
import { env } from "@/env";
import Image from "next/image";

async function getShopContext(domain: string) {
  try {
    const shopResponse = await fetch(`${env.NEXT_PUBLIC_API_URL}/api/v1/storefront/shop/${domain}`, {
      next: { revalidate: 60 }
    })
    if (shopResponse.ok) {
      const shopPayload = await shopResponse.json()
      return shopPayload.data || shopPayload
    }
  } catch (err) {}
  return null
}

async function getProducts(shopId: string) {
  try {
    const productsResponse = await fetch(`${env.NEXT_PUBLIC_API_URL}/api/v1/catalog/products`, {
      headers: { "X-Shop-ID": shopId },
      next: { revalidate: 60 }
    })
    if (productsResponse.ok) {
      const productsPayload = await productsResponse.json()
      return productsPayload.data || []
    }
  } catch (err) {}
  return []
}

async function getCategories(shopId: string) {
  try {
    const categoriesResponse = await fetch(`${env.NEXT_PUBLIC_API_URL}/api/v1/catalog/categories`, {
      headers: { "X-Shop-ID": shopId },
      next: { revalidate: 60 }
    })
    if (categoriesResponse.ok) {
      const categoriesPayload = await categoriesResponse.json()
      return categoriesPayload.data || []
    }
  } catch (err) {}
  return []
}

export default async function CollectionPage({ params }: { params: Promise<{ domain: string; slug: string }> }) {
  const { domain: rawDomain, slug } = await params
  const domain = decodeURIComponent(rawDomain)
  
  const shop = await getShopContext(domain);
  if (!shop) {
    return <div className="p-12 text-center text-zinc-500">Shop not found</div>
  }

  const [allProducts, categories] = await Promise.all([
    getProducts(shop.shop_id),
    getCategories(shop.shop_id)
  ]);

  const categoryMap = new Map<string, any>(categories.map((cat: any) => [cat.id, cat]));

  const activeCategory = categories.find((cat: any) => 
    cat.slug?.toLowerCase() === slug.toLowerCase() || cat.name?.toLowerCase() === slug.toLowerCase()
  );

  const displayProducts = allProducts.filter((product: any) => {
    if (slug === 'all') return true;
    if (activeCategory && product.category_id === activeCategory.id) return true;
    if (product.category?.name?.toLowerCase() === slug.toLowerCase() || product.category?.slug?.toLowerCase() === slug.toLowerCase()) return true;
    return false;
  });

  const categoryName = slug === 'all'
    ? 'All Footwear & Essentials'
    : (activeCategory?.name || (slug ? slug.charAt(0).toUpperCase() + slug.slice(1).replace('-', ' ') : 'Collection'));

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12 text-zinc-900 dark:text-zinc-50">
      <div className="flex flex-col md:flex-row gap-10 lg:gap-12">
        <aside className="w-full md:w-64 flex-shrink-0">
          <div className="rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-white/60 dark:bg-zinc-900/40 p-6 backdrop-blur-md">
            <h2 className="text-base font-bold text-zinc-900 dark:text-zinc-100 mb-6">Collections</h2>
            
            <div className="mb-6">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-400 mb-3">Categories</h3>
              <ul className="space-y-2">
                <li>
                  <Link
                    href="/collections/all"
                    className={`text-sm block py-1.5 px-3 rounded-lg transition-colors ${
                      slug === 'all'
                        ? 'font-bold bg-zinc-900 text-white dark:bg-white dark:text-zinc-950 shadow-sm'
                        : 'text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                    }`}
                  >
                    All Products
                  </Link>
                </li>
                {categories.map((cat: any) => {
                  const isActive = slug.toLowerCase() === cat.slug.toLowerCase();
                  return (
                    <li key={cat.id}>
                      <Link
                        href={`/collections/${cat.slug}`}
                        className={`text-sm block py-1.5 px-3 rounded-lg transition-colors ${
                          isActive
                            ? 'font-bold bg-zinc-900 text-white dark:bg-white dark:text-zinc-950 shadow-sm'
                            : 'text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                        }`}
                      >
                        {cat.name}
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </div>
          </div>
        </aside>

        <div className="flex-1 min-w-0">
          <div className="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4 pb-6 border-b border-zinc-200 dark:border-zinc-800">
            <div>
              <h1 className="text-3xl font-extrabold tracking-tight text-zinc-900 dark:text-zinc-50">{categoryName}</h1>
              <p className="mt-1.5 text-sm text-zinc-600 dark:text-zinc-400">
                Showing {displayProducts.length} authenticated items
              </p>
            </div>
            <div>
              <select className="border border-zinc-200 dark:border-zinc-800 text-zinc-900 dark:text-zinc-100 bg-white dark:bg-zinc-900 rounded-xl text-xs font-medium focus:ring-1 focus:ring-zinc-700 py-2 pl-3 pr-8 shadow-sm">
                <option>Sort by: Featured</option>
                <option>Price: Low to High</option>
                <option>Price: High to Low</option>
                <option>Newest Arrivals</option>
              </select>
            </div>
          </div>

          {displayProducts.length === 0 ? (
            <div className="py-24 text-center">
              <p className="text-zinc-500 text-base">No items found in this collection.</p>
              <Link href="/collections/all" className="mt-4 inline-block text-xs font-semibold underline text-zinc-900 dark:text-zinc-100">
                Browse all products
              </Link>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-y-10 gap-x-6">
              {displayProducts.map((product: any) => {
                const primaryImage = product.images?.[0]?.url || '/seed-prod-1.jpg';
                const prodCategory = categoryMap.get(product.category_id)?.name || activeCategory?.name || 'Footwear';
                const formattedPrice = new Intl.NumberFormat('en-EG', {
                  style: 'currency',
                  currency: 'EGP',
                  maximumFractionDigits: 0,
                }).format(Number(product.price || 0));

                return (
                  <Link href={`/product/${product.id}`} key={product.id} className="group flex flex-col">
                    <div className="aspect-[4/3] w-full overflow-hidden rounded-2xl bg-zinc-100 dark:bg-zinc-900 relative border border-zinc-200 dark:border-white/[0.08]">
                      <img
                        src={primaryImage}
                        alt={product.title || product.name}
                        className="h-full w-full object-cover object-center group-hover:scale-105 transition-transform duration-500 ease-out absolute inset-0"
                      />
                    </div>
                    <div className="mt-3.5 flex flex-col gap-1">
                      <div className="flex justify-between items-start gap-2">
                        <h3 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 group-hover:underline underline-offset-4 line-clamp-2 leading-snug">
                          {product.title || product.name}
                        </h3>
                        <p className="text-sm font-bold font-mono text-zinc-900 dark:text-zinc-100 flex-shrink-0">
                          {formattedPrice}
                        </p>
                      </div>
                      <p className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                        {prodCategory}
                      </p>
                    </div>
                  </Link>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
