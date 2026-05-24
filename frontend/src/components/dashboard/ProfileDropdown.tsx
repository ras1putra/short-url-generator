"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { useLogout } from "@/hooks/useAuth";
import { useUserStore } from "@/store/useUserStore";
import { useBalance, useDisconnect } from "wagmi";
import { useWalletConnection } from "@/hooks/wallet/useWalletConnection";
import { LogOut, Settings, ChevronDown, Copy, Wallet, Check } from "lucide-react";
import { AxiosError } from "axios";
import { toast } from "sonner";
import { classifyWalletError } from "@/lib/wallet";
import type { ApiErrorResponse } from "@/types/api";
import { ROUTE_SETTINGS } from "@/lib/constants";


export default function ProfileDropdown() {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const logoutMutation = useLogout();
  const user = useUserStore((state) => state.user);

  const { connectWallet, isConnecting, isConnected, address } = useWalletConnection();
  const { data: balanceData } = useBalance({ address });
  const { mutate: disconnectWallet } = useDisconnect();

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const handleCopy = async () => {
    if (!address) return;
    await navigator.clipboard.writeText(address);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  const handleConnect = async () => {
    try {
      await connectWallet();
    } catch (e) {
      toast.error(classifyWalletError(e, "connect"));
    }
  };

  const handleDisconnect = () => {
    disconnectWallet();
    setOpen(false);
  };

  const displayName = user?.name || user?.email?.split("@")[0] || "User";
  const initial = displayName[0].toUpperCase();
  const formattedBalance = balanceData
    ? `${(Number(balanceData.value) / 10 ** balanceData.decimals).toFixed(4)} ${balanceData.symbol}`
    : null;

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2 text-white/50 hover:text-white px-3 py-2 rounded-lg text-sm font-medium transition-colors cursor-pointer"
      >
        <div className="h-8 w-8 rounded-full bg-[#6EE7B7]/20 flex items-center justify-center text-sm font-bold text-[#6EE7B7]">
          {initial}
        </div>
        <ChevronDown size={14} className={`transition-transform ${open ? "rotate-180" : ""}`} />
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-2 w-72 rounded-xl border border-white/[0.08] bg-[#0A0A0A] shadow-2xl overflow-hidden z-50">
          <div className="p-4 border-b border-white/[0.06]">
            <p className="text-sm font-medium text-white">{displayName}</p>
            <p className="text-sm text-white/40 mt-0.5">{user?.email}</p>
          </div>

          <div className="p-4 border-b border-white/[0.06]">
            <p className="text-sm font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-2">Wallet</p>
            {isConnected && address ? (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-mono text-white/70">
                    {address.slice(0, 6)}...{address.slice(-4)}
                  </span>
                  <button onClick={handleCopy} className="text-white/30 hover:text-white/70 transition-colors cursor-pointer">
                    {copied ? <Check size={14} /> : <Copy size={14} />}
                  </button>
                </div>
                <div className="flex items-center gap-2 text-sm text-white/60">
                  <Wallet size={14} />
                  <span>{formattedBalance || "—"}</span>
                </div>
                <button
                  onClick={handleDisconnect}
                  className="text-sm text-white/30 hover:text-red-400 transition-colors cursor-pointer"
                >
                  Disconnect
                </button>
              </div>
            ) : (
              <button
                onClick={handleConnect}
                disabled={isConnecting}
                className="flex items-center gap-2 text-sm font-medium text-[#6EE7B7] hover:text-[#6EE7B7]/80 transition-colors cursor-pointer disabled:opacity-50"
              >
                <Wallet size={14} />
                {isConnecting ? "Connecting..." : "Connect Wallet"}
              </button>
            )}
          </div>

          <div className="p-2">
            <Link
              href={ROUTE_SETTINGS}
              onClick={() => setOpen(false)}
              className="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm text-white/50 hover:text-white hover:bg-white/[0.04] transition-colors"
            >
              <Settings size={16} />
              Settings
            </Link>
            <button
              onClick={() => {
                setOpen(false);
                logoutMutation.mutate(undefined, {
                  onError: (err) => {
                    const axiosErr = err as AxiosError<ApiErrorResponse>;
                    toast.error(axiosErr?.response?.data?.message || "Failed to sign out");
                  },
                });
              }}
              disabled={logoutMutation.isPending}
              className="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm text-white/50 hover:text-white hover:bg-white/[0.04] transition-colors cursor-pointer disabled:opacity-50"
            >
              <LogOut size={16} />
              Sign out
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
