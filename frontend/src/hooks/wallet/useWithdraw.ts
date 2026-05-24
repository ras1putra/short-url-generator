"use client";

import { useState, useCallback } from "react";
import { createWalletClient, custom, type EIP1193Provider } from "viem";
import { WITHDRAWER_ABI } from "@/lib/paymentGateway";
import { useConfigStore } from "@/store/useConfigStore";
import { definePaymentChain } from "@/lib/wagmi";
import { useWalletConnection } from "./useWalletConnection";
import { useConnection } from "wagmi";
import { api } from "@/lib/api";
import { API_WALLET } from "@/lib/constants";

export type WithdrawStatus = "idle" | "signing" | "pending" | "success";

export interface WithdrawalPermit {
  wallet: string;
  amount: string;
  nonce: string;
  deadline: string;
  signature: string;
  contract: string;
  chain_id: number;
}

export function useWithdraw() {
  const { connectWallet } = useWalletConnection();
  const { address } = useConnection();
  const [status, setStatus] = useState<WithdrawStatus>("idle");
  const [txHash, setTxHash] = useState<string | null>(null);

  const withdraw = useCallback(async (amountInETH: string): Promise<string> => {
    const cfg = useConfigStore.getState().config;
    if (!cfg?.contract_withdrawer || !cfg?.payment_chain) {
      throw new Error("Web3 config not loaded properly");
    }

    setStatus("signing");

    const connector = await connectWallet();
    const provider = await connector!.getProvider();
    const chain = definePaymentChain(cfg.payment_chain);

    const walletClient = createWalletClient({
      transport: custom(provider as EIP1193Provider),
      chain,
    });

    const account = address as `0x${string}` | undefined;
    if (!account) throw new Error("No account found");

    // 1. Get signed permit from backend
    const permitResp = await api.post(`${API_WALLET}/withdraw`, {
      amount: amountInETH,
      wallet_addr: account,
    });

    const permit = permitResp.data.data as WithdrawalPermit;

    if (!permit || !permit.signature) {
      throw new Error("Failed to get withdrawal permit from server");
    }

    // 2. Submit withdrawal to the contract
    setStatus("pending");

    const hash = await walletClient.writeContract({
      address: cfg.contract_withdrawer as `0x${string}`,
      abi: WITHDRAWER_ABI,
      functionName: "withdraw",
      args: [
        BigInt(permit.amount),
        BigInt(permit.nonce),
        BigInt(permit.deadline),
        permit.signature as `0x${string}`,
      ],
      account,
      chain,
    });

    setTxHash(hash);
    setStatus("success");
    return hash;
  }, [connectWallet, address]);

  return { withdraw, status, txHash };
}
