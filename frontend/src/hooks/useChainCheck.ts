"use client";

import { useCallback } from "react";
import { useConnection } from "wagmi";
import { useConfigStore } from "@/store/useConfigStore";
import { getEthereum } from "@/lib/ethereum";
import { CHAIN_NOT_ADDED } from "@/lib/constants";

export function useChainCheck() {
  const { chainId } = useConnection();
  const cfg = useConfigStore.getState().config;
  const targetChainId = cfg?.payment_chain?.chain_id ?? null;
  const targetChainName = cfg?.payment_chain?.chain_name ?? "";
  const isCorrectChain = targetChainId === null || chainId === targetChainId;

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
