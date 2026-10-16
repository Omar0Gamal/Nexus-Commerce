"use client";

import React, { useState } from "react";
import { useParams } from "next/navigation";
import { customInstance } from "@/lib/axios";

export default function TrackOrderPage() {
  const params = useParams();
  const domain = params.domain as string;

  const [trackingNumber, setTrackingNumber] = useState("");
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleTrack = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setStatus(null);
    setLoading(true);

    try {
      const shopResponse = await customInstance<{ data: { id: string, name: string } }>({
        url: `/api/v1/storefront/shop/${domain}`,
        method: "GET",
      });
      
      const shopId = shopResponse.data.id;

      const trackResponse = await customInstance<{ data: { status: string } }>({
        url: `/api/v1/storefront/orders/track?tracking_number=${trackingNumber}&email=${encodeURIComponent(email)}`,
        method: "GET",
        headers: {
          "X-Shop-ID": shopId,
          "X-Gateway-Secret": "nexus-dev-secret-2026",
        },
      });

      setStatus(trackResponse.data.status);
    } catch (err: any) {
      setError(err?.response?.data?.error || err?.response?.data?.message || "Order not found or access denied.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-xl mx-auto py-12">
      <h1 className="text-3xl font-bold mb-8 text-center">Track Your Order</h1>
      <form onSubmit={handleTrack} className="space-y-6 bg-white dark:bg-zinc-900 p-8 rounded-xl shadow-sm border border-zinc-200 dark:border-zinc-800">
        <div>
          <label htmlFor="trackingNumber" className="block text-sm font-medium mb-2">Order / Tracking Number</label>
          <input
            id="trackingNumber"
            type="text"
            required
            className="w-full px-4 py-2 border rounded-md dark:bg-zinc-800 dark:border-zinc-700 focus:ring-2 focus:ring-zinc-900 outline-none"
            value={trackingNumber}
            onChange={(e) => setTrackingNumber(e.target.value)}
            placeholder="e.g. 1001"
          />
        </div>
        <div>
          <label htmlFor="email" className="block text-sm font-medium mb-2">Email Address</label>
          <input
            id="email"
            type="email"
            required
            className="w-full px-4 py-2 border rounded-md dark:bg-zinc-800 dark:border-zinc-700 focus:ring-2 focus:ring-zinc-900 outline-none"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="email@example.com"
          />
        </div>
        <button
          type="submit"
          disabled={loading}
          className="w-full py-3 px-4 bg-zinc-900 hover:bg-zinc-800 text-white font-medium rounded-md disabled:opacity-50 transition-colors"
        >
          {loading ? "Searching..." : "Track Order"}
        </button>
      </form>

      {error && (
        <div className="mt-8 p-4 bg-red-50 text-red-600 rounded-md border border-red-100 text-center">
          {error}
        </div>
      )}

      {status && (
        <div className="mt-8 p-6 bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-300 rounded-md border border-green-200 dark:border-green-900/50 text-center space-y-2">
          <h2 className="text-lg font-medium">Order Found</h2>
          <p className="text-2xl font-bold uppercase tracking-wider">{status}</p>
        </div>
      )}
    </div>
  );
}
