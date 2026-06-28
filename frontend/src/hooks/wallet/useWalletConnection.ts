import { useCallback } from "react";
import { useAccount, useConnect } from "wagmi";
import { useConnectModal } from "@rainbow-me/rainbowkit";
import { useConfigStore } from "@/store/useConfigStore";
import { getEthereum } from "@/lib/ethereum";

export function useWalletConnection() {
  const { address, isConnected } = useAccount();
  const { isPending: isConnecting } = useConnect();
  const { openConnectModal } = useConnectModal();

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
    if (openConnectModal) {
      openConnectModal();
    }
  }, [openConnectModal]);

  return { connectWallet, addToken, isConnecting, isConnected, address };
}
