"use client";

import { format } from "date-fns";
import { ArrowDownToLine, ExternalLink } from "lucide-react";
import { useConfigStore } from "@/store/useConfigStore";
import type { Transaction } from "@/types/ads";
import {
  BLOCK_CONFIRMATIONS,
  TX_STATUS_PENDING,
  TX_STATUS_FAILED,
  TX_TYPE_DEPOSIT,
  TX_TYPE_EARNING,
  TX_TYPE_AD_EARNING,
  TX_TYPE_AD_SPEND,
  TX_TYPE_WITHDRAWAL,
  TX_TYPE_WITHDRAWAL_FEE,
} from "@/lib/constants";
import DataTable, { Column } from "@/components/ui/DataTable";
import { formatBalance } from "@/lib/wallet";

interface TransactionTableProps {
  transactions: Transaction[] | undefined;
  confirmations: Record<string, number>;
  isLoading: boolean;
  isFetching?: boolean;
  symbol?: string;
  page?: number;
  totalPages?: number;
  total?: number;
  search?: string;
  sortBy?: string;
  sortDir?: string;
  onPageChange?: (page: number) => void;
  onPerPageChange?: (perPage: number) => void;
  onSearchChange?: (search: string) => void;
  onSortChange?: (col: string) => void;
}

export default function TransactionTable({
  transactions, confirmations, isLoading, isFetching = false, symbol = "TK",
  page = 1, totalPages = 1, total = 0,
  search = "", sortBy = "created_at", sortDir = "desc",
  onPageChange, onSearchChange, onSortChange,
}: TransactionTableProps) {
  const explorerUrl = useConfigStore((s) => s.config)?.payment_chain?.explorer_url;

  const getRowClassName = (tx: Transaction) => {
    const isPending = tx.status === TX_STATUS_PENDING;
    const isFailed = tx.status === TX_STATUS_FAILED;
    if (isPending) {
      return "bg-white/[0.01] border-l-2 border-yellow-500/50 animate-pulse";
    }
    if (isFailed) {
      return "opacity-60 border-l-2 border-red-500/50";
    }
    return "";
  };

  const columns: Column<Transaction>[] = [
    {
      header: "Type",
      accessorKey: "type",
      sortable: true,
      cell: (tx) => {
        const isPending = tx.status === TX_STATUS_PENDING;
        const isFailed = tx.status === TX_STATUS_FAILED;
        if (isPending) {
          const currentConf = confirmations[tx.tx_hash ?? ""] ?? 0;
          return (
            <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold bg-yellow-500/15 text-yellow-400 ring-1 ring-yellow-500/25">
              {tx.type} (Confirming {currentConf}/{BLOCK_CONFIRMATIONS})
            </span>
          );
        }
        if (isFailed) {
          return (
            <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold bg-red-500/15 text-red-400 ring-1 ring-red-500/25 animate-pulse">
              {tx.type} (FAILED)
            </span>
          );
        }
        return (
          <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold ${tx.type === TX_TYPE_DEPOSIT
              ? "bg-green-500/15 text-green-400 ring-1 ring-green-500/25"
              : tx.type === TX_TYPE_EARNING || tx.type === TX_TYPE_AD_EARNING
                ? "bg-blue-500/15 text-blue-400 ring-1 ring-blue-500/25"
                : tx.type === TX_TYPE_AD_SPEND
                  ? "bg-red-500/15 text-red-400 ring-1 ring-red-500/25"
                  : tx.type === TX_TYPE_WITHDRAWAL || tx.type === TX_TYPE_WITHDRAWAL_FEE
                    ? "bg-orange-500/15 text-orange-400 ring-1 ring-orange-500/25"
                    : "bg-white/10 text-white/50 ring-1 ring-white/20"
            }`}>
            {tx.type.replace("_", " ")}
          </span>
        );
      },
    },
    {
      header: "Amount",
      accessorKey: "amount",
      sortable: true,
      className: "text-right",
      cell: (tx) => {
        const isFailed = tx.status === TX_STATUS_FAILED;
        const amount = Number(tx.amount);
        return (
          <span className={`font-mono-dm font-medium ${isFailed ? "text-red-400 line-through" : amount >= 0 ? "text-green-400" : "text-red-400"
            }`}>
            {amount >= 0 ? "+" : "-"}{formatBalance(Math.abs(amount))} {symbol}
          </span>
        );
      },
    },
    {
      header: "Date",
      accessorKey: "created_at",
      sortable: true,
      className: "text-right",
      cell: (tx) => (
        <span className="text-white/50 font-mono-dm">
          {format(new Date(tx.created_at), "MMM d, yyyy HH:mm")}
        </span>
      ),
    },
    {
      header: "Tx Hash",
      className: "text-right",
      cell: (tx) => (
        <div className="flex justify-end font-mono-dm">
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
        </div>
      ),
    },
  ];

  return (
    <DataTable
      columns={columns}
      data={transactions || []}
      isLoading={isLoading}
      isFetching={isFetching}
      sortBy={sortBy}
      sortDir={sortDir}
      onSort={onSortChange}
      page={page}
      totalPages={totalPages}
      totalItems={total}
      onPageChange={onPageChange}
      perPage={page === 1 ? undefined : undefined}
      searchPlaceholder="Search by type..."
      searchValue={search}
      onSearchChange={onSearchChange}
      getRowClassName={getRowClassName}
      emptyIcon={<ArrowDownToLine className="h-8 w-8 text-white/30" />}
      emptyTitle="No transactions yet"
      emptyDescription="Deposit funds to your wallet to get started."
    />
  );
}
