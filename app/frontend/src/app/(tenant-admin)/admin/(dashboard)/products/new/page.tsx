"use client"

import * as React from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import {
  ArrowLeft,
  Save,
  UploadCloud,
  X,
  Sparkles,
  Plus,
  Trash2,
  DollarSign,
  Package,
  Layers,
  Globe,
  CheckCircle2,
  Image as ImageIcon
} from "lucide-react"
import { uploadToR2 } from "@/lib/upload"
import { customInstance } from "@/lib/axios"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"

const productSchema = z.object({
  title: z.string().min(2, "Product title is required (min 2 characters)"),
  description: z.string().optional(),
  price: z.coerce.number().min(0, "Price must be greater than or equal to 0"),
  compare_at_price: z.coerce.number().optional(),
  cost_per_item: z.coerce.number().optional(),
  sku: z.string().optional(),
  stock_quantity: z.coerce.number().int().min(0, "Stock cannot be negative").default(0),
  low_stock_threshold: z.coerce.number().int().min(0).default(5),
  status: z.enum(["draft", "active", "archived"]).default("active"),
  category: z.string().optional(),
  vendor: z.string().optional(),
  tags: z.string().optional(),
})

type ProductFormValues = z.infer<typeof productSchema>

interface VariantRow {
  id: string
  name: string
  sku: string
  price: number
  stock: number
}

