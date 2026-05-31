"use client";

import Loading from "@/components/ui/Loading";
import LinkStatCharts from "@/components/links/LinkStatCharts";
import { useLinkDetail, useUpdateLink } from "@/hooks/useLinks";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, ExternalLink, Copy, QrCode, Pencil, X, Loader2, Link2, DollarSign } from "lucide-react";
import { useState } from "react";
import { format } from "date-fns";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { AxiosError } from "axios";
import { toast } from "sonner";
import type { ApiErrorResponse } from "@/types/api";
import { Link as LinkType } from "@/types/link";
import { editSchema, type EditForm } from "@/lib/validators";
import { EXPIRY_UNIT_DAYS, EXPIRY_UNITS, ROUTE_LINKS } from "@/lib/constants";
import { useCategories } from "@/hooks/useCategories";


export default function LinkDetailClient({ slug }: { slug: string }) {
  const { data: link, isLoading, error } = useLinkDetail(slug);
  const [copied, setCopied] = useState(false);
  const [isEditing, setIsEditing] = useState(false);

  const copyToClipboard = (url: string) => {
    navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (isLoading) return <Loading height="h-64" />;

  if (error) {
    return (
      <div className="rounded-2xl bg-red-900/20 p-6 border border-red-900/50 text-red-300">
        {(error as AxiosError<ApiErrorResponse>)?.response?.data?.message || "Failed to load link details"}
      </div>
    );
  }

  if (!link) {
    return (
      <div className="rounded-2xl bg-red-900/20 p-6 border border-red-900/50 text-red-300">
        Link not found
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <Link href={ROUTE_LINKS} className="inline-flex items-center text-sm font-medium text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors mb-4">
          <ArrowLeft className="mr-1 h-4 w-4" />
          Back to Links
        </Link>
        <div className="flex items-center gap-3 mb-2">
          <div className="h-8 w-8 rounded-lg bg-[#6EE7B7]/10 flex items-center justify-center">
            <Link2 size={16} className="text-[#6EE7B7]" />
          </div>
          <h1 className="text-3xl font-black tracking-tight text-white">Link Details</h1>
        </div>
        <p className="mt-2 text-white/50 font-mono-dm text-sm">Detailed info for <span className="font-semibold text-[#6EE7B7]">/{slug}</span></p>
      </div>

      <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 mb-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-bold text-white/90">Info</h2>
          <button
            onClick={() => setIsEditing(!isEditing)}
            className="flex items-center gap-1.5 text-sm font-medium transition-colors cursor-pointer text-white/50 hover:text-[#6EE7B7]"
          >
            {isEditing ? (
              <>
                <X className="h-4 w-4" />
                Cancel
              </>
            ) : (
              <>
                <Pencil className="h-4 w-4" />
                Edit
              </>
            )}
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Short Link</p>
            <div className="flex items-center">
              <span className="text-[#6EE7B7] font-medium mr-2">{link.short_url}</span>
              <button
                onClick={() => copyToClipboard(link.short_url)}
                className="text-white/30 hover:text-white transition-colors cursor-pointer"
              >
                {copied ? <span className="text-xs text-[#6EE7B7] font-medium">Copied!</span> : <Copy className="h-4 w-4" />}
              </button>
            </div>
          </div>

          <div>
            <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Original URL</p>
            <div className="flex items-center">
              <span className="truncate text-white/60 text-sm mr-2">{link.original}</span>
              <a href={link.original} target="_blank" rel="noopener noreferrer" className="text-white/30 hover:text-[#6EE7B7] flex-shrink-0 transition-colors">
                <ExternalLink className="h-4 w-4" />
              </a>
            </div>
          </div>

          <div>
            <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Created</p>
            <span className="text-white/60 text-sm">{format(new Date(link.created_at), 'MMM d, yyyy HH:mm')}</span>
          </div>

          <div>
            <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Expires</p>
            <span className="text-white/60 text-sm">{link.expires_at ? format(new Date(link.expires_at), 'MMM d, yyyy HH:mm') : "Never"}</span>
          </div>

          <div>
            <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">QR Code</p>
            <a href={link.qr_url} target="_blank" rel="noopener noreferrer" className="inline-flex items-center text-[#6EE7B7] hover:text-[#A7F3D0] text-sm transition-colors">
              <QrCode className="mr-1.5 h-4 w-4" />
              View QR Code
            </a>
          </div>

          <div>
            <p className="text-xs font-bold text-white/50 uppercase tracking-widest font-mono-dm mb-1">Monetization</p>
            <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold ${
              link.is_monetized
                ? "bg-[#6EE7B7]/15 text-[#6EE7B7] ring-1 ring-[#6EE7B7]/25"
                : "bg-white/10 text-white/50 ring-1 ring-white/20"
            }`}>
              {link.is_monetized ? (link.allowed_categories?.length ? `Active (${link.allowed_categories.length} categories)` : "Active") : "Off"}
            </span>
          </div>
        </div>

        {isEditing && <EditSection link={link} slug={slug} onClose={() => setIsEditing(false)} />}
      </div>

      <div className="mb-8">
        <h2 className="text-xl font-bold text-white/90 mb-6">Statistics</h2>
        <LinkStatCharts slug={slug} />
      </div>
    </div>
  );
}

function EditSection({ link, slug, onClose }: { link: LinkType; slug: string; onClose: () => void }) {
  const updateMutation = useUpdateLink(slug);
  const router = useRouter();

  const [expiresUnit, setExpiresUnit] = useState<typeof EXPIRY_UNITS[number]>(EXPIRY_UNIT_DAYS);
  const [editingMonetized, setEditingMonetized] = useState(link.is_monetized);
  const [editingCategories, setEditingCategories] = useState<string[]>(link.allowed_categories || []);
  const { data: categories } = useCategories();

  const toggleCategory = (cat: string) => {
    setEditingCategories((prev) =>
      prev.includes(cat) ? prev.filter((c) => c !== cat) : [...prev, cat]
    );
  };

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<EditForm>({
    resolver: zodResolver(editSchema),
    defaultValues: { custom_slug: "", expires_value: NaN, expires_unit: undefined as typeof EXPIRY_UNITS[number] | undefined },
  });

  const onSubmit = (data: EditForm) => {
    const payload: Record<string, string | number | boolean | string[]> = {};
    if (data.custom_slug) payload.custom_slug = data.custom_slug;
    if (data.expires_value && !isNaN(data.expires_value)) {
      payload.expires_value = Number(data.expires_value);
      payload.expires_unit = data.expires_unit || EXPIRY_UNIT_DAYS;
    }
    payload.is_monetized = editingMonetized;
    payload.allowed_categories = editingCategories;

    if (Object.keys(payload).length === 0) {
      toast.error("No changes to save");
      return;
    }

    updateMutation.mutate(payload as Record<string, string | number>, {
      onSuccess: (response) => {
        toast.success("Link updated successfully");
        reset();
        onClose();
        const newSlug = response.data.slug;
        if (newSlug && newSlug !== slug) {
          router.replace(`/dashboard/links/${newSlug}`);
        }
      },
      onError: (error) => {
        toast.error(error.response?.data?.message || "Failed to update link");
      },
    });
  };

  return (
    <>
      <div className="border-t border-white/[0.06] my-6" />
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm">
              Custom Alias
            </label>
            <div className="flex rounded-xl shadow-sm">
              <span className="inline-flex items-center rounded-l-xl border border-r-0 border-white/10 bg-white/[0.03] px-3 text-white/30 font-mono-dm text-sm">
                /
              </span>
              <input
                type="text"
                className="block w-full min-w-0 flex-1 rounded-none rounded-r-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none sm:text-sm transition-all"
                placeholder={link.slug}
                {...register("custom_slug")}
              />
            </div>
            {errors.custom_slug && <p className="mt-1 text-xs text-red-400 font-medium">{errors.custom_slug.message}</p>}
            <p className="mt-1 text-xs text-white/30">Leave empty to keep current</p>
          </div>

          <div>
            <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm">
              Expiry
            </label>
<div className="flex gap-2">
                <input
                  type="number"
                  min="1"
                  onKeyDown={(e) => ["e", "E", "+", "-"].includes(e.key) && e.preventDefault()}
                  className="block min-w-0 flex-1 appearance-none rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none sm:text-sm transition-all"
                  placeholder={link.expires_at ? "Set new expiry" : "No expiry set"}
                  {...register("expires_value", { valueAsNumber: true })}
                />
                <div className="shrink-0 flex rounded-xl border border-white/10 bg-white/[0.03] overflow-hidden">
                  {EXPIRY_UNITS.map((unit) => (
                    <button
                      key={unit}
                      type="button"
                      onClick={() => { setExpiresUnit(unit); setValue("expires_unit", unit); }}
                      className={`px-3 py-2.5 text-xs font-medium tracking-wide transition-colors cursor-pointer ${
                        expiresUnit === unit
                          ? "bg-[#6EE7B7]/15 text-[#6EE7B7]"
                          : "text-white/40 hover:text-white/70"
                      }`}
                    >
                      {unit.charAt(0).toUpperCase() + unit.slice(1)}
                    </button>
                  ))}
                </div>
              </div>
            {(errors.expires_value || errors.expires_unit) && (
              <p className="mt-1 text-xs text-red-400 font-medium">
                {errors.expires_value?.message || errors.expires_unit?.message}
              </p>
            )}
            <p className="mt-1 text-xs text-white/30">Leave empty to keep current</p>
          </div>
        </div>

        <div className="border-t border-white/[0.06] pt-4">
          <div className="flex items-center justify-between mb-3">
            <label className="text-xs font-bold text-[#6EE7B7] uppercase tracking-widest font-mono-dm flex items-center gap-1.5">
              <DollarSign size={12} />
              Monetization
            </label>
            <button
              type="button"
              role="switch"
              aria-checked={editingMonetized}
              onClick={() => setEditingMonetized(!editingMonetized)}
              className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                editingMonetized ? "bg-[#6EE7B7]" : "bg-white/10"
              }`}
            >
              <span className={`pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow transition duration-200 ease-in-out ${
                editingMonetized ? "translate-x-4" : "translate-x-0"
              }`} />
            </button>
          </div>
          <p className="text-xs text-white/40 mb-3 leading-relaxed">
            Earn tokens when visitors see relevant ads before being redirected.
          </p>
          {editingMonetized && categories && categories.length > 0 && (
            <div>
              <label className="block text-xs font-bold text-white/40 mb-2 uppercase tracking-widest font-mono-dm">
                Allowed Ad Categories
              </label>
              <div className="flex flex-wrap gap-2">
                {categories.map((cat) => {
                  const isSelected = editingCategories.includes(cat.category);
                  return (
                    <button
                      key={cat.category}
                      type="button"
                      onClick={() => toggleCategory(cat.category)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all cursor-pointer ${
                        isSelected
                          ? "bg-[#6EE7B7]/15 text-[#6EE7B7] ring-1 ring-[#6EE7B7]/30"
                          : "bg-white/[0.03] text-white/40 hover:text-white/70 ring-1 ring-white/10"
                      }`}
                    >
                      {cat.label}
                    </button>
                  );
                })}
              </div>
              {editingCategories.length === 0 && (
                <p className="mt-1.5 text-xs text-white/30">Select at least one category to enable monetization.</p>
              )}
            </div>
          )}
        </div>

        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={() => { reset(); onClose(); }}
            className="px-4 py-2.5 text-sm text-white/50 hover:text-white transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="btn-primary flex items-center justify-center px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {updateMutation.isPending ? <Loader2 className="animate-spin h-5 w-5" /> : "Save Changes"}
          </button>
        </div>
      </form>
    </>
  );
}
