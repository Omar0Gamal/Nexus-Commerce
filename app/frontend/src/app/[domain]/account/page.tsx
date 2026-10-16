"use client"

import * as React from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useQueryClient } from "@tanstack/react-query"
import {
  Package,
  MapPin,
  User,
  LogOut,
  Plus,
  Trash2,
  CheckCircle2,
  AlertCircle,
  Loader2,
  ExternalLink,
} from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { useCustomerAuthStore } from "@/store/useCustomerAuthStore"
import {
  useGetApiV1OrdersMy,
  getGetApiV1OrdersMyQueryKey,
} from "@/api/endpoints/orders/orders"
import {
  useGetApiV1Addresses,
  usePostApiV1Addresses,
  useDeleteApiV1AddressesId,
  usePatchApiV1AddressesIdDefault,
  getGetApiV1AddressesQueryKey,
} from "@/api/endpoints/addresses/addresses"
import {
  useGetApiV1AuthCustomerMe,
  usePatchApiV1AuthCustomerMe,
  getGetApiV1AuthCustomerMeQueryKey,
} from "@/api/endpoints/auth/auth"

export default function AccountPage() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const { customerToken, logoutCustomer } = useCustomerAuthStore()

  const [mounted, setMounted] = React.useState(false)
  const [token, setToken] = React.useState<string | null>(null)

  React.useEffect(() => {
    setMounted(true)
    const storedToken =
      customerToken ||
      (typeof window !== "undefined"
        ? localStorage.getItem("customer_access_token")
        : null)
    setToken(storedToken)
  }, [customerToken])

  const isAuthEnabled = Boolean(token)

  const {
    data: rawOrdersData,
    isLoading: ordersLoading,
    error: ordersError,
  } = useGetApiV1OrdersMy({
    query: { enabled: isAuthEnabled },
  })

  const {
    data: rawAddressesData,
    isLoading: addressesLoading,
  } = useGetApiV1Addresses({
    query: { enabled: isAuthEnabled },
  })

  const {
    data: rawMeData,
    isLoading: meLoading,
  } = useGetApiV1AuthCustomerMe({
    query: { enabled: isAuthEnabled },
  })

  const orders = React.useMemo(() => {
    if (!rawOrdersData) return []
    const ordersPayload = (rawOrdersData as any).data ?? rawOrdersData
    return Array.isArray(ordersPayload) ? ordersPayload : []
  }, [rawOrdersData])

  const addresses = React.useMemo(() => {
    if (!rawAddressesData) return []
    const addressesPayload = (rawAddressesData as any).data ?? rawAddressesData
    return Array.isArray(addressesPayload) ? addressesPayload : []
  }, [rawAddressesData])

  const customerMe = React.useMemo(() => {
    if (!rawMeData) return null
    return (rawMeData as any).data ?? rawMeData
  }, [rawMeData])

  const [isAddAddressOpen, setIsAddAddressOpen] = React.useState(false)
  const [newLabel, setNewLabel] = React.useState("Home")
  const [newFirstName, setNewFirstName] = React.useState("")
  const [newLastName, setNewLastName] = React.useState("")
  const [newAddress1, setNewAddress1] = React.useState("")
  const [newAddress2, setNewAddress2] = React.useState("")
  const [newCity, setNewCity] = React.useState("Cairo")
  const [newState, setNewState] = React.useState("Cairo")
  const [newZipCode, setNewZipCode] = React.useState("")
  const [newCountry, setNewCountry] = React.useState("Egypt")
  const [newPhone, setNewPhone] = React.useState("")
  const [newIsDefault, setNewIsDefault] = React.useState(false)
  const [addressError, setAddressError] = React.useState<string | null>(null)

  const createAddressMutation = usePostApiV1Addresses({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetApiV1AddressesQueryKey() })
        setIsAddAddressOpen(false)
        resetAddressForm()
      },
      onError: (err: any) => {
        setAddressError(err?.response?.data?.message || "Failed to save address.")
      },
    },
  })

  const deleteAddressMutation = useDeleteApiV1AddressesId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetApiV1AddressesQueryKey() })
      },
    },
  })

  const setDefaultAddressMutation = usePatchApiV1AddressesIdDefault({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetApiV1AddressesQueryKey() })
      },
    },
  })

  const [editFirstName, setEditFirstName] = React.useState("")
  const [editLastName, setEditLastName] = React.useState("")
  const [editPhone, setEditPhone] = React.useState("")
  const [profileSuccess, setProfileSuccess] = React.useState(false)
  const [profileError, setProfileError] = React.useState<string | null>(null)

  React.useEffect(() => {
    if (customerMe) {
      setEditFirstName(customerMe.first_name || "")
      setEditLastName(customerMe.last_name || "")
      setEditPhone(customerMe.phone || "")
    }
  }, [customerMe])

  const updateProfileMutation = usePatchApiV1AuthCustomerMe({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetApiV1AuthCustomerMeQueryKey() })
        setProfileSuccess(true)
        setProfileError(null)
        setTimeout(() => setProfileSuccess(false), 3000)
      },
      onError: (err: any) => {
        setProfileError(err?.response?.data?.message || "Failed to update profile.")
      },
    },
  })

  const resetAddressForm = () => {
    setNewLabel("Home")
    setNewFirstName("")
    setNewLastName("")
    setNewAddress1("")
    setNewAddress2("")
    setNewCity("Cairo")
    setNewState("Cairo")
    setNewZipCode("")
    setNewCountry("Egypt")
    setNewPhone("")
    setNewIsDefault(false)
    setAddressError(null)
  }

  const handleAddAddress = (e: React.FormEvent) => {
    e.preventDefault()
    setAddressError(null)
    createAddressMutation.mutate({
      data: {
        label: newLabel,
        first_name: newFirstName,
        last_name: newLastName,
        address1: newAddress1,
        address2: newAddress2 || undefined,
        city: newCity,
        state: newState,
        country: newCountry,
        zip_code: newZipCode || undefined,
        phone: newPhone || undefined,
        is_default: newIsDefault,
      },
    })
  }

  const handleUpdateProfile = (e: React.FormEvent) => {
    e.preventDefault()
    setProfileSuccess(false)
    setProfileError(null)
    updateProfileMutation.mutate({
      data: {
        first_name: editFirstName,
        last_name: editLastName,
        phone: editPhone,
      },
    })
  }

  const handleLogout = () => {
    logoutCustomer()
    router.push("/login")
  }

  const formatStatusVariant = (status: string) => {
    const normalizedStatus = (status || "").toLowerCase()
    if (normalizedStatus === "delivered" || normalizedStatus === "paid" || normalizedStatus === "completed") return "default"
    if (normalizedStatus === "shipped") return "secondary"
    return "outline"
  }

  if (!mounted) {
    return (
      <div className="container mx-auto max-w-5xl px-4 py-16 min-h-[60vh] flex items-center justify-center">
        <Loader2 className="w-6 h-6 animate-spin text-zinc-400" />
      </div>
    )
  }

  if (!token) {
    return (
      <div className="container mx-auto max-w-xl px-4 py-16 min-h-[60vh] flex flex-col items-center justify-center text-center">
        <div className="w-16 h-16 rounded-full bg-zinc-100 dark:bg-zinc-900 flex items-center justify-center mb-6">
          <User className="w-8 h-8 text-zinc-500" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50 mb-2">
          Customer Portal
        </h1>
        <p className="text-zinc-500 text-sm max-w-md mb-8">
          Please sign in to view your orders, manage saved shipping addresses, and update your personal details.
        </p>
        <Link
          href="/login"
          className="inline-flex h-12 items-center justify-center rounded-xl bg-zinc-900 dark:bg-zinc-100 px-8 text-sm font-semibold text-zinc-50 dark:text-zinc-900 shadow hover:opacity-90 transition-opacity"
        >
          Sign In or Register
        </Link>
      </div>
    )
  }

  return (
    <div className="container mx-auto max-w-5xl px-4 py-10 min-h-screen">
      <div className="mb-8 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
            My Account
          </h1>
          <p className="text-zinc-500 text-sm mt-1">
            {customerMe?.email ? `Signed in as ${customerMe.email}` : "Manage your orders and address book."}
          </p>
        </div>
        <button
          onClick={handleLogout}
          className="inline-flex items-center gap-2 self-start sm:self-auto text-sm font-medium text-zinc-500 hover:text-red-600 dark:hover:text-red-400 transition-colors cursor-pointer"
        >
          <LogOut className="w-4 h-4" />
          <span>Sign Out</span>
        </button>
      </div>

      <Tabs defaultValue="orders" className="flex flex-col md:flex-row gap-8" orientation="vertical">
        <div className="md:w-64 shrink-0">
          <TabsList className="flex flex-col w-full h-auto bg-transparent p-0 space-y-1.5" variant="line">
            <TabsTrigger
              value="orders"
              className="justify-start gap-3 px-4 py-3 w-full data-[active]:bg-zinc-100 dark:data-[active]:bg-zinc-800 rounded-xl font-medium text-sm transition-all"
            >
              <Package className="w-4 h-4 opacity-70" />
              <span>Order History</span>
            </TabsTrigger>
            <TabsTrigger
              value="addresses"
              className="justify-start gap-3 px-4 py-3 w-full data-[active]:bg-zinc-100 dark:data-[active]:bg-zinc-800 rounded-xl font-medium text-sm transition-all"
            >
              <MapPin className="w-4 h-4 opacity-70" />
              <span>Saved Addresses</span>
            </TabsTrigger>
            <TabsTrigger
              value="settings"
              className="justify-start gap-3 px-4 py-3 w-full data-[active]:bg-zinc-100 dark:data-[active]:bg-zinc-800 rounded-xl font-medium text-sm transition-all"
            >
              <User className="w-4 h-4 opacity-70" />
              <span>Account Settings</span>
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1">
          <TabsContent value="orders" className="m-0">
            <Card className="border-zinc-200 dark:border-zinc-800 shadow-sm">
              <CardHeader>
                <CardTitle>Order History</CardTitle>
                <CardDescription>View your orders and track their delivery status.</CardDescription>
              </CardHeader>
              <CardContent>
                {ordersLoading ? (
                  <div className="py-12 flex justify-center items-center">
                    <Loader2 className="w-6 h-6 animate-spin text-zinc-400" />
                  </div>
                ) : ordersError ? (
                  <Alert variant="destructive">
                    <AlertCircle className="w-4 h-4 mr-2" />
                    <AlertDescription>Failed to load orders. Please try again later.</AlertDescription>
                  </Alert>
                ) : orders.length === 0 ? (
                  <div className="py-12 flex flex-col items-center justify-center text-center gap-3">
                    <Package className="w-12 h-12 text-zinc-300 dark:text-zinc-700 stroke-[1.5]" />
                    <p className="font-semibold text-zinc-800 dark:text-zinc-200">No orders placed yet</p>
                    <p className="text-sm text-zinc-500 max-w-sm">
                      When you order items from our store, your order history and live shipping updates will appear here.
                    </p>
                    <Link
                      href="/"
                      className="mt-2 inline-flex h-10 items-center justify-center rounded-xl bg-zinc-900 dark:bg-zinc-100 px-5 text-sm font-medium text-zinc-50 dark:text-zinc-900 shadow hover:opacity-90"
                    >
                      Start Shopping
                    </Link>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {orders.map((order: any) => {
                      const orderNum = order.order_number ? `#${order.order_number}` : (order.id ? `#${order.id.slice(0, 8)}` : "#Order")
                      const formattedDate = order.created_at
                        ? new Date(order.created_at).toLocaleDateString("en-US", {
                            year: "numeric",
                            month: "short",
                            day: "numeric",
                          })
                        : "Recent"

                      return (
                        <div
                          key={order.id}
                          className="flex flex-col p-5 rounded-xl border border-zinc-200/80 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/50 hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors gap-4"
                        >
                          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-zinc-100 dark:border-zinc-800/80 pb-3">
                            <div>
                              <div className="flex items-center gap-2">
                                <span className="font-semibold text-zinc-900 dark:text-zinc-100">{orderNum}</span>
                                <Badge variant={formatStatusVariant(order.status)} className="capitalize">
                                  {order.status || "Pending"}
                                </Badge>
                              </div>
                              <p className="text-xs text-zinc-500 mt-0.5">{formattedDate}</p>
                            </div>
                            <div className="flex items-center gap-4">
                              <span className="font-mono font-semibold text-zinc-900 dark:text-zinc-100 text-sm">
                                EGP {order.total_price}
                              </span>
                              <Link
                                href={`/track?tracking_number=${order.order_number || order.id}`}
                                className="inline-flex items-center gap-1 text-xs font-medium text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 transition-colors"
                              >
                                <span>Track</span>
                                <ExternalLink className="w-3 h-3" />
                              </Link>
                            </div>
                          </div>

                          {Array.isArray(order.items) && order.items.length > 0 && (
                            <div className="space-y-1.5 text-xs text-zinc-600 dark:text-zinc-400">
                              {order.items.map((item: any, idx: number) => (
                                <div key={item.id || idx} className="flex justify-between items-center">
                                  <span>
                                    {item.name || item.title || "Item"} &times; {item.quantity || 1}
                                  </span>
                                  <span className="font-mono">EGP {item.price}</span>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      )
                    })}
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="addresses" className="m-0">
            <Card className="border-zinc-200 dark:border-zinc-800 shadow-sm">
              <CardHeader className="flex flex-row items-center justify-between pb-4">
                <div>
                  <CardTitle>Saved Addresses</CardTitle>
                  <CardDescription>Manage your delivery destinations for quick checkout.</CardDescription>
                </div>
                <Dialog open={isAddAddressOpen} onOpenChange={setIsAddAddressOpen}>
                  <DialogTrigger className="inline-flex items-center gap-1.5 h-9 px-3.5 rounded-lg bg-zinc-900 dark:bg-zinc-100 text-zinc-50 dark:text-zinc-900 text-xs font-semibold hover:opacity-90 transition-opacity cursor-pointer">
                    <Plus className="w-3.5 h-3.5" />
                    <span>Add Address</span>
                  </DialogTrigger>
                  <DialogContent className="max-w-md bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800">
                    <DialogHeader>
                      <DialogTitle>Add New Address</DialogTitle>
                    </DialogHeader>

                    {addressError && (
                      <Alert variant="destructive" className="my-2">
                        <AlertCircle className="w-4 h-4 mr-2" />
                        <AlertDescription>{addressError}</AlertDescription>
                      </Alert>
                    )}

                    <form onSubmit={handleAddAddress} className="space-y-4 pt-2">
                      <div className="grid grid-cols-2 gap-3">
                        <div>
                          <label className="text-xs font-medium text-zinc-500 mb-1 block">Label</label>
                          <input
                            type="text"
                            required
                            value={newLabel}
                            onChange={(e) => setNewLabel(e.target.value)}
                            placeholder="Home / Work"
                            className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium text-zinc-500 mb-1 block">Country</label>
                          <select
                            value={newCountry}
                            onChange={(e) => setNewCountry(e.target.value)}
                            className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                          >
                            <option value="Egypt">Egypt</option>
                            <option value="United Arab Emirates">UAE</option>
                            <option value="Saudi Arabia">Saudi Arabia</option>
                          </select>
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-3">
                        <div>
                          <label className="text-xs font-medium text-zinc-500 mb-1 block">First Name</label>
                          <input
                            type="text"
                            required
                            value={newFirstName}
                            onChange={(e) => setNewFirstName(e.target.value)}
                            className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium text-zinc-500 mb-1 block">Last Name</label>
                          <input
                            type="text"
                            required
                            value={newLastName}
                            onChange={(e) => setNewLastName(e.target.value)}
                            className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                          />
                        </div>
                      </div>

                      <div>
                        <label className="text-xs font-medium text-zinc-500 mb-1 block">Address</label>
                        <input
                          type="text"
                          required
                          value={newAddress1}
                          onChange={(e) => setNewAddress1(e.target.value)}
                          placeholder="Street name, building"
                          className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                        />
                      </div>

                      <div>
                        <label className="text-xs font-medium text-zinc-500 mb-1 block">Apartment / Suite (optional)</label>
                        <input
                          type="text"
                          value={newAddress2}
                          onChange={(e) => setNewAddress2(e.target.value)}
                          placeholder="Apt, floor, suite"
                          className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                        />
                      </div>

                      <div className="grid grid-cols-3 gap-3">
                        <div>
                          <label className="text-xs font-medium text-zinc-500 mb-1 block">City</label>
                          <input
                            type="text"
                            required
                            value={newCity}
                            onChange={(e) => setNewCity(e.target.value)}
                            className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium text-zinc-500 mb-1 block">State</label>
                          <input
                            type="text"
                            value={newState}
                            onChange={(e) => setNewState(e.target.value)}
                            className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                          />
                        </div>
                        <div>
                          <label className="text-xs font-medium text-zinc-500 mb-1 block">Zip Code</label>
                          <input
                            type="text"
                            value={newZipCode}
                            onChange={(e) => setNewZipCode(e.target.value)}
                            className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                          />
                        </div>
                      </div>

                      <div>
                        <label className="text-xs font-medium text-zinc-500 mb-1 block">Phone Number</label>
                        <input
                          type="tel"
                          value={newPhone}
                          onChange={(e) => setNewPhone(e.target.value)}
                          placeholder="+20 100 000 0000"
                          className="h-10 w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 text-sm"
                        />
                      </div>

                      <div className="flex items-center gap-2 pt-2">
                        <input
                          type="checkbox"
                          id="is_default"
                          checked={newIsDefault}
                          onChange={(e) => setNewIsDefault(e.target.checked)}
                          className="rounded border-zinc-300 dark:border-zinc-700 accent-zinc-900"
                        />
                        <label htmlFor="is_default" className="text-xs text-zinc-600 dark:text-zinc-400">
                          Set as default delivery address
                        </label>
                      </div>

                      <button
                        type="submit"
                        disabled={createAddressMutation.isPending}
                        className="w-full h-11 rounded-xl bg-zinc-900 dark:bg-zinc-100 text-zinc-50 dark:text-zinc-900 font-semibold text-sm hover:opacity-90 transition-opacity disabled:opacity-50"
                      >
                        {createAddressMutation.isPending ? "Saving..." : "Save Address"}
                      </button>
                    </form>
                  </DialogContent>
                </Dialog>
              </CardHeader>

              <CardContent>
                {addressesLoading ? (
                  <div className="py-12 flex justify-center items-center">
                    <Loader2 className="w-6 h-6 animate-spin text-zinc-400" />
                  </div>
                ) : addresses.length === 0 ? (
                  <div className="py-12 flex flex-col items-center justify-center text-center gap-3">
                    <MapPin className="w-12 h-12 text-zinc-300 dark:text-zinc-700 stroke-[1.5]" />
                    <p className="font-semibold text-zinc-800 dark:text-zinc-200">No saved addresses</p>
                    <p className="text-sm text-zinc-500 max-w-sm">
                      Save your shipping address for one-click checkout across your purchases.
                    </p>
                    <button
                      onClick={() => setIsAddAddressOpen(true)}
                      className="mt-2 inline-flex h-10 items-center justify-center gap-1.5 rounded-xl bg-zinc-900 dark:bg-zinc-100 px-5 text-sm font-medium text-zinc-50 dark:text-zinc-900 shadow hover:opacity-90 cursor-pointer"
                    >
                      <Plus className="w-4 h-4" />
                      <span>Add Your First Address</span>
                    </button>
                  </div>
                ) : (
                  <div className="grid gap-4 sm:grid-cols-2">
                    {addresses.map((addr: any) => (
                      <div
                        key={addr.id}
                        className="p-5 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/50 flex flex-col justify-between"
                      >
                        <div>
                          <div className="flex items-center justify-between mb-3">
                            <span className="text-xs font-semibold uppercase tracking-wider text-zinc-500">
                              {addr.label || "Address"}
                            </span>
                            {addr.is_default && (
                              <Badge variant="default" className="text-[10px] h-5">
                                Default
                              </Badge>
                            )}
                          </div>
                          <p className="font-medium text-sm text-zinc-900 dark:text-zinc-100">
                            {addr.first_name} {addr.last_name}
                          </p>
                          <p className="text-sm text-zinc-600 dark:text-zinc-400 mt-1">
                            {addr.address1}
                            {addr.address2 ? `, ${addr.address2}` : ""}
                          </p>
                          <p className="text-xs text-zinc-500 mt-0.5">
                            {addr.city}, {addr.state} {addr.zip_code || ""}
                          </p>
                          <p className="text-xs text-zinc-500">{addr.country}</p>
                          {addr.phone && <p className="text-xs text-zinc-500 mt-1 font-mono">{addr.phone}</p>}
                        </div>

                        <div className="flex items-center justify-between border-t border-zinc-200/60 dark:border-zinc-800 pt-3 mt-4">
                          {!addr.is_default ? (
                            <button
                              onClick={() => setDefaultAddressMutation.mutate({ id: addr.id, data: {} })}
                              className="text-xs text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors cursor-pointer"
                            >
                              Set as default
                            </button>
                          ) : (
                            <span className="text-xs text-zinc-400">Primary delivery</span>
                          )}
                          <button
                            onClick={() => deleteAddressMutation.mutate({ id: addr.id })}
                            className="text-zinc-400 hover:text-red-500 transition-colors p-1"
                            title="Delete address"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="settings" className="m-0">
            <Card className="border-zinc-200 dark:border-zinc-800 shadow-sm">
              <CardHeader>
                <CardTitle>Account Settings</CardTitle>
                <CardDescription>Update your personal information and contact details.</CardDescription>
              </CardHeader>
              <CardContent>
                {meLoading ? (
                  <div className="py-12 flex justify-center items-center">
                    <Loader2 className="w-6 h-6 animate-spin text-zinc-400" />
                  </div>
                ) : (
                  <form onSubmit={handleUpdateProfile} className="space-y-6 max-w-lg">
                    {profileSuccess && (
                      <Alert className="bg-emerald-50 text-emerald-800 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-900">
                        <CheckCircle2 className="w-4 h-4 mr-2 text-emerald-600 dark:text-emerald-400" />
                        <AlertDescription>Your account profile was updated successfully.</AlertDescription>
                      </Alert>
                    )}

                    {profileError && (
                      <Alert variant="destructive">
                        <AlertCircle className="w-4 h-4 mr-2" />
                        <AlertDescription>{profileError}</AlertDescription>
                      </Alert>
                    )}

                    <div>
                      <label className="text-xs font-medium text-zinc-500 mb-1.5 block">Email Address</label>
                      <input
                        type="email"
                        disabled
                        value={customerMe?.email || ""}
                        className="h-11 w-full rounded-xl border border-zinc-200 dark:border-zinc-800 bg-zinc-100 dark:bg-zinc-900 px-4 text-sm text-zinc-500 cursor-not-allowed"
                      />
                      <span className="text-[11px] text-zinc-400 mt-1 block">Email address cannot be changed.</span>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="text-xs font-medium text-zinc-500 mb-1.5 block">First Name</label>
                        <input
                          type="text"
                          required
                          value={editFirstName}
                          onChange={(e) => setEditFirstName(e.target.value)}
                          className="h-11 w-full rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-4 text-sm text-zinc-900 dark:text-zinc-100"
                        />
                      </div>
                      <div>
                        <label className="text-xs font-medium text-zinc-500 mb-1.5 block">Last Name</label>
                        <input
                          type="text"
                          required
                          value={editLastName}
                          onChange={(e) => setEditLastName(e.target.value)}
                          className="h-11 w-full rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-4 text-sm text-zinc-900 dark:text-zinc-100"
                        />
                      </div>
                    </div>

                    <div>
                      <label className="text-xs font-medium text-zinc-500 mb-1.5 block">Phone Number</label>
                      <input
                        type="tel"
                        value={editPhone}
                        onChange={(e) => setEditPhone(e.target.value)}
                        placeholder="+20 100 000 0000"
                        className="h-11 w-full rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 px-4 text-sm text-zinc-900 dark:text-zinc-100"
                      />
                    </div>

                    <button
                      type="submit"
                      disabled={updateProfileMutation.isPending}
                      className="inline-flex h-11 items-center justify-center rounded-xl bg-zinc-900 dark:bg-zinc-100 px-6 text-sm font-semibold text-zinc-50 dark:text-zinc-900 shadow hover:opacity-90 transition-opacity disabled:opacity-50 cursor-pointer"
                    >
                      {updateProfileMutation.isPending ? "Saving..." : "Save Changes"}
                    </button>
                  </form>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  )
}
