"use client";

import { ReactNode, useMemo } from "react";
import { WagmiProvider } from "wagmi";
import { RainbowKitProvider, darkTheme } from "@rainbow-me/rainbowkit";
import "@rainbow-me/rainbowkit/styles.css";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getWagmiConfig } from "@/lib/wagmi";
import { useConfig } from "@/hooks/useConfig";
import { AlertCircle, RefreshCw } from "lucide-react";
import { usePathname } from "next/navigation";

const queryClient = new QueryClient();

function WagmiAppWrapper({ children }: { children: ReactNode }) {
  const { isError, data: appConfig } = useConfig();
  const pathname = usePathname();

  const wagmiConfig = useMemo(() => {
    if (!appConfig) return null;
    try {
      return getWagmiConfig(appConfig);
    } catch (e) {
      console.error("Failed to generate Wagmi config:", e);
      return null;
    }
  }, [appConfig]);

  const isDashboard = pathname?.startsWith("/dashboard");

  if (isError) {
    if (isDashboard) {
      return (
        <div className="flex h-screen w-screen items-center justify-center bg-[#0A0A0A] p-4 sm:p-6">
          <div className="max-w-md w-full rounded-2xl bg-white/[0.02] border border-white/[0.08] overflow-hidden">
            <div className="p-8 text-center">
              <div className="mx-auto h-12 w-12 rounded-full bg-red-500/10 flex items-center justify-center mb-4">
                <AlertCircle className="h-6 w-6 text-red-400" />
              </div>
              <h2 className="text-xl font-black tracking-tight text-white mb-2">Config failed to load</h2>
              <p className="text-sm text-white/50 font-mono-dm">
                Could not connect to the server. Please try again.
              </p>
            </div>
            <div className="border-t border-white/[0.06] px-8">
              <button
                onClick={() => window.location.reload()}
                className="btn-primary w-full flex items-center justify-center gap-2 px-4 py-2.5 text-sm tracking-wider uppercase cursor-pointer"
              >
                <RefreshCw className="h-4 w-4" />
                Retry
              </button>
            </div>
          </div>
        </div>
      );
    }
    return <>{children}</>;
  }

  if (!appConfig || !wagmiConfig) {
    if (isDashboard) {
      return (
        <div className="flex h-screen w-screen items-center justify-center bg-[#0A0A0A]">
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="h-10 w-10 animate-spin rounded-full border-[3px] border-[#6EE7B7] border-t-transparent" />
            <p className="text-sm text-white/50 animate-pulse font-mono-dm">Loading configuration...</p>
          </div>
        </div>
      );
    }
    return <>{children}</>;
  }

  return (
    <WagmiProvider config={wagmiConfig}>
      <RainbowKitProvider theme={darkTheme()}>
        {children}
      </RainbowKitProvider>
    </WagmiProvider>
  );
}

export default function Providers({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <WagmiAppWrapper>
        {children}
      </WagmiAppWrapper>
    </QueryClientProvider>
  );
}
