"use client";

import dynamic from "next/dynamic";
import { useAggregateStatsQuery } from "@/hooks/useAggregateStats";
import { useLinksQuery } from "@/hooks/useLinks";
import Loading from "@/components/ui/Loading";
import { AxiosError } from "axios";
import { Globe2, Link2, MousePointerClick, Users } from "lucide-react";
import { format } from "date-fns";
import { getCountryName } from "@/lib/countries";
import type { ApiErrorResponse } from "@/types/api";

const GlobeView = dynamic(() => import("@/components/ui/GlobeView"), { ssr: false });

export default function DashboardGlobe() {
  const { data: linksData } = useLinksQuery(1, 1);
  const { data: stats, isLoading, error } = useAggregateStatsQuery();

  if (isLoading) {
    return <Loading height="h-96" />;
  }

  if (error) {
    return (
      <div className="rounded-2xl bg-red-900/20 p-4 sm:p-6 border border-red-900/50 text-red-300">
        {(error as AxiosError<ApiErrorResponse>)?.response?.data?.message || "Failed to load statistics"}
      </div>
    );
  }

  if (!stats || stats.total_clicks === 0) {
    return (
      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-12 text-center">
        <Globe2 size={40} className="mx-auto mb-4 text-white/20" />
        <h2 className="text-xl font-bold text-white/60 mb-2">No clicks yet</h2>
        <p className="text-white/40 text-sm max-w-md mx-auto">Share your links to start seeing global traffic here.</p>
      </div>
    );
  }

  const globePoints = stats.top_countries.map((c) => ({
    country: c.country,
    count: c.count,
  }));

  const latestLink = linksData?.links?.[0];

  return (
    <div className="space-y-4 sm:space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 sm:gap-6">
        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-4 sm:p-6 flex items-center">
          <div className="h-12 w-12 rounded-xl bg-[#6EE7B7]/10 flex items-center justify-center mr-4">
            <MousePointerClick className="h-6 w-6 text-[#6EE7B7]" />
          </div>
          <div>
            <p className="text-xs font-bold text-white/50 font-mono-dm uppercase tracking-widest">Total Clicks</p>
            <p className="text-3xl font-black text-white stat-number">{stats.total_clicks.toLocaleString()}</p>
          </div>
        </div>

        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-4 sm:p-6 flex items-center">
          <div className="h-12 w-12 rounded-xl bg-[#6EE7B7]/10 flex items-center justify-center mr-4">
            <Users className="h-6 w-6 text-[#6EE7B7]" />
          </div>
          <div>
            <p className="text-xs font-bold text-white/50 font-mono-dm uppercase tracking-widest">Unique Visitors</p>
            <p className="text-3xl font-black text-white stat-number">{stats.unique_clicks.toLocaleString()}</p>
          </div>
        </div>

        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-4 sm:p-6 flex items-center">
          <div className="h-12 w-12 rounded-xl bg-[#6EE7B7]/10 flex items-center justify-center mr-4">
            <Link2 className="h-6 w-6 text-[#6EE7B7]" />
          </div>
          <div>
            <p className="text-xs font-bold text-white/50 font-mono-dm uppercase tracking-widest">Active Links</p>
            <p className="text-3xl font-black text-white stat-number">{linksData?.total?.toLocaleString() ?? 0}</p>
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-4 sm:p-6">
        <h3 className="text-lg font-bold text-white mb-4 flex items-center">
          <Globe2 className="mr-2 h-5 w-5 text-[#6EE7B7]" />
          Global Reach
        </h3>
        <GlobeView points={globePoints} height={420} />
        {stats.top_countries.length > 0 && (
          <div className="mt-4 flex flex-wrap gap-3">
            {stats.top_countries.slice(0, 7).map((c) => (
              <span key={c.country} className="bg-[#6EE7B7]/10 text-[#6EE7B7] text-xs font-bold px-3 py-1.5 rounded-full border border-[#6EE7B7]/20">
                {getCountryName(c.country)} &middot; {c.count.toLocaleString()}
              </span>
            ))}
          </div>
        )}
      </div>

      {latestLink && stats.clicks_per_day.length > 0 && (
        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-4 sm:p-6">
          <h3 className="text-sm font-medium text-white/50 mb-2">Last activity: {format(new Date(stats.clicks_per_day[0].date),'MMM d, yyyy')}</h3>
        </div>
      )}
    </div>
  );
}