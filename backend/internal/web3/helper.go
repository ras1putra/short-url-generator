package web3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	eip712DomainTypeHash = crypto.Keccak256Hash([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))
	faucetClaimTypeHash  = crypto.Keccak256Hash([]byte("FaucetClaim(address wallet,uint256 amount,uint256 nonce,uint256 deadline)"))
)

func computeDomainSeparator(d EIP712Domain) [32]byte {
	buf := new(bytes.Buffer)
	buf.Write(eip712DomainTypeHash.Bytes())
	buf.Write(crypto.Keccak256Hash([]byte(d.Name)).Bytes())
	buf.Write(crypto.Keccak256Hash([]byte(d.Version)).Bytes())
	buf.Write(common.LeftPadBytes(d.ChainID.Bytes(), 32))
	buf.Write(common.LeftPadBytes(d.VerifyingContract.Bytes(), 32))

	return crypto.Keccak256Hash(buf.Bytes())
}

func encodeFaucetClaim(wallet common.Address, amount, nonce, deadline *big.Int) []byte {
	buf := new(bytes.Buffer)
	buf.Write(faucetClaimTypeHash.Bytes())
	buf.Write(common.LeftPadBytes(wallet.Bytes(), 32))
	buf.Write(common.LeftPadBytes(amount.Bytes(), 32))
	buf.Write(common.LeftPadBytes(nonce.Bytes(), 32))
	buf.Write(common.LeftPadBytes(deadline.Bytes(), 32))
	return buf.Bytes()
}

func rpcCall(ctx context.Context, rpcURL string, body interface{}, result interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("rpc marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("rpc request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("rpc do error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("rpc read error: %w", err)
	}

	var rpcResp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		return fmt.Errorf("rpc response parse error: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error (%d): %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if rpcResp.Result == nil || string(rpcResp.Result) == "null" {
		return nil
	}

	return json.Unmarshal(rpcResp.Result, result)
}

func decodeAddress(hex string) string {
	if len(hex) >= 42 {
		return "0x" + hex[len(hex)-40:]
	}
	return hex
}

func decodeBytes32(hex string) string {
	if len(hex) >= 66 {
		return "0x" + hex[2:66]
	}
	return hex
}

func decodeUint256(data string) *big.Int {
	if len(data) < 66 {
		return big.NewInt(0)
	}
	val := new(big.Int)
	val.SetString(data[2:66], 16)
	return val
}

func parseHexUint64(hex string) uint64 {
	if len(hex) < 3 {
		return 0
	}
	val := new(big.Int)
	val.SetString(hex[2:], 16)
	return val.Uint64()
}
