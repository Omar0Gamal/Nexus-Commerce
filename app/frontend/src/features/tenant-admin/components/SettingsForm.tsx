"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { useGetApiV1ShopsSettings, usePatchApiV1ShopsSettings } from "@/api/endpoints/auth/auth";

const settingsSchema = z.object({
  name: z.string().min(1, "Store name is required"),
  currency: z.string().min(3, "Currency code is required"),
  custom_domain: z.string().optional(),
  seo_title: z.string().optional(),
  seo_description: z.string().optional(),
})

type SettingsFormValues = z.infer<typeof settingsSchema>

export function SettingsForm() {
  const { data: settingsEnvelope, isLoading: isFetching } = useGetApiV1ShopsSettings()
  const { mutate: updateSettings, isPending: isUpdating } = usePatchApiV1ShopsSettings()

  const [message, setMessage] = React.useState<{ type: 'success' | 'error', text: string } | null>(null)

  const settings = settingsEnvelope?.data
  
  const form = useForm<SettingsFormValues>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      name: "",
      currency: "EGP",
      custom_domain: "",
      seo_title: "",
      seo_description: "",
    },
  })

  React.useEffect(() => {
    if (settings) {
      form.reset({
        name: settings.name || "",
        currency: settings.currency || "EGP",
        custom_domain: settings.custom_domain || "",
        seo_title: settings.seo_title || "",
        seo_description: settings.seo_description || "",
      })
    }
  }, [settings, form])

  if (isFetching) {
    return <div className="text-zinc-500 animate-pulse text-sm">Loading configuration...</div>
  }

  const onSubmit = (formValues: SettingsFormValues) => {
    setMessage(null)
    updateSettings({ data: formValues }, {
      onSuccess: () => {
        setMessage({ type: 'success', text: 'Settings updated successfully.' })
      },
      onError: (err: any) => {
        setMessage({ type: 'error', text: err.response?.data?.error || 'Failed to update settings.' })
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
        <div className="flex flex-col gap-1">
          <h2 className="text-lg font-medium text-zinc-100">General Information</h2>
          <p className="text-sm text-zinc-500">Your store's core identity and base configuration.</p>
        </div>
        
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="name" className="text-sm font-medium text-zinc-400">Store Name</Label>
            <Input
              id="name"
              {...form.register("name")}
              className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 transition-all"
            />
            {form.formState.errors.name && (
              <p className="text-xs text-red-400 mt-1">{form.formState.errors.name.message}</p>
            )}
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="currency" className="text-sm font-medium text-zinc-400">Base Currency</Label>
            <Input
              id="currency"
              {...form.register("currency")}
              className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 transition-all"
            />
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-6 p-6 rounded-2xl bg-zinc-900/50 border border-white/[0.05] shadow-[inset_0_1px_1px_rgba(255,255,255,0.02)]">
        <div className="flex flex-col gap-1">
          <h2 className="text-lg font-medium text-zinc-100">Domain Configuration</h2>
          <p className="text-sm text-zinc-500">Connect a custom domain to your storefront.</p>
        </div>
        
        <div className="flex flex-col gap-2">
          <Label htmlFor="custom_domain" className="text-sm font-medium text-zinc-400">Custom Domain</Label>
          <div className="relative">
            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-600 text-sm font-mono">https://</span>
            <Input
              id="custom_domain"
              placeholder="www.yourbrand.com"
              {...form.register("custom_domain")}
              className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 pl-20 focus-visible:ring-zinc-700 transition-all font-mono text-sm placeholder:text-zinc-700"
            />
          </div>
          <p className="text-xs text-zinc-600 mt-1">Requires a Pro or Premium plan. You will also need to configure a CNAME record pointing to shops.egyptcommerce.com.</p>
        </div>
      </div>

      <div className="flex flex-col gap-6 p-6 rounded-2xl bg-zinc-900/50 border border-white/[0.05] shadow-[inset_0_1px_1px_rgba(255,255,255,0.02)]">
        <div className="flex flex-col gap-1">
          <h2 className="text-lg font-medium text-zinc-100">Search Engine Optimization (SEO)</h2>
          <p className="text-sm text-zinc-500">Control how your store appears on Google and social media.</p>
        </div>
        
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="seo_title" className="text-sm font-medium text-zinc-400">Meta Title</Label>
            <Input
              id="seo_title"
              placeholder="e.g. Acme Studio | High-End Ceramics"
              {...form.register("seo_title")}
              className="h-11 rounded-xl bg-zinc-950 border-zinc-800 text-zinc-100 focus-visible:ring-zinc-700 transition-all"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="seo_description" className="text-sm font-medium text-zinc-400">Meta Description</Label>
            <textarea
              id="seo_description"
              {...form.register("seo_description")}
              placeholder="e.g. Discover our latest collection of handcrafted ceramics..."
              className="min-h-[100px] p-3 rounded-xl bg-zinc-950 border border-zinc-800 text-zinc-100 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-700 transition-all resize-none"
            />
          </div>
        </div>
      </div>

      <div className="flex justify-end pt-4">
        <button
          type="submit"
          disabled={isUpdating}
          className="group flex items-center justify-center rounded-full bg-zinc-100 px-8 py-2.5 text-sm font-semibold text-zinc-900 transition-all hover:bg-white disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {isUpdating ? "Saving..." : "Save Configuration"}
        </button>
      </div>
    </form>
  )
}