package web3

import (
	"context"
	"math/big"

	"github.com/redis/go-redis/v9"

	"urlshortener/internal/config"
	"urlshortener/pkg/constants"
)

type ETHClient struct {
	rpcURL  string
	redis   *redis.Client
	chainID *big.Int
}

func NewETHClient(cfg *config.Config, rdb *redis.Client) *ETHClient {
	return &ETHClient{
		rpcURL: cfg.NodeRPCURL,
		redis:  rdb,
	}
}

func (c *ETHClient) RPCURL() string {
	return c.rpcURL
}

func (c *ETHClient) Redis() *redis.Client {
	return c.redis
}

func (c *ETHClient) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	val, err := c.redis.Get(ctx, constants.RedisPrefixWeb3Block).Uint64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *ETHClient) SetLastProcessedBlock(ctx context.Context, block uint64) error {
	return c.redis.Set(ctx, constants.RedisPrefixWeb3Block, block, 0).Err()
}
