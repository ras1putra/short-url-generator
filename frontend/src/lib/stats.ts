import { formatBrowserData, formatDeviceData } from "./stats-utils";

export { formatBrowserData, formatDeviceData };

export function formatStatsSelect<T extends {
  browsers: Record<string, number>;
  devices: Record<string, number>;
}>(data: T) {
  return {
    ...data,
    formatted: {
      browsers: formatBrowserData(data.browsers),
      devices: formatDeviceData(data.devices),
    },
  };
}
