"use client";

import Script from "next/script";
import { useConfigStore } from "@/store/useConfigStore";

export default function TurnstileWidget() {
  const siteKey = useConfigStore((s) => s.config?.turnstile_site_key);

  if (!siteKey) return null;

  return (
    <>
      <Script
        src="https://challenges.cloudflare.com/turnstile/v0/api.js"
        strategy="afterInteractive"
      />
      <div className="cf-turnstile" data-sitekey={siteKey} data-theme="dark" />
    </>
  );
}
