"use client";

import { useCallback } from "react";
import { useConnection, useConnect, useConnectors } from "wagmi";
import { useConfigStore } from "@/store/useConfigStore";
import { CHAIN_NOT_ADDED } from "@/lib/constants";
import { getEthereum } from "@/lib/ethereum";

export function useWalletConnection() {
  const { address, isConnected, connector } = useConnection();
  const { mutateAsync, isPending: isConnecting } = useConnect();
  const connectors = useConnectors();

  const ensureChain = useCallback(async () => {
    const cfg = useConfigStore.getState().config;
    const ethereum = getEthereum();
    if (!ethereum || !cfg?.payment_chain) return;

    const chainIdHex = `0x${cfg.payment_chain.chain_id.toString(16)}`;
    try {
      await ethereum.request({ method: "wallet_switchEthereumChain", params: [{ chainId: chainIdHex }] });
    } catch (e: unknown) {
      if ((e as { code?: number }).code === CHAIN_NOT_ADDED) {
        await ethereum.request({
          method: "wallet_addEthereumChain",
          params: [{
            chainId: chainIdHex,
            chainName: cfg.payment_chain.chain_name,
            rpcUrls: [cfg.payment_chain.rpc_url],
            blockExplorerUrls: cfg.payment_chain.explorer_url ? [cfg.payment_chain.explorer_url] : undefined,
            nativeCurrency: cfg.payment_chain.currency,
          }],
        });
      }
    }
  }, []);

  const addToken = useCallback(async () => {
    const cfg = useConfigStore.getState().config;
    const ethereum = getEthereum();
    if (!ethereum || !cfg?.contract_token) return;

    await ethereum.request({
      method: "wallet_watchAsset",
      params: {
        type: "ERC20",
        options: {
          address: cfg.contract_token,
          symbol: cfg.token_symbol,
          decimals: cfg.token_decimals,
        },
      },
    });
  }, []);

  const connectWallet = useCallback(async () => {
    if (!getEthereum()) {
      throw new Error("No wallet detected. Please install MetaMask or another wallet extension.");
    }

    const activeConnector = connector ?? connectors[0];
    if (!activeConnector) throw new Error("No wallet connector found");
    if (!isConnected) await mutateAsync({ connector: activeConnector });

    await ensureChain();

    return activeConnector;
  }, [isConnected, connector, mutateAsync, connectors, ensureChain]);

  return { connectWallet, addToken, isConnecting, isConnected, address };
}
