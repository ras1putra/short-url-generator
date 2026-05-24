"use client";

import RequireRole from "@/components/auth/RequireRole";
import { useAds, useCreateAd, useDeleteAd, useAdTypes } from "@/hooks/useAds";
import { useWallet } from "@/hooks/wallet/useWallet";
import { useConfigStore } from "@/store/useConfigStore";
import { useCategories } from "@/hooks/useCategories";
import { Megaphone, Plus, Loader2, Trash2, BarChart3, X } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { createAdSchema, CreateAdForm } from "@/lib/validators";
import {
  ROLE_ADVERTISER,
  ROLE_ADMIN,
} from "@/lib/constants";


export default function CampaignsPage() {
  const { data: campaigns, isLoading } = useAds();
  const deleteAd = useDeleteAd();
  const [showCreate, setShowCreate] = useState(false);
  const { data: wallet } = useWallet();
  const cfg = useConfigStore((s) => s.config);
  const symbol = cfg?.token_symbol || "SURL";

  return (
    <RequireRole roles={[ROLE_ADVERTISER, ROLE_ADMIN]}>
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <div className="h-8 w-8 rounded-lg bg-[#22D3EE]/10 flex items-center justify-center">
                <Megaphone size={16} className="text-[#22D3EE]" />
              </div>
              <h1 className="text-3xl font-black tracking-tight text-white">Campaigns</h1>
            </div>
            <p className="mt-2 text-white/50 font-mono-dm text-sm">{"// Manage your ad campaigns"}</p>
          </div>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="btn-primary flex items-center gap-2 px-4 py-2.5 text-sm tracking-wider uppercase cursor-pointer"
          >
            {showCreate ? <X size={16} /> : <Plus size={16} />}
            {showCreate ? "Cancel" : "New Campaign"}
          </button>
        </div>
      </div>

      {showCreate && (
        <div className="mb-8 rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
          <h2 className="text-xl font-bold text-white/90 mb-6">Create Campaign</h2>
          <CreateCampaignForm
            onSuccess={() => setShowCreate(false)}
            walletBalance={Number(wallet?.balance || 0)}
            symbol={symbol}
          />
        </div>
      )}

      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="animate-spin h-8 w-8 text-white/30" />
        </div>
      ) : campaigns && campaigns.length > 0 ? (
        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-white/[0.06]">
              <thead className="bg-white/[0.02]">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Campaign</th>
                  <th className="px-6 py-4 text-left text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Category</th>
                  <th className="px-6 py-4 text-right text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Budget</th>
                  <th className="px-6 py-4 text-right text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">CPM</th>
                  <th className="px-6 py-4 text-center text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Status</th>
                  <th className="px-6 py-4 text-right text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/[0.06]">
                {campaigns.map((ad) => (
                  <tr key={ad.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="px-6 py-4">
                      <div>
                        <p className="text-white font-medium">{ad.title}</p>
                        {ad.description && (
                          <p className="text-white/40 text-sm truncate max-w-xs">{ad.description}</p>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-white/60 text-sm capitalize">{ad.category}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <p className="text-white font-mono-dm">{Number(ad.remaining_budget).toFixed(2)} {symbol}</p>
                      <p className="text-white/30 text-xs font-mono-dm">of {Number(ad.total_budget).toFixed(2)} {symbol}</p>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-white/60 font-mono-dm">
                      {Number(ad.cpm).toFixed(2)} {symbol}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-center">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold ${ad.status === "active"
                        ? "bg-green-500/15 text-green-400 ring-1 ring-green-500/25"
                        : ad.status === "paused"
                          ? "bg-yellow-500/15 text-yellow-400 ring-1 ring-yellow-500/25"
                          : "bg-white/10 text-white/50 ring-1 ring-white/20"
                        }`}>
                        {ad.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <div className="flex justify-end gap-3">
                        <Link
                          href={`/dashboard/campaigns/${ad.id}`}
                          className="text-[#22D3EE] hover:text-[#67E8F9] transition-colors"
                          title="View Details"
                        >
                          <BarChart3 className="h-5 w-5" />
                        </Link>
                        <button
                          onClick={() => deleteAd.mutate(ad.id, {
                            onSuccess: () => toast.success("Campaign deleted"),
                            onError: (err) => toast.error(err.response?.data?.message || "Delete failed"),
                          })}
                          className="text-white/40 hover:text-red-400 transition-colors cursor-pointer disabled:opacity-50"
                          title="Delete"
                          disabled={deleteAd.isPending}
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
        </div>
      ) : (
        <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-12 text-center">
          <Megaphone size={40} className="mx-auto mb-4 text-white/20" />
          <h2 className="text-xl font-bold text-white/60 mb-2">No campaigns yet</h2>
          <p className="text-white/40 text-sm max-w-md mx-auto">
            Create your first campaign to start advertising.
          </p>
        </div>
      )}
    </RequireRole>
  );
}

import { MediaUploader } from "@/components/campaigns/MediaUploader";

interface CreateCampaignFormProps {
  onSuccess: () => void;
  walletBalance: number;
  symbol: string;
}

function CreateCampaignForm({ onSuccess, walletBalance, symbol }: CreateCampaignFormProps) {
  const createAd = useCreateAd();
  const { data: categories } = useCategories();
  const { data: adTypes } = useAdTypes();
  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<CreateAdForm>({
    resolver: zodResolver(createAdSchema),
    defaultValues: {
      title: "",
      description: "",
      image_url: "",
      target_url: "",
      category: "regular",
      total_budget: "" as unknown as number,
      ad_type: "BANNER",
    },
  });

  // eslint-disable-next-line react-hooks/incompatible-library
  const adType = watch("ad_type");
  const category = watch("category");
  const imageUrl = watch("image_url");
  const adTypeKey = adType || "BANNER";

  const dynamicAdType = adTypes?.find((t) => t.ad_type === adTypeKey);
  const targetRatio = dynamicAdType ? parseFloat(dynamicAdType.aspect_ratio) : 1;
  const recommendedResolution = dynamicAdType ? dynamicAdType.recommended_resolution : "";

  const baseCpm = dynamicAdType ? parseFloat(dynamicAdType.cpm) : 0;
  const selectedCategory = categories?.find((c) => c.category === category);
  const categoryMultiplier = selectedCategory ? parseFloat(selectedCategory.multiplier) : 1;
  const effectiveCpm = baseCpm * categoryMultiplier;

  useEffect(() => {
    if (!categories || categories.length === 0) return;

    const currentCategory = category?.trim();
    const exists = !!currentCategory && categories.some((c) => c.category === currentCategory);
    if (exists) return;

    const fallback = categories.find((c) => c.category === "regular")?.category || categories[0].category;
    setValue("category", fallback, { shouldValidate: true });
  }, [categories, category, setValue]);

  const onSubmit = (data: CreateAdForm) => {
    if (Number(data.total_budget) > walletBalance) {
      toast.error(`Insufficient platform balance. Your wallet balance is only ${walletBalance.toFixed(2)} ${symbol}, but you entered a campaign budget of ${Number(data.total_budget).toFixed(2)} ${symbol}.`);
      return;
    }

    createAd.mutate(data, {
      onSuccess: () => {
        toast.success("Campaign created successfully!");
        onSuccess();
      },
      onError: (err) => {
        toast.error(err.response?.data?.message || "Failed to create campaign");
      },
    });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="flex items-center justify-between p-3.5 rounded-xl bg-white/[0.03] border border-white/[0.06] mb-4">
        <span className="text-xs text-white/40 font-mono-dm">{"// platform billing wallet"}</span>
        <span className="text-xs font-bold font-mono text-[#22D3EE]">
          Available Balance: {walletBalance.toFixed(2)} {symbol}
        </span>
      </div>

      <div className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Title</label>
            <input type="text" {...register("title")} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all" placeholder="Summer Campaign" />
            {errors.title && <p className="mt-1 text-xs text-red-400">{errors.title.message}</p>}
          </div>
          <div>
            <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Category</label>
            <select value={category || ""} {...register("category")} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all">
              {categories?.map((cat) => (
                <option key={cat.category} value={cat.category} className="bg-[#0A0A0A]">
                  {cat.label}
                </option>
              ))}
            </select>
          </div>
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

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Target URL</label>
            <input type="url" {...register("target_url")} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all" placeholder="https://example.com/landing" />
            {errors.target_url && <p className="mt-1 text-xs text-red-400">{errors.target_url.message}</p>}
          </div>
          <div>
            <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Description</label>
            <input type="text" {...register("description")} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all" placeholder="Optional description" />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Ad Format Type</label>
            <select {...register("ad_type")} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all">
              {adTypes && adTypes.length > 0 ? (
                adTypes.map((t) => (
                  <option key={t.ad_type} value={t.ad_type} className="bg-[#0A0A0A]">{t.label}</option>
                ))
              ) : (
                <option className="bg-[#0A0A0A]">Loading ad formats...</option>
              )}
            </select>
          </div>
          <div>
            <label className="block text-xs font-bold text-[#22D3EE] mb-2 uppercase tracking-widest font-mono-dm">Total Budget ({symbol})</label>
            <input type="number" step="0.01" {...register("total_budget", { valueAsNumber: true })} className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#22D3EE]/50 focus:outline-none sm:text-sm transition-all" placeholder="100.00" />
            {errors.total_budget && <p className="mt-1 text-xs text-red-400">{errors.total_budget.message}</p>}
          </div>
          <div>
            <label className="block text-xs font-bold text-white/40 mb-2 uppercase tracking-widest font-mono-dm">Platform CPM ({symbol})</label>
            <div className="w-full rounded-xl border border-white/10 bg-white/[0.05] px-3 py-2.5 text-white/70 font-mono sm:text-sm leading-relaxed">
              <span>{baseCpm.toFixed(2)}</span>
              {categoryMultiplier !== 1 && (
                <span> × {categoryMultiplier.toFixed(2)} ({selectedCategory?.label || category})</span>
              )}
              <span className="text-cyan-400"> = {effectiveCpm.toFixed(2)} {symbol}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="flex justify-end pt-4">
        <button
          type="submit"
          disabled={createAd.isPending}
          className="btn-primary flex items-center gap-2 px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50"
        >
          {createAd.isPending ? <Loader2 className="animate-spin h-4 w-4" /> : <Megaphone size={16} />}
          Create Campaign
        </button>
      </div>
    </form>
  );
}
