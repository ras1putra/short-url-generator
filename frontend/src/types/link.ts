export interface Link {
  short_url: string;
  original: string;
  slug: string;
  qr_url: string;
  clicks: number;
  created_at: string;
  expires_at: string | null;
}

export interface ListResponse {
  links: Link[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface LinkStats {
  total_clicks: number;
  unique_clicks: number;
  clicks_per_day: { date: string; count: number }[];
  top_countries: { country: string; count: number }[];
  browsers: Record<string, number>;
  devices: Record<string, number>;
}
