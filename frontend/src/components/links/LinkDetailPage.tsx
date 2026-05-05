"use client";

import Loading from "@/components/ui/Loading";
import LinkStatCharts from "@/components/links/LinkStatCharts";
import { useLinkDetail, useUpdateLink } from "@/hooks/useLinks";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, ExternalLink, Copy, QrCode, Pencil, X, Loader2 } from "lucide-react";
import { useState } from "react";
import { format } from "date-fns";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Link as LinkType } from "@/types/link";

const editSchema = z.object({
  custom_slug: z.string().regex(/^[a-zA-Z0-9-]*$/, "Only letters, numbers, and dashes allowed").min(3).max(20).optional().or(z.literal("")),
  expires_value: z.number().min(1, "Minimum value is 1").optional().or(z.nan()),
  expires_unit: z.enum(["minutes", "hours", "days"]).optional(),
}).refine(
  (data) => {
    if (data.expires_value && !isNaN(data.expires_value) && !data.expires_unit) {
      return false;
    }
    return true;
  },
  { message: "Please select a unit", path: ["expires_unit"] }
);

type EditForm = z.infer<typeof editSchema>;

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

  if (error || !link) {
    return (
      <div className="rounded-2xl bg-red-900/20 p-6 border border-red-900/50 text-red-300">
        Failed to load link details. The link might not exist.
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <Link href="/dashboard" className="inline-flex items-center text-sm font-medium text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors mb-4">
          <ArrowLeft className="mr-1 h-4 w-4" />
          Back to Dashboard
        </Link>
        <h1 className="text-3xl font-black tracking-tight text-white">Link Details</h1>
        <p className="mt-2 text-white/50 font-mono-dm text-sm">Detailed info for <span className="font-semibold text-[#6EE7B7]">/{slug}</span></p>
      </div>

      <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-6 mb-8">
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

  const [expiresUnit, setExpiresUnit] = useState<"minutes" | "hours" | "days">("days");

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<EditForm>({
    resolver: zodResolver(editSchema),
    defaultValues: { custom_slug: "", expires_value: NaN, expires_unit: undefined as "minutes" | "hours" | "days" | undefined },
  });

  const onSubmit = (data: EditForm) => {
    const payload: Record<string, string | number> = {};
    if (data.custom_slug) payload.custom_slug = data.custom_slug;
    if (data.expires_value && !isNaN(data.expires_value)) {
      payload.expires_value = Number(data.expires_value);
      payload.expires_unit = data.expires_unit || "days";
    }

    if (Object.keys(payload).length === 0) {
      toast.error("No changes to save");
      return;
    }

    updateMutation.mutate(payload, {
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
                  className="block min-w-0 flex-1 appearance-none rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none sm:text-sm transition-all"
                  placeholder={link.expires_at ? "Set new expiry" : "No expiry set"}
                  {...register("expires_value", { valueAsNumber: true })}
                />
                <div className="shrink-0 flex rounded-xl border border-white/10 bg-white/[0.03] overflow-hidden">
                  {(["days", "hours", "minutes"] as const).map((unit) => (
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
