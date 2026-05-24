import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { ApiResponse } from "@/types/api";
import { LinkStats } from "@/types/link";
import { formatStatsSelect } from "@/lib/stats";
import { API_LINKS } from "@/lib/constants";

export function useLinkStatsQuery(slug: string) {
  return useQuery({
    queryKey: ["link-stats", slug],
    queryFn: async () => {
      const response = await api.get<ApiResponse<LinkStats>>(`${API_LINKS}/${slug}/stats`);
      return response.data.data;
    },
    select: formatStatsSelect,
  });
}
