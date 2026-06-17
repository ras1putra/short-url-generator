"use client";

import { useCallback } from "react";
import { useConnection, useConnect, useConnectors } from "wagmi";
import { useConfigStore } from "@/store/useConfigStore";
import { getEthereum } from "@/lib/ethereum";
import { useChainCheck } from "@/hooks/useChainCheck";

export function useWalletConnection() {
  const { address, isConnected, connector } = useConnection();
  const { mutateAsync, isPending: isConnecting } = useConnect();
  const connectors = useConnectors();
  const { switchToCorrectChain } = useChainCheck();

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

    await switchToCorrectChain();

    return activeConnector;
  }, [isConnected, connector, mutateAsync, connectors, switchToCorrectChain]);

  return { connectWallet, addToken, isConnecting, isConnected, address };
}
