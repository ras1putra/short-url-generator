import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { ApiResponse } from "@/types/api";
import { Wallet } from "@/types/ads";
import { API_WALLET, DEFAULT_PAGE_SIZE } from "@/lib/constants";

export function useWallet(page: number = 1, perPage: number = DEFAULT_PAGE_SIZE, search?: string, sortBy?: string, sortDir?: string) {
  return useQuery({
    queryKey: ["wallet", page, perPage, search, sortBy, sortDir],
    queryFn: async () => {
      const params: Record<string, unknown> = { page, per_page: perPage };
      if (search) params.q = search;
      if (sortBy && sortBy !== "created_at") params.sort_by = sortBy;
      if (sortDir && sortDir !== "desc") params.sort_dir = sortDir;
      const response = await api.get<ApiResponse<Wallet>>(API_WALLET, { params });
      return response.data.data;
    },
    placeholderData: keepPreviousData,
  });
}
