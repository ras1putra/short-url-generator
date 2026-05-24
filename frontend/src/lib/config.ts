export interface ChainConfig {
  chain_id: number;
  chain_name: string;
  rpc_url: string;
  explorer_url?: string;
  currency: {
    name: string;
    symbol: string;
    decimals: number;
  };
}

export interface AppConfig {
  contract_payment: string;
  contract_token?: string;
  contract_faucet?: string;
  contract_withdrawer?: string;
  token_symbol: string;
  token_decimals: number;
  payment_chain: ChainConfig;
}
