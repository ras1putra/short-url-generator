"use client";

import { useState } from "react";
import { useCampaignDetail } from "@/hooks/useCampaignDetail";
import { Megaphone, ArrowLeft, ExternalLink, Pencil, X, Loader2, Save, FileVideo, Image as ImageIcon, Coins } from "lucide-react";
import Link from "next/link";
import { format } from "date-fns";
import { MediaUploader } from "@/components/campaigns/MediaUploader";
import { useCategories } from "@/hooks/useCategories";

import { ROUTE_CAMPAIGNS } from "@/lib/constants";
import { useConfigStore } from "@/store/useConfigStore";
import { isVideoUrl, isGifUrl } from "@/lib/media";
import { useAdTypes } from "@/hooks/useAds";

export default function CampaignDetailClient({ id }: { id: string }) {
  const cfg = useConfigStore((s) => s.config);
  const symbol = cfg?.token_symbol || "SURL";
  const { data: categories } = useCategories();
  const { data: adTypes } = useAdTypes();

  const [topUpVal, setTopUpVal] = useState("");

  const {
    ad,
    stats,
    wallet,
    isEditing,
    setIsEditing,
    form: {
      register,
      handleSubmit,
      setValue,
      watch,
      formState: { errors },
    },
    handleToggleStatus,
    onSave,
    handleTopUp,
    topUpAd,
    isLoading,
    error,
    updateAd,
  } = useCampaignDetail(id);

  const handleTopUpSubmit = (e: React.SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault();
    handleTopUp(parseFloat(topUpVal), () => setTopUpVal(""));
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="animate-spin h-8 w-8 text-white/30" />
      </div>
    );
  }

  if (error || !ad) {
    return (
      <div className="rounded-2xl bg-red-900/20 p-6 border border-red-900/50 text-red-300">
        Campaign not found.
      </div>
    );
  }

  const imageUrl = watch("image_url");
  const adTypeValue = watch("ad_type");
  const category = watch("category");
  const adTypeKey = adTypeValue || "BANNER";

  const dynamicAdType = adTypes?.find((t) => t.ad_type === adTypeKey);
  const targetRatio = dynamicAdType ? parseFloat(dynamicAdType.aspect_ratio) : 1;
  const recommendedResolution = dynamicAdType ? dynamicAdType.recommended_resolution : "";

  const selectedCategory = categories?.find((c) => c.category === category);
  const storedCpm = ad ? Number(ad.cpm) : 0;

  return (
    <div>
      <div className="mb-8">
        <Link href={ROUTE_CAMPAIGNS} className="inline-flex items-center text-sm font-medium text-[#22D3EE] hover:text-[#67E8F9] transition-colors mb-4">
          <ArrowLeft className="mr-1 h-4 w-4" />
          Back to Campaigns
        </Link>
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <div className="h-8 w-8 rounded-lg bg-[#22D3EE]/10 flex items-center justify-center">
                <Megaphone size={16} className="text-[#22D3EE]" />
              </div>
              <h1 className="text-3xl font-black tracking-tight text-white">{ad.title}</h1>
            </div>
            <p className="mt-2 text-white/50 font-mono-dm text-sm">{"// Campaign ID: "}{ad.id}</p>
          </div>
          <div className="flex gap-3">
            <button
              onClick={handleToggleStatus}
              disabled={updateAd.isPending}
              className="btn-primary flex items-center gap-2 px-4 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50"
            >
              {updateAd.isPending ? <Loader2 className="animate-spin h-4 w-4" /> : null}
              {ad.status === "active" ? "Pause" : "Activate"}
            </button>
          </div>
        </div>
      </div>

      {/* Ad Creative Preview Card */}
      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 mb-6">
        <h3 className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-4">Ad Creative Preview</h3>
        {imageUrl ? (
          <div className="relative group overflow-hidden rounded-xl border border-white/10 bg-white/[0.01] max-h-[320px] flex items-center justify-center p-0"
            style={{ aspectRatio: targetRatio, maxWidth: "100%" }}
          >
            {isVideoUrl(imageUrl) ? (
              <div className="relative w-full h-full">
                <video
                  src={imageUrl}
                  muted
                  playsInline
                  preload="metadata"
                  className="absolute inset-0 w-full h-full rounded-lg object-cover pointer-events-none"
                />
                <div className="absolute top-2 left-2 px-2 py-0.5 rounded bg-black/60 backdrop-blur-md border border-white/10 flex items-center gap-1.5 text-[10px] font-mono text-cyan-400">
                  <FileVideo size={10} />
                  VIDEO AD
                </div>
              </div>
            ) : isGifUrl(imageUrl) ? (
              <div className="relative w-full h-full">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={imageUrl}
                  alt={ad?.title || "Campaign"}
                  className="absolute inset-0 w-full h-full rounded-lg object-cover transition-transform duration-500 group-hover:scale-105"
                />
                <div className="absolute top-2 left-2 px-2 py-0.5 rounded bg-black/60 backdrop-blur-md border border-white/10 flex items-center gap-1.5 text-[10px] font-mono text-cyan-400">
                  <ImageIcon size={10} />
                  GIF AD
                </div>
              </div>
            ) : (
              <div className="relative w-full h-full">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={imageUrl}
                  alt={ad?.title || "Campaign"}
                  className="absolute inset-0 w-full h-full rounded-lg object-cover transition-transform duration-500 group-hover:scale-105"
                  onError={(e) => {
                    (e.target as HTMLImageElement).src = "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=600&auto=format&fit=crop";
                  }}
                />
                <div className="absolute top-2 left-2 px-2 py-0.5 rounded bg-black/60 backdrop-blur-md border border-white/10 flex items-center gap-1.5 text-[10px] font-mono text-cyan-400">
                  <ImageIcon size={10} />
                  IMAGE AD
                </div>
              </div>
            )}
            <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-end p-4">
              <a
                href={ad.target_url}
                target="_blank"
                rel="noopener noreferrer"
                className="w-full flex items-center justify-center gap-1.5 py-2 px-4 rounded-lg bg-[#22D3EE] text-black font-bold text-xs uppercase tracking-wider hover:bg-[#67E8F9] transition-all"
              >
                Test Link
                <ExternalLink size={12} />
              </a>
            </div>
          </div>
        ) : (
          <div className="rounded-xl border border-dashed border-white/10 bg-white/[0.01] flex items-center justify-center p-8 min-h-[120px] text-center">
            <Megaphone className="h-10 w-10 text-white/20 mb-3" />
            <p className="text-sm text-white/40">No creative uploaded</p>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        {[
          { label: "Impressions", value: stats?.impressions || 0, color: "text-[#22D3EE]" },
          { label: "Clicks", value: stats?.clicks || 0, color: "text-green-400" },
          { label: "Completions", value: stats?.completions || 0, color: "text-yellow-400" },
        ].map((stat) => (
          <div key={stat.label} className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
            <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-2">{stat.label}</p>
            <p className={`text-3xl font-black ${stat.color}`}>{stat.value.toLocaleString()}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Left Column: Details Card */}
        <div className="lg:col-span-2 rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-bold text-white/90 mb-6">{isEditing ? "Edit Campaign" : "Details"}</h2>
            <button
              onClick={() => setIsEditing(!isEditing)}
              className="flex items-center gap-1.5 text-sm font-medium transition-colors cursor-pointer text-white/50 hover:text-[#22D3EE]"
            >
              {isEditing ? <X className="h-4 w-4" /> : <Pencil className="h-4 w-4" />}
              {isEditing ? "Cancel" : "Edit"}
            </button>
          </div>

          {isEditing ? (
            <form onSubmit={handleSubmit(onSave)} className="space-y-4">
              <div className="space-y-4">
                <div>
                  <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Title</label>
                  <input type="text" {...register("title")} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all" placeholder="Campaign title" />
                  {errors.title && <p className="mt-1 text-xs text-red-400">{errors.title.message}</p>}
                </div>

                <div>
                  <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Ad Creative Media</label>
                  <MediaUploader
                    value={imageUrl}
                    onChange={(url) => setValue("image_url", url, { shouldValidate: !!url })}
                    targetRatio={targetRatio}
                    recommendedResolution={recommendedResolution}
                  />
                  {errors.image_url && <p className="mt-1 text-xs text-red-400">{errors.image_url.message}</p>}
                </div>

                <div>
                  <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Target URL</label>
                  <input type="url" {...register("target_url")} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all" />
                  {errors.target_url && <p className="mt-1 text-xs text-red-400">{errors.target_url.message}</p>}
                </div>

                <div>
                  <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Description</label>
                  <textarea {...register("description")} rows={2} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all placeholder-white/20" placeholder="Optional description..." />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div>
                    <label className="block text-xs font-bold text-white/40 mb-2 uppercase tracking-widest font-mono-dm">Ad Format</label>
                    <div className="w-full rounded-xl border border-white/10 bg-white/[0.05] px-3 py-2.5 text-white/70 font-mono sm:text-sm">
                      {adTypes?.find((t) => t.ad_type === ad?.ad_type)?.label || ad?.ad_type}
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-bold text-white/40 mb-2 uppercase tracking-widest font-mono-dm">Category</label>
                    <div className="w-full rounded-xl border border-white/10 bg-white/[0.05] px-3 py-2.5 text-white/70 font-mono sm:text-sm">
                      {selectedCategory?.label || category}
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-bold text-white/40 mb-2 uppercase tracking-widest font-mono-dm">Platform CPM ({symbol})</label>
                    <div className="w-full rounded-xl border border-white/10 bg-white/[0.05] px-3 py-2.5 text-white/70 font-mono sm:text-sm leading-relaxed">
                      <span>{storedCpm.toFixed(2)} {symbol}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div className="flex justify-end pt-4">
                <button
                  type="submit"
                  disabled={updateAd.isPending}
                  className="btn-primary flex items-center gap-2 px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50"
                >
                  {updateAd.isPending ? <Loader2 className="animate-spin h-4 w-4" /> : <Save size={16} />}
                  Save Changes
                </button>
              </div>
            </form>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Category</p>
                <span className="text-white/60 text-sm capitalize">{ad.category}</span>
              </div>
              <div>
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Status</p>
                <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold ${ad.status === "active"
                    ? "bg-green-500/15 text-green-400 ring-1 ring-green-500/25"
                    : ad.status === "paused"
                      ? "bg-yellow-500/15 text-yellow-400 ring-1 ring-yellow-500/25"
                      : "bg-white/10 text-white/50 ring-1 ring-white/20"
                  }`}>
                  {ad.status}
                </span>
              </div>
              <div>
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Ad Format Type</p>
                <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-mono font-bold bg-[#22D3EE]/15 text-[#22D3EE] ring-1 ring-[#22D3EE]/25 uppercase">
                  {ad.ad_type || "BANNER"}
                </span>
              </div>
              <div>
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Total Budget</p>
                <span className="text-white/60 font-mono-dm">{Number(ad.total_budget).toFixed(2)} {symbol}</span>
              </div>
              <div>
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Remaining Budget</p>
                <span className="text-white/60 font-mono-dm">{Number(ad.remaining_budget).toFixed(2)} {symbol}</span>
              </div>
              <div>
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">CPM</p>
                <span className="text-white/60 font-mono-dm">{Number(ad.cpm).toFixed(2)} {symbol}</span>
              </div>
              <div className="md:col-span-2">
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Description</p>
                <span className="text-white/60 text-sm">{ad.description || <span className="text-white/20 font-mono-dm">{"// none"}</span>}</span>
              </div>
              <div className="md:col-span-2">
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Target URL</p>
                <a href={ad.target_url} target="_blank" rel="noopener noreferrer" className="inline-flex items-center text-[#22D3EE] hover:text-[#67E8F9] text-sm transition-colors break-all">
                  <ExternalLink className="mr-1.5 h-4 w-4 shrink-0" />
                  {ad.target_url}
                </a>
              </div>
              <div>
                <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Created</p>
                <span className="text-white/60 text-sm">{format(new Date(ad.created_at), "MMM d, yyyy HH:mm")}</span>
              </div>
            </div>
          )}
        </div>

        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
          <div className="flex items-center gap-2 mb-4">
            <div className="p-2 rounded-lg bg-[#22D3EE]/10">
              <Coins className="h-5 w-5 text-[#22D3EE]" />
            </div>
            <h2 className="text-lg font-bold text-white/90">Top-Up Budget</h2>
          </div>

          <p className="text-xs text-white/50 mb-6 leading-relaxed">
            Inject {symbol} tokens directly into this campaign&apos;s budget to instantly boost or resume ad delivery.
          </p>

          <form onSubmit={handleTopUpSubmit} className="space-y-4">
            <div>
              <div className="flex justify-between items-center mb-2">
                <label className="text-xs font-bold text-[#22D3EE] uppercase tracking-widest font-mono-dm">
                  Amount to Add
                </label>
                {wallet && (
                  <span className="text-xs text-white/40">
                    Wallet: {Number(wallet.balance).toFixed(2)} {symbol}
                  </span>
                )}
              </div>
              <div className="relative rounded-xl border border-white/10 bg-white/[0.03] focus-within:border-[#22D3EE]/50 transition-all">
                <input
                  type="number"
                  step="any"
                  min="0.01"
                  required
                  value={topUpVal}
                  onChange={(e) => setTopUpVal(e.target.value)}
                  placeholder="0.00"
                  className="w-full bg-transparent border-0 pl-3 pr-16 py-2.5 text-white placeholder-white/20 focus:outline-none sm:text-sm font-mono-dm"
                />
                <div className="absolute inset-y-0 right-3 flex items-center pointer-events-none text-xs font-bold text-white/40 font-mono-dm">
                  {symbol}
                </div>
              </div>
            </div>

            <button
              type="submit"
              disabled={topUpAd.isPending}
              className="btn-primary w-full flex items-center justify-center gap-2 py-2.5 px-4 text-xs tracking-wider uppercase cursor-pointer disabled:opacity-50"
            >
              {topUpAd.isPending ? (
                <Loader2 className="animate-spin h-4 w-4" />
              ) : (
                <Coins size={14} />
              )}
              Add Budget
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
