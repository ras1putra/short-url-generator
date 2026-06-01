"use client";

import { useState, useCallback } from "react";
import { createWalletClient, createPublicClient, custom, parseEther, type EIP1193Provider } from "viem";
import { PAYMENT_GATEWAY_ABI, ERC20_ABI } from "@/lib/paymentGateway";
import { useConfigStore } from "@/store/useConfigStore";
import { definePaymentChain } from "@/lib/wagmi";
import { useWalletConnection } from "./useWalletConnection";
import { useConnection } from "wagmi";

export type DepositStatus = "idle" | "pending" | "success";

export function useDeposit() {
  const { connectWallet } = useWalletConnection();
  const { address } = useConnection();
  const [status, setStatus] = useState<DepositStatus>("idle");

  const getOnChainBalance = useCallback(async (): Promise<string> => {
    const cfg = useConfigStore.getState().config;
    if (!cfg?.payment_chain || !cfg?.contract_token) return "0.00";

    try {
      const connector = await connectWallet();
      const provider = await connector!.getProvider();
      const chain = definePaymentChain(cfg.payment_chain);

      const publicClient = createPublicClient({
        chain,
        transport: custom(provider as EIP1193Provider),
      });

      const account = address as `0x${string}` | undefined;
      if (!account) return "0.00";

      const balance = await publicClient.readContract({
        address: cfg.contract_token as `0x${string}`,
        abi: ERC20_ABI,
        functionName: "balanceOf",
        args: [account],
      });

      return (Number(balance) / 1e18).toFixed(2);
    } catch (e) {
      console.error("Failed to fetch token balance from wallet", e);
      return "0.00";
    }
  }, [connectWallet, address]);

  const deposit = useCallback(async (refId: `0x${string}`, amountInETH: string) => {
    const cfg = useConfigStore.getState().config;
    if (!cfg?.payment_chain || !cfg?.contract_payment || !cfg?.contract_token) {
      throw new Error("Web3 config not loaded properly");
    }

    setStatus("pending");
    try {
      const connector = await connectWallet();

      const provider = await connector!.getProvider();
      const chain = definePaymentChain(cfg.payment_chain);

      const walletClient = createWalletClient({
        transport: custom(provider as EIP1193Provider),
        chain,
      });

      const publicClient = createPublicClient({
        chain,
        transport: custom(provider as EIP1193Provider),
      });

      const accounts = await connector.getAccounts();
      const account = accounts[0] as `0x${string}` | undefined;
      if (!account) throw new Error("No account found");

      const amountBig = parseEther(amountInETH);

      const currentAllowance = await publicClient.readContract({
        address: cfg.contract_token as `0x${string}`,
        abi: ERC20_ABI,
        functionName: "allowance",
        args: [account, cfg.contract_payment as `0x${string}`],
      });

      if (currentAllowance < amountBig) {
        const approveHash = await walletClient.writeContract({
          address: cfg.contract_token as `0x${string}`,
          abi: ERC20_ABI,
          functionName: "approve",
          args: [cfg.contract_payment as `0x${string}`, amountBig],
          account,
          chain,
        });

        await publicClient.waitForTransactionReceipt({ hash: approveHash });
      }

      // Call PaymentGateway deposit
      const hash = await walletClient.writeContract({
        address: cfg.contract_payment as `0x${string}`,
        abi: PAYMENT_GATEWAY_ABI,
        functionName: "deposit",
        args: [refId, amountBig],
        account,
        chain,
      });

      setStatus("success");
      return hash;
    } catch (e) {
      setStatus("idle");
      throw e;
    }
  }, [connectWallet, address]);

  return { deposit, status, getOnChainBalance };
}
