"use client";

import { useLinksQuery, useDeleteLink } from "@/hooks/useLinks";
import { formatDistanceToNowStrict } from "date-fns";
import { BarChart3, Copy, ExternalLink, QrCode, Trash2, Clock, DollarSign } from "lucide-react";
import Link from "next/link";
import { useState, useEffect } from "react";
import { DEFAULT_PAGE_SIZE } from "@/lib/constants";
import DataTable, { Column } from "@/components/ui/DataTable";
import type { Link as LinkType } from "@/types/link";

function timeAgo(date: string) {
  return formatDistanceToNowStrict(new Date(date), { addSuffix: true });
}

export default function LinkTable() {
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(DEFAULT_PAGE_SIZE);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [isMonetized, setIsMonetized] = useState<string>("");
  const [sortBy, setSortBy] = useState<string>("created_at");
  const [sortDir, setSortDir] = useState<string>("desc");
  const [copiedSlug, setCopiedSlug] = useState<string | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  const monetizedFilter = isMonetized === "" ? undefined : isMonetized === "true";

  const { data, isLoading, isFetching } = useLinksQuery(
    page,
    perPage,
    debouncedSearch || undefined,
    monetizedFilter,
    sortBy,
    sortDir
  );
  const deleteMutation = useDeleteLink();

  const copyToClipboard = (url: string, slug: string) => {
    navigator.clipboard.writeText(url);
    setCopiedSlug(slug);
    setTimeout(() => setCopiedSlug(null), 2000);
  };

  const handleSort = (columnId: string) => {
    setPage(1);
    if (sortBy === columnId) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortBy(columnId);
      setSortDir("desc");
    }
  };

  const links = data?.links || [];
  const totalPages = data?.total_pages || 1;
  const totalItems = data?.total || 0;
  const hasActiveFilter = debouncedSearch !== "" || isMonetized !== "";

  const columns: Column<LinkType>[] = [
    {
      header: "Short Link",
      accessorKey: "short_url",
      sortable: true,
      cell: (link) => (
        <div className="flex items-center group">
          <span className="text-[#6EE7B7] font-medium mr-2">{link.short_url}</span>
          <button
            onClick={() => copyToClipboard(link.short_url, link.slug)}
            className="text-white/30 hover:text-white transition-colors opacity-0 group-hover:opacity-100 hover:cursor-pointer"
          >
            {copiedSlug === link.slug ? (
              <span className="text-xs text-[#6EE7B7] font-medium ml-1">Copied!</span>
            ) : (
              <Copy className="h-4 w-4" />
            )}
          </button>
        </div>
      ),
    },
    {
      header: "Original URL",
      accessorKey: "original",
      sortable: true,
      cell: (link) => (
        <div className="flex items-center max-w-xs sm:max-w-md">
          <span className="truncate text-white/60 text-sm">{link.original}</span>
          <a
            href={link.original}
            target="_blank"
            rel="noopener noreferrer"
            className="ml-2 text-white/30 hover:text-[#6EE7B7] flex-shrink-0 transition-colors"
          >
            <ExternalLink className="h-4 w-4" />
          </a>
        </div>
      ),
    },
    {
      header: "Created",
      accessorKey: "created_at",
      sortable: true,
      cell: (link) => (
        <div className="flex items-center gap-1.5 text-sm text-white/50 font-mono-dm">
          <Clock className="h-3.5 w-3.5 text-white/30" />
          {timeAgo(link.created_at)}
        </div>
      ),
    },
    {
      header: "Expires",
      accessorKey: "expires_at",
      sortable: true,
      cell: (link) => (
        <div className="text-sm font-mono-dm">
          {link.expires_at ? (
            new Date(link.expires_at) < new Date() ? (
              <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold bg-red-500/15 text-red-400 ring-1 ring-red-500/25">
                Expired
              </span>
            ) : (
              <span className="text-white/50">{timeAgo(link.expires_at)}</span>
            )
          ) : (
            <span className="text-white/30">Never</span>
          )}
        </div>
      ),
    },
    {
      header: "Monetize",
      accessorKey: "is_monetized",
      sortable: true,
      className: "text-center",
      cell: (link) => (
        <div className="flex justify-center">
          {link.is_monetized ? (
            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-semibold bg-[#6EE7B7]/15 text-[#6EE7B7] ring-1 ring-[#6EE7B7]/25">
              <DollarSign size={10} />
              On
            </span>
          ) : (
            <span className="text-white/30 text-xs">Off</span>
          )}
        </div>
      ),
    },
    {
      header: "Actions",
      className: "text-right",
      cell: (link) => (
        <div className="flex justify-end gap-3">
          <a
            href={link.qr_url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-white/40 hover:text-[#6EE7B7] transition-colors"
            title="View QR Code"
          >
            <QrCode className="h-5 w-5" />
          </a>
          <Link
            href={`/dashboard/links/${link.slug}`}
            className="text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors"
            title="View Details"
          >
            <BarChart3 className="h-5 w-5" />
          </Link>
          <button
            onClick={() => deleteMutation.mutate(link.slug)}
            className="text-white/40 hover:text-red-400 hover:cursor-pointer transition-colors disabled:opacity-50"
            title="Delete Link"
            disabled={deleteMutation.isPending}
          >
            <Trash2 className="h-5 w-5" />
          </button>
        </div>
      ),
    },
  ];

  return (
    <DataTable
      columns={columns}
      data={links}
      isLoading={isLoading}
      isFetching={isFetching}
      sortBy={sortBy}
      sortDir={sortDir}
      onSort={handleSort}
      page={page}
      totalPages={totalPages}
      totalItems={totalItems}
      onPageChange={setPage}
      perPage={perPage}
      onPerPageChange={(p) => {
        setPerPage(p);
        setPage(1);
      }}
      searchPlaceholder="Search by URL or slug..."
      searchValue={search}
      onSearchChange={setSearch}
      extraFilterSlot={
        <select
          value={isMonetized}
          onChange={(e) => {
            setIsMonetized(e.target.value);
            setPage(1);
          }}
          className="bg-white/[0.03] border border-white/[0.08] text-white/70 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-[#6EE7B7]/50 transition-all cursor-pointer"
        >
          <option value="" className="bg-[#0A0A0A]">All Links</option>
          <option value="true" className="bg-[#0A0A0A]">Monetized</option>
          <option value="false" className="bg-[#0A0A0A]">Non-Monetized</option>
        </select>
      }
      emptyIcon={<BarChart3 size={40} />}
      emptyTitle={hasActiveFilter ? "No links match your filter" : "No links yet"}
      emptyDescription={
        hasActiveFilter
          ? "Try adjusting your search or filter to find what you're looking for."
          : "Create your first short link using the form above."
      }
    />
  );
}
