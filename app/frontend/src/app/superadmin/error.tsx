"use client" // Error components must be Client Components

import { useEffect } from "react"
import { AlertCircle, RotateCcw } from "lucide-react"
import { Button } from "@/components/ui/button"

export default function SuperadminError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    // Log the error to an error reporting service like Sentry here
    console.error("Superadmin Dashboard Error:", error)
  }, [error])

  return (
    <div className="flex h-[calc(100vh-100px)] w-full flex-col items-center justify-center p-8 text-center bg-background">
      <div className="flex size-20 items-center justify-center rounded-full bg-destructive/10 text-destructive mb-6 shadow-sm ring-1 ring-destructive/20">
        <AlertCircle className="size-10" />
      </div>
      <h2 className="text-2xl font-bold tracking-tight text-foreground mb-3">Something went wrong!</h2>
      <p className="text-muted-foreground mb-8 max-w-md mx-auto leading-relaxed">
        We encountered an unexpected error while trying to load the dashboard. Our engineering team has been notified.
      </p>
      <Button 
        onClick={() => reset()}
        className="gap-2 shadow-sm"
        size="lg"
      >
        <RotateCcw className="size-4" />
        Try Again
      </Button>
    </div>
  )
}
