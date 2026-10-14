"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Switch } from "@/components/ui/switch"
import { useGetApiV1ShopsSettings, usePatchApiV1ShopsSettings } from "@/api/endpoints/auth/auth";

const storefrontConfigSchema = z.object({
  hero: z.object({
    enabled: z.boolean(),
    headline: z.string().optional(),
    subheadline: z.string().optional(),
    bg_image: z.string().optional(),
  }),
  featured_categories: z.object({
    enabled: z.boolean(),
    title: z.string().optional(),
  }),
  brand_story: z.object({
    enabled: z.boolean(),
    text: z.string().optional(),
  })
})

type StorefrontConfigValues = z.infer<typeof storefrontConfigSchema>

export function StorefrontEditorForm() {
  const { data: settingsEnvelope, isLoading: isFetching } = useGetApiV1ShopsSettings()
  const { mutate: updateSettings, isPending: isUpdating } = usePatchApiV1ShopsSettings()

  const [message, setMessage] = React.useState<{ type: 'success' | 'error', text: string } | null>(null)

  const settings = settingsEnvelope?.data
  
  const form = useForm<StorefrontConfigValues>({
    resolver: zodResolver(storefrontConfigSchema),
    defaultValues: {
      hero: {
        enabled: true,
        headline: "Welcome to our store",
        subheadline: "The best products in the world.",
        bg_image: "",
      },
      featured_categories: {
        enabled: true,
        title: "Shop by Category",
      },
      brand_story: {
        enabled: false,
        text: "We started this company to...",
      }
    },
  })

  React.useEffect(() => {
    if (settings && (settings as any).storefront_config) {
      const config = (settings as any).storefront_config as any
      form.reset({
        hero: {
          enabled: config.hero?.enabled ?? true,
          headline: config.hero?.headline || "",
          subheadline: config.hero?.subheadline || "",
          bg_image: config.hero?.bg_image || "",
        },
        featured_categories: {
          enabled: config.featured_categories?.enabled ?? true,
          title: config.featured_categories?.title || "",
        },
        brand_story: {
          enabled: config.brand_story?.enabled ?? false,
          text: config.brand_story?.text || "",
        }
      })
    }
  }, [settings, form])

  if (isFetching) {
    return <div className="text-zinc-500 animate-pulse text-sm">Loading editor...</div>
  }

  const onSubmit = (formValues: StorefrontConfigValues) => {
    setMessage(null)
    updateSettings({ data: { storefront_config: formValues } as any }, {
      onSuccess: () => {
        setMessage({ type: 'success', text: 'Storefront updated successfully.' })
      },
      onError: (err: any) => {
        setMessage({ type: 'error', text: err.response?.data?.error || 'Failed to update storefront.' })
      }
    })
  }

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="flex flex-col gap-8 max-w-2xl">
      {message && (
        <Alert variant={message.type === 'error' ? 'destructive' : 'default'} className={`rounded-xl border ${message.type === 'error' ? 'bg-red-500/10 border-red-500/20 text-red-400' : 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400'}`}>
          <AlertDescription>{message.text}</AlertDescription>
        </Alert>
      )}

      <div className="flex flex-col gap-6 p-6 rounded-2xl bg-zinc-900/50 border border-white/[0.05] shadow-[inset_0_1px_1px_rgba(255,255,255,0.02)]">
        <div className="flex items-center justify-between">
          <div className="flex flex-col gap-1">
            <h2 className="text-lg font-medium text-zinc-100">Hero Section</h2>
            <p className="text-sm text-zinc-500">The main banner at the top of your store.</p>
          </div>
          <div className="flex items-center space-x-2">
            <Switch 
              id="hero.enabled" 
              checked={form.watch("hero.enabled")} 
              onCheckedChange={(enabled) => form.setValue("hero.enabled", enabled)} 
            />
          </div>
        </div>
        
        {form.watch("hero.enabled") && (
          <div className="flex flex-col gap-4 mt-2">
            <div className="flex flex-col gap-2">
              <Label htmlFor="hero.headline" className="text-sm font-medium text-zinc-400">Headline</Label>
              <Input
                id="hero.headline"
                {...form.register("hero.headline")}
                className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 transition-all"
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="hero.subheadline" className="text-sm font-medium text-zinc-400">Sub-headline</Label>
              <Input
                id="hero.subheadline"
                {...form.register("hero.subheadline")}
                className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 transition-all"
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="hero.bg_image" className="text-sm font-medium text-zinc-400">Background Image URL</Label>
              <Input
                id="hero.bg_image"
                {...form.register("hero.bg_image")}
                placeholder="https://..."
                className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 transition-all font-mono text-sm"
              />
            </div>
          </div>
        )}
      </div>

      <div className="flex flex-col gap-6 p-6 rounded-2xl bg-zinc-900/50 border border-white/[0.05] shadow-[inset_0_1px_1px_rgba(255,255,255,0.02)]">
        <div className="flex items-center justify-between">
          <div className="flex flex-col gap-1">
            <h2 className="text-lg font-medium text-zinc-100">Featured Categories</h2>
            <p className="text-sm text-zinc-500">Show a grid of your top categories.</p>
          </div>
          <div className="flex items-center space-x-2">
            <Switch 
              id="featured_categories.enabled" 
              checked={form.watch("featured_categories.enabled")} 
              onCheckedChange={(enabled) => form.setValue("featured_categories.enabled", enabled)} 
            />
          </div>
        </div>
        
        {form.watch("featured_categories.enabled") && (
          <div className="flex flex-col gap-4 mt-2">
            <div className="flex flex-col gap-2">
              <Label htmlFor="featured_categories.title" className="text-sm font-medium text-zinc-400">Section Title</Label>
              <Input
                id="featured_categories.title"
                {...form.register("featured_categories.title")}
                className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 transition-all"
              />
            </div>
          </div>
        )}
      </div>

      <div className="flex flex-col gap-6 p-6 rounded-2xl bg-zinc-900/50 border border-white/[0.05] shadow-[inset_0_1px_1px_rgba(255,255,255,0.02)]">
        <div className="flex items-center justify-between">
          <div className="flex flex-col gap-1">
            <h2 className="text-lg font-medium text-zinc-100">Brand Story</h2>
            <p className="text-sm text-zinc-500">Tell your customers about your mission.</p>
          </div>
          <div className="flex items-center space-x-2">
            <Switch 
              id="brand_story.enabled" 
              checked={form.watch("brand_story.enabled")} 
              onCheckedChange={(enabled) => form.setValue("brand_story.enabled", enabled)} 
            />
          </div>
        </div>
        
        {form.watch("brand_story.enabled") && (
          <div className="flex flex-col gap-4 mt-2">
            <div className="flex flex-col gap-2">
              <Label htmlFor="brand_story.text" className="text-sm font-medium text-zinc-400">Story Text</Label>
              <textarea
                id="brand_story.text"
                {...form.register("brand_story.text")}
                className="min-h-[120px] p-3 rounded-xl bg-zinc-950 border border-zinc-800 text-zinc-100 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-700 transition-all resize-none"
              />
            </div>
          </div>
        )}
      </div>

      <div className="flex justify-end pt-4">
        <button
          type="submit"
          disabled={isUpdating}
          className="group flex items-center justify-center rounded-full bg-zinc-100 px-8 py-2.5 text-sm font-semibold text-zinc-900 transition-all hover:bg-white disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {isUpdating ? "Saving..." : "Save Appearance"}
        </button>
      </div>
    </form>
  )
}