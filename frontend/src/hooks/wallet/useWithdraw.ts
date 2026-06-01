"use client";

import { useState, useCallback } from "react";
import { api } from "@/lib/api";
import { API_WALLET } from "@/lib/constants";

export type WithdrawStatus = "idle" | "pending" | "success";

export function useWithdraw() {
  const [status, setStatus] = useState<WithdrawStatus>("idle");
  const [txHash, setTxHash] = useState<string | null>(null);

  const withdraw = useCallback(async (amountInETH: string, walletAddr: string): Promise<{ hash: string }> => {
    setStatus("pending");
    try {
      const resp = await api.post(`${API_WALLET}/withdraw`, {
        amount: amountInETH,
        wallet_addr: walletAddr,
      });

      const txHash = resp.data.data.tx_hash as string;
      setTxHash(txHash);
      setStatus("success");
      return { hash: txHash };
    } catch (e) {
      setStatus("idle");
      throw e;
    }
  }, []);

  return { withdraw, status, txHash };
}
