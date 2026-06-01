"use client";

import { useState, useEffect } from "react";
import { useWriteContract, useWaitForTransactionReceipt } from "wagmi";
import RequireRole from "@/components/auth/RequireRole";
import { useWalletConnection } from "@/hooks/wallet/useWalletConnection";
import { useFaucetClaim, useFaucetConfirm, useDevETHClaim, useFaucetHistory } from "@/hooks/useFaucet";
import type { FaucetHistoryItem } from "@/hooks/useFaucet";
import { FAUCET_ABI } from "@/lib/faucet";
import { Droplets, Wallet, Loader2, CheckCircle, ExternalLink, Plus, Flame, Clock, ChevronLeft, ChevronRight, Search, ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react";
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from "@/lib/constants";
import { toast } from "sonner";
import { classifyWalletError } from "@/lib/wallet";
import { ROLES, FAUCET_AMOUNT } from "@/lib/constants";
import { useConfigStore } from "@/store/useConfigStore";
import { formatDistanceToNowStrict } from "date-fns";


export default function FaucetPage() {
  const cfg = useConfigStore((s) => s.config);
  const { address, isConnected, connectWallet, addToken, isConnecting } = useWalletConnection();
  const claimMutation = useFaucetClaim();
  const confirmMutation = useFaucetConfirm();
  const devETHMutation = useDevETHClaim();
  const [historyPage, setHistoryPage] = useState(1);
  const [historyPerPage, setHistoryPerPage] = useState(DEFAULT_PAGE_SIZE);
  const [historySearch, setHistorySearch] = useState("");
  const [historySortBy, setHistorySortBy] = useState("claimed_at");
  const [historySortDir, setHistorySortDir] = useState("desc");
  const { data: historyData, isLoading: isHistoryLoading } = useFaucetHistory(historyPage, historyPerPage, historySearch || undefined, historySortBy, historySortDir);
  const history = historyData?.claims || [];
  const historyTotalPages = historyData?.total_pages || 1;
  const { mutateAsync, isPending: isWritePending } = useWriteContract();
  const [txHash, setTxHash] = useState<string | null>(null);

  const { isLoading: isConfirming, isSuccess: isConfirmed } = useWaitForTransactionReceipt({
    hash: txHash as `0x${string}` | undefined,
  });

  const handleHistorySort = (col: string) => {
    setHistoryPage(1);
    if (historySortBy === col) {
      setHistorySortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setHistorySortBy(col);
      setHistorySortDir("desc");
    }
  };

  const renderSortIcon = (col: string) => {
    if (historySortBy !== col) return <ArrowUpDown className="h-3 w-3 inline ml-1 opacity-30" />;
    return historySortDir === "asc" ? <ArrowUp className="h-3 w-3 inline ml-1 text-[#6EE7B7]" /> : <ArrowDown className="h-3 w-3 inline ml-1 text-[#6EE7B7]" />;
  };

  const symbol = cfg?.token_symbol ?? "";
  const explorerUrl = cfg?.payment_chain?.explorer_url;

  const isPending = claimMutation.isPending || isWritePending || isConfirming || confirmMutation.isPending;
  const isComplete = isConfirmed && !confirmMutation.isPending;

  useEffect(() => {
    if (isConfirmed && txHash && address && !confirmMutation.isSuccess && !confirmMutation.isPending) {
      confirmMutation.mutateAsync({ txHash, walletAddr: address }).catch((e) => {
        toast.error(classifyWalletError(e));
      });
    }
  }, [isConfirmed, txHash, address, confirmMutation]);

  const handleClaim = async () => {
    if (!address) return;

    try {
      const payload = await claimMutation.mutateAsync(address);

      const hash = await mutateAsync({
        address: payload.faucet_addr as `0x${string}`,
        abi: FAUCET_ABI,
        functionName: "requestTokens",
        args: [
          payload.wallet as `0x${string}`,
          BigInt(payload.amount),
          BigInt(payload.nonce),
          BigInt(payload.deadline),
          payload.signature as `0x${string}`,
        ],
      });

      setTxHash(hash);
      toast.success(`Transaction submitted! ${FAUCET_AMOUNT} ${symbol} tokens on the way.`);
    } catch (e) {
      toast.error(classifyWalletError(e));
    }
  };
  const handleDevETHClaim = async () => {
    if (!address) return;
    try {
      const payload = await devETHMutation.mutateAsync(address);
      toast.success(
        <span>
          Success! 1.0 Dev ETH sent to your wallet.
          {payload.tx_hash && (
            <>
              <br />
              Transaction:{" "}
              <a
                href={`${explorerUrl}/tx/${payload.tx_hash}`}
                target="_blank"
                rel="noopener noreferrer"
                className="underline text-[#22D3EE] hover:text-[#67E8F9]"
              >
                {payload.tx_hash.slice(0, 10)}...{payload.tx_hash.slice(-8)}
              </a>
            </>
          )}
        </span>,
        { duration: 10000 }
      );
    } catch (e) {
      toast.error(classifyWalletError(e));
    }
  };
  return (
    <RequireRole roles={[...ROLES]}>
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="h-8 w-8 rounded-lg bg-[#22D3EE]/10 flex items-center justify-center">
            <Droplets size={16} className="text-[#22D3EE]" />
          </div>
          <h1 className="text-3xl font-black tracking-tight text-white">Faucet</h1>
        </div>
        <p className="mt-2 text-white/50 font-mono-dm text-sm">
          {"// Get free test tokens to explore the platform"}
        </p>
      </div>

      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] overflow-hidden">
        <div className="p-6 md:p-8">
          <div className="flex flex-col md:flex-row md:items-center gap-6">
            <div className="h-16 w-16 rounded-2xl bg-[#22D3EE]/10 flex items-center justify-center shrink-0">
              <Droplets size={32} className="text-[#22D3EE]" />
            </div>

            <div className="flex-1">
              <p className="text-3xl font-black text-white">{FAUCET_AMOUNT} {symbol}</p>
              <p className="text-sm text-white/50 mt-1">
                Free test tokens &middot; One claim per 24 hours &middot; You pay gas
              </p>
            </div>

            <div className="shrink-0">
              {isConnected && address ? (
                <button
                  onClick={handleClaim}
                  disabled={isPending || isComplete}
                  className="btn-primary flex items-center gap-2 px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isPending ? (
                    <Loader2 className="animate-spin h-4 w-4" />
                  ) : isComplete ? (
                    <CheckCircle className="h-4 w-4" />
                  ) : (
                    <Droplets className="h-4 w-4" />
                  )}
                  {isPending
                    ? "Confirming..."
                    : isComplete
                      ? "Claimed"
                      : `Claim ${FAUCET_AMOUNT} ${symbol}`}
                </button>
              ) : (
                <button
                  onClick={() => connectWallet().catch((e) => toast.error(classifyWalletError(e, "connect")))}
                  disabled={isConnecting}
                  className="btn-primary flex items-center gap-2 px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50"
                >
                  {isConnecting ? <Loader2 className="animate-spin h-4 w-4" /> : <Wallet className="h-4 w-4" />}
                  {isConnecting ? "Connecting..." : "Connect Wallet"}
                </button>
              )}
            </div>
          </div>
        </div>

        <div className="border-t border-white/[0.06] px-6 md:px-8 py-4 flex items-center justify-between flex-wrap gap-3">
          <div className="flex items-center gap-2 text-sm text-white/50">
            <Wallet className="h-4 w-4" />
            {address ? (
              <span className="font-mono text-white/70">
                {address.slice(0, 6)}...{address.slice(-4)}
              </span>
            ) : (
              <span>No wallet connected</span>
            )}
          </div>
          <div className="flex items-center gap-3">
            {address && cfg?.contract_token && (
              <button
                onClick={() => addToken().catch((e) => toast.error(classifyWalletError(e)))}
                className="btn-primary flex items-center gap-1.5 px-3 py-1.5 text-xs tracking-wider uppercase cursor-pointer"
              >
                <Plus className="h-3 w-3" />
                Add {cfg.token_symbol}
              </button>
            )}
            {address && explorerUrl && (
              <a
                href={`${explorerUrl}/address/${address}`}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1.5 text-xs text-[#22D3EE] hover:text-[#67E8F9] transition-colors"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                View on Explorer
              </a>
            )}
          </div>
        </div>
      </div>

      {process.env.NEXT_PUBLIC_APP_MODE === "development" && (
        <div className="mt-6 rounded-2xl border border-amber-500/20 bg-amber-500/[0.02] overflow-hidden shadow-[0_0_20px_rgba(245,158,11,0.05)] relative group">
          <div className="absolute top-0 right-0 w-64 h-64 bg-amber-500/10 rounded-full blur-[80px] -mr-20 -mt-20 pointer-events-none transition-all group-hover:bg-amber-500/15" />
          <div className="p-6 md:p-8">
            <div className="flex flex-col md:flex-row md:items-center gap-6">
              <div className="h-16 w-16 rounded-2xl bg-amber-500/10 flex items-center justify-center shrink-0 border border-amber-500/20 animate-pulse">
                <Flame size={32} className="text-amber-400" />
              </div>

              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <p className="text-2xl font-black text-white">1.0 Dev ETH</p>
                  <span className="px-2 py-0.5 rounded-full text-xs font-bold uppercase tracking-wider bg-amber-500/20 text-amber-300 border border-amber-500/30">
                    Gas Faucet
                  </span>
                </div>
                <p className="text-sm text-white/50 mt-2">
                  Need gas to claim tokens? Request 1.0 test ETH for your MetaMask wallet on the local dev network.
                </p>
              </div>

              <div className="shrink-0 z-10">
                {isConnected && address ? (
                  <button
                    onClick={handleDevETHClaim}
                    disabled={devETHMutation.isPending}
                    className="flex items-center gap-2 px-6 py-3 text-sm font-bold tracking-wider uppercase rounded-xl border border-amber-500/30 bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 hover:shadow-[0_0_15px_rgba(245,158,11,0.2)]"
                  >
                    {devETHMutation.isPending ? (
                      <Loader2 className="animate-spin h-4 w-4" />
                    ) : (
                      <Flame className="h-4 w-4" />
                    )}
                    {devETHMutation.isPending ? "Funding..." : "Claim 1.0 ETH"}
                  </button>
                ) : (
                  <button
                    onClick={() => connectWallet().catch((e) => toast.error(classifyWalletError(e, "connect")))}
                    disabled={isConnecting}
                    className="btn-primary flex items-center gap-2 px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50"
                  >
                    {isConnecting ? <Loader2 className="animate-spin h-4 w-4" /> : <Wallet className="h-4 w-4" />}
                    {isConnecting ? "Connecting..." : "Connect Wallet"}
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="mt-12">
        <h2 className="text-xl font-bold text-white/90 mb-6">Claim History</h2>

        <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] overflow-hidden">
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3 px-6 py-4 border-b border-white/[0.06]">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-white/30" />
              <input
                type="text"
                value={historySearch}
                onChange={(e) => { setHistorySearch(e.target.value); setHistoryPage(1); }}
                placeholder="Search by tx hash..."
                className="w-full bg-white/[0.03] border border-white/[0.08] text-white/70 text-sm rounded-lg pl-9 pr-3 py-2 focus:outline-none focus:border-[#6EE7B7]/50 transition-all placeholder:text-white/20"
              />
            </div>
          </div>

          {isHistoryLoading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="animate-spin h-8 w-8 text-white/30" />
            </div>
          ) : history.length > 0 ? (
            <>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-white/[0.06]">
                  <thead className="bg-white/[0.02]">
                    <tr>
                      <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm cursor-pointer select-none hover:text-white/70 transition-colors" onClick={() => handleHistorySort("amount")}>
                        Amount {renderSortIcon("amount")}
                      </th>
                      <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm cursor-pointer select-none hover:text-white/70 transition-colors" onClick={() => handleHistorySort("tx_hash")}>
                        Transaction Hash {renderSortIcon("tx_hash")}
                      </th>
                      <th scope="col" className="px-6 py-4 text-right text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm cursor-pointer select-none hover:text-white/70 transition-colors" onClick={() => handleHistorySort("claimed_at")}>
                        Claimed {renderSortIcon("claimed_at")}
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/[0.06]">
                    {history.map((item: FaucetHistoryItem) => (
                      <tr key={item.id} className="hover:bg-white/[0.02] transition-colors">
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className="text-green-400 font-mono-dm">{Number(item.amount) / 1e18} {symbol}</span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          {explorerUrl ? (
                            <a
                              href={`${explorerUrl}/tx/${item.tx_hash}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="inline-flex items-center gap-1.5 font-mono text-sm text-[#22D3EE] transition-colors"
                            >
                              {item.tx_hash.slice(0, 10)}...{item.tx_hash.slice(-8)}
                              <ExternalLink className="h-3 w-3 shrink-0" />
                            </a>
                          ) : (
                            <span className="text-white/50 font-mono text-sm">
                              {item.tx_hash.slice(0, 10)}...{item.tx_hash.slice(-8)}
                            </span>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-right">
                          <div className="flex items-center justify-end gap-1.5 text-sm text-white/50 font-mono-dm">
                            <Clock className="h-3.5 w-3.5 text-white/30" />
                            {formatDistanceToNowStrict(new Date(item.claimed_at), { addSuffix: true })}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="flex flex-col sm:flex-row items-center justify-between px-6 py-4 border-t border-white/[0.06] gap-4">
                <div className="flex items-center gap-4 order-2 sm:order-1">
                  <span className="text-sm text-white/50 whitespace-nowrap">
                    Page {historyPage} of {historyTotalPages}
                  </span>
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-white/30 uppercase tracking-widest font-mono-dm">Show</span>
                    <select
                      value={historyPerPage}
                      onChange={(e) => { setHistoryPerPage(Number(e.target.value)); setHistoryPage(1); }}
                      className="bg-white/[0.03] border border-white/[0.08] text-white/70 text-xs rounded-lg px-2 py-1 focus:outline-none focus:border-[#6EE7B7]/50 transition-all cursor-pointer"
                    >
                      {PAGE_SIZE_OPTIONS.map((size) => (
                        <option key={size} value={size} className="bg-[#0A0A0A]">{size}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <div className="flex gap-2 order-1 sm:order-2">
                  <button
                    onClick={() => setHistoryPage((p) => Math.max(1, p - 1))}
                    disabled={historyPage === 1}
                    className="flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg border border-white/[0.08] bg-white/[0.03] text-white/70 hover:bg-white/[0.06] hover:text-white transition-colors disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
                  >
                    <ChevronLeft className="h-4 w-4" />
                    Prev
                  </button>
                  <button
                    onClick={() => setHistoryPage((p) => Math.min(historyTotalPages, p + 1))}
                    disabled={historyPage === historyTotalPages || historyTotalPages === 0}
                    className="flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg border border-white/[0.08] bg-white/[0.03] text-white/70 hover:bg-white/[0.06] hover:text-white transition-colors disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
                  >
                    Next
                    <ChevronRight className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </>
          ) : (
            <div className="px-6 py-16 text-center">
              <Droplets size={40} className="mx-auto mb-4 text-white/20" />
              <h2 className="text-xl font-bold text-white/60 mb-2">
                {historySearch ? "No claims match your search" : "No claims yet"}
              </h2>
              <p className="text-white/40 text-sm max-w-md mx-auto">
                {historySearch
                  ? "Try adjusting your search to find what you're looking for."
                  : `Claim your first ${symbol} tokens using the faucet above.`}
              </p>
            </div>
          )}
        </div>
      </div>
    </RequireRole>
  );
}
