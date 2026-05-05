"use client";

import dynamic from "next/dynamic";
import { useAggregateStatsQuery } from "@/hooks/useAggregateStats";
import { useLinksQuery } from "@/hooks/useLinks";
import Loading from "@/components/ui/Loading";
import { Globe2, Link2, MousePointerClick } from "lucide-react";
import { format } from "date-fns";
import { getCountryName } from "@/lib/countries";

const GlobeView = dynamic(() => import("@/components/ui/GlobeView"), { ssr: false });

export default function DashboardGlobe() {
  const { data: linksData } = useLinksQuery(1, 1);
  const { data: stats, isLoading, error } = useAggregateStatsQuery();

  if (isLoading) {
    return <Loading height="h-96" />;
  }

  if (error || !stats || stats.total_clicks === 0) {
    return (
      <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-12 text-center">
        <Globe2 className="h-12 w-12 text-white/20 mx-auto mb-4" />
        <h3 className="text-xl font-bold text-white mb-2">No clicks yet</h3>
        <p className="text-white/50">Share your links to start seeing global traffic here.</p>
      </div>
    );
  }

  const globePoints = stats.top_countries.map((c) => ({
    country: c.country,
    count: c.count,
  }));

  const latestLink = linksData?.links?.[0];

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-6 flex items-center">
          <div className="h-12 w-12 rounded-xl bg-[#6EE7B7]/10 flex items-center justify-center mr-4">
            <MousePointerClick className="h-6 w-6 text-[#6EE7B7]" />
          </div>
          <div>
            <p className="text-sm font-medium text-white/50 font-mono-dm uppercase tracking-widest">Total Clicks</p>
            <p className="text-3xl font-black text-white stat-number">{stats.total_clicks.toLocaleString()}</p>
          </div>
        </div>

        <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-6 flex items-center">
          <div className="h-12 w-12 rounded-xl bg-[#6EE7B7]/10 flex items-center justify-center mr-4">
            <MousePointerClick className="h-6 w-6 text-[#6EE7B7]" />
          </div>
          <div>
            <p className="text-sm font-medium text-white/50 font-mono-dm uppercase tracking-widest">Unique Visitors</p>
            <p className="text-3xl font-black text-white stat-number">{stats.unique_clicks.toLocaleString()}</p>
          </div>
        </div>

        <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-6 flex items-center">
          <div className="h-12 w-12 rounded-xl bg-[#6EE7B7]/10 flex items-center justify-center mr-4">
            <Link2 className="h-6 w-6 text-[#6EE7B7]" />
          </div>
          <div>
            <p className="text-sm font-medium text-white/50 font-mono-dm uppercase tracking-widest">Active Links</p>
            <p className="text-3xl font-black text-white stat-number">{linksData?.total?.toLocaleString() ?? 0}</p>
          </div>
        </div>
      </div>

      <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-6">
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
        <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-6">
          <h3 className="text-sm font-medium text-white/50 mb-2">Last activity: {format(new Date(stats.clicks_per_day[0].date), 'MMM d, yyyy')}</h3>
        </div>
      )}
    </div>
  );
}