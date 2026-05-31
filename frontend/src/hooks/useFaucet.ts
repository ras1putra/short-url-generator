import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { ApiResponse } from "@/types/api";
import { API_FAUCET_CLAIM, API_FAUCET_CONFIRM, API_FAUCET_DEV_ETH, API_FAUCET_HISTORY } from "@/lib/constants";

export interface FaucetClaimPayload {
  wallet: string;
  amount: string;
  nonce: string;
  deadline: string;
  signature: string;
  faucet_addr: string;
  chain_id: number;
}

export interface FaucetConfirmPayload {
  status: string;
  tx_hash: string;
}

export function useFaucetClaim() {
  return useMutation({
    mutationFn: async (walletAddr: string) => {
      const response = await api.post<ApiResponse<FaucetClaimPayload>>(API_FAUCET_CLAIM, {
        wallet_addr: walletAddr,
      });
      return response.data.data;
    },
  });
}

export function useFaucetConfirm() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: { txHash: string; walletAddr: string }) => {
      const response = await api.post<ApiResponse<FaucetConfirmPayload>>(API_FAUCET_CONFIRM, {
        tx_hash: params.txHash,
        wallet_addr: params.walletAddr,
      });
      return response.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["faucet-history"] });
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    },
  });
}

export function useDevETHClaim() {
  return useMutation({
    mutationFn: async (walletAddr: string) => {
      const response = await api.post<ApiResponse<{ tx_hash: string }>>(API_FAUCET_DEV_ETH, {
        wallet_addr: walletAddr,
      });
      return response.data.data;
    },
  });
}

export interface FaucetHistoryItem {
  id: string;
  amount: string;
  tx_hash: string;
  claimed_at: string;
}

export interface FaucetHistoryResponse {
  claims: FaucetHistoryItem[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export function useFaucetHistory(page: number = 1, perPage: number = 10, search?: string, sortBy?: string, sortDir?: string) {
  return useQuery({
    queryKey: ["faucet-history", page, perPage, search, sortBy, sortDir],
    queryFn: async () => {
      const params: Record<string, string | number> = { page, per_page: perPage };
      if (search) params.q = search;
      if (sortBy && sortBy !== "claimed_at") params.sort_by = sortBy;
      if (sortDir && sortDir !== "desc") params.sort_dir = sortDir;
      const response = await api.get<ApiResponse<FaucetHistoryResponse>>(API_FAUCET_HISTORY, { params });
      return response.data.data;
    },
  });
}
