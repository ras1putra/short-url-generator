"use client";

import Loading from "@/components/ui/Loading";
import { useLinksQuery, useDeleteLink } from "@/hooks/useLinks";
import { formatDistanceToNowStrict } from "date-fns";
import { BarChart3, Copy, ExternalLink, QrCode, Trash2, ChevronLeft, ChevronRight, Clock, DollarSign } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS } from "@/lib/constants";

function timeAgo(date: string) {
  return formatDistanceToNowStrict(new Date(date), { addSuffix: true });
}

export default function LinkTable() {
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(DEFAULT_PAGE_SIZE);
  const [copiedSlug, setCopiedSlug] = useState<string | null>(null);
  const { data, isLoading } = useLinksQuery(page, perPage);
  const deleteMutation = useDeleteLink();

  const copyToClipboard = (url: string, slug: string) => {
    navigator.clipboard.writeText(url);
    setCopiedSlug(slug);
    setTimeout(() => setCopiedSlug(null), 2000);
  };

  if (isLoading) {
    return <Loading height="h-auto" className="p-16" />;
  }

  const links = data?.links || [];
  const totalPages = data?.total_pages || 1;

  if (links.length === 0) {
    const isFirstPage = page === 1;
    return (
      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-12 text-center">
        <BarChart3 size={40} className="mx-auto mb-4 text-white/20" />
        <h2 className="text-xl font-bold text-white/60 mb-2">
          {isFirstPage ? "No links yet" : "Nothing here"}
        </h2>
        <p className="text-white/40 text-sm max-w-md mx-auto">
          {isFirstPage
            ? "Create your first short link using the form above."
            : "No links on this page."}
        </p>
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-white/[0.06]">
          <thead className="bg-white/[0.02]">
            <tr>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Short Link</th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Original URL</th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Created</th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Expires</th>
              <th scope="col" className="px-6 py-4 text-center text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Monetize</th>
              <th scope="col" className="px-6 py-4 text-right text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/[0.06]">
            {links.map((link) => (
              <tr key={link.slug} className="hover:bg-white/[0.02] transition-colors">
                <td className="px-6 py-4 whitespace-nowrap">
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
                </td>
                <td className="px-6 py-4">
                  <div className="flex items-center max-w-xs sm:max-w-md">
                    <span className="truncate text-white/60 text-sm">{link.original}</span>
                    <a href={link.original} target="_blank" rel="noopener noreferrer" className="ml-2 text-white/30 hover:text-[#6EE7B7] flex-shrink-0 transition-colors">
                      <ExternalLink className="h-4 w-4" />
                    </a>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center gap-1.5 text-sm text-white/50 font-mono-dm">
                    <Clock className="h-3.5 w-3.5 text-white/30" />
                    {timeAgo(link.created_at)}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-mono-dm">
                  {link.expires_at ? (
                    new Date(link.expires_at) < new Date() ? (
                      <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold bg-red-500/15 text-red-400 ring-1 ring-red-500/25">Expired</span>
                    ) : (
                      <span className="text-white/50">{timeAgo(link.expires_at)}</span>
                    )
                  ) : (
                    <span className="text-white/30">Never</span>
                  )}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-center">
                  {link.is_monetized ? (
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-semibold bg-[#6EE7B7]/15 text-[#6EE7B7] ring-1 ring-[#6EE7B7]/25">
                      <DollarSign size={10} />
                      On
                    </span>
                  ) : (
                    <span className="text-white/30 text-xs">Off</span>
                  )}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div className="flex justify-end gap-3">
                    <a href={link.qr_url} target="_blank" rel="noopener noreferrer" className="text-white/40 hover:text-[#6EE7B7] transition-colors" title="View QR Code">
                      <QrCode className="h-5 w-5" />
                    </a>
                    <Link href={`/dashboard/links/${link.slug}`} className="text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors" title="View Details">
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
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {(totalPages > 1 || links.length > 0) && (
        <div className="flex flex-col sm:flex-row items-center justify-between px-6 py-4 border-t border-white/[0.06] gap-4">
          <div className="flex items-center gap-4 order-2 sm:order-1">
            <span className="text-sm text-white/50 whitespace-nowrap">
              Page {page} of {totalPages}
            </span>
            <div className="flex items-center gap-2">
              <span className="text-xs font-bold text-white/30 uppercase tracking-widest font-mono-dm">Show</span>
              <select
                value={perPage}
                onChange={(e) => {
                  setPerPage(Number(e.target.value));
                  setPage(1);
                }}
                className="bg-white/[0.03] border border-white/[0.08] text-white/70 text-xs rounded-lg px-2 py-1 focus:outline-none focus:border-[#6EE7B7]/50 transition-all cursor-pointer"
              >
                {PAGE_SIZE_OPTIONS.map((size) => (
                  <option key={size} value={size} className="bg-[#0A0A0A]">
                    {size}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div className="flex gap-2 order-1 sm:order-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg border border-white/[0.08] bg-white/[0.03] text-white/70 hover:bg-white/[0.06] hover:text-white transition-colors disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
            >
              <ChevronLeft className="h-4 w-4" />
              Prev
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages || totalPages === 0}
              className="flex items-center gap-1 px-3 py-1.5 text-sm rounded-lg border border-white/[0.08] bg-white/[0.03] text-white/70 hover:bg-white/[0.06] hover:text-white transition-colors disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}