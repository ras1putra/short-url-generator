"use client";

import { AlertTriangle } from "lucide-react";
import { useChainCheck } from "@/hooks/useChainCheck";

export default function WrongChainBanner() {
  const { isCorrectChain, targetChainName, switchToCorrectChain } = useChainCheck();

  if (isCorrectChain) return null;

  return (
    <div className="flex items-center gap-3 px-4 py-3 bg-amber-500/10 border-b border-amber-500/20">
      <AlertTriangle size={16} className="text-amber-400 shrink-0" />
      <p className="text-sm text-amber-300/90 flex-1">
        Wrong network detected. Please switch to <span className="font-bold">{targetChainName}</span>.
      </p>
      <button
        onClick={switchToCorrectChain}
        className="text-xs font-bold uppercase tracking-wider px-3 py-1.5 rounded-lg bg-amber-500/20 hover:bg-amber-500/30 text-amber-300 transition-colors cursor-pointer shrink-0"
      >
        Switch
      </button>
    </div>
  );
}
