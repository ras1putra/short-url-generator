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
}

export interface Wallet {
  balance: number;
  available?: number;
  frozen?: number;
  updated_at: string;
  transactions: Transaction[];
}

export interface Transaction {
  id: string;
  user_id: string;
  amount: number;
  type: string;
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

export interface PendingTransaction {
  tx_hash: string;
  amount: number;
  created_at: string;
  type: "DEPOSIT" | "WITHDRAWAL" | "APPROVE";
  confirmations: number;
  target_confirmations: number;
}
