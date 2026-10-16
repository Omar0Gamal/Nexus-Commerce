import { Metadata } from "next"
import Link from "next/link"
import { AnalyticsTracker } from "@/features/analytics/components/AnalyticsTracker"
import { CartDrawer } from "@/features/storefront/components/CartDrawer"
import { StorefrontContextSync } from "@/features/storefront/components/StorefrontContextSync"
import { StorefrontMobileNav } from "@/features/storefront/components/StorefrontMobileNav"

async function getShopContext(domain: string) {
  try {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"
    const shopResponse = await fetch(`${apiUrl}/api/v1/storefront/shop/${domain}`, {
      next: { revalidate: 60 },
    })
    if (shopResponse.ok) {
      const shopPayload = await shopResponse.json()
      return shopPayload.data || shopPayload
    }
  } catch (err) {}
  return null
}

export async function generateMetadata({ params }: { params: Promise<{ domain: string }> }): Promise<Metadata> {
  const resolvedParams = await params;
  const domain = decodeURIComponent(resolvedParams.domain);
  const shop = await getShopContext(domain);
  return {
    title: shop?.name ? `${shop.name} | Storefront` : `Storefront | ${domain}`,
    description: "Powered by Nexus Commerce",
  }
}

export default async function StorefrontLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ domain: string }>
}) {
  const resolvedParams = await params;
  const domain = decodeURIComponent(resolvedParams.domain)
  const shop = await getShopContext(domain)

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 font-sans text-zinc-900 dark:text-zinc-50 selection:bg-zinc-900 selection:text-zinc-50 dark:selection:bg-zinc-50 dark:selection:text-zinc-900">
      <AnalyticsTracker />
      <StorefrontContextSync shopId={shop?.shop_id} shopName={shop?.name} />

      <header className="sticky top-0 z-50 w-full border-b border-zinc-200 dark:border-zinc-800/50 bg-white/80 dark:bg-zinc-950/80 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8 gap-4">
          <div className="flex items-center gap-2 sm:gap-3 min-w-0">
            <StorefrontMobileNav shopName={shop?.name || domain} />
            <Link href="/" className="font-bold tracking-tight text-base sm:text-lg hover:opacity-80 transition-opacity truncate whitespace-nowrap">
              {shop?.name || domain}
            </Link>
          </div>
          <nav className="flex items-center gap-3 sm:gap-6 shrink-0">
            <div className="hidden md:flex items-center gap-6">
              <Link href="/" className="text-sm font-medium text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-50 transition-colors">
                Products
              </Link>
              <Link href="/track" className="text-sm font-medium text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-50 transition-colors">
                Track Order
              </Link>
              <Link href="/account" className="text-sm font-medium text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-50 transition-colors">
                Account
              </Link>
            </div>
            <CartDrawer />
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12 md:py-24">
        {children}
      </main>

      <footer className="border-t border-zinc-200 dark:border-zinc-800/50 mt-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12 flex flex-col items-center justify-center text-center">
          <p className="text-sm text-zinc-500 dark:text-zinc-500">
            &copy; {new Date().getFullYear()} {shop?.name || domain}. All rights reserved.
          </p>
          <a href="https://egyptcommerce.com" className="mt-4 text-xs text-zinc-400 dark:text-zinc-600 hover:underline">
            Powered by Nexus Commerce
          </a>
        </div>
      </footer>
    </div>
  )
}
