"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import Loading from "@/components/ui/Loading";
import { useLinkStatsQuery } from "@/hooks/useLinkStats";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie } from "recharts";
import { AxiosError } from "axios";
import { Activity, Globe2, MonitorSmartphone, Compass } from "lucide-react";
import { getCountryName } from "@/lib/countries";
import type { ApiErrorResponse } from "@/types/api";
import type { FormattedEntry } from "@/lib/stats-utils";
import type { AdEventItem } from "@/types/link";
import DataTable, { Column } from "@/components/ui/DataTable";
import {
  REJECT_REASON_HONEYPOT_HIT,
  REJECT_REASON_TOO_FAST,
  REJECT_REASON_NO_MOUSE_MOVEMENT,
  REJECT_REASON_DUPLICATE_IP,
  REJECT_REASON_DUPLICATE_FINGERPRINT,
  AD_EVENT_IMPRESSION,
  AD_EVENT_CLICK,
  AD_EVENT_COMPLETION,
  AD_EVENT_SKIP,
  DEFAULT_PAGE_SIZE,
} from "@/lib/constants";

const tooltipStyle = { backgroundColor: '#111', borderColor: 'rgba(255,255,255,0.1)', borderRadius: '0.75rem', color: '#fff' };

const GlobeView = dynamic(() => import("@/components/ui/GlobeView"), { ssr: false });

const formatRejectionReason = (reason?: string) => {
  if (!reason) return '—';
  const mappings: Record<string, string> = {
    [REJECT_REASON_DUPLICATE_IP]: 'Duplicate IP (24h)',
    [REJECT_REASON_DUPLICATE_FINGERPRINT]: 'Duplicate Browser (24h)',
    [REJECT_REASON_TOO_FAST]: 'Session Too Fast',
    [REJECT_REASON_HONEYPOT_HIT]: 'Bot Honeypot Triggered',
    [REJECT_REASON_NO_MOUSE_MOVEMENT]: 'No Mouse Movement'
  };
  return mappings[reason] || reason.replace(/_/g, ' ').toLowerCase().replace(/\b\w/g, c => c.toUpperCase());
};

export default function LinkStatCharts({ slug }: { slug: string }) {
  const [eventPage, setEventPage] = useState(1);
  const [eventPerPage, setEventPerPage] = useState(DEFAULT_PAGE_SIZE);
  const [eventSortBy, setEventSortBy] = useState("time");
  const [eventSortDir, setEventSortDir] = useState("desc");
  const { data: stats, isLoading, error, isFetching } = useLinkStatsQuery(slug, eventPage, eventPerPage, eventSortBy, eventSortDir);

  const handleEventSort = (col: string) => {
    setEventPage(1);
    if (eventSortBy === col) {
      setEventSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setEventSortBy(col);
      setEventSortDir("desc");
    }
  };

  const eventColumns: Column<AdEventItem>[] = [
    {
      header: "Time",
      accessorKey: "time",
      sortable: true,
      cell: (e) => <span className="text-white/60 text-xs">{new Date(e.time).toLocaleString()}</span>,
    },
    {
      header: "Event",
      accessorKey: "event_type",
      sortable: true,
      cell: (e) => (
        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${e.event_type === AD_EVENT_IMPRESSION ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20' :
          e.event_type === AD_EVENT_CLICK ? 'bg-green-500/10 text-green-400 border border-green-500/20' :
            e.event_type === AD_EVENT_COMPLETION ? 'bg-purple-500/10 text-purple-400 border border-purple-500/20' :
              e.event_type === AD_EVENT_SKIP ? 'bg-yellow-500/10 text-yellow-400 border border-yellow-500/20' :
                e.event_type === 'REJECTION' ? 'bg-red-500/10 text-red-400 border border-red-500/20' :
                  'bg-white/10 text-white/60 border border-white/10'
        }`}>
          {e.event_type}
        </span>
      ),
    },
    {
      header: "Ad",
      accessorKey: "ad_title",
      sortable: true,
      cell: (e) => <span className="text-white/80 max-w-[200px] truncate block">{e.ad_title}</span>,
    },
    {
      header: "Type",
      accessorKey: "ad_type",
      cell: (e) => <span className="text-white/60 capitalize">{e.ad_type}</span>,
    },
    {
      header: "Valid",
      accessorKey: "is_valid",
      cell: (e) => (
        <span>
          {e.event_type === AD_EVENT_COMPLETION ? (
            e.is_valid ? (
              <span className="text-green-400 text-xs font-medium">&#10003; Valid</span>
            ) : (
              <span className="text-red-400 text-xs font-medium">&#10007; Invalid</span>
            )
          ) : (
            <span className="text-white/40 text-xs">—</span>
          )}
        </span>
      ),
    },
    {
      header: "Quality",
      accessorKey: "quality_score",
      cell: (e) => (
        <span className="text-white/60 text-xs">
          {e.event_type === AD_EVENT_COMPLETION ? (e.quality_score || '—') : '—'}
        </span>
      ),
    },
    {
      header: "Reason",
      accessorKey: "rejection_reason",
      cell: (e) => (
        <span className="text-white/40 text-xs max-w-[150px] truncate block">
          {e.event_type === AD_EVENT_COMPLETION ? formatRejectionReason(e.rejection_reason) : '—'}
        </span>
      ),
    },
  ];

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

  const hasClicks = stats && stats.total_clicks > 0;
  const hasEvents = stats && stats.events && stats.events.length > 0;

  if (!stats || (!hasClicks && !hasEvents)) {
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

  const pagination = stats.event_pagination;

  return (
    <div className="space-y-6">
      {!hasClicks && (
        <div className="rounded-2xl border border-cyan-500/20 bg-cyan-500/5 p-6 flex flex-col md:flex-row items-center gap-4">
          <div className="h-10 w-10 rounded-xl bg-cyan-500/10 flex items-center justify-center shrink-0">
            <Activity className="h-5 w-5 text-cyan-400 animate-pulse" />
          </div>
          <div>
            <h4 className="text-sm font-bold text-white mb-1">Live Ad Activity Detected!</h4>
            <p className="text-xs text-white/60 leading-relaxed">
              Visitors are currently hitting your link and viewing advertisements. Interactive charts and geographical maps will populate as soon as the first visitor successfully completes the redirect flow.
            </p>
          </div>
        </div>
      )}

      {hasClicks && (
        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
          <h3 className="text-lg font-bold text-white mb-4 flex items-center">
            <Globe2 className="mr-2 h-5 w-5 text-[#6EE7B7]" />
            Click Distribution
          </h3>
          <GlobeView points={globePoints} height={420} />
        </div>
      )}

      {hasClicks && (
        <>
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
              <div className="h-48 w-full min-w-0">
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
              <div className="h-48 w-full min-w-0">
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
        </>
      )}

      {stats.events && stats.events.length > 0 && (
        <DataTable
          columns={eventColumns}
          data={stats.events}
          isLoading={isLoading}
          isFetching={isFetching}
          sortBy={eventSortBy}
          sortDir={eventSortDir}
          onSort={handleEventSort}
          page={eventPage}
          totalPages={pagination?.total_pages || 1}
          totalItems={pagination?.total || 0}
          onPageChange={setEventPage}
          perPage={eventPerPage}
          onPerPageChange={(p) => {
            setEventPerPage(p);
            setEventPage(1);
          }}
          title={
            <h3 className="text-lg font-bold text-white flex items-center py-2">
              <Activity className="mr-2 h-5 w-5 text-[#6EE7B7]" />
              Recent Activity
            </h3>
          }
        />
      )}
    </div>
  );
}
