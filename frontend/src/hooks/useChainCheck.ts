"use client";

import { useCallback, useEffect, useState } from "react";
import { useAccount } from "wagmi";
import { useConfigStore } from "@/store/useConfigStore";
import { getEthereum } from "@/lib/ethereum";
import { CHAIN_NOT_ADDED } from "@/lib/constants";

export function useChainCheck() {
  const { isConnected } = useAccount();
  const cfg = useConfigStore((s) => s.config);
  const [walletChainId, setWalletChainId] = useState<number | null>(null);

  const targetChainId = cfg?.payment_chain?.chain_id ?? null;
  const targetChainName = cfg?.payment_chain?.chain_name ?? "";

  useEffect(() => {
    if (!isConnected || typeof window === "undefined") {
      setTimeout(() => setWalletChainId(null), 0);
      return;
    }

    const ethereum = getEthereum() as {
      request?: (args: { method: string }) => Promise<string | number>;
      on?: (event: string, handler: (...args: unknown[]) => void) => void;
      removeListener?: (event: string, handler: (...args: unknown[]) => void) => void;
    } | undefined;

    if (!ethereum?.request) return;

    ethereum.request({ method: "eth_chainId" })
      .then((id) => {
        setWalletChainId(parseInt(id as string, 16));
      })
      .catch(() => {});

    const handleChainChanged = (chainId: unknown) => {
      setWalletChainId(parseInt(chainId as string, 16));
    };

    ethereum.on?.("chainChanged", handleChainChanged);

    return () => {
      ethereum.removeListener?.("chainChanged", handleChainChanged);
      setTimeout(() => setWalletChainId(null), 0);
    };
  }, [isConnected]);

  const isCorrectChain = targetChainId === null 
    || !isConnected 
    || walletChainId === null 
    || walletChainId === targetChainId;

  const switchToCorrectChain = useCallback(async () => {
    const ethereum = getEthereum();
    const config = useConfigStore.getState().config;
    if (!ethereum || !config?.payment_chain) return;

    const chainIdHex = `0x${config.payment_chain.chain_id.toString(16)}`;
    try {
      await ethereum.request({ method: "wallet_switchEthereumChain", params: [{ chainId: chainIdHex }] });
    } catch (e: unknown) {
      if ((e as { code?: number }).code === CHAIN_NOT_ADDED) {
        await ethereum.request({
          method: "wallet_addEthereumChain",
          params: [{
            chainId: chainIdHex,
            chainName: config.payment_chain.chain_name,
            rpcUrls: [config.payment_chain.rpc_url],
            blockExplorerUrls: config.payment_chain.explorer_url ? [config.payment_chain.explorer_url] : undefined,
            nativeCurrency: config.payment_chain.currency,
          }],
        });
      } else {
        throw e;
      }
    }
  }, []);

  return { isCorrectChain, targetChainId, targetChainName, switchToCorrectChain };
}
