import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { AxiosError } from "axios";
import { ApiResponse, ApiErrorResponse } from "@/types/api";
import { Ad, AdStats, CreateAdPayload, UpdateAdPayload, CampaignListResponse } from "@/types/ads";
import { API_ADS, DEFAULT_PAGE_SIZE } from "@/lib/constants";

export function useAds(page: number = 1, perPage: number = DEFAULT_PAGE_SIZE, search?: string, sortBy?: string, sortDir?: string) {
  return useQuery({
    queryKey: ["ads", page, perPage, search, sortBy, sortDir],
    queryFn: async () => {
      const params: Record<string, unknown> = { page, per_page: perPage };
      if (search) params.q = search;
      if (sortBy && sortBy !== "created_at") params.sort_by = sortBy;
      if (sortDir && sortDir !== "desc") params.sort_dir = sortDir;
      const response = await api.get<ApiResponse<CampaignListResponse>>(API_ADS, { params });
      return response.data.data;
    },
    placeholderData: keepPreviousData,
  });
}

export function useAd(id: string) {
  return useQuery({
    queryKey: ["ad", id],
    queryFn: async () => {
      const response = await api.get<ApiResponse<Ad>>(`${API_ADS}/${id}`);
      return response.data.data;
    },
    enabled: !!id,
  });
}

export function useAdStats(id: string) {
  return useQuery({
    queryKey: ["ad", id, "stats"],
    queryFn: async () => {
      const response = await api.get<ApiResponse<AdStats>>(`${API_ADS}/${id}/stats`);
      return response.data.data;
    },
    enabled: !!id,
  });
}

export function useCreateAd() {
  const queryClient = useQueryClient();
  return useMutation<ApiResponse<Ad>, AxiosError<ApiErrorResponse>, CreateAdPayload>({
    mutationFn: async (payload) => {
      const response = await api.post<ApiResponse<Ad>>(API_ADS, payload);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ads"] });
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    },
  });
}

export function useUpdateAd(id: string) {
  const queryClient = useQueryClient();
  return useMutation<ApiResponse<Ad>, AxiosError<ApiErrorResponse>, UpdateAdPayload>({
    mutationFn: async (payload) => {
      const response = await api.patch<ApiResponse<Ad>>(`${API_ADS}/${id}`, payload);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ads"] });
      queryClient.invalidateQueries({ queryKey: ["ad", id] });
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    },
  });
}

export function useDeleteAd() {
  const queryClient = useQueryClient();
  return useMutation<void, AxiosError<ApiErrorResponse>, string>({
    mutationFn: async (id: string) => {
      await api.delete(`${API_ADS}/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ads"] });
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    },
  });
}

export function useTopUpAd(id: string) {
  const queryClient = useQueryClient();
  return useMutation<ApiResponse<Ad>, AxiosError<ApiErrorResponse>, { amount: number }>({
    mutationFn: async (payload) => {
      const response = await api.post<ApiResponse<Ad>>(`${API_ADS}/${id}/topup`, payload);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ads"] });
      queryClient.invalidateQueries({ queryKey: ["ad", id] });
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    },
  });
}

export interface AdType {
  ad_type: string;
  cpm: string;
  label: string;
  aspect_ratio: string;
  recommended_resolution: string;
}

export function useAdTypes() {
  return useQuery({
    queryKey: ["adTypes"],
    queryFn: async () => {
      const response = await api.get<ApiResponse<AdType[]>>("/api/ads/types");
      return response.data.data;
    },
  });
}
