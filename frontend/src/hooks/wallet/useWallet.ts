import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { ApiResponse } from "@/types/api";
import { Wallet } from "@/types/ads";
import { API_WALLET } from "@/lib/constants";

export function useWallet(refetchIntervalMs: number | false = false) {
  return useQuery({
    queryKey: ["wallet"],
    queryFn: async () => {
      const response = await api.get<ApiResponse<Wallet>>(API_WALLET);
      return response.data.data;
    },
    refetchInterval: refetchIntervalMs,
  });
}
