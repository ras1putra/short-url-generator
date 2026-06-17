"use client";

import { useState, useEffect } from "react";
import dynamic from "next/dynamic";
import { Sparkles, Loader2, Check } from "lucide-react";
import { useNFTPass, ACTION_APPROVE, ACTION_MINT } from "@/hooks/useNFTPass";
import { useConfig } from "@/hooks/useConfig";
import { useChainCheck } from "@/hooks/useChainCheck";

const NFTCard3D = dynamic(() => import("./NFTCard3D"), {
  ssr: false,
  loading: () => (
    <div className="w-full h-[320px] md:h-[400px] flex flex-col items-center justify-center bg-white/[0.01] rounded-2xl border border-white/[0.05]">
      <div className="w-8 h-8 border-2 border-[#6EE7B7]/30 border-t-[#6EE7B7] rounded-full animate-spin mb-2" />
      <span className="text-xs font-mono-dm text-white/40 uppercase tracking-widest">Loading Viewer...</span>
    </div>
  ),
});

function NFTMintButton() {
  const [mounted, setMounted] = useState(false);
  const {
    isConnected,
    isConnecting,
    isConfigLoading,
    isError,
    tokenSymbol,
    displayPrice,
    hasNFT,
    isBusy,
    pendingAction,
    handleMintPass,
  } = useNFTPass();
  const { isCorrectChain, targetChainName, switchToCorrectChain } = useChainCheck();

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMounted(true);
  }, []);

  const isLoading = !mounted || isConfigLoading;

  const buttonState = (() => {
    if (isLoading) return { text: "Preparing Mint...", icon: <Loader2 size={16} className="animate-spin" />, disabled: true, onClick: undefined };
    if (isError) return { text: "Minting Unavailable", icon: <Sparkles size={16} />, disabled: true, onClick: undefined };
    if (hasNFT) return { text: "Pass Active (Already Owned)", icon: <Check size={16} className="text-[#0A0A0A]" />, disabled: true, onClick: undefined };
    if (isConnecting) return { text: "Connecting Wallet...", icon: <Loader2 size={16} className="animate-spin" />, disabled: true, onClick: undefined };
    if (pendingAction === ACTION_APPROVE) return { text: `Approving ${tokenSymbol}...`, icon: <Loader2 size={16} className="animate-spin" />, disabled: true, onClick: undefined };
    if (pendingAction === ACTION_MINT) return { text: "Minting Pass...", icon: <Loader2 size={16} className="animate-spin" />, disabled: true, onClick: undefined };
    if (!isConnected) return { text: "Connect & Mint Pass", icon: <Sparkles size={16} />, disabled: false, onClick: handleMintPass };
    if (!isCorrectChain) return { text: `Switch to ${targetChainName}`, icon: <Sparkles size={16} />, disabled: false, onClick: switchToCorrectChain };
    return { text: `Mint Pass (${displayPrice} ${tokenSymbol})`, icon: <Sparkles size={16} />, disabled: isBusy, onClick: handleMintPass };
  })();

  const baseBtnClasses = "px-8 py-4 rounded-xl text-base font-extrabold tracking-tight inline-flex items-center justify-center gap-2 transition-all duration-300 w-full sm:w-auto";
  const buttonClasses = hasNFT
    ? `${baseBtnClasses} bg-white/20 text-white/50 border border-white/10 cursor-not-allowed`
    : buttonState.disabled
      ? `${baseBtnClasses} bg-white/5 text-white/40 border border-white/5 cursor-not-allowed animate-pulse`
      : `${baseBtnClasses} bg-[#6EE7B7] text-[#0A0A0A] hover:bg-[#34D399] hover:-translate-y-0.5 hover:shadow-[0_8px_30px_rgba(110,231,183,0.3)] cursor-pointer`;

  return (
    <button
      disabled={buttonState.disabled}
      className={buttonClasses}
      onClick={buttonState.onClick}
    >
      {buttonState.icon}
      <span>{buttonState.text}</span>
    </button>
  );
}

export default function NFTSection() {
  const { data: appConfig, isError } = useConfig();

  return (
    <section id="nft-pass"className="relative py-12 sm:py-16 lg:py-24 px-4 sm:px-6 md:px-12 max-w-7xl mx-auto overflow-hidden">
      {/* Background glow blobs */}
      <div className="absolute top-1/2 left-1/4 w-80 h-80 rounded-full pointer-events-none bg-[radial-gradient(circle,rgba(110,231,183,0.04)_0%,transparent_70%)] -translate-y-1/2" />

      <div className="grid grid-cols-1 md:grid-cols-12 gap-12 md:gap-16 items-center">
        {/* Left Column: 3D Showcase Card */}
        <div className="md:col-span-5 flex items-center justify-center md:justify-start relative">
          <div className="absolute w-72 h-72 rounded-full border border-white/[0.03] animate-pulse pointer-events-none" />

          <div className="w-full relative max-w-md rounded-3xl overflow-hidden bg-gradient-to-b from-white/[0.04] to-white/[0.005] border border-white/[0.08] p-4 shadow-[0_24px_50px_-12px_rgba(0,0,0,0.5)] hover:border-[#6EE7B7]/30 transition-all duration-300 group">
            <NFTCard3D />
          </div>
        </div>

        {/* Right Column: Copy and CTA */}
        <div className="md:col-span-7 space-y-4 sm:space-y-6 text-left">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-mono-dm border border-[#6EE7B7]/30 text-[#6EE7B7] bg-[#6EE7B7]/5">
            <Sparkles size={12} className="animate-pulse" />
            <span>MINT PASS</span>
          </div>

          <h2 className="text-4xl md:text-5xl font-black tracking-tight leading-tight text-glow text-white">
            BYPASS ADS WITH<br />
            <span className="stat-number">THE NFT PASS</span>
          </h2>

          <p className="text-white/70 text-base md:text-lg leading-relaxed max-w-lg">
            Hold this pass in your wallet to skip redirect ads instantly. Verification is gasless and takes a single click.
          </p>

          <div className="pt-2">
            {appConfig ? (
              <NFTMintButton />
            ) : (
              <button
                disabled
                className="px-8 py-4 rounded-xl text-base font-extrabold tracking-tight inline-flex items-center gap-2 transition-all duration-300 bg-white/5 text-white/40 border border-white/5 cursor-not-allowed animate-pulse"
              >
                {isError ? <Sparkles size={16} /> : <Loader2 size={16} className="animate-spin" />}
                <span>{isError ? "Minting Unavailable" : "Preparing Mint..."}</span>
              </button>
            )}

            {process.env.NEXT_PUBLIC_SWAP_LINK && (
              <p className="mt-3 text-xs text-white/40 font-mono-dm text-center sm:text-left">
                {"//"} Don&apos;t have enough {appConfig?.token_symbol ?? "tokens"}?{" "}
                <a
                  href={process.env.NEXT_PUBLIC_SWAP_LINK}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-[#6EE7B7] hover:text-[#34D399] transition-colors underline underline-offset-2"
                >
                  Buy it here
                </a>
              </p>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
