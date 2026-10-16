import { env } from "@/env"
import type { Metadata } from "next"
import { ProductDetailClient } from "@/features/storefront/components/ProductDetailClient"

export async function generateMetadata({ params }: { params: Promise<{ domain: string; id: string }> }): Promise<Metadata> {
  const { domain: rawDomain, id } = await params
  const domain = decodeURIComponent(rawDomain)
  const shop = await getShopContext(domain)
  if (!shop) return { title: "Product Not Found" }

  const product = await getProduct(shop.shop_id, id)
  if (!product) return { title: "Product Not Found" }

  return {
    title: `${product.title} | ${shop.name}`,
    description: product.description || "Authenticated luxury footwear and grails.",
    openGraph: {
      title: `${product.title} | ${shop.name}`,
      description: product.description,
      images: product.images?.[0]?.url ? [product.images[0].url] : [],
    },
  }
}

async function getProduct(shopId: string, productId: string) {
  try {
    const productResponse = await fetch(`${env.NEXT_PUBLIC_API_URL}/api/v1/catalog/products/${productId}`, {
      headers: { "X-Shop-ID": shopId },
      next: { revalidate: 60 }
    })
    if (productResponse.ok) {
      const productPayload = await productResponse.json()
      return productPayload.data
    }
  } catch (err) {}
  
  return null
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
  } catch (err) {}
  
  return null
}

export default async function ProductPage({ params }: { params: Promise<{ domain: string; id: string }> }) {
  const { domain: rawDomain, id } = await params
  const domain = decodeURIComponent(rawDomain)
  const shop = await getShopContext(domain)
  
  if (!shop) {
    return (
      <div className="py-20 text-center text-zinc-500">
        Shop not found for domain {domain}.
      </div>
    )
  }

  const product = await getProduct(shop.shop_id, id)
  
  if (!product) {
    return (
      <div className="py-20 text-center text-zinc-500">
        Product not found in this catalog.
      </div>
    )
  }

  return (
    <div className="animate-in fade-in duration-500 pb-16">
      <ProductDetailClient product={product} shop={shop} />
    </div>
  )
}