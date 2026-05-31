export interface Ad {
  id: string;
  advertiser_id: string;
  title: string;
  description?: string;
  image_url: string;
  target_url: string;
  category: string;
  total_budget: number;
  remaining_budget: number;
  cpm: number;
  status: string;
  ad_type: string;
  created_at: string;
  updated_at: string;
}

export interface AdStats {
  ad_id: string;
  impressions: number;
  clicks: number;
  completions: number;
  valid_completions: number;
  invalid_completions: number;
  skips: number;
  avg_quality_score: number;
}

export interface Wallet {
  balance: number;
  available?: number;
  frozen?: number;
  updated_at: string;
  transactions: Transaction[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface Transaction {
  id: string;
  user_id: string;
  amount: number;
  type: string;
  status: "PENDING" | "CONFIRMED" | "FAILED";
  tx_hash?: string;
  metadata?: unknown;
  created_at: string;
}

export interface CreateAdPayload {
  title: string;
  description?: string;
  image_url: string;
  target_url: string;
  category: string;
  total_budget: number;
  ad_type: string;
}

export interface CampaignListResponse {
  campaigns: Ad[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface UpdateAdPayload {
  title?: string;
  description?: string;
  image_url?: string;
  target_url?: string;
  category?: string;
  status?: string;
  total_budget?: number;
  ad_type?: string;
}
