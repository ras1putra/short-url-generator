"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useCreateLink } from "@/hooks/useLinks";
import { useCategories } from "@/hooks/useCategories";
import { linkSchema } from "@/lib/validators";
import { EXPIRY_UNIT_DAYS, EXPIRY_UNITS } from "@/lib/constants";
import { useState } from "react";
import { Loader2, Settings2, PlusCircle, DollarSign } from "lucide-react";
import { toast } from "sonner";

type LinkForm = z.infer<typeof linkSchema>;
type ExpiresUnit = typeof EXPIRY_UNITS[number];


export default function CreateLinkForm() {
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [expiresUnit, setExpiresUnit] = useState<ExpiresUnit>(EXPIRY_UNIT_DAYS);
  const [isMonetized, setIsMonetized] = useState(false);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const { data: categories } = useCategories();

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
      expires_unit: EXPIRY_UNIT_DAYS,
    }
  });

  const createMutation = useCreateLink();

  const toggleCategory = (cat: string) => {
    setSelectedCategories((prev) =>
      prev.includes(cat) ? prev.filter((c) => c !== cat) : [...prev, cat]
    );
  };

  const onSubmit = (data: LinkForm) => {
    const url = /^https?:\/\//i.test(data.url) ? data.url : `https://${data.url}`;
    const payload: Record<string, string | number | boolean | string[]> = { url };
    if (data.custom_slug) payload.custom_slug = data.custom_slug;
    if (data.expires_value && !isNaN(data.expires_value)) {
      payload.expires_value = Number(data.expires_value);
      payload.expires_unit = data.expires_unit || EXPIRY_UNIT_DAYS;
    }
    payload.is_monetized = isMonetized;
    payload.allowed_categories = selectedCategories;
    createMutation.mutate(payload as Record<string, string | number>, {
      onSuccess: (res) => {
        toast.success(`Successfully shortened: ${res.data.short_url}`);
        reset();
        setIsMonetized(false);
        setSelectedCategories([]);
      },
      onError: (error) => {
        toast.error(error.response?.data?.message || "Failed to create short link.");
      }
    });
  };

  return (
    <div className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 mb-8">
      <h2 className="text-xl font-bold text-white/90 mb-6">Create New Link</h2>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <div className="flex rounded-xl border border-white/10 bg-white/[0.03] focus-within:border-[#6EE7B7]/50 focus-within:outline-none overflow-hidden transition-all">
            <input
              type="text"
              placeholder="https://example.com/very/long/url"
              className="block w-full bg-transparent px-3 py-2.5 text-white placeholder-white/20 focus:outline-none sm:text-sm"
              {...register("url")}
            />
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="btn-primary flex items-center justify-center px-6 py-2.5 text-sm tracking-wider uppercase cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {createMutation.isPending ? <Loader2 className="animate-spin h-4 w-4" /> : <span className="flex items-center"><PlusCircle className="mr-2 h-4 w-4" /> Shorten</span>}
            </button>
          </div>
          {errors.url && <p className="mt-1 text-xs text-red-400 font-medium">{errors.url.message}</p>}
        </div>

        <div className="flex items-center">
          <button
            type="button"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="text-sm text-white/50 hover:text-[#6EE7B7] flex items-center transition-colors cursor-pointer"
          >
            <Settings2 className="mr-1 h-4 w-4" />
            {showAdvanced ? "Hide Advanced Options" : "Show Advanced Options"}
          </button>
        </div>

        {showAdvanced && (
          <div className="space-y-4 p-4 rounded-2xl border border-white/[0.08] bg-white/[0.02]">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
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
                    min="1"
                    placeholder="e.g. 7"
                    onKeyDown={(e) => ["e", "E", "+", "-"].includes(e.key) && e.preventDefault()}
                    className="block min-w-0 flex-1 appearance-none rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2.5 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none sm:text-sm transition-all"
                    {...register("expires_value", { valueAsNumber: true })}
                  />
                  <div className="shrink-0 flex rounded-xl border border-white/10 bg-white/[0.03] overflow-hidden">
                    {EXPIRY_UNITS.map((unit) => (
                      <button
                        key={unit}
                        type="button"
                        onClick={() => { setExpiresUnit(unit); setValue("expires_unit", unit); }}
                        className={`px-3 py-2.5 text-xs font-medium tracking-wide transition-colors cursor-pointer ${expiresUnit === unit
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

            <div className="border-t border-white/[0.06] pt-4">
              <div className="flex items-center justify-between mb-3">
                <label className="text-xs font-bold text-[#6EE7B7] uppercase tracking-widest font-mono-dm flex items-center gap-1.5">
                  <DollarSign size={12} />
                  Monetization
                </label>
                <button
                  type="button"
                  role="switch"
                  aria-checked={isMonetized}
                  onClick={() => setIsMonetized(!isMonetized)}
                  className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${isMonetized ? "bg-[#6EE7B7]" : "bg-white/10"
                    }`}
                >
                  <span className={`pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow transition duration-200 ease-in-out ${isMonetized ? "translate-x-4" : "translate-x-0"
                    }`} />
                </button>
              </div>
              <p className="text-xs text-white/40 mb-3 leading-relaxed">
                Earn tokens when visitors see relevant ads before redirecting. Select which ad categories are allowed below.
              </p>
              {isMonetized && categories && categories.length > 0 && (
                <div>
                  <label className="block text-xs font-bold text-white/40 mb-2 uppercase tracking-widest font-mono-dm">
                    Allowed Ad Categories
                  </label>
                  <div className="flex flex-wrap gap-2">
                    {categories.map((cat) => {
                      const isSelected = selectedCategories.includes(cat.category);
                      return (
                        <button
                          key={cat.category}
                          type="button"
                          onClick={() => toggleCategory(cat.category)}
                          className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all cursor-pointer ${isSelected
                              ? "bg-[#6EE7B7]/15 text-[#6EE7B7] ring-1 ring-[#6EE7B7]/30"
                              : "bg-white/[0.03] text-white/40 hover:text-white/70 ring-1 ring-white/10"
                            }`}
                        >
                          {cat.label}
                        </button>
                      );
                    })}
                  </div>
                  {selectedCategories.length === 0 && (
                    <p className="mt-1.5 text-xs text-white/30">Select at least one category to enable monetization.</p>
                  )}
                </div>
              )}
            </div>
          </div>
        )}
      </form>
    </div>
  );
}
