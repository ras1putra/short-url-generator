package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"urlshortener/pkg/constants"
)

type Config struct {
	Port               string  `mapstructure:"PORT"`
	Env                string  `mapstructure:"ENV"`
	DBHost             string  `mapstructure:"DB_HOST"`
	DBPort             string  `mapstructure:"DB_PORT"`
	DBName             string  `mapstructure:"DB_NAME"`
	DBUser             string  `mapstructure:"DB_USER"`
	DBPassword         string  `mapstructure:"DB_PASSWORD"`
	DBURL              string  `mapstructure:"DB_URL"`
	RedisAddr          string  `mapstructure:"REDIS_ADDR"`
	RedisPassword      string  `mapstructure:"REDIS_PASSWORD"`
	BaseURL            string  `mapstructure:"BASE_URL"`
	JWTAccessSecret    string  `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret   string  `mapstructure:"JWT_REFRESH_SECRET"`
	RateLimitRedirect  int     `mapstructure:"RATE_LIMIT_REDIRECT"`
	RateLimitCreate    int     `mapstructure:"RATE_LIMIT_CREATE"`
	GeoIPDBPath        string  `mapstructure:"GEOIP_DB_PATH"`
	AllowedOrigins     string  `mapstructure:"ALLOWED_ORIGINS"`
	NodeRPCURL         string  `mapstructure:"NODE_RPC_URL"`
	ChainRPCURL        string  `mapstructure:"CHAIN_RPC_URL"`
	ContractPayment    string  `mapstructure:"CONTRACT_PAYMENT"`
	ContractToken      string  `mapstructure:"CONTRACT_TOKEN"`
	TokenSymbol        string  `mapstructure:"TOKEN_SYMBOL"`
	TokenDecimals      int     `mapstructure:"TOKEN_DECIMALS"`
	OwnerAddress       string  `mapstructure:"OWNER_ADDRESS"`
	ContractFaucet     string  `mapstructure:"CONTRACT_FAUCET"`
	ContractWithdrawer string  `mapstructure:"CONTRACT_WITHDRAWER"`
	FaucetSignerKey    string  `mapstructure:"FAUCET_SIGNER_KEY"`
	ChainID            int     `mapstructure:"CHAIN_ID"`
	ChainName          string  `mapstructure:"CHAIN_NAME"`
	ExplorerURL        string  `mapstructure:"EXPLORER_URL"`
	S3Endpoint         string  `mapstructure:"S3_ENDPOINT"`
	S3AccessKey        string  `mapstructure:"S3_ACCESS_KEY"`
	S3SecretKey        string  `mapstructure:"S3_SECRET_KEY"`
	S3Bucket           string  `mapstructure:"S3_BUCKET"`
	S3PublicURL        string  `mapstructure:"S3_PUBLIC_URL"`
	S3Region           string  `mapstructure:"S3_REGION"`
	BridgeHMACSecret   string  `mapstructure:"BRIDGE_HMAC_SECRET"`
	QualityMinScore    float64 `mapstructure:"QUALITY_MIN_SCORE"`
	MinSessionMs       int64   `mapstructure:"MIN_SESSION_MS"`
}

func (c *Config) IsDev() bool {
	return c.Env == constants.EnvDevelopment
}

// Load reads the environment variables into the Config struct.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("ENV", constants.EnvDevelopment)

	bindEnvs(Config{})

	if err := viper.ReadInConfig(); err != nil {
		zap.L().Warn("No .env file found or error reading it, using environment variables only", zap.Error(err))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		zap.L().Error("Failed to unmarshal config", zap.Error(err))
		return nil, err
	}

	if cfg.DBURL == "" {
		cfg.DBURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	}

	if cfg.IsDev() {
		cfg.ChainID = 31337
		cfg.ChainName = "Hardhat"
		cfg.ChainRPCURL = "http://127.0.0.1:8545"
		cfg.ExplorerURL = "http://localhost:5100"
		cfg.FaucetSignerKey = "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"
		cfg.ContractToken = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"
		cfg.ContractPayment = "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9"
		cfg.ContractFaucet = "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9"
		cfg.ContractWithdrawer = "0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0"
		cfg.S3Endpoint = "http://garage:3900"
		cfg.S3Bucket = "short-url-ads"
		cfg.S3PublicURL = "http://short-url-ads.web.localhost:3902"
		cfg.S3Region = "garage"
	}

	return &cfg, nil
}

func bindEnvs(iface interface{}) {
	t := reflect.TypeOf(iface)

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("mapstructure")
		if tag != "" {
			viper.BindEnv(tag)
		}
	}
}
