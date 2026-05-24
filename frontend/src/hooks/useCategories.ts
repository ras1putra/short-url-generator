"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { API_CATEGORIES } from "@/lib/constants";
import type { ApiResponse } from "@/types/api";

export interface Category {
  category: string;
  label: string;
  multiplier: string;
}

export function useCategories() {
  return useQuery({
    queryKey: ["ad-categories"],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Category[]>>(API_CATEGORIES);
      return res.data.data;
    },
  });
}
