import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { ApiResponse } from "@/types/api";
import { LinkStats } from "@/types/link";

const COLORS = ['#6366f1', '#8b5cf6', '#ec4899', '#f43f5e', '#f97316'];

export function useLinkStatsQuery(slug: string) {
  return useQuery({
    queryKey: ["link-stats", slug],
    queryFn: async () => {
      const response = await api.get<ApiResponse<LinkStats>>(`/api/links/${slug}/stats`);
      return response.data.data;
    },
    select: (data) => {
      const browserData = Object.entries(data.browsers)
        .map(([name, value], index) => ({
          name,
          value,
          fill: COLORS[(index + 2) % COLORS.length]
        }))
        .filter((item) => item.value > 0);

      const deviceData = Object.entries(data.devices)
        .map(([name, value], index) => ({
          name,
          value,
          fill: COLORS[index % COLORS.length]
        }))
        .filter((item) => item.value > 0);

      return {
        ...data,
        formatted: {
          browsers: browserData,
          devices: deviceData
        }
      };
    }
  });
}
