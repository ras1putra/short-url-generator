import { http, createConfig } from "wagmi";
import { defineChain, type Chain } from "viem";
import { connectorsForWallets } from "@rainbow-me/rainbowkit";
import {
  rainbowWallet,
  metaMaskWallet,
  walletConnectWallet,
  okxWallet,
  safeWallet,
  trustWallet,
  rabbyWallet,
  ledgerWallet,
  zerionWallet,
} from "@rainbow-me/rainbowkit/wallets";
import type { AppConfig } from "./config";

export function definePaymentChain(payment_chain: AppConfig["payment_chain"]): Chain {
  return defineChain({
    id: payment_chain.chain_id,
    name: payment_chain.chain_name,
    nativeCurrency: payment_chain.currency,
    rpcUrls: {
      default: { http: [payment_chain.rpc_url] },
    },
    blockExplorers: payment_chain.explorer_url
      ? {
        default: {
          name: "Explorer",
          url: payment_chain.explorer_url,
        },
      }
      : undefined,
  });
}

export function getWagmiConfig(appConfig?: AppConfig | null) {
  if (!appConfig) {
    throw new Error("Web3 configuration is missing. Internal server error, please contact administrators.");
  }

  const dynamicChain = definePaymentChain(appConfig.payment_chain);
  const projectId = process.env.NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID || "7d5df2f7cde7b483c66d21469e8e01bd";

  const connectors = connectorsForWallets(
    [
      {
        groupName: "Popular",
        wallets: [
          rainbowWallet,
          metaMaskWallet,
          okxWallet,
          safeWallet,
          trustWallet,
          rabbyWallet,
          ledgerWallet,
          zerionWallet,
          walletConnectWallet,
        ],
      },
    ],
    { appName: "Short URL Generator", projectId }
  );

  return createConfig({
    chains: [dynamicChain],
    connectors,
    transports: {
      [dynamicChain.id]: http(appConfig.payment_chain.rpc_url),
    },
    ssr: true,
  });
}
