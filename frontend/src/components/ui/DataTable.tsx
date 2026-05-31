import React, { useState, useEffect } from "react";
import { ChevronLeft, ChevronRight, Search, ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react";
import { PAGE_SIZE_OPTIONS } from "@/lib/constants";

export interface Column<T> {
  header: React.ReactNode;
  accessorKey?: keyof T | string;
  sortable?: boolean;
  sortId?: string;
  className?: string;
  cell?: (row: T, index: number) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  isLoading?: boolean;
  isFetching?: boolean;

  // Sorting
  sortBy?: string;
  sortDir?: string;
  onSort?: (colId: string) => void;

  // Header Title
  title?: React.ReactNode;

  // Pagination
  page?: number;
  totalPages?: number;
  totalItems?: number;
  onPageChange?: (page: number) => void;
  perPage?: number;
  onPerPageChange?: (perPage: number) => void;
  perPageOptions?: number[];

  // Search
  searchPlaceholder?: string;
  searchValue?: string;
  onSearchChange?: (val: string) => void;
  extraFilterSlot?: React.ReactNode;

  // Row styling
  getRowClassName?: (row: T, index: number) => string;

  // Empty State
  emptyIcon?: React.ReactNode;
  emptyTitle?: string;
  emptyDescription?: string;
}

export default function DataTable<T>({
  columns,
  data,
  isLoading = false,
  isFetching = false,
  sortBy,
  sortDir,
  onSort,
  title,
  page = 1,
  totalPages = 1,
  totalItems = 0,
  onPageChange,
  perPage,
  onPerPageChange,
  perPageOptions = PAGE_SIZE_OPTIONS,
  searchPlaceholder = "Search...",
  searchValue,
  onSearchChange,
  extraFilterSlot,
  getRowClassName,
  emptyIcon,
  emptyTitle = "No data available",
  emptyDescription = "There are no items to display.",
}: DataTableProps<T>) {
  const [localSearch, setLocalSearch] = useState(searchValue ?? "");

  useEffect(() => {
    if (searchValue !== undefined) {
      setLocalSearch(searchValue);
    }
  }, [searchValue]);

  const handleSearchChange = (val: string) => {
    setLocalSearch(val);
    if (onSearchChange) {
      onSearchChange(val);
    }
  };

  const renderSortIcon = (column: Column<T>) => {
    if (!column.sortable || !onSort) return null;
    const colId = column.sortId || (column.accessorKey as string);
    if (sortBy !== colId) {
      return <ArrowUpDown className="h-3 w-3 inline ml-1 opacity-30" />;
    }
    return sortDir === "asc" ? (
      <ArrowUp className="h-3 w-3 inline ml-1 text-[#6EE7B7]" />
    ) : (
      <ArrowDown className="h-3 w-3 inline ml-1 text-[#6EE7B7]" />
    );
  };

  const handleHeaderClick = (column: Column<T>) => {
    if (column.sortable && onSort) {
      const colId = column.sortId || (column.accessorKey as string);
      onSort(colId);
    }
  };

  if (isLoading) {
    return (
      <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-16 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white/30"></div>
      </div>
    );
  }

  const showSearchOrFilter = onSearchChange || extraFilterSlot;
  const hasData = data && data.length > 0;

  return (
    <div
      className={`rounded-2xl border border-white/[0.08] bg-white/[0.02] overflow-hidden transition-opacity duration-300 ${
        isFetching ? "opacity-50 pointer-events-none" : "opacity-100"
      }`}
    >
      {title && (
        <div className="px-6 pt-6 pb-2 border-b border-white/[0.04]">
          {title}
        </div>
      )}
      {showSearchOrFilter && (
        <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3 px-6 py-4 border-b border-white/[0.06]">
          {onSearchChange && (
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-white/30" />
              <input
                type="text"
                value={localSearch}
                onChange={(e) => handleSearchChange(e.target.value)}
                placeholder={searchPlaceholder}
                className="w-full bg-white/[0.03] border border-white/[0.08] text-white/70 text-sm rounded-lg pl-9 pr-3 py-2 focus:outline-none focus:border-[#6EE7B7]/50 transition-all placeholder:text-white/20"
              />
            </div>
          )}
          {extraFilterSlot}
        </div>
      )}

      {!hasData ? (
        <div className="px-6 py-16 text-center flex flex-col items-center justify-center">
          {emptyIcon && <div className="mb-4 text-white/20">{emptyIcon}</div>}
          <h2 className="text-xl font-bold text-white/60 mb-2">{emptyTitle}</h2>
          <p className="text-white/40 text-sm max-w-md mx-auto">{emptyDescription}</p>
        </div>
      ) : (
        <>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-white/[0.06]">
              <thead className="bg-white/[0.02]">
                <tr>
                  {columns.map((col, idx) => (
                    <th
                      key={idx}
                      scope="col"
                      className={`px-6 py-4 text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm ${
                        col.sortable ? "cursor-pointer select-none hover:text-white/70 transition-colors" : ""
                      } ${col.className ?? "text-left"}`}
                      onClick={() => handleHeaderClick(col)}
                    >
                      {col.header}
                      {renderSortIcon(col)}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-white/[0.06]">
                {data.map((row, rowIdx) => (
                  <tr
                    key={rowIdx}
                    className={`hover:bg-white/[0.02] transition-colors ${
                      getRowClassName ? getRowClassName(row, rowIdx) : ""
                    }`}
                  >
                    {columns.map((col, colIdx) => (
                      <td key={colIdx} className={`px-6 py-4 whitespace-nowrap text-sm ${col.className ?? "text-left"}`}>
                        {col.cell
                          ? col.cell(row, rowIdx)
                          : col.accessorKey
                          ? String(row[col.accessorKey as keyof T] ?? "")
                          : null}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {onPageChange && (totalPages > 1 || hasData) && (
            <div className="flex flex-col sm:flex-row items-center justify-between px-6 py-4 border-t border-white/[0.06] gap-4">
              <div className="flex items-center gap-4 order-2 sm:order-1">
                <span className="text-sm text-white/50 whitespace-nowrap">
                  Page {page} of {totalPages} {totalItems > 0 && `(${totalItems} total)`}
                </span>
                {onPerPageChange && perPage && (
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-white/30 uppercase tracking-widest font-mono-dm">Show</span>
                    <select
                      value={perPage}
                      onChange={(e) => {
                        onPerPageChange(Number(e.target.value));
                      }}
                      className="bg-white/[0.03] border border-white/[0.08] text-white/70 text-xs rounded-lg px-2 py-1 focus:outline-none focus:border-[#6EE7B7]/50 transition-all cursor-pointer"
                    >
                      {perPageOptions.map((size) => (
                        <option key={size} value={size} className="bg-[#0A0A0A]">
                          {size}
                        </option>
                      ))}
                    </select>
                  </div>
                )}
              </div>
              <div className="flex gap-2 order-1 sm:order-2">
                <button
                  onClick={() => onPageChange(Math.max(1, page - 1))}
                  disabled={page === 1}
                  className="flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg border border-white/[0.08] bg-white/[0.03] text-white/70 hover:bg-white/[0.06] hover:text-white transition-colors disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
                >
                  <ChevronLeft className="h-4 w-4" />
                  Prev
                </button>
                <button
                  onClick={() => onPageChange(Math.min(totalPages, page + 1))}
                  disabled={page === totalPages || totalPages === 0}
                  className="flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg border border-white/[0.08] bg-white/[0.03] text-white/70 hover:bg-white/[0.06] hover:text-white transition-colors disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
                >
                  Next
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
