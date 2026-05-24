"use client";

import dynamic from "next/dynamic";
import Loading from "@/components/ui/Loading";
import { useLinkStatsQuery } from "@/hooks/useLinkStats";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie } from "recharts";
import { AxiosError } from "axios";
import { Activity, Globe2, MonitorSmartphone, Compass } from "lucide-react";
import { getCountryName } from "@/lib/countries";
import type { ApiErrorResponse } from "@/types/api";
import type { FormattedEntry } from "@/lib/stats-utils";

const tooltipStyle = { backgroundColor: '#111', borderColor: 'rgba(255,255,255,0.1)', borderRadius: '0.75rem', color: '#fff' };

const GlobeView = dynamic(() => import("@/components/ui/GlobeView"), { ssr: false });

export default function LinkStatCharts({ slug }: { slug: string }) {
  const { data: stats, isLoading, error } = useLinkStatsQuery(slug);

  if (isLoading) {
    return <Loading height="h-64" />;
  }

  if (error) {
    return (
      <div className="rounded-2xl bg-red-900/20 p-6 border border-red-900/50 text-red-300">
        {(error as AxiosError<ApiErrorResponse>)?.response?.data?.message || "Failed to load statistics"}
      </div>
    );
  }

  if (!stats || stats.total_clicks === 0) {
    return (
      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-12 text-center">
        <Activity size={40} className="mx-auto mb-4 text-white/20" />
        <h2 className="text-xl font-bold text-white/60 mb-2">No clicks yet</h2>
        <p className="text-white/40 text-sm max-w-md mx-auto">Share your link to start seeing statistics here.</p>
      </div>
    );
  }

  const { browsers: browserData, devices: deviceData } = stats.formatted as { browsers: FormattedEntry[]; devices: FormattedEntry[] };

  const globePoints = stats.top_countries.map((c) => ({
    country: c.country,
    count: c.count,
  }));

  return (
    <div className="space-y-6">
      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
        <h3 className="text-lg font-bold text-white mb-4 flex items-center">
          <Globe2 className="mr-2 h-5 w-5 text-[#6EE7B7]" />
          Click Distribution
        </h3>
        <GlobeView points={globePoints} height={420} />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 flex items-center">
          <div className="h-12 w-12 rounded-xl bg-[#6EE7B7]/10 flex items-center justify-center mr-4">
            <Activity className="h-6 w-6 text-[#6EE7B7]" />
          </div>
          <div>
            <p className="text-xs font-bold text-white/50 font-mono-dm uppercase tracking-widest">Total Clicks</p>
            <p className="text-3xl font-black text-white stat-number">{stats.total_clicks}</p>
          </div>
        </div>

        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 flex items-center">
          <div className="h-12 w-12 rounded-xl bg-[#6EE7B7]/10 flex items-center justify-center mr-4">
            <Activity className="h-6 w-6 text-[#6EE7B7]" />
          </div>
          <div>
            <p className="text-xs font-bold text-white/50 font-mono-dm uppercase tracking-widest">Unique Visitors</p>
            <p className="text-3xl font-black text-white stat-number">{stats.unique_clicks}</p>
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
        <h3 className="text-lg font-bold text-white mb-6 flex items-center">
          <Activity className="mr-2 h-5 w-5 text-[#6EE7B7]" />
          Clicks Over Time
        </h3>
        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={stats.clicks_per_day} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" vertical={false} />
              <XAxis dataKey="date" stroke="rgba(255,255,255,0.3)" tick={{ fill: 'rgba(255,255,255,0.3)', fontSize: 12 }} />
              <YAxis stroke="rgba(255,255,255,0.3)" tick={{ fill: 'rgba(255,255,255,0.3)', fontSize: 12 }} />
              <Tooltip contentStyle={tooltipStyle} itemStyle={{ color: '#6EE7B7' }} />
              <Bar dataKey="count" fill="#6EE7B7" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
          <h3 className="text-lg font-bold text-white mb-6 flex items-center">
            <Globe2 className="mr-2 h-5 w-5 text-[#6EE7B7]" />
            Top Countries
          </h3>
          <ul className="space-y-4">
            {stats.top_countries.map((country) => (
              <li key={country.country} className="flex justify-between items-center">
                <span className="text-white/70 font-medium">{getCountryName(country.country) || "Unknown"}</span>
                <span className="bg-[#6EE7B7]/10 text-[#6EE7B7] text-xs font-bold px-2.5 py-1 rounded-full border border-[#6EE7B7]/20">
                  {country.count}
                </span>
              </li>
            ))}
          </ul>
        </div>

        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
          <h3 className="text-lg font-bold text-white mb-6 flex items-center">
            <MonitorSmartphone className="mr-2 h-5 w-5 text-[#6EE7B7]" />
            Devices
          </h3>
          <div className="h-48 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={deviceData}
                  cx="50%"
                  cy="50%"
                  innerRadius={50}
                  outerRadius={80}
                  paddingAngle={5}
                  dataKey="value"
                />
                <Tooltip contentStyle={tooltipStyle} />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="flex justify-center gap-4 mt-2 flex-wrap">
            {deviceData.map((entry) => (
              <div key={entry.name} className="flex items-center text-xs text-white/50 capitalize">
                <span className="w-3 h-3 rounded-full mr-1.5" style={{ backgroundColor: entry.fill }}></span>
                {entry.name}
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
          <h3 className="text-lg font-bold text-white mb-6 flex items-center">
            <Compass className="mr-2 h-5 w-5 text-[#6EE7B7]" />
            Browsers
          </h3>
          <div className="h-48 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={browserData}
                  cx="50%"
                  cy="50%"
                  innerRadius={50}
                  outerRadius={80}
                  paddingAngle={5}
                  dataKey="value"
                />
                <Tooltip contentStyle={tooltipStyle} />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="flex justify-center gap-4 mt-2 flex-wrap">
            {browserData.map((entry) => (
              <div key={entry.name} className="flex items-center text-xs text-white/50 capitalize">
                <span className="w-3 h-3 rounded-full mr-1.5" style={{ backgroundColor: entry.fill }}></span>
                {entry.name}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}