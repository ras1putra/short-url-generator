"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useCreateLink } from "@/hooks/useLinks";
import { linkSchema } from "@/lib/validators";
import { useState } from "react";
import { Loader2, Link as LinkIcon, Settings2, PlusCircle } from "lucide-react";
import { toast } from "sonner";

type LinkForm = z.infer<typeof linkSchema>;
type ExpiresUnit = "minutes" | "hours" | "days";

export default function CreateLinkForm() {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [formKey, setFormKey] = useState(0);
  const [expiresUnit, setExpiresUnit] = useState<ExpiresUnit>("days");

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<LinkForm>({
    resolver: zodResolver(linkSchema),
    defaultValues: {
      url: "",
      custom_slug: "",
      expires_unit: "days",
    }
  });

  const createMutation = useCreateLink();

  const onSubmit = (data: LinkForm) => {
    const url = /^https?:\/\//i.test(data.url) ? data.url : `https://${data.url}`;
    const payload: Record<string, string | number> = { url };
    if (data.custom_slug) payload.custom_slug = data.custom_slug;
    if (data.expires_value && !isNaN(data.expires_value)) {
      payload.expires_value = Number(data.expires_value);
      payload.expires_unit = data.expires_unit || "days";
    }
    createMutation.mutate(payload, {
      onSuccess: (res) => {
        toast.success(`Successfully shortened: ${res.data.short_url}`);
        setFormKey(prev => prev + 1);
        reset();
      },
      onError: (error) => {
        toast.error(error.response?.data?.message || "Failed to create short link.");
      }
    });
  };

  return (
    <div className="rounded-2xl bg-white/[0.02] border border-white/[0.08] p-6 mb-8" key={formKey}>
      <h2 className="text-xl font-bold text-white mb-6 flex items-center">
        <LinkIcon className="mr-2 h-5 w-5 text-[#6EE7B7]" />
        Create New Link
      </h2>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <div className="flex rounded-xl border border-white/10 bg-white/[0.03] focus-within:border-[#6EE7B7]/50 focus-within:ring-1 focus-within:ring-[#6EE7B7]/50 overflow-hidden transition-all">
            <input
              type="text"
              placeholder="https://example.com/very/long/url"
              className="block w-full bg-transparent px-4 py-3 text-white placeholder-white/20 focus:outline-none sm:text-sm"
              {...register("url")}
            />
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="btn-primary flex items-center justify-center px-6 py-3 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {createMutation.isPending ? <Loader2 className="animate-spin h-5 w-5" /> : <span className="flex items-center"><PlusCircle className="mr-2 h-4 w-4" /> Shorten</span>}
            </button>
          </div>
          {errors.url && <p className="mt-1 text-xs text-red-400 font-medium">{errors.url.message}</p>}
        </div>

        <div className="flex items-center">
          <button
            type="button"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="text-sm text-white/50 hover:text-[#6EE7B7] flex items-center transition-colors font-mono-dm cursor-pointer"
          >
            <Settings2 className="mr-1 h-4 w-4" />
            {showAdvanced ? "Hide Advanced Options" : "Show Advanced Options"}
          </button>
        </div>

        {showAdvanced && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 p-4 bg-white/[0.02] rounded-xl border border-white/[0.06]">
            <div>
              <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm">
                Custom Alias (Optional)
              </label>
              <div className="flex rounded-xl shadow-sm">
                <span className="inline-flex items-center rounded-l-xl border border-r-0 border-white/10 bg-white/[0.03] px-3 text-white/30 font-mono-dm text-sm">
                  /
                </span>
                <input
                  type="text"
                  className="block w-full min-w-0 flex-1 rounded-none rounded-r-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none sm:text-sm transition-all"
                  placeholder="my-campaign"
                  {...register("custom_slug")}
                />
              </div>
              {errors.custom_slug && <p className="mt-1 text-xs text-red-400 font-medium">{errors.custom_slug.message}</p>}
            </div>

            <div>
              <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm">
                Expiration (Optional)
              </label>
              <div className="flex gap-2">
                <input
                  type="number"
                  placeholder="e.g. 7"
                  className="block min-w-0 flex-1 appearance-none rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none sm:text-sm transition-all"
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
            </div>
          </div>
        )}
      </form>
    </div>
  );
}
