const COLORS = ["#6EE7B7", "#60A5FA", "#FBBF24", "#F472B6", "#A78BFA", "#34D399", "#F87171", "#FB923C"];

export interface FormattedEntry {
  name: string;
  value: number;
  fill: string;
}

export function formatBrowserData(browsers: Record<string, number>): FormattedEntry[] {
  return Object.entries(browsers).map(([name, value], i) => ({
    name,
    value,
    fill: COLORS[i % COLORS.length],
  }));
}

export function formatDeviceData(devices: Record<string, number>): FormattedEntry[] {
  return Object.entries(devices).map(([name, value], i) => ({
    name,
    value,
    fill: COLORS[i % COLORS.length],
  }));
}
