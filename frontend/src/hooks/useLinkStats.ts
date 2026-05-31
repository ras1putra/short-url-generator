import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { ApiResponse } from "@/types/api";
import { LinkStats } from "@/types/link";
import { formatStatsSelect } from "@/lib/stats";
import { API_LINKS, DEFAULT_PAGE_SIZE } from "@/lib/constants";

export function useLinkStatsQuery(slug: string, eventPage?: number, eventPerPage?: number, eventSortBy?: string, eventSortDir?: string) {
  return useQuery({
    queryKey: ["link-stats", slug, eventPage ?? 1, eventPerPage ?? DEFAULT_PAGE_SIZE, eventSortBy ?? "time", eventSortDir ?? "desc"],
    queryFn: async () => {
      const params = new URLSearchParams();
      params.set("event_page", String(eventPage ?? 1));
      params.set("event_per_page", String(eventPerPage ?? DEFAULT_PAGE_SIZE));
      if (eventSortBy && eventSortBy !== "time") params.set("event_sort_by", eventSortBy);
      if (eventSortDir && eventSortDir !== "desc") params.set("event_sort_dir", eventSortDir);
      const response = await api.get<ApiResponse<LinkStats>>(`${API_LINKS}/${slug}/stats?${params.toString()}`);
      return response.data.data;
    },
    select: formatStatsSelect,
    placeholderData: keepPreviousData,
  });
}
