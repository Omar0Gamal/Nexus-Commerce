"use client";

import { useEffect } from "react";
import { usePathname } from "next/navigation";
import { customInstance } from "@/lib/axios";

export function AnalyticsTracker() {
  const pathname = usePathname();

  useEffect(() => {
    const host = window.location.host;
    const trackEvent = async () => {
      try {
        let shopDomain = host;
        
        if (pathname?.startsWith("/")) {
          const parts = pathname.split("/");
          if (parts.length > 1 && parts[1]) {
            shopDomain = parts[1];
          }
        }
        
        const sessionId = localStorage.getItem("nexus_session_id") || crypto.randomUUID();
        localStorage.setItem("nexus_session_id", sessionId);

        await customInstance({
          url: "/api/v1/analytics/ingest",
          method: "POST",
          data: {
            event: "page_view",
            shop_id: shopDomain,
            session_id: sessionId,
            payload: {
              pathname: pathname,
              referrer: document.referrer,
              userAgent: navigator.userAgent
            }
          }
        });
      } catch (err) {
        console.error("Analytics tracking failed:", err);
      }
    };

    trackEvent();
  }, [pathname]);

  return null;
}
