"use client"

import * as React from "react"
import Link from "next/link"
import { Menu, Package, Truck, User } from "lucide-react"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"

export function StorefrontMobileNav({ shopName }: { shopName: string }) {
  const [open, setOpen] = React.useState(false)

  return (
    <div className="md:hidden flex items-center">
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger className="p-2 rounded-xl text-zinc-700 dark:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors focus:outline-none">
          <Menu className="w-5 h-5" />
          <span className="sr-only">Toggle mobile menu</span>
        </SheetTrigger>
        <SheetContent side="left" className="w-[300px] bg-white dark:bg-zinc-950 p-6 flex flex-col gap-6 border-r border-zinc-200 dark:border-zinc-800">
          <SheetHeader className="text-left border-b border-zinc-200 dark:border-zinc-800/80 pb-4">
            <SheetTitle className="text-lg font-bold tracking-tight text-zinc-900 dark:text-zinc-50 truncate">
              {shopName}
            </SheetTitle>
          </SheetHeader>
          <nav className="flex flex-col gap-2">
            <Link
              href="/"
              onClick={() => setOpen(false)}
              className="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-semibold text-zinc-800 dark:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-900 transition-colors"
            >
              <Package className="w-4 h-4 text-emerald-500" />
              Products
            </Link>
            <Link
              href="/track"
              onClick={() => setOpen(false)}
              className="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-semibold text-zinc-800 dark:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-900 transition-colors"
            >
              <Truck className="w-4 h-4 text-emerald-500" />
              Track Order
            </Link>
            <Link
              href="/account"
              onClick={() => setOpen(false)}
              className="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-semibold text-zinc-800 dark:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-900 transition-colors"
            >
              <User className="w-4 h-4 text-emerald-500" />
              Account
            </Link>
          </nav>
        </SheetContent>
      </Sheet>
    </div>
  )
}
