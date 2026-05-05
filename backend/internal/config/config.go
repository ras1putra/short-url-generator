package config

import (
	"fmt"
	"reflect"

	"go.uber.org/zap"

	"github.com/spf13/viper"
)

type Config struct {
	Port              string `mapstructure:"PORT"`
	Env               string `mapstructure:"ENV"`
	DBHost            string `mapstructure:"DB_HOST"`
	DBPort            string `mapstructure:"DB_PORT"`
	DBName            string `mapstructure:"DB_NAME"`
	DBUser            string `mapstructure:"DB_USER"`
	DBPassword        string `mapstructure:"DB_PASSWORD"`
	DBURL             string `mapstructure:"DB_URL"`
	DBMaxConns        int    `mapstructure:"DB_MAX_CONNS"`
	RedisAddr         string `mapstructure:"REDIS_ADDR"`
	RedisPassword     string `mapstructure:"REDIS_PASSWORD"`
	BaseURL           string `mapstructure:"BASE_URL"`
	JWTAccessSecret   string `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret  string `mapstructure:"JWT_REFRESH_SECRET"`
	RateLimitRedirect int    `mapstructure:"RATE_LIMIT_REDIRECT"`
	RateLimitCreate   int    `mapstructure:"RATE_LIMIT_CREATE"`
	GeoIPDBPath       string `mapstructure:"GEOIP_DB_PATH"`
	AllowedOrigins    string `mapstructure:"ALLOWED_ORIGINS"`
}

func (c *Config) IsDev() bool {
	return c.Env == "development" || c.Env == "dev"
}

// Load reads the environment variables into the Config struct.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("ENV", "development")

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

	return &cfg, nil
}

func bindEnvs(iface interface{}) {
	v := reflect.ValueOf(iface)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("mapstructure")
		if tag != "" {
			viper.BindEnv(tag)
		}
	}
}