export default function NewProductPage() {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [isAiGenerating, setIsAiGenerating] = React.useState(false)
  const [isUploading, setIsUploading] = React.useState(false)
  const [imageUrlInput, setImageUrlInput] = React.useState("")
  const [images, setImages] = React.useState<string[]>([
    "https://images.unsplash.com/photo-1552346154-21d32810aba3?auto=format&fit=crop&w=800&q=80"
  ])

  const [hasVariants, setHasVariants] = React.useState(true)
  const [variantOptionName, setVariantOptionName] = React.useState("Size")
  const [variantValuesInput, setVariantValuesInput] = React.useState("US 8.5, US 9, US 9.5, US 10, US 10.5, US 11")
  const [variants, setVariants] = React.useState<VariantRow[]>([
    { id: "v1", name: "US 8.5", sku: "KICK-85", price: 8500, stock: 4 },
    { id: "v2", name: "US 9", sku: "KICK-90", price: 8500, stock: 8 },
    { id: "v3", name: "US 9.5", sku: "KICK-95", price: 8500, stock: 5 },
    { id: "v4", name: "US 10", sku: "KICK-100", price: 8500, stock: 6 },
    { id: "v5", name: "US 10.5", sku: "KICK-105", price: 8500, stock: 2 },
    { id: "v6", name: "US 11", sku: "KICK-110", price: 8500, stock: 3 },
  ])

  const form = useForm<any>({
    resolver: zodResolver(productSchema) as any,
    defaultValues: {
      title: "",
      description: "",
      price: 8500,
      compare_at_price: 10500,
      cost_per_item: 4200,
      sku: "KICK-OG-01",
      stock_quantity: 28,
      low_stock_threshold: 5,
      status: "active",
      category: "Sneakers",
      vendor: "Nike / Jordan",
      tags: "retro, high-top, limited, premium",
    },
  })

  const watchPrice = form.watch("price") || 0
  const watchCost = form.watch("cost_per_item") || 0
  const watchTitle = form.watch("title") || ""
  const watchDescription = form.watch("description") || ""

  const profit = Math.max(0, watchPrice - watchCost)
  const marginPercent = watchPrice > 0 ? Math.round((profit / watchPrice) * 100) : 0

  const handleGenerateAiDescription = async () => {
    const title = form.getValues("title")
    if (!title) {
      alert("Please enter a product title first so the AI can craft a matching luxury description.")
      return
    }

    setIsAiGenerating(true)
    try {
      const res = await customInstance<{ data: { description: string } }>({
        url: "/api/v1/ai/generate-description",
        method: "POST",
        data: { product_title: title },
      })
      const desc = res?.data?.description || (res as any)?.description
      if (desc) {
        form.setValue("description", desc, { shouldValidate: true, shouldDirty: true })
      }
    } catch (err) {
      console.error("AI Generation error:", err)
      const fallbackDesc = `Engineered for the discerning collector, the ${title} represents the pinnacle of craftsmanship and heritage design. Featuring premium full-grain leather uppers, responsive cushioning, and signature aesthetic lines that effortlessly bridge high-performance athletics with modern luxury streetwear. Hand-finished and authenticated.`
      form.setValue("description", fallbackDesc, { shouldValidate: true, shouldDirty: true })
    } finally {
      setIsAiGenerating(false)
    }
  }

  const handleAddImageUrl = () => {
    if (!imageUrlInput.trim()) return
    setImages(prev => [...prev, imageUrlInput.trim()])
    setImageUrlInput("")
  }

  const handleRemoveImage = (indexToRemove: number) => {
    setImages(prev => prev.filter((_, i) => i !== indexToRemove))
  }

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    try {
      setIsUploading(true)
      const url = await uploadToR2(file)
      if (url) {
        setImages(prev => [...prev, url])
      }
    } catch (err) {
      console.error("Upload error", err)
    } finally {
      setIsUploading(false)
    }
  }

  const generateVariantsFromInput = () => {
    const values = variantValuesInput.split(",").map(v => v.trim()).filter(Boolean)
    const baseSku = form.getValues("sku") || "SKU"
    const basePrice = form.getValues("price") || 0
    const newVariants: VariantRow[] = values.map((variantValue, idx) => ({
      id: `v-${Date.now()}-${idx}`,
      name: variantValue,
      sku: `${baseSku}-${variantValue.replace(/[^a-zA-Z0-9]/g, "")}`,
      price: basePrice,
      stock: 5,
    }))
    setVariants(newVariants)
  }

  const onSubmit = async (productFormData: any) => {
    setIsSubmitting(true)
    try {
      const slug = productFormData.title
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/(^-|-$)+/g, "")

      const productRes = await customInstance<{ data: any; id?: string }>({
        url: "/api/v1/catalog/products",
        method: "POST",
        data: {
          title: productFormData.title,
          slug: slug,
          price: productFormData.price,
          description: productFormData.description,
          status: productFormData.status,
          compare_at_price: productFormData.compare_at_price,
          sku: productFormData.sku,
          stock_quantity: hasVariants
            ? variants.reduce((acc, v) => acc + v.stock, 0)
            : productFormData.stock_quantity,
        },
      })

      const createdProduct = productRes?.data || productRes
      const productId = createdProduct?.id

      if (productId) {
        for (let i = 0; i < images.length; i++) {
          try {
            await customInstance({
              url: `/api/v1/catalog/products/${productId}/images/link`,
              method: "POST",
              data: {
                url: images[i],
                is_primary: i === 0,
              },
            })
          } catch (imgErr) {
            console.warn("Could not link image:", images[i], imgErr)
          }
        }

        if (hasVariants && variants.length > 0) {
          for (const variant of variants) {
            try {
              await customInstance({
                url: `/api/v1/catalog/products/${productId}/variants`,
                method: "POST",
                data: {
                  title: variant.name,
                  sku: variant.sku,
                  price: variant.price,
                  inventory_quantity: variant.stock,
                },
              })
            } catch (varErr) {
              console.warn("Could not create variant:", variant.name, varErr)
            }
          }
        }
      }

      router.push("/admin/products")
    } catch (e: any) {
      console.error("Failed to save product", e)
      alert(e.message || "Failed to save product.")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form
      onSubmit={form.handleSubmit(onSubmit)}
      className="flex flex-col gap-8 w-full max-w-5xl mx-auto pb-24 animate-in fade-in duration-500"
    >
      <div className="flex items-center justify-between border-b border-white/[0.05] pb-6">
        <div className="flex items-center gap-4">
          <Link
            href="/admin/products"
            className="flex items-center justify-center size-9 rounded-xl bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
          </Link>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-zinc-50">Create Product</h1>
            <p className="text-xs text-zinc-400">Configure catalog information, variants, and pricing</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => router.push("/admin/products")}
            className="px-4 h-9 rounded-xl font-medium text-xs text-zinc-400 hover:text-zinc-100 transition-colors"
          >
            Discard
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="flex h-9 items-center justify-center gap-2 rounded-xl bg-zinc-100 hover:bg-white px-5 text-xs font-semibold text-zinc-950 transition-all shadow-sm disabled:opacity-50"
          >
            <Save className="w-3.5 h-3.5" />
            {isSubmitting ? "Publishing..." : "Save Product"}
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 flex flex-col gap-6">
          
          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-6 flex flex-col gap-5">
              <h2 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                <Package className="size-4 text-emerald-400" /> General Information
              </h2>

              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-zinc-300">
                  Product Title <span className="text-red-400">*</span>
                </label>
                <input
                  type="text"
                  {...form.register("title")}
                  placeholder="e.g. Air Jordan 1 Retro High 'Lost & Found'"
                  className="h-10 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3.5 text-xs text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                />
                {form.formState.errors.title && (
                  <span className="text-[11px] text-red-400">
                    {String(form.formState.errors.title.message)}
                  </span>
                )}
              </div>

              <div className="flex flex-col gap-1.5">
                <div className="flex items-center justify-between">
                  <label className="text-xs font-medium text-zinc-300">Product Description</label>
                  <button
                    type="button"
                    onClick={handleGenerateAiDescription}
                    disabled={isAiGenerating}
                    className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 hover:bg-emerald-500/20 text-xs font-medium transition-colors disabled:opacity-50"
                  >
                    <Sparkles className="size-3 animate-spin" />
                    {isAiGenerating ? "Generating Copy..." : "✨ AI Copy Assistant"}
                  </button>
                </div>
                <textarea
                  rows={4}
                  {...form.register("description")}
                  placeholder="Detailed material breakdown, sizing advice, and luxury storytelling..."
                  className="w-full rounded-xl bg-zinc-950 border border-zinc-800 p-3.5 text-xs text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700 resize-none leading-relaxed"
                />
              </div>
            </div>
          </div>

          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-6 flex flex-col gap-5">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                  <ImageIcon className="size-4 text-blue-400" /> Media & Photography
                </h2>
                <span className="text-[11px] text-zinc-400">{images.length} images linked</span>
              </div>

              {images.length > 0 && (
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                  {images.map((img, idx) => (
                    <div
                      key={idx}
                      className="group relative h-28 rounded-xl overflow-hidden border border-zinc-800 bg-zinc-950"
                    >
                      <img src={img} alt="Product" className="w-full h-full object-cover" />
                      {idx === 0 && (
                        <span className="absolute top-2 left-2 rounded-md bg-zinc-900/90 border border-white/10 px-1.5 py-0.5 text-[9px] font-semibold text-zinc-200">
                          Primary
                        </span>
                      )}
                      <button
                        type="button"
                        onClick={() => handleRemoveImage(idx)}
                        className="absolute top-2 right-2 p-1 rounded-md bg-red-950/80 text-red-400 border border-red-800/60 opacity-0 group-hover:opacity-100 transition-opacity"
                      >
                        <X className="size-3" />
                      </button>
                    </div>
                  ))}
                </div>
              )}

              <div className="flex flex-col sm:flex-row gap-2">
                <input
                  type="url"
                  value={imageUrlInput}
                  onChange={(e) => setImageUrlInput(e.target.value)}
                  placeholder="Paste direct image URL (Unsplash, Shopify CDN, etc)..."
                  className="h-9 flex-1 rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200 placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                />
                <button
                  type="button"
                  onClick={handleAddImageUrl}
                  className="h-9 px-3.5 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium transition-colors"
                >
                  Add Image
                </button>
                <label className="h-9 px-3.5 rounded-xl bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 text-zinc-300 text-xs font-medium transition-colors flex items-center justify-center gap-1.5 cursor-pointer">
                  <UploadCloud className="size-3.5" />
                  <span>{isUploading ? "Uploading..." : "Upload File"}</span>
                  <input
                    type="file"
                    accept="image/*"
                    onChange={handleImageUpload}
                    className="hidden"
                    disabled={isUploading}
                  />
                </label>
              </div>
            </div>
          </div>

          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-6 flex flex-col gap-5">
              <h2 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                <DollarSign className="size-4 text-amber-400" /> Pricing & Unit Economics
              </h2>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-zinc-300">Price (EGP)</label>
                  <input
                    type="number"
                    step="any"
                    {...form.register("price")}
                    className="h-10 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3.5 text-xs font-mono text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-zinc-300">Compare-at Price (MSRP)</label>
                  <input
                    type="number"
                    step="any"
                    {...form.register("compare_at_price")}
                    className="h-10 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3.5 text-xs font-mono text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-medium text-zinc-300">Cost per Item (EGP)</label>
                  <input
                    type="number"
                    step="any"
                    {...form.register("cost_per_item")}
                    className="h-10 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3.5 text-xs font-mono text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700"
                  />
                </div>
              </div>

              <div className="rounded-xl bg-zinc-950 border border-zinc-800/80 p-3.5 flex items-center justify-between text-xs">
                <div className="flex items-center gap-2 text-zinc-400">
                  <span>Unit Margin:</span>
                  <span className="font-bold text-emerald-400 font-mono">{marginPercent}%</span>
                </div>
                <div className="flex items-center gap-2 text-zinc-400">
                  <span>Unit Profit:</span>
                  <span className="font-bold text-zinc-200 font-mono">
                    EGP {profit.toLocaleString()}
                  </span>
                </div>
                <div className="flex items-center gap-1.5 text-zinc-500 text-[11px]">
                  <CheckCircle2 className="size-3.5 text-emerald-400" />
                  <span>Healthy Gross Margin</span>
                </div>
              </div>
            </div>
          </div>

          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-6 flex flex-col gap-5">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                    <Layers className="size-4 text-violet-400" /> Variant Matrix
                  </h2>
                  <p className="text-xs text-zinc-500">Configure size, color, or style SKU variations</p>
                </div>
                <label className="flex items-center gap-2 cursor-pointer text-xs text-zinc-300">
                  <input
                    type="checkbox"
                    checked={hasVariants}
                    onChange={(e) => setHasVariants(e.target.checked)}
                    className="rounded border-zinc-700 bg-zinc-950 text-emerald-500"
                  />
                  <span>Has multiple options</span>
                </label>
              </div>

              {hasVariants && (
                <div className="flex flex-col gap-4">
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
                    <input
                      type="text"
                      value={variantOptionName}
                      onChange={(e) => setVariantOptionName(e.target.value)}
                      placeholder="Option (e.g. Size)"
                      className="h-9 rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                    />
                    <input
                      type="text"
                      value={variantValuesInput}
                      onChange={(e) => setVariantValuesInput(e.target.value)}
                      placeholder="Values (comma separated)"
                      className="sm:col-span-2 h-9 rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                    />
                  </div>
                  <button
                    type="button"
                    onClick={generateVariantsFromInput}
                    className="self-start px-3 h-8 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium transition-colors"
                  >
                    Regenerate Matrix
                  </button>

                  <div className="rounded-xl border border-zinc-800 bg-zinc-950 overflow-hidden mt-2">
                    <table className="w-full text-xs text-left">
                      <thead className="bg-zinc-900/80 text-zinc-400 border-b border-zinc-800">
                        <tr>
                          <th className="px-3 py-2.5 font-medium">Variant</th>
                          <th className="px-3 py-2.5 font-medium">SKU</th>
                          <th className="px-3 py-2.5 font-medium">Price (EGP)</th>
                          <th className="px-3 py-2.5 font-medium">Stock</th>
                          <th className="px-3 py-2.5 w-8"></th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-zinc-800/60">
                        {variants.map((v, idx) => (
                          <tr key={v.id}>
                            <td className="px-3 py-2 font-medium text-zinc-200">{v.name}</td>
                            <td className="px-3 py-2">
                              <input
                                type="text"
                                value={v.sku}
                                onChange={(e) => {
                                  const updated = [...variants]
                                  updated[idx].sku = e.target.value
                                  setVariants(updated)
                                }}
                                className="h-7 w-28 rounded-md bg-zinc-900 border border-zinc-800 px-2 font-mono text-[11px] text-zinc-300"
                              />
                            </td>
                            <td className="px-3 py-2">
                              <input
                                type="number"
                                value={v.price}
                                onChange={(e) => {
                                  const updated = [...variants]
                                  updated[idx].price = Number(e.target.value)
                                  setVariants(updated)
                                }}
                                className="h-7 w-24 rounded-md bg-zinc-900 border border-zinc-800 px-2 font-mono text-[11px] text-zinc-300"
                              />
                            </td>
                            <td className="px-3 py-2">
                              <input
                                type="number"
                                value={v.stock}
                                onChange={(e) => {
                                  const updated = [...variants]
                                  updated[idx].stock = Number(e.target.value)
                                  setVariants(updated)
                                }}
                                className="h-7 w-20 rounded-md bg-zinc-900 border border-zinc-800 px-2 font-mono text-[11px] text-zinc-300"
                              />
                            </td>
                            <td className="px-3 py-2 text-center">
                              <button
                                type="button"
                                onClick={() => setVariants(prev => prev.filter((_, i) => i !== idx))}
                                className="text-zinc-500 hover:text-red-400"
                              >
                                <Trash2 className="size-3.5" />
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </div>
          </div>

          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-6 flex flex-col gap-4">
              <h2 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
                <Globe className="size-4 text-emerald-400" /> Search Engine Listing (SERP) Preview
              </h2>
              <div className="rounded-xl bg-zinc-950 border border-zinc-800/80 p-4 font-sans">
                <div className="text-[11px] text-zinc-500 flex items-center gap-1.5">
                  <span>https://kicksandco.eg</span>
                  <span>›</span>
                  <span>products</span>
                  <span>›</span>
                  <span className="text-zinc-400">
                    {watchTitle ? watchTitle.toLowerCase().replace(/[^a-z0-9]+/g, "-") : "product-slug"}
                  </span>
                </div>
                <div className="text-sm font-medium text-blue-400 hover:underline cursor-pointer mt-1">
                  {watchTitle || "Product Title Goes Here | Kicks & Co. Cairo"}
                </div>
                <div className="text-xs text-zinc-400 mt-1 line-clamp-2 leading-relaxed">
                  {watchDescription || "Discover our curated collection of authenticated luxury footwear with same-day Cairo dispatch, flexible installments, and full size exchange."}
                </div>
              </div>
            </div>
          </div>

        </div>

        <div className="flex flex-col gap-6">
          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-5 flex flex-col gap-4">
              <h3 className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">
                Product Status
              </h3>
              <select
                {...form.register("status")}
                className="h-10 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-700"
              >
                <option value="active">Active (Visible on Storefront)</option>
                <option value="draft">Draft (Hidden)</option>
                <option value="archived">Archived</option>
              </select>
            </div>
          </div>

          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-5 flex flex-col gap-4">
              <h3 className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">
                Inventory Rules
              </h3>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-zinc-400">Master SKU</label>
                <input
                  type="text"
                  {...form.register("sku")}
                  className="h-9 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3 font-mono text-xs text-zinc-200"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-zinc-400">Total Stock Quantity</label>
                <input
                  type="number"
                  {...form.register("stock_quantity")}
                  className="h-9 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3 font-mono text-xs text-zinc-200"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-zinc-400">Low Stock Alert Threshold</label>
                <input
                  type="number"
                  {...form.register("low_stock_threshold")}
                  className="h-9 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3 font-mono text-xs text-zinc-200"
                />
              </div>
            </div>
          </div>

          <div className="p-1 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
            <div className="rounded-[calc(1.5rem-2px)] bg-zinc-900/60 p-5 flex flex-col gap-4">
              <h3 className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">
                Classification
              </h3>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-zinc-400">Category</label>
                <select
                  {...form.register("category")}
                  className="h-9 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                >
                  <option value="Sneakers">Sneakers</option>
                  <option value="Streetwear">Streetwear</option>
                  <option value="Accessories">Accessories</option>
                  <option value="Care & Protection">Care & Protection</option>
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-zinc-400">Vendor / Brand</label>
                <input
                  type="text"
                  {...form.register("vendor")}
                  placeholder="e.g. Nike, Jordan, Yeezy"
                  className="h-9 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-zinc-400">Tags (comma separated)</label>
                <input
                  type="text"
                  {...form.register("tags")}
                  placeholder="limited, high-top, rare"
                  className="h-9 w-full rounded-xl bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </form>
  )
}
