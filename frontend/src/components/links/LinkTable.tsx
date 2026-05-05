"use client";

import Loading from "@/components/ui/Loading";
import { useLinksQuery, useLinkActions } from "@/hooks/useLinks";
import { formatDistanceToNowStrict } from "date-fns";
import { BarChart3, Copy, ExternalLink, QrCode, Trash2, ChevronLeft, ChevronRight, Clock } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

const PER_PAGE = 5;

function timeAgo(date: string) {
  return formatDistanceToNowStrict(new Date(date), { addSuffix: true });
}

export default function LinkTable() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useLinksQuery(page, PER_PAGE);
  const { copiedSlug, copyToClipboard, deleteLink, isDeleting } = useLinkActions();

  if (isLoading) {
    return <Loading height="h-auto" className="p-16" />;
  }

  const links = data?.links || [];
  const totalPages = data?.total_pages || 1;

  if (links.length === 0 && page === 1) {
    return (
      <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-12 flex flex-col items-center justify-center text-center">
        <div className="h-16 w-16 rounded-full bg-white/[0.04] flex items-center justify-center mb-4">
          <BarChart3 className="h-8 w-8 text-white/30" />
        </div>
        <h3 className="text-lg font-bold text-white mb-1">No links yet</h3>
        <p className="text-white/50">Create your first short link using the form above.</p>
      </div>
    );
  }

  return (
    <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-white/[0.06]">
          <thead className="bg-white/[0.02]">
            <tr>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Short Link</th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Original URL</th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Created</th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Expires</th>
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
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div className="flex justify-end gap-3">
                    <a href={link.qr_url} target="_blank" rel="noopener noreferrer" className="text-white/40 hover:text-[#6EE7B7] transition-colors" title="View QR Code">
                      <QrCode className="h-5 w-5" />
                    </a>
                    <Link href={`/dashboard/links/${link.slug}`} className="text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors" title="View Details">
                      <BarChart3 className="h-5 w-5" />
                    </Link>
                    <button
                      onClick={() => deleteLink(link.slug)}
                      className="text-white/40 hover:text-red-400 hover:cursor-pointer transition-colors disabled:opacity-50"
                      title="Delete Link"
                      disabled={isDeleting}
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

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-6 py-4 border-t border-white/[0.06]">
          <span className="text-sm text-white/50">
            Page {page} of {totalPages}
          </span>
          <div className="flex gap-2">
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
              disabled={page === totalPages}
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