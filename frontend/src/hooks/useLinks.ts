import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { AxiosError } from "axios";
import { ApiResponse, ApiErrorResponse } from "@/types/api";
import { Link, ListResponse } from "@/types/link";
import { useState } from "react";

export function useLinksQuery(page: number = 1, perPage: number = 5) {
  return useQuery({
    queryKey: ["links", page, perPage],
    queryFn: async () => {
      const response = await api.get<ApiResponse<ListResponse>>("/api/links", {
        params: { page, per_page: perPage },
      });
      return response.data.data;
    },
  });
}

export function useCreateLink() {
  const queryClient = useQueryClient();
  return useMutation<ApiResponse<Link>, AxiosError<ApiErrorResponse>, Record<string, string | number>>({
    mutationFn: async (payload) => {
      const response = await api.post<ApiResponse<Link>>("/api/links", payload);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] });
    },
  });
}

export function useLinkDetail(slug: string) {
  return useQuery({
    queryKey: ["link", slug],
    queryFn: async () => {
      const response = await api.get<ApiResponse<Link>>(`/api/links/${slug}`);
      return response.data.data;
    },
    enabled: !!slug,
  });
}

export function useUpdateLink(slug: string) {
  const queryClient = useQueryClient();
  return useMutation<ApiResponse<Link>, AxiosError<ApiErrorResponse>, { custom_slug?: string; expires_value?: number; expires_unit?: string }>({
    mutationFn: async (data) => {
      const response = await api.patch<ApiResponse<Link>>(`/api/links/${slug}`, data);
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
      await api.delete(`/api/links/${slug}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] });
    },
  });
}

export function useLinkActions() {
  const [copiedSlug, setCopiedSlug] = useState<string | null>(null);
  const deleteMutation = useDeleteLink();

  const copyToClipboard = (url: string, slug: string) => {
    navigator.clipboard.writeText(url);
    setCopiedSlug(slug);
    setTimeout(() => setCopiedSlug(null), 2000);
  };

  return {
    copiedSlug,
    copyToClipboard,
    deleteLink: deleteMutation.mutate,
    isDeleting: deleteMutation.isPending,
  };
}
