"use client";

import RequireRole from "@/components/auth/RequireRole";
import { useWallet } from "@/hooks/wallet/useWallet";
import { useDeposit } from "@/hooks/wallet/useDeposit";
import { useWithdraw } from "@/hooks/wallet/useWithdraw";
import { useWalletConnection } from "@/hooks/wallet/useWalletConnection";
import TransactionTable from "@/components/wallet/TransactionTable";
import { useUserStore } from "@/store/useUserStore";
import { Wallet, ArrowDownToLine, ArrowUpFromLine, Loader2, Plus } from "lucide-react";
import { useState, useEffect } from "react";
import { toast } from "sonner";
import { padHex } from "viem";
import { useReadContract } from "wagmi";
import { AxiosError } from "axios";
import Decimal from "decimal.js";
import { classifyWalletError, formatBalance } from "@/lib/wallet";
import { useConfigStore } from "@/store/useConfigStore";
import type { ApiErrorResponse } from "@/types/api";
import { useQueryClient } from "@tanstack/react-query";
import { createPublicClient, custom, type EIP1193Provider } from "viem";
import { definePaymentChain } from "@/lib/wagmi";
import { ERC20_ABI } from "@/lib/paymentGateway";
import { api } from "@/lib/api";

import { ROLE_USER, ROLE_ADVERTISER, ROLE_ADMIN, API_WALLET, DEFAULT_PAGE_SIZE, TX_TYPE_DEPOSIT } from "@/lib/constants";

