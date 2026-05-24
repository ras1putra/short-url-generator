package dto

type CurrencyResponse struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

type PaymentChainResponse struct {
	ChainID     int              `json:"chain_id"`
	ChainName   string           `json:"chain_name"`
	RPCURL      string           `json:"rpc_url"`
	ExplorerURL string           `json:"explorer_url"`
	Currency    CurrencyResponse `json:"currency"`
}

type ConfigResponse struct {
	ContractPayment    string               `json:"contract_payment"`
	ContractToken      string               `json:"contract_token"`
	ContractFaucet     string               `json:"contract_faucet"`
	ContractWithdrawer string               `json:"contract_withdrawer"`
	TokenSymbol        string               `json:"token_symbol"`
	TokenDecimals      int                  `json:"token_decimals"`
	PaymentChain       PaymentChainResponse `json:"payment_chain"`
}
