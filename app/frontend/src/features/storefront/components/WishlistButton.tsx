"use client"
import { useState } from "react"
import { Heart } from "lucide-react"
import { useCustomerAuthStore } from "@/store/useCustomerAuthStore"
import { useQueryClient } from "@tanstack/react-query"

export function WishlistButton({ productId, shopId }: { productId: string, shopId: string }) {
  const { customerToken } = useCustomerAuthStore()
  const [wishlisted, setWishlisted] = useState(false)
  const [loading, setLoading] = useState(false)
  const queryClient = useQueryClient()
  
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || ''

  const toggle = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!customerToken) return
    setLoading(true)
    try {
      if (wishlisted) {
        await fetch(`${apiUrl}/api/v1/wishlist/items/${productId}`, {
          method: 'DELETE',
          headers: { 'Authorization': `Bearer ${customerToken}`, 'X-Shop-ID': shopId }
        })
        setWishlisted(false)
      } else {
        await fetch(`${apiUrl}/api/v1/wishlist/items`, {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${customerToken}`, 'X-Shop-ID': shopId, 'Content-Type': 'application/json' },
          body: JSON.stringify({ product_id: productId })
        })
        setWishlisted(true)
      }
      queryClient.invalidateQueries({ queryKey: ["wishlist"] })
    } catch (e) {}
    setLoading(false)
  }
  
  if (!customerToken) return null
  
  return (
    <button
      onClick={toggle}
      disabled={loading}
      className={`p-3 rounded-full transition-all shadow-sm backdrop-blur-md border ${
        wishlisted 
          ? 'bg-red-500/10 text-red-500 border-red-500/20' 
          : 'bg-white/80 dark:bg-black/50 text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white border-zinc-200 dark:border-white/10 hover:border-zinc-300'
      }`}
      aria-label={wishlisted ? 'Remove from wishlist' : 'Add to wishlist'}
    >
      <Heart className={`w-4 h-4 ${wishlisted ? 'fill-current' : ''}`} />
    </button>
  )
}