export default function WalletPage() {
  const cfg = useConfigStore((s) => s.config);
  const symbol = cfg?.token_symbol || "TK";
  const queryClient = useQueryClient();

  const [confirmations, setConfirmations] = useState<Record<string, number>>({});
  const [txPage, setTxPage] = useState(1);
  const [txPerPage, setTxPerPage] = useState(DEFAULT_PAGE_SIZE);
  const [txSearch, setTxSearch] = useState("");
  const [txSortBy, setTxSortBy] = useState("created_at");
  const [txSortDir, setTxSortDir] = useState("desc");
  const { data: wallet, isLoading, error, isFetching } = useWallet(txPage, txPerPage, txSearch || undefined, txSortBy, txSortDir);

  const txHandleSort = (col: string) => {
    setTxPage(1);
    if (txSortBy === col) {
      setTxSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setTxSortBy(col);
      setTxSortDir("desc");
    }
  };
  const { deposit, status: depositStatus } = useDeposit();
  const { withdraw: withdrawOnChain, status: withdrawStatus } = useWithdraw();
  const { addToken, address, isConnected, connectWallet } = useWalletConnection();
  const user = useUserStore((s) => s.user);
  const [showDeposit, setShowDeposit] = useState(false);
  const [showWithdraw, setShowWithdraw] = useState(false);
  const [depositAmount, setDepositAmount] = useState("");
  const [withdrawAmount, setWithdrawAmount] = useState("");

  const isAdvertiser = user?.role === ROLE_ADVERTISER || user?.role === ROLE_ADMIN;

  const { data: rawBalance, isFetching: isBalanceFetching } = useReadContract({
    address: cfg?.contract_token as `0x${string}`,
    abi: ERC20_ABI,
    functionName: "balanceOf",
    args: address ? [address as `0x${string}`] : undefined,
    query: {
      refetchInterval: 10_000,
      enabled: !!address && !!cfg?.contract_token,
    },
  });

  const onChainBalance = rawBalance !== undefined ? (Number(rawBalance) / 1e18).toFixed(8) : null;

  useEffect(() => {
    const pendingTxs = wallet?.transactions?.filter((tx) => tx.status === "PENDING") || [];
    if (pendingTxs.length === 0 || !cfg) return;

    let active = true;
    const interval = setInterval(async () => {
      try {
        const connector = await connectWallet();
        if (!connector) return;
        const provider = await connector.getProvider();
        const chain = definePaymentChain(cfg.payment_chain);
        const publicClient = createPublicClient({
          chain,
          transport: custom(provider as EIP1193Provider),
        });

        for (const tx of pendingTxs) {
          if (!tx.tx_hash) continue;
          try {
            const confs = await publicClient.getTransactionConfirmations({
              hash: tx.tx_hash as `0x${string}`,
            });
            if (active) {
              setConfirmations((prev) => ({ ...prev, [tx.tx_hash!]: Number(confs) }));
            }
          } catch (e: unknown) {
            console.error("Failed to check confirmations for", tx.tx_hash, e);
          }
        }
      } catch (err) {
        console.error("Failed to poll confirmations", err);
      }
    }, 4000);

    return () => {
      active = false;
      clearInterval(interval);
    };
  }, [wallet?.transactions, cfg, connectWallet, queryClient]);

  const handleDeposit = async () => {
    const num = parseFloat(depositAmount);
    if (isNaN(num) || num <= 0) {
      toast.error("Enter a valid amount");
      return;
    }
    if (onChainBalance === null) {
      toast.error("Wallet balance is still loading. Please wait...");
      return;
    }
    const balNum = parseFloat(onChainBalance);
    if (num > balNum) {
      toast.error(`Insufficient wallet balance. You only have ${onChainBalance} ${symbol} in your wallet.`);
      return;
    }
    try {
      const uuidHex = `0x${user!.id.replace(/-/g, "")}` as `0x${string}`;
      const refId = padHex(uuidHex, { size: 32, dir: "right" });
      const hash = await deposit(refId, depositAmount);
      toast.success(`Transaction sent: ${hash.slice(0, 10)}...`);

      await api.post(`${API_WALLET}/pending`, {
        amount: num,
        type: TX_TYPE_DEPOSIT,
        tx_hash: hash,
      });
      queryClient.invalidateQueries({ queryKey: ["wallet"] });

      setShowDeposit(false);
      setDepositAmount("");
    } catch (e) {
      toast.error(classifyWalletError(e, "deposit"));
      setShowDeposit(true);
    }
  };

  const handleWithdraw = async () => {
    const num = parseFloat(withdrawAmount);
    if (isNaN(num) || num <= 0) {
      toast.error("Enter a valid amount");
      return;
    }
    const fee = cfg?.platform_fee ?? 0;
    if (new Decimal(num).add(fee).gt(wallet?.available ?? 0)) {
      toast.error(`Insufficient available balance. You need enough balance to cover the withdrawal amount and the platform fee of ${fee} ${symbol}.`);
      return;
    }
    try {
      const { hash } = await withdrawOnChain(withdrawAmount, address || "");
      toast.success(`Withdrawal successful: ${hash.slice(0, 10)}...`);

      setWithdrawAmount("");
      setShowWithdraw(false);
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    } catch (e) {
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
      toast.error(classifyWalletError(e, "withdraw"));
    }
  };

  return (
    <RequireRole roles={[ROLE_USER, ROLE_ADVERTISER, ROLE_ADMIN]}>
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <div className="h-8 w-8 rounded-lg bg-[#22D3EE]/10 flex items-center justify-center">
                <Wallet size={16} className="text-[#22D3EE]" />
              </div>
              <h1 className="text-3xl font-black tracking-tight text-white">Wallet</h1>
            </div>
            <p className="mt-2 text-white/50 font-mono-dm text-sm">{"// Manage your balance and transactions"}</p>
          </div>
        </div>
      </div>

      {error && !isLoading ? (
        <div className="rounded-2xl bg-red-900/20 p-6 border border-red-900/50 text-red-300">
          {(error as AxiosError<ApiErrorResponse>)?.response?.data?.message || "Failed to load wallet"}
        </div>
      ) : null}

      <div className="space-y-6">
        {isAdvertiser ? (
          <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 md:p-8">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
              <div className="border-b md:border-b-0 md:border-r border-white/10 pb-4 md:pb-0 md:pr-6">
                <p className="text-xs font-bold text-[#6EE7B7] uppercase tracking-widest font-mono-dm mb-1">Available Balance</p>
                <p className="text-3xl font-black text-white">{formatBalance(wallet?.available ?? wallet?.balance ?? 0)} {symbol}</p>
                <p className="text-xs text-white/40 mt-1 font-mono-dm">{"// Ready for campaigns & withdrawals"}</p>
              </div>
              <div className="border-b md:border-b-0 md:border-r border-white/10 pb-4 md:pb-0 md:pr-6">
                <p className="text-xs font-bold text-amber-400 uppercase tracking-widest font-mono-dm mb-1">Frozen Balance</p>
                <p className="text-3xl font-black text-white">{formatBalance(wallet?.frozen ?? 0)} {symbol}</p>
                <p className="text-xs text-white/40 mt-1 font-mono-dm">{"// Allocated to active campaign budgets"}</p>
              </div>
              <div>
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Total Account Balance</p>
                <p className="text-3xl font-black text-white">{formatBalance(wallet?.balance ?? 0)} {symbol}</p>
                <p className="text-xs text-white/40 mt-1 font-mono-dm">{"// Sum of available and allocated"}</p>
              </div>
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowDeposit(!showDeposit);
                  setShowWithdraw(false);
                }}
                className="btn-primary flex items-center gap-2 px-4 py-2.5 text-sm tracking-wider uppercase cursor-pointer"
              >
                <Plus size={16} />
                Deposit Funds
              </button>
              <button
                onClick={() => {
                  setShowWithdraw(!showWithdraw);
                  setShowDeposit(false);
                }}
                className="btn-primary flex items-center gap-2 px-4 py-2.5 text-sm tracking-wider uppercase cursor-pointer"
              >
                <ArrowUpFromLine size={16} />
                Withdraw Funds
              </button>
              {cfg?.contract_token && (
                <button
                  onClick={() => addToken().catch((e) => toast.error(classifyWalletError(e)))}
                  className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-xs tracking-wider uppercase cursor-pointer ml-auto"
                >
                  <Plus className="h-3 w-3" />
                  Add {cfg.token_symbol}
                </button>
              )}
            </div>
            {showDeposit && (
              <div className="mt-4 p-4 rounded-xl border border-white/[0.08] bg-white/[0.03]">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs text-white/50">Deposit {symbol} from Wallet</span>
                  {isConnected && (
                    <button
                      type="button"
                      onClick={() => onChainBalance && setDepositAmount(onChainBalance)}
                      className="text-xs text-[#22D3EE] hover:text-[#67E8F9] transition-colors font-mono cursor-pointer flex items-center gap-0.5"
                    >
                      Wallet Balance: {isBalanceFetching && onChainBalance === null ? "Loading..." : `${formatBalance(onChainBalance)} ${symbol}`} <span className="text-[10px] opacity-75">(MAX)</span>
                    </button>
                  )}
                </div>
                <div className="flex gap-3">
                  <input
                    type="number"
                    step="0.001"
                    min="0.001"
                    onKeyDown={(e) => ["e", "E", "+", "-"].includes(e.key) && e.preventDefault()}
                    placeholder={`Amount in ${symbol}...`}
                    value={depositAmount}
                    onChange={(e) => setDepositAmount(e.target.value)}
                    className="flex-1 rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all"
                  />
                  <button
                    onClick={handleDeposit}
                    disabled={depositStatus === "pending"}
                    className="btn-primary flex items-center gap-2 px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50"
                  >
                    {depositStatus === "pending" ? (
                      <Loader2 className="animate-spin h-4 w-4" />
                    ) : (
                      <ArrowDownToLine size={16} />
                    )}
                    {depositStatus === "pending" ? "Processing..." : "Deposit"}
                  </button>
                </div>
              </div>
            )}
            {showWithdraw && (
              <div className="mt-4 p-4 rounded-xl border border-white/[0.08] bg-white/[0.03]">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs text-white/50">Withdraw {symbol} to Web3 Wallet</span>
                  <button
                    type="button"
                    onClick={() => {
                      const fee = cfg?.platform_fee ?? 0;
                      const maxWithdrawable = Number(Decimal.max(0, Decimal.sub(wallet?.available ?? 0, fee)).toFixed(8));
                      setWithdrawAmount(String(maxWithdrawable));
                    }}
                    className="text-xs text-[#6EE7B7] hover:text-[#34D399] transition-colors font-mono cursor-pointer flex items-center gap-0.5"
                  >
                    Available: {formatBalance(wallet?.available ?? 0)} {symbol} <span className="text-[10px] opacity-75">(MAX)</span>
                  </button>
                </div>
                <div className="flex gap-3 mb-2">
                  <input
                    type="number"
                    step="0.001"
                    min="0.001"
                    onKeyDown={(e) => ["e", "E", "+", "-"].includes(e.key) && e.preventDefault()}
                    placeholder={`Amount in ${symbol}...`}
                    value={withdrawAmount}
                    onChange={(e) => setWithdrawAmount(e.target.value)}
                    className="flex-1 rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all"
                  />
                  <button
                    onClick={handleWithdraw}
                    disabled={withdrawStatus !== "idle" && withdrawStatus !== "success"}
                    className="btn-primary flex items-center gap-2 px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50"
                  >
                    {withdrawStatus === "pending" ? (
                      <Loader2 className="animate-spin h-4 w-4" />
                    ) : (
                      <ArrowUpFromLine size={16} />
                    )}
                    {withdrawStatus === "pending" ? "Processing..." : "Withdraw"}
                  </button>
                </div>
                {cfg?.platform_fee !== undefined && (
                  <p className="text-[11px] text-white/40 font-mono-dm">
                    {"//"} Platform withdrawal fee: <span className="text-amber-400 font-semibold">{cfg.platform_fee} {symbol}</span> (deducted from balance, payout amount will be exact signed value)
                  </p>
                )}
              </div>
            )}
          </div>
        ) : (
          <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 md:p-8">
            <div className="mb-6">
              <p className="text-xs font-bold text-[#6EE7B7] uppercase tracking-widest font-mono-dm mb-1">Balance</p>
              <p className="text-3xl font-black text-white">{formatBalance(wallet?.available ?? wallet?.balance ?? 0)} {symbol}</p>
              <p className="text-xs text-white/40 mt-1 font-mono-dm">{"// Available to withdraw"}</p>
            </div>
            <div className="mt-4 p-4 rounded-xl border border-white/[0.08] bg-white/[0.03]">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-white/50">Withdraw {symbol} to Web3 Wallet</span>
                <button
                  type="button"
                  onClick={() => {
                    const fee = cfg?.platform_fee ?? 0;
                    const maxWithdrawable = Number(Decimal.max(0, Decimal.sub(wallet?.available ?? 0, fee)).toFixed(8));
                    setWithdrawAmount(String(maxWithdrawable));
                  }}
                  className="text-xs text-[#6EE7B7] hover:text-[#34D399] transition-colors font-mono cursor-pointer flex items-center gap-0.5"
                >
                  Available: {formatBalance(wallet?.available ?? 0)} {symbol} <span className="text-[10px] opacity-75">(MAX)</span>
                </button>
              </div>
              <div className="flex gap-3 mb-2">
                <input
                  type="number"
                  step="0.001"
                  min="0.001"
                  onKeyDown={(e) => ["e", "E", "+", "-"].includes(e.key) && e.preventDefault()}
                  placeholder={`Amount in ${symbol}...`}
                  value={withdrawAmount}
                  onChange={(e) => setWithdrawAmount(e.target.value)}
                  className="flex-1 rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all"
                />
                <button
                  onClick={handleWithdraw}
                  disabled={withdrawStatus !== "idle" && withdrawStatus !== "success"}
                  className="btn-primary flex items-center gap-2 px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50"
                >
                  {withdrawStatus === "pending" ? (
                    <Loader2 className="animate-spin h-4 w-4" />
                  ) : (
                    <ArrowUpFromLine size={16} />
                  )}
                  {withdrawStatus === "pending" ? "Processing..." : "Withdraw"}
                </button>
              </div>
              {cfg?.platform_fee !== undefined && (
                <p className="text-[11px] text-white/40 font-mono-dm">
                  {"//"} Platform withdrawal fee: <span className="text-amber-400 font-semibold">{cfg.platform_fee} {symbol}</span> (deducted from balance, payout amount will be exact signed value)
                </p>
              )}
            </div>
          </div>
        )}

        <div>
          <h2 className="text-xl font-bold text-white/90 mb-6">Transaction History</h2>
          <TransactionTable
            transactions={wallet?.transactions}
            confirmations={confirmations}
            isLoading={isLoading}
            isFetching={isFetching}
            symbol={symbol}
            page={txPage}
            totalPages={wallet?.total_pages}
            total={wallet?.total}
            search={txSearch}
            sortBy={txSortBy}
            sortDir={txSortDir}
            onPageChange={setTxPage}
            onPerPageChange={(p) => { setTxPerPage(p); setTxPage(1); }}
            onSearchChange={(s) => { setTxSearch(s); setTxPage(1); }}
            onSortChange={txHandleSort}
          />
        </div>
      </div>
    </RequireRole>
  );
}
