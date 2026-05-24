"use client";

import Loading from "@/components/ui/Loading";
import { format } from "date-fns";
import { ArrowDownToLine, ExternalLink } from "lucide-react";
import { useConfigStore } from "@/store/useConfigStore";
import type { Transaction, PendingTransaction } from "@/types/ads";

interface TransactionTableProps {
  transactions: Transaction[] | undefined;
  pendingTransactions?: PendingTransaction[];
  isLoading: boolean;
  symbol?: string;
}

export default function TransactionTable({ transactions, pendingTransactions, isLoading, symbol = "TK" }: TransactionTableProps) {
  const explorerUrl = useConfigStore((s) => s.config)?.payment_chain?.explorer_url;

  if (isLoading) {
    return <Loading height="h-auto" className="p-16" />;
  }

  const confirmedTxHashes = new Set(transactions?.map((tx) => tx.tx_hash).filter(Boolean));
  const filteredPending = pendingTransactions?.filter((tx) => !confirmedTxHashes.has(tx.tx_hash)) ?? [];
  const hasPending = filteredPending.length > 0;
  const hasTransactions = transactions && transactions.length > 0;

  if (!hasTransactions && !hasPending) {
    return (
      <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-12 flex flex-col items-center justify-center text-center">
        <div className="h-16 w-16 rounded-full bg-white/[0.04] flex items-center justify-center mb-4">
          <ArrowDownToLine className="h-8 w-8 text-white/30" />
        </div>
        <h3 className="text-lg font-bold text-white mb-1">No transactions yet</h3>
        <p className="text-white/50">Deposit funds to your wallet to get started.</p>
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-white/[0.06]">
          <thead className="bg-white/[0.02]">
            <tr>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Type</th>
              <th scope="col" className="px-6 py-4 text-right text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Amount</th>
              <th scope="col" className="px-6 py-4 text-right text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Date</th>
              <th scope="col" className="px-6 py-4 text-right text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Tx Hash</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/[0.06]">
            {/* Pending Transactions */}
            {hasPending && filteredPending.map((tx) => (
              <tr key={tx.tx_hash} className="bg-white/[0.01] hover:bg-white/[0.02] transition-colors border-l-2 border-yellow-500/50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold bg-yellow-500/15 text-yellow-400 ring-1 ring-yellow-500/25 animate-pulse">
                    Deposit (Confirming {tx.confirmations}/{tx.target_confirmations})
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right font-mono-dm font-medium text-yellow-400">
                  +{tx.amount.toFixed(2)} {symbol}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-white/50 font-mono-dm">
                  {format(new Date(tx.created_at), "MMM d, yyyy HH:mm")}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-mono-dm">
                  {tx.tx_hash && explorerUrl ? (
                    <a
                      href={`${explorerUrl}/tx/${tx.tx_hash}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-[#22D3EE] hover:text-[#67E8F9] transition-colors"
                    >
                      {tx.tx_hash.slice(0, 10)}...
                      <ExternalLink className="h-3.5 w-3.5" />
                    </a>
                  ) : (
                    <span className="text-white/30">—</span>
                  )}
                </td>
              </tr>
            ))}

            {/* Confirmed Transactions */}
            {hasTransactions && transactions.map((tx) => (
              <tr key={tx.id} className="hover:bg-white/[0.02] transition-colors">
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold ${tx.type === "DEPOSIT"
                      ? "bg-green-500/15 text-green-400 ring-1 ring-green-500/25"
                      : tx.type === "EARNING"
                        ? "bg-blue-500/15 text-blue-400 ring-1 ring-blue-500/25"
                        : tx.type === "AD_SPEND"
                          ? "bg-red-500/15 text-red-400 ring-1 ring-red-500/25"
                          : "bg-white/10 text-white/50 ring-1 ring-white/20"
                    }`}>
                    {tx.type.replace("_", " ")}
                  </span>
                </td>
                <td className={`px-6 py-4 whitespace-nowrap text-right font-mono-dm font-medium ${tx.amount >= 0 ? "text-green-400" : "text-red-400"
                  }`}>
                  {tx.amount >= 0 ? "+" : ""}{Math.abs(tx.amount).toFixed(2)} {symbol}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-white/50 font-mono-dm">
                  {format(new Date(tx.created_at), "MMM d, yyyy HH:mm")}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-mono-dm">
                  {tx.tx_hash && explorerUrl ? (
                    <a
                      href={`${explorerUrl}/tx/${tx.tx_hash}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-[#22D3EE] hover:text-[#67E8F9] transition-colors"
                    >
                      {tx.tx_hash.slice(0, 10)}...
                      <ExternalLink className="h-3.5 w-3.5" />
                    </a>
                  ) : (
                    <span className="text-white/30">—</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
