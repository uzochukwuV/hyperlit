package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Corrected Hyperliquid URLs
	HyperliquidAPIURL       string
	HyperliquidTestnetURL   string
	HyperliquidWSURL        string
	HyperliquidTestnetWSURL string

	// Database and other services
	DatabaseURL        string
	RedisURL           string
	PythonAnalyticsURL string

	// Environment and logging
	Environment string
	LogLevel    string

	// Trading Configuration
	MaxFollowersPerLeader int
	DefaultRiskPercentage float64
	MaxOrderBatchSize     int
	OrderBatchInterval    time.Duration
	MaxPositionSize       float64

	// Hyperliquid-specific settings
	MaxWebSocketSubscriptions int
	ReconnectAttempts         int
	ReconnectDelay            time.Duration
	OrderTimeout              time.Duration

	// Rate Limiting (aligned with Hyperliquid limits)
	MaxAPIRequestsPerMinute int
	MaxWebSocketMessages    int

	// Security
	APIWalletPrivateKeys map[string]string
	SignatureChainID     int64 // 42161 for Arbitrum
}

func Load() *Config {
	return &Config{
		// Fixed URLs based on official documentation
		HyperliquidAPIURL:       getEnv("HYPERLIQUID_API_URL", "https://api.hyperliquid.xyz"),
		HyperliquidTestnetURL:   getEnv("HYPERLIQUID_TESTNET_URL", "https://api.hyperliquid-testnet.xyz"),
		HyperliquidWSURL:        getEnv("HYPERLIQUID_WS_URL", "wss://api.hyperliquid.xyz/ws"),
		HyperliquidTestnetWSURL: getEnv("HYPERLIQUID_TESTNET_WS_URL", "wss://api.hyperliquid-testnet.xyz/ws"),

		DatabaseURL:        getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/copytrading?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		PythonAnalyticsURL: getEnv("PYTHON_ANALYTICS_URL", "http://localhost:8001"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),

		MaxFollowersPerLeader: getEnvInt("MAX_FOLLOWERS_PER_LEADER", 100),
		DefaultRiskPercentage: getEnvFloat("DEFAULT_RISK_PERCENTAGE", 0.02),
		MaxOrderBatchSize:     getEnvInt("MAX_ORDER_BATCH_SIZE", 50),
		OrderBatchInterval:    time.Duration(getEnvInt("ORDER_BATCH_INTERVAL_MS", 100)) * time.Millisecond,
		MaxPositionSize:       getEnvFloat("MAX_POSITION_SIZE", 100000.0),

		// Hyperliquid API limits
		MaxWebSocketSubscriptions: getEnvInt("MAX_WEBSOCKET_SUBSCRIPTIONS", 1000), // Hyperliquid limit
		ReconnectAttempts:         getEnvInt("RECONNECT_ATTEMPTS", 5),
		ReconnectDelay:            time.Duration(getEnvInt("RECONNECT_DELAY_MS", 5000)) * time.Millisecond,
		OrderTimeout:              time.Duration(getEnvInt("ORDER_TIMEOUT_MS", 10000)) * time.Millisecond,

		// Rate limits aligned with Hyperliquid
		MaxAPIRequestsPerMinute: getEnvInt("MAX_API_REQUESTS_PER_MINUTE", 1200), // Hyperliquid limit
		MaxWebSocketMessages:    getEnvInt("MAX_WEBSOCKET_MESSAGES", 2000),      // Hyperliquid limit

		APIWalletPrivateKeys: map[string]string{
			"default": getEnv("API_WALLET_PRIVATE_KEY", ""),
		},
		SignatureChainID: getEnvInt64("SIGNATURE_CHAIN_ID", 42161), // Arbitrum
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
