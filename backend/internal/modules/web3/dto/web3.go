package dto

type FaucetClaimRequest struct {
	WalletAddr string `json:"wallet_addr" validate:"required,eth_addr"`
}

type FaucetClaimResponse struct {
	Wallet     string `json:"wallet"`
	Amount     string `json:"amount"`
	Nonce      string `json:"nonce"`
	Deadline   string `json:"deadline"`
	Signature  string `json:"signature"`
	FaucetAddr string `json:"faucet_addr"`
	ChainID    int64  `json:"chain_id"`
}

type FaucetConfirmRequest struct {
	TxHash     string `json:"tx_hash" validate:"required,eth_tx_hash"`
	WalletAddr string `json:"wallet_addr" validate:"required,eth_addr"`
}

type FaucetConfirmResponse struct {
	Status string `json:"status"`
	TxHash string `json:"tx_hash"`
}

type FaucetHistoryItem struct {
	ID        string `json:"id"`
	Amount    string `json:"amount"`
	TxHash    string `json:"tx_hash"`
	ClaimedAt string `json:"claimed_at"`
}

type FaucetHistoryResponse struct {
	Claims     []FaucetHistoryItem `json:"claims"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
	TotalPages int                 `json:"total_pages"`
}

type DevETHResponse struct {
	TxHash string `json:"tx_hash"`
}

type DepositStatusResponse struct {
	Contract  string `json:"contract"`
	Status    string `json:"status"`
	LastBlock uint64 `json:"last_block"`
}
