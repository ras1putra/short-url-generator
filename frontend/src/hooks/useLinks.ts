import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { AxiosError } from "axios";
import { ApiResponse, ApiErrorResponse } from "@/types/api";
import { Link, ListResponse } from "@/types/link";
import { API_LINKS } from "@/lib/constants";

export function useLinksQuery(
  page: number = 1,
  perPage: number = 5,
  search?: string,
  isMonetized?: boolean,
  sortBy?: string,
  sortDir?: string,
) {
  return useQuery({
    queryKey: ["links", page, perPage, search, isMonetized, sortBy, sortDir],
    queryFn: async () => {
      const params: Record<string, unknown> = { page, per_page: perPage };
      if (search) params.q = search;
      if (isMonetized !== undefined) params.is_monetized = isMonetized;
      if (sortBy && sortBy !== "created_at") params.sort_by = sortBy;
      if (sortDir && sortDir !== "desc") params.sort_dir = sortDir;
      const response = await api.get<ApiResponse<ListResponse>>(API_LINKS, { params });
      return response.data.data;
    },
    placeholderData: keepPreviousData,
  });
}

export function useCreateLink() {
  const queryClient = useQueryClient();
  return useMutation<ApiResponse<Link>, AxiosError<ApiErrorResponse>, Record<string, string | number>>({
    mutationFn: async (payload) => {
      const response = await api.post<ApiResponse<Link>>(API_LINKS, payload);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] });
      queryClient.invalidateQueries({ queryKey: ["aggregate-stats"] });
    },
  });
}

export function useLinkDetail(slug: string) {
  return useQuery({
    queryKey: ["link", slug],
    queryFn: async () => {
      const response = await api.get<ApiResponse<Link>>(`${API_LINKS}/${slug}`);
      return response.data.data;
    },
    enabled: !!slug,
  });
}

export function useUpdateLink(slug: string) {
  const queryClient = useQueryClient();
  return useMutation<ApiResponse<Link>, AxiosError<ApiErrorResponse>, Record<string, string | number | boolean | string[]>>({
    mutationFn: async (data) => {
      const response = await api.patch<ApiResponse<Link>>(`${API_LINKS}/${slug}`, data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] });
      queryClient.invalidateQueries({ queryKey: ["link", slug] });
    },
  });
}

export function useDeleteLink() {
  const queryClient = useQueryClient();
  return useMutation<void, AxiosError<ApiErrorResponse>, string>({
    mutationFn: async (slug: string) => {
      await api.delete(`${API_LINKS}/${slug}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] });
      queryClient.invalidateQueries({ queryKey: ["aggregate-stats"] });
    },
  });
}
