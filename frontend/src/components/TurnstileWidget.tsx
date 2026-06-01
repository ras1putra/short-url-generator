"use client";

import { useEffect, useRef, useCallback } from "react";
import Script from "next/script";
import { useConfigStore } from "@/store/useConfigStore";

export default function TurnstileWidget() {
  const siteKey = useConfigStore((s) => s.config?.turnstile_site_key);
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetIdRef = useRef<string | null>(null);

  const initializeTurnstile = useCallback(() => {
    if (!siteKey || !containerRef.current || !window.turnstile) {
      return;
    }

    // Clean up any existing widget first
    if (widgetIdRef.current) {
      try {
        window.turnstile.remove(widgetIdRef.current);
      } catch {
        // Ignore errors if widget was already removed
      }
      widgetIdRef.current = null;
    }

    try {
      widgetIdRef.current = window.turnstile.render(containerRef.current, {
        sitekey: siteKey,
        theme: "dark",
      });
    } catch {
    }
  }, [siteKey]);

  // Run on mount, unmount, or when siteKey changes
  useEffect(() => {
    if (typeof window !== "undefined" && window.turnstile && siteKey) {
      initializeTurnstile();
    }

    return () => {
      if (widgetIdRef.current && window.turnstile) {
        try {
          window.turnstile.remove(widgetIdRef.current);
        } catch {
          // Ignore unmount cleanup errors
        }
        widgetIdRef.current = null;
      }
    };
  }, [siteKey, initializeTurnstile]);

  if (!siteKey) {
    return null;
  }

  return (
    <>
      <Script
        src="https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit"
        strategy="afterInteractive"
        onReady={initializeTurnstile}
      />
      <div ref={containerRef} className="my-4 min-h-[65px] flex justify-center items-center" />
    </>
  );
}
