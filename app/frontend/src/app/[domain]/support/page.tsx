import Link from "next/link"
import { ArrowLeft, Search, BookOpen, Truck, RotateCcw, ShieldQuestion } from "lucide-react"

export default function SupportPage() {
  return (
    <div className="flex flex-col min-h-screen bg-zinc-50 dark:bg-zinc-950">
      <header className="flex items-center justify-between px-6 md:px-12 h-20 border-b border-zinc-200 dark:border-white/[0.05]">
        <Link href={`/`} className="flex items-center gap-2 text-sm font-medium text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-50 transition-colors">
          <ArrowLeft className="w-4 h-4" />
          Back to Store
        </Link>
      </header>

      <main className="flex-1 max-w-5xl mx-auto w-full px-6 py-12 md:py-24">
        <div className="text-center mb-16">
          <h1 className="text-4xl md:text-5xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50 mb-6">
            How can we help?
          </h1>
          <div className="max-w-2xl mx-auto relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-zinc-400" />
            <input 
              type="text" 
              placeholder="Search for articles, tracking, or return policies..."
              className="w-full h-14 pl-12 pr-4 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 text-lg focus:outline-none focus:ring-2 focus:ring-zinc-900 dark:focus:ring-zinc-100 transition-all shadow-sm"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-16">
          <div className="p-6 rounded-2xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors cursor-pointer">
            <Truck className="w-8 h-8 text-zinc-900 dark:text-zinc-100 mb-4" />
            <h3 className="text-lg font-semibold mb-2">Shipping & Delivery</h3>
            <p className="text-zinc-500 text-sm">Track your order, shipping times, and delivery policies.</p>
          </div>
          <div className="p-6 rounded-2xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors cursor-pointer">
            <RotateCcw className="w-8 h-8 text-zinc-900 dark:text-zinc-100 mb-4" />
            <h3 className="text-lg font-semibold mb-2">Returns & Refunds</h3>
            <p className="text-zinc-500 text-sm">How to return an item, refund processing times.</p>
          </div>
          <div className="p-6 rounded-2xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors cursor-pointer">
            <ShieldQuestion className="w-8 h-8 text-zinc-900 dark:text-zinc-100 mb-4" />
            <h3 className="text-lg font-semibold mb-2">FAQs</h3>
            <p className="text-zinc-500 text-sm">Common questions about products, accounts, and more.</p>
          </div>
        </div>

        <div className="border-t border-zinc-200 dark:border-zinc-800 pt-12">
          <h2 className="text-2xl font-bold tracking-tight mb-8">Popular Articles</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Link href="#" className="flex items-start gap-4 p-4 rounded-xl hover:bg-zinc-100 dark:hover:bg-zinc-900 transition-colors group">
              <BookOpen className="w-5 h-5 text-zinc-400 group-hover:text-zinc-900 dark:group-hover:text-zinc-100 shrink-0 mt-0.5" />
              <div>
                <h4 className="font-medium text-zinc-900 dark:text-zinc-100">How long does shipping take?</h4>
                <p className="text-sm text-zinc-500 line-clamp-2 mt-1">Standard shipping takes 3-5 business days. Express shipping is available for 1-2 day delivery.</p>
              </div>
            </Link>
            <Link href="#" className="flex items-start gap-4 p-4 rounded-xl hover:bg-zinc-100 dark:hover:bg-zinc-900 transition-colors group">
              <BookOpen className="w-5 h-5 text-zinc-400 group-hover:text-zinc-900 dark:group-hover:text-zinc-100 shrink-0 mt-0.5" />
              <div>
                <h4 className="font-medium text-zinc-900 dark:text-zinc-100">What is your return policy?</h4>
                <p className="text-sm text-zinc-500 line-clamp-2 mt-1">We accept returns within 30 days of purchase for a full refund, provided items are unworn and in original packaging.</p>
              </div>
            </Link>
          </div>
        </div>
      </main>
    </div>
  )
}
