"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useAd, useAdStats, useUpdateAd, useTopUpAd } from "@/hooks/useAds";
import { useWallet } from "@/hooks/wallet/useWallet";
import { createAdSchema, CreateAdForm } from "@/lib/validators";

export function useCampaignDetail(id: string) {
  const { data: ad, isLoading, error } = useAd(id);
  const { data: stats } = useAdStats(id);
  const { data: wallet } = useWallet();
  const updateAd = useUpdateAd(id);
  const topUpAd = useTopUpAd(id);
  const [isEditing, setIsEditing] = useState(false);

  const form = useForm<CreateAdForm>({
    resolver: zodResolver(createAdSchema),
    defaultValues: {
      title: "",
      description: "",
      image_url: "",
      target_url: "",
      category: "regular",
      total_budget: 0,
      ad_type: "BANNER",
    },
  });

  const { reset } = form;

  useEffect(() => {
    if (ad && !isEditing) {
      reset({
        title: ad.title,
        description: ad.description || "",
        image_url: ad.image_url,
        target_url: ad.target_url,
        category: ad.category,
        total_budget: Number(ad.total_budget),
        ad_type: ad.ad_type || "BANNER",
      });
    }
  }, [ad, reset, isEditing]);

  const handleToggleStatus = () => {
    const newStatus = ad?.status === "active" ? "paused" : "active";
    updateAd.mutate(
      { status: newStatus },
      {
        onSuccess: () => toast.success(`Campaign ${newStatus}`),
        onError: (err) => toast.error(err.response?.data?.message || "Update failed"),
      }
    );
  };

  const onSave = (data: CreateAdForm) => {
    if (!ad) return;

    updateAd.mutate(
      {
        title: data.title,
        description: data.description,
        image_url: data.image_url,
        target_url: data.target_url,
      },
      {
        onSuccess: () => {
          toast.success("Campaign updated successfully");
          setIsEditing(false);
        },
        onError: (err) => {
          toast.error(err.response?.data?.message || "Failed to update campaign");
        },
      }
    );
  };

  const handleTopUp = (amount: number, onSuccess?: () => void) => {
    if (isNaN(amount) || amount <= 0) {
      toast.error("Please enter a valid amount greater than 0");
      return;
    }

    topUpAd.mutate(
      { amount },
      {
        onSuccess: () => {
          toast.success("Campaign budget topped up successfully!");
          if (onSuccess) onSuccess();
        },
        onError: (err) => {
          toast.error(err.response?.data?.message || "Failed to top up campaign");
        },
      }
    );
  };

  return {
    ad,
    stats,
    wallet,
    updateAd,
    topUpAd,
    isEditing,
    setIsEditing,
    form,
    handleToggleStatus,
    onSave,
    handleTopUp,
    isLoading,
    error,
  };
}
