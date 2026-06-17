import { http, createConfig } from "wagmi";
import { injected, metaMask } from "wagmi/connectors";
import { defineChain, type Chain } from "viem";
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

  return createConfig({
    chains: [dynamicChain],
    connectors: [injected(), metaMask()],
    transports: {
      [dynamicChain.id]: http(appConfig.payment_chain.rpc_url),
    },
    ssr: true,
  });
}
