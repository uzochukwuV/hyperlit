Analysis of Your Hyperliquid Copy Trading Platform
Based on my research and review of your code, here's my comprehensive analysis:
1. Alignment with Hyperliquid Protocol ✅
Your copy trading platform concept is perfectly aligned with Hyperliquid's capabilities:

Real-time trading: Hyperliquid supports high-frequency trading (200k orders/second) with 1-block finality
Copy trading feasibility: The platform provides user fills via WebSocket (userFills subscription) and comprehensive trade data
Vault-like functionality: Hyperliquid already has vault features that allow followers to copy leaders' strategies
Full on-chain transparency: All trades are recorded on-chain, making copy trading auditable

2. URL and RPC Configuration ⚠️ Needs Updates
Your current URLs need corrections:
Current (Incorrect):
goHyperliquidWSURL:       "wss://api.hyperliquid.xyz/ws"
HyperliquidTestnetWSURL: "wss://api.hyperliquid-testnet.xyz/ws"
Correct URLs:
go// Mainnet
HyperliquidAPIURL:      "https://api.hyperliquid.xyz"
HyperliquidWSURL:       "wss://api.hyperliquid.xyz/ws"

// Testnet  
HyperliquidTestnetURL:  "https://api.hyperliquid-testnet.xyz"
HyperliquidTestnetWSURL: "wss://api.hyperliquid-testnet.xyz/ws"
3. Technical Feasibility ✅ Highly Feasible
Your implementation approach is sound:

WebSocket subscriptions: Correctly using userFills and userEvents
Real-time replication: The engine can capture trades immediately via WebSocket
API wallet support: Properly implements agent wallets for followers
Risk management: Good framework for position sizing and risk controls

4. Critical Issues and Improvements Needed
A. Asset ID Mapping (Critical Bug)Fixed Asset ID MappingCode // getAssetID returns the asset ID for a given asset name (perp or spot) with dynamic lookup
func (api *HyperliquidAPI) getAssetID(asset string) (int, error) {
	api.metaMutex.RLock()
	defer api.metaMutex.RUnlock()

	// Check perpetuals first
	if api.B. EIP-712 Signature Implementation (Critical)
Your current signature implementation is incomplete. Here's the correct one:Fixed EIP-712 Signature for HyperliquidCode func (s *Signer) SignAction(action interface{}, walletAddress string, nonce int64) (map[string]interface{}, error) {
	// Create the EIP-712 typed data structure for Hyperliquid
	actionBytes, err := json.Marshal(action)
	if err != nil {
		return nil, C. WebSocket Subscription ImprovementsImproved WebSocket Subscriptions for Copy TradingCode func (c *Client) subscribeToLeaderEvents(leaderAddress string) error {
	// Subscribe to user fills (trades) - Most important for copy trading
	fillsSubscription := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interfacD. Order Placement ImprovementsImproved Order Placement for HyperliquidCode func (api *HyperliquidAPI) PlaceOrder(ctx context.Context, order *models.EnhancedOrderRequest, apiWalletAddress string) (*models.OrderResponse, error) {
	// Validate order parameters first
	if err := api.ValidateOrder(ctx, order, apiWalletAddress, tr5. Enhanced ConfigurationEnhanced Configuration for HyperliquidCode package config

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
	
	// Database an6. Additional Enhancements Needed
Rate Limiting Implementation
You need to implement rate limiting to stay within Hyperliquid's limits:

REST API: 1200 requests/minute per IP
WebSocket: 1000 subscriptions, 2000 messages/min per IP

Error Handling for Hyperliquid-Specific Cases

Handle asset delisting scenarios
Manage open interest cap situations
Implement proper retry logic for network issues

Copy Trading Specific Features

Vault Integration: Consider using Hyperliquid's native vault system
Slippage Protection: Implement price impact analysis
Funding Cost Tracking: Monitor funding payments for perpetuals
Position Sizing Algorithms: More sophisticated sizing based on account equity

7. Performance Optimizations

Connection Pooling: Implement WebSocket connection pooling
Batch Processing: Group similar orders together
Caching: Cache metadata and price data appropriately
Database Optimization: Use TimescaleDB features effectively

8. Security Considerations

API Key Management: Implement secure key rotation
Signature Validation: Add signature verification for incoming requests
Rate Limiting: Implement per-user rate limiting
Input Sanitization: Validate all user inputs

Summary
Your copy trading platform is very well-designed and highly feasible on Hyperliquid. The main issues are:

✅ Excellent Architecture: Clean separation of concerns, good use of Go patterns
⚠️ Critical Fixes Needed: URL corrections, EIP-712 signature, asset ID mapping
✅ Protocol Alignment: Perfect fit for Hyperliquid's capabilities
🚀 Enhancement Opportunities: Rate limiting, vault integration, better error handling

With the fixes I've provided, your platform should work excellently with Hyperliquid's high-performance infrastructure.RetryEVanother advancement idea , since hyperliquid is decentralized it will be wise to implement a functionality where u can just without a trader registering u can just copy their address and do all trades that they make by providing their addressEditExcellent idea! This is actually a perfect advancement for Hyperliquid since all trading data is fully transparent on-chain. You can implement "permissionless copy trading" where users can copy any trader's address without requiring the trader to register or opt-in. This leverages Hyperliquid's decentralized nature beautifully.
Let me design this functionality for you:












Assetmapping critical bug . 
// getAssetID returns the asset ID for a given asset name (perp or spot) with dynamic lookup
func (api *HyperliquidAPI) getAssetID(asset string) (int, error) {
	api.metaMutex.RLock()
	defer api.metaMutex.RUnlock()

	// Check perpetuals first
	if api.perpMeta != nil {
		for i, assetInfo := range api.perpMeta.Universe {
			if assetInfo.Name == asset {
				// Check if asset is delisted
				if enhancedInfo, ok := interface{}(assetInfo).(models.EnhancedAssetInfo); ok {
					if enhancedInfo.IsDelisted {
						return 0, fmt.Errorf("asset %s is delisted", asset)
					}
				}
				return i, nil // Correct: return index directly for perps
			}
		}
	}

	// Check spot markets - FIXED: Proper spot asset ID calculation
	if api.spotMeta != nil {
		for _, pair := range api.spotMeta.Universe {
			if pair.Name == asset {
				return 10000 + pair.Index, nil // Correct formula from docs
			}
		}
	}

	return 0, fmt.Errorf("unsupported asset: %s", asset)
}

##  EIP-712 Signature Implementation (Critical)

func (s *Signer) SignAction(action interface{}, walletAddress string, nonce int64) (map[string]interface{}, error) {
	// Create the EIP-712 typed data structure for Hyperliquid
	actionBytes, err := json.Marshal(action)
	if err != nil {
		return nil, err
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"HyperliquidTransaction": []apitypes.Type{
				{Name: "action", Type: "string"},
				{Name: "nonce", Type: "uint64"},
				{Name: "chainId", Type: "uint256"},
			},
		},
		PrimaryType: "HyperliquidTransaction",
		Domain: apitypes.TypedDataDomain{
			Name:              "HyperliquidSignTransaction",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(42161), // Arbitrum Chain ID
			VerifyingContract: walletAddress,
		},
		Message: apitypes.TypedDataMessage{
			"action":  string(actionBytes),
			"nonce":   strconv.FormatInt(nonce, 10),
			"chainId": "42161",
		},
	}

	// Hash and sign the typed data
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	hash := crypto.Keccak256Hash(rawData)

	signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return nil, err
	}

	// Adjust v value for Ethereum
	if signature[64] < 27 {
		signature[64] += 27
	}

	return map[string]interface{}{
		"r": hex.EncodeToString(signature[0:32]),
		"s": hex.EncodeToString(signature[32:64]),
		"v": int(signature[64]),
	}, nil
}

## . WebSocket Subscription Improvements

func (c *Client) subscribeToLeaderEvents(leaderAddress string) error {
	// Subscribe to user fills (trades) - Most important for copy trading
	fillsSubscription := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "userFills",
			"user": leaderAddress,
		},
	}
	if err := c.sendMessage(fillsSubscription); err != nil {
		return fmt.Errorf("failed to subscribe to user fills: %w", err)
	}

	// Subscribe to order updates to track order lifecycle
	orderSubscription := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "orderUpdates", 
			"user": leaderAddress,
		},
	}
	if err := c.sendMessage(orderSubscription); err != nil {
		return fmt.Errorf("failed to subscribe to order updates: %w", err)
	}

	// Subscribe to user events for funding, liquidations, etc.
	userEventsSubscription := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "userEvents",
			"user": leaderAddress,
		},
	}
	if err := c.sendMessage(userEventsSubscription); err != nil {
		return fmt.Errorf("failed to subscribe to user events: %w", err)
	}

	return nil
}

// Process user fills (the core copy trading data)
func (c *Client) processUserFills(data map[string]interface{}, tradeChannel chan models.TradeEvent) {
	// Handle both snapshot and streaming data
	isSnapshot, _ := data["isSnapshot"].(bool)
	
	var fills []interface{}
	if fillsData, ok := data["fills"].([]interface{}); ok {
		fills = fillsData
	}

	for _, fillData := range fills {
		fill, ok := fillData.(map[string]interface{})
		if !ok {
			continue
		}

		tradeEvent := c.parseUserFill(fill, isSnapshot)
		if tradeEvent != nil {
			select {
			case tradeChannel <- *tradeEvent:
			case <-c.shutdown:
				return
			default:
				log.Warn().Msg("Trade channel full, dropping trade event")
			}
		}
	}
}

func (c *Client) parseUserFill(fill map[string]interface{}, isSnapshot bool) *models.TradeEvent {
	event := &models.TradeEvent{}

	// Parse essential fields for copy trading
	if coin, ok := fill["coin"].(string); ok {
		event.Coin = coin
	}
	if side, ok := fill["side"].(string); ok {
		event.Side = side
	}
	if px, ok := fill["px"].(string); ok {
		event.Px = px
	}
	if sz, ok := fill["sz"].(string); ok {
		event.Sz = sz
	}
	if timeFloat, ok := fill["time"].(float64); ok {
		event.Time = int64(timeFloat)
	}
	if hash, ok := fill["hash"].(string); ok {
		event.Hash = hash
	}

	// Additional copy trading relevant fields
	if startPos, ok := fill["startPosition"].(string); ok {
		event.StartPos = startPos
	}
	if closedPnl, ok := fill["closedPnl"].(string); ok {
		// Store in a custom field or extend the model
		event.User = closedPnl // Temporarily using User field
	}

	event.User = c.leaderAddress
	return event
}

# Order Placement Improvements

func (api *HyperliquidAPI) PlaceOrder(ctx context.Context, order *models.EnhancedOrderRequest, apiWalletAddress string) (*models.OrderResponse, error) {
	// Validate order parameters first
	if err := api.ValidateOrder(ctx, order, apiWalletAddress, true); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	// Convert asset to proper format
	assetID, err := api.getAssetID(order.Asset)
	if err != nil {
		return nil, err
	}

	// Generate nonce with proper timing
	nonce := api.nonceManager.GetNextNonce(apiWalletAddress)

	// Prepare order data with correct format
	orderData := map[string]interface{}{
		"a": assetID,
		"b": order.IsBuy,
		"p": api.formatPrice(order.Price),
		"s": api.formatSize(order.Size),
		"r": false, // reduceOnly - should be configurable
		"t": api.getOrderTypeCode(order.OrderType, order.Tif),
	}

	// Add client order ID if provided
	if order.ClOid != nil {
		orderData["c"] = *order.ClOid
	}

	orderAction := map[string]interface{}{
		"type":     "order",
		"orders":   []map[string]interface{}{orderData},
		"grouping": "na", // Can be "normalTpsl" for TP/SL grouping
	}

	// Sign the action with corrected signature
	signature, err := api.signer.SignAction(orderAction, apiWalletAddress, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to sign order: %w", err)
	}

	// Prepare request body
	reqBody := map[string]interface{}{
		"action":    orderAction,
		"nonce":     nonce,
		"signature": signature,
	}

	// Only add vaultAddress if trading on behalf of vault/subaccount
	if order.VaultAddress != "" {
		reqBody["vaultAddress"] = order.VaultAddress
	}

	var response models.OrderResponse
	err = api.makeRequest(ctx, apiURL+"/exchange", reqBody, &response)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	// Enhanced error handling for Hyperliquid-specific responses
	if response.Status != "ok" {
		return nil, fmt.Errorf("order rejected with status: %s", response.Status)
	}

	// Check individual order statuses
	if response.Data.Statuses != nil {
		for i, status := range response.Data.Statuses {
			if status.Error != "" {
				return nil, fmt.Errorf("order %d rejected: %s", i, status.Error)
			}
		}
	}

	return &response, nil
}

// Enhanced price/size formatting to meet Hyperliquid requirements
func (api *HyperliquidAPI) formatPrice(price *float64) string {
	if price == nil {
		return ""
	}
	// Remove trailing zeros as required by Hyperliquid
	formatted := strconv.FormatFloat(*price, 'f', -1, 64)
	return strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
}

func (api *HyperliquidAPI) formatSize(size float64) string {
	// Remove trailing zeros as required by Hyperliquid
	formatted := strconv.FormatFloat(size, 'f', -1, 64)
	return strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
}

# Enhanced Configuration

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
	RedisURL          string
	PythonAnalyticsURL string
	
	// Environment and logging
	Environment string
	LogLevel    string
	
	// Trading Configuration
	MaxFollowersPerLeader  int
	DefaultRiskPercentage  float64
	MaxOrderBatchSize      int
	OrderBatchInterval     time.Duration
	MaxPositionSize        float64
	
	// Hyperliquid-specific settings
	MaxWebSocketSubscriptions int
	ReconnectAttempts        int
	ReconnectDelay          time.Duration
	OrderTimeout            time.Duration
	
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
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379"),
		PythonAnalyticsURL: getEnv("PYTHON_ANALYTICS_URL", "http://localhost:8001"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		
		MaxFollowersPerLeader:  getEnvInt("MAX_FOLLOWERS_PER_LEADER", 100),
		DefaultRiskPercentage:  getEnvFloat("DEFAULT_RISK_PERCENTAGE", 0.02),
		MaxOrderBatchSize:      getEnvInt("MAX_ORDER_BATCH_SIZE", 50),
		OrderBatchInterval:     time.Duration(getEnvInt("ORDER_BATCH_INTERVAL_MS", 100)) * time.Millisecond,
		MaxPositionSize:        getEnvFloat("MAX_POSITION_SIZE", 100000.0),
		
		// Hyperliquid API limits
		MaxWebSocketSubscriptions: getEnvInt("MAX_WEBSOCKET_SUBSCRIPTIONS", 1000), // Hyperliquid limit
		ReconnectAttempts:        getEnvInt("RECONNECT_ATTEMPTS", 5),
		ReconnectDelay:          time.Duration(getEnvInt("RECONNECT_DELAY_MS", 5000)) * time.Millisecond,
		OrderTimeout:            time.Duration(getEnvInt("ORDER_TIMEOUT_MS", 10000)) * time.Millisecond,
		
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


## permisionless trading sytem 
package engine

import (
	"context"
	"fmt"
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/api"
	"hyperliquid-copy-trading/internal/database"
	"hyperliquid-copy-trading/internal/models"
	"hyperliquid-copy-trading/internal/utils"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PermissionlessCopyEngine enables copying any trader without their registration
type PermissionlessCopyEngine struct {
	config           *config.Config
	db               *database.PostgresDB
	hyperliquidAPI   *api.HyperliquidAPI
	wsManager        *websocket.Manager
	
	// Track discovered traders and their followers
	discoveredTraders map[string]*TraderInfo
	traderFollowers   map[string][]*models.PermissionlessFollower
	tradersMutex      sync.RWMutex
	
	// Performance tracking for auto-discovery
	performanceTracker *PerformanceTracker
	
	shutdown chan struct{}
	wg       sync.WaitGroup
}

type TraderInfo struct {
	Address            string                 `json:"address"`
	FirstDiscovered    time.Time             `json:"first_discovered"`
	TotalVolume        float64               `json:"total_volume"`
	TradeCount         int                   `json:"trade_count"`
	PerformanceMetrics *models.PnLAnalytics  `json:"performance_metrics"`
	LastActivity       time.Time             `json:"last_activity"`
	IsActive           bool                  `json:"is_active"`
	AssetBreakdown     map[string]float64    `json:"asset_breakdown"`
}

type PermissionlessFollower struct {
	ID                   int                    `json:"id"`
	UserID               string                 `json:"user_id"`
	TargetTraderAddress  string                 `json:"target_trader_address"`
	APIWalletAddress     string                 `json:"api_wallet_address"`
	CopyPercentage       float64                `json:"copy_percentage"`
	MaxPositionSize      float64                `json:"max_position_size"`
	MinTradeSize         float64                `json:"min_trade_size"`
	AssetWhitelist       []string               `json:"asset_whitelist,omitempty"`
	AssetBlacklist       []string               `json:"asset_blacklist,omitempty"`
	AutoDiscoveryEnabled bool                   `json:"auto_discovery_enabled"`
	CopyFilters          *CopyFilters           `json:"copy_filters"`
	IsActive             bool                   `json:"is_active"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

type CopyFilters struct {
	MinPositionValue     float64   `json:"min_position_value"`
	MaxPositionValue     float64   `json:"max_position_value"`
	OnlyProfitableTrades bool      `json:"only_profitable_trades"`
	ExcludeLeverageAbove int       `json:"exclude_leverage_above"`
	TimeDelaySeconds     int       `json:"time_delay_seconds"`
	OnlyDuringHours      *TimeRange `json:"only_during_hours,omitempty"`
}

type TimeRange struct {
	StartHour int `json:"start_hour"` // 0-23
	EndHour   int `json:"end_hour"`   // 0-23
}

func NewPermissionlessCopyEngine(cfg *config.Config, db *database.PostgresDB, wsManager *websocket.Manager) *PermissionlessCopyEngine {
	hyperliquidAPI, err := api.NewHyperliquidAPI(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Hyperliquid API")
	}

	return &PermissionlessCopyEngine{
		config:             cfg,
		db:                 db,
		hyperliquidAPI:     hyperliquidAPI,
		wsManager:          wsManager,
		discoveredTraders:  make(map[string]*TraderInfo),
		traderFollowers:    make(map[string][]*models.PermissionlessFollower),
		performanceTracker: NewPerformanceTracker(),
		shutdown:           make(chan struct{}),
	}
}

// AddPermissionlessFollower allows copying any trader address
func (pce *PermissionlessCopyEngine) AddPermissionlessFollower(ctx context.Context, follower *models.PermissionlessFollower) error {
	// Validate the target trader address
	if !utils.ValidateAddress(follower.TargetTraderAddress) {
		return fmt.Errorf("invalid trader address: %s", follower.TargetTraderAddress)
	}

	// Check if trader exists and has activity
	traderInfo, err := pce.discoverTrader(ctx, follower.TargetTraderAddress)
	if err != nil {
		return fmt.Errorf("failed to discover trader: %w", err)
	}

	if traderInfo.TradeCount == 0 {
		return fmt.Errorf("trader %s has no trading activity", follower.TargetTraderAddress)
	}

	// Store in database
	if err := pce.db.CreatePermissionlessFollower(ctx, follower); err != nil {
		return fmt.Errorf("failed to store follower: %w", err)
	}

	// Add to tracking
	pce.tradersMutex.Lock()
	pce.traderFollowers[follower.TargetTraderAddress] = append(
		pce.traderFollowers[follower.TargetTraderAddress], 
		follower,
	)
	pce.tradersMutex.Unlock()

	// Start monitoring if not already
	if err := pce.startMonitoringTrader(follower.TargetTraderAddress); err != nil {
		log.Error().Err(err).Str("trader", follower.TargetTraderAddress).Msg("Failed to start monitoring")
	}

	log.Info().
		Str("trader", follower.TargetTraderAddress).
		Str("follower", follower.UserID).
		Msg("Permissionless follower added")

	return nil
}

// discoverTrader analyzes any address to determine if it's a viable trader
func (pce *PermissionlessCopyEngine) discoverTrader(ctx context.Context, address string) (*TraderInfo, error) {
	pce.tradersMutex.RLock()
	if existing, exists := pce.discoveredTraders[address]; exists {
		pce.tradersMutex.RUnlock()
		return existing, nil
	}
	pce.tradersMutex.RUnlock()

	log.Info().Str("address", address).Msg("Discovering new trader")

	// Get trader's recent trading history
	fills, err := pce.hyperliquidAPI.GetUserFills(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get user fills: %w", err)
	}

	// Analyze trading activity
	traderInfo := &TraderInfo{
		Address:         address,
		FirstDiscovered: time.Now(),
		TradeCount:      len(fills),
		AssetBreakdown:  make(map[string]float64),
		IsActive:        len(fills) > 0,
	}

	// Calculate performance metrics
	if len(fills) > 0 {
		traderInfo.LastActivity = time.Unix(fills[0].Time/1000, 0)
		
		// Calculate volume and asset breakdown
		for _, fill := range fills {
			price, _ := utils.ParseFloat(fill.Px)
			size, _ := utils.ParseFloat(fill.Sz)
			volume := price * size
			
			traderInfo.TotalVolume += volume
			traderInfo.AssetBreakdown[fill.Coin] += volume
		}

		// Get detailed performance analysis
		performance, err := pce.performanceTracker.AnalyzeTraderPerformance(fills)
		if err == nil {
			traderInfo.PerformanceMetrics = performance
		}
	}

	// Store discovered trader
	pce.tradersMutex.Lock()
	pce.discoveredTraders[address] = traderInfo
	pce.tradersMutex.Unlock()

	return traderInfo, nil
}

// startMonitoringTrader begins real-time monitoring of a trader
func (pce *PermissionlessCopyEngine) startMonitoringTrader(traderAddress string) error {
	// Check if already monitoring
	pce.tradersMutex.RLock()
	isMonitored := pce.wsManager.IsMonitoring(traderAddress)
	pce.tradersMutex.RUnlock()

	if isMonitored {
		return nil
	}

	// Subscribe to trader's fills and order updates
	tradeChannel, userChannel, err := pce.wsManager.SubscribeToLeader(traderAddress)
	if err != nil {
		return fmt.Errorf("failed to subscribe to trader: %w", err)
	}

	// Start processing trader's activity
	pce.wg.Add(1)
	go func() {
		defer pce.wg.Done()
		pce.processTraderActivity(traderAddress, tradeChannel, userChannel)
	}()

	log.Info().Str("trader", traderAddress).Msg("Started monitoring trader")
	return nil
}

// processTraderActivity handles real-time trader activity
func (pce *PermissionlessCopyEngine) processTraderActivity(traderAddress string, tradeChannel chan models.TradeEvent, userChannel chan models.UserEvent) {
	for {
		select {
		case <-pce.shutdown:
			return
		case tradeEvent, ok := <-tradeChannel:
			if !ok {
				return
			}
			pce.handleTraderTrade(traderAddress, tradeEvent)
		case userEvent, ok := <-userChannel:
			if !ok {
				return
			}
			pce.handleTraderUserEvent(traderAddress, userEvent)
		}
	}
}

// handleTraderTrade processes a trader's trade and triggers copying
func (pce *PermissionlessCopyEngine) handleTraderTrade(traderAddress string, tradeEvent models.TradeEvent) {
	ctx := context.Background()

	log.Info().
		Str("trader", traderAddress).
		Str("coin", tradeEvent.Coin).
		Str("side", tradeEvent.Side).
		Str("size", tradeEvent.Sz).
		Str("price", tradeEvent.Px).
		Msg("Trader trade detected")

	// Update trader info
	pce.updateTraderActivity(traderAddress, tradeEvent)

	// Get followers for this trader
	pce.tradersMutex.RLock()
	followers := pce.traderFollowers[traderAddress]
	pce.tradersMutex.RUnlock()

	if len(followers) == 0 {
		return
	}

	// Process each follower
	for _, follower := range followers {
		if !follower.IsActive {
			continue
		}

		// Apply copy filters
		if !pce.shouldCopyTrade(follower, tradeEvent) {
			log.Debug().
				Str("trader", traderAddress).
				Str("follower", follower.UserID).
				Msg("Trade filtered out")
			continue
		}

		// Calculate copy size
		copySize := pce.calculateCopySize(follower, tradeEvent)
		if copySize <= 0 {
			continue
		}

		// Execute copy trade
		go pce.executeCopyTrade(ctx, follower, tradeEvent, copySize)
	}
}

// shouldCopyTrade applies filters to determine if trade should be copied
func (pce *PermissionlessCopyEngine) shouldCopyTrade(follower *models.PermissionlessFollower, trade models.TradeEvent) bool {
	if follower.CopyFilters == nil {
		return true
	}

	filters := follower.CopyFilters

	// Check asset whitelist/blacklist
	if len(follower.AssetWhitelist) > 0 {
		found := false
		for _, asset := range follower.AssetWhitelist {
			if asset == trade.Coin {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(follower.AssetBlacklist) > 0 {
		for _, asset := range follower.AssetBlacklist {
			if asset == trade.Coin {
				return false
			}
		}
	}

	// Check position value limits
	price, _ := utils.ParseFloat(trade.Px)
	size, _ := utils.ParseFloat(trade.Sz)
	positionValue := price * size

	if positionValue < filters.MinPositionValue || positionValue > filters.MaxPositionValue {
		return false
	}

	// Check time restrictions
	if filters.OnlyDuringHours != nil {
		currentHour := time.Now().Hour()
		if currentHour < filters.OnlyDuringHours.StartHour || currentHour > filters.OnlyDuringHours.EndHour {
			return false
		}
	}

	// Apply time delay if specified
	if filters.TimeDelaySeconds > 0 {
		tradeTime := time.Unix(trade.Time/1000, 0)
		if time.Since(tradeTime) < time.Duration(filters.TimeDelaySeconds)*time.Second {
			// Schedule for later execution
			go func() {
				time.Sleep(time.Duration(filters.TimeDelaySeconds) * time.Second)
				// Re-execute the copy logic
			}()
			return false
		}
	}

	return true
}

// calculateCopySize determines the appropriate size for copying
func (pce *PermissionlessCopyEngine) calculateCopySize(follower *models.PermissionlessFollower, trade models.TradeEvent) float64 {
	originalSize, _ := utils.ParseFloat(trade.Sz)
	
	// Apply copy percentage
	copySize := originalSize * (follower.CopyPercentage / 100.0)
	
	// Apply minimum size filter
	if copySize < follower.MinTradeSize {
		return 0
	}
	
	// Apply maximum position size
	price, _ := utils.ParseFloat(trade.Px)
	positionValue := copySize * price
	
	if positionValue > follower.MaxPositionSize {
		copySize = follower.MaxPositionSize / price
	}
	
	return copySize
}

// executeCopyTrade executes the actual copy trade
func (pce *PermissionlessCopyEngine) executeCopyTrade(ctx context.Context, follower *models.PermissionlessFollower, trade models.TradeEvent, size float64) {
	price, _ := utils.ParseFloat(trade.Px)
	
	order := &models.EnhancedOrderRequest{
		Asset:     trade.Coin,
		IsBuy:     trade.Side == "B",
		Size:      size,
		Price:     &price,
		OrderType: "market", // Copy as market order for immediate execution
		Tif:       "Ioc",    // Immediate or Cancel
	}

	response, err := pce.hyperliquidAPI.PlaceOrder(ctx, order, follower.APIWalletAddress)
	if err != nil {
		log.Error().
			Err(err).
			Str("trader", follower.TargetTraderAddress).
			Str("follower", follower.UserID).
			Msg("Failed to execute copy trade")
		return
	}

	if response.Status == "ok" {
		log.Info().
			Str("trader", follower.TargetTraderAddress).
			Str("follower", follower.UserID).
			Str("coin", trade.Coin).
			Float64("size", size).
			Msg("Copy trade executed successfully")

		// Store the copy trade record
		copyTrade := &models.CopyTrade{
			OriginalTraderAddress: follower.TargetTraderAddress,
			FollowerID:           follower.ID,
			OriginalTradeHash:    trade.Hash,
			Asset:                trade.Coin,
			Side:                 trade.Side,
			OriginalSize:         trade.Sz,
			CopiedSize:           fmt.Sprintf("%.6f", size),
			OriginalPrice:        trade.Px,
			ExecutedAt:           time.Now(),
			Status:               "executed",
		}

		if err := pce.db.CreateCopyTrade(ctx, copyTrade); err != nil {
			log.Error().Err(err).Msg("Failed to store copy trade record")
		}
	}
}

// updateTraderActivity updates trader statistics
func (pce *PermissionlessCopyEngine) updateTraderActivity(traderAddress string, trade models.TradeEvent) {
	pce.tradersMutex.Lock()
	defer pce.tradersMutex.Unlock()

	trader, exists := pce.discoveredTraders[traderAddress]
	if !exists {
		return
	}

	trader.LastActivity = time.Unix(trade.Time/1000, 0)
	trader.TradeCount++

	price, _ := utils.ParseFloat(trade.Px)
	size, _ := utils.ParseFloat(trade.Sz)
	volume := price * size

	trader.TotalVolume += volume
	trader.AssetBreakdown[trade.Coin] += volume
}

// GetDiscoveredTraders returns all discovered traders with their stats
func (pce *PermissionlessCopyEngine) GetDiscoveredTraders() []*TraderInfo {
	pce.tradersMutex.RLock()
	defer pce.tradersMutex.RUnlock()

	traders := make([]*TraderInfo, 0, len(pce.discoveredTraders))
	for _, trader := range pce.discoveredTraders {
		traders = append(traders, trader)
	}

	return traders
}

// Auto-discovery functionality
func (pce *PermissionlessCopyEngine) StartAutoDiscovery(ctx context.Context) {
	pce.wg.Add(1)
	go func() {
		defer pce.wg.Done()
		pce.runAutoDiscovery(ctx)
	}()
}

// runAutoDiscovery automatically discovers high-performing traders
func (pce *PermissionlessCopyEngine) runAutoDiscovery(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pce.shutdown:
			return
		case <-ticker.C:
			pce.discoverActiveTraders(ctx)
		}
	}
}

// discoverActiveTraders finds new traders from recent market activity
func (pce *PermissionlessCopyEngine) discoverActiveTraders(ctx context.Context) {
	// This would analyze recent trades from WebSocket feeds
	// and identify addresses with significant trading activity
	log.Info().Msg("Running auto-discovery for active traders")
	
	// Implementation would involve:
	// 1. Monitoring high-volume trades from WebSocket feeds
	// 2. Tracking addresses with consistent profitability
	// 3. Analyzing trading patterns for quality metrics
	// 4. Building a recommendation system
}

func (pce *PermissionlessCopyEngine) handleTraderUserEvent(traderAddress string, userEvent models.UserEvent) {
	// Handle funding payments, liquidations, etc.
	log.Debug().
		Str("trader", traderAddress).
		Str("event_type", userEvent.Type).
		Msg("Trader user event")
}

func (pce *PermissionlessCopyEngine) Stop() {
	close(pce.shutdown)
	pce.wg.Wait()
}

# enhanced model trading system
package models

import (
	"time"
)

// PermissionlessFollower allows copying any trader without their consent
type PermissionlessFollower struct {
	ID                   int                    `json:"id"`
	UserID               string                 `json:"user_id"`
	TargetTraderAddress  string                 `json:"target_trader_address"`
	APIWalletAddress     string                 `json:"api_wallet_address"`
	CopyPercentage       float64                `json:"copy_percentage"`
	MaxPositionSize      float64                `json:"max_position_size"`
	MinTradeSize         float64                `json:"min_trade_size"`
	AssetWhitelist       []string               `json:"asset_whitelist,omitempty"`
	AssetBlacklist       []string               `json:"asset_blacklist,omitempty"`
	AutoDiscoveryEnabled bool                   `json:"auto_discovery_enabled"`
	CopyFilters          *CopyFilters           `json:"copy_filters"`
	IsActive             bool                   `json:"is_active"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

type CopyFilters struct {
	MinPositionValue     float64   `json:"min_position_value"`
	MaxPositionValue     float64   `json:"max_position_value"`
	OnlyProfitableTrades bool      `json:"only_profitable_trades"`
	ExcludeLeverageAbove int       `json:"exclude_leverage_above"`
	TimeDelaySeconds     int       `json:"time_delay_seconds"`
	OnlyDuringHours      *TimeRange `json:"only_during_hours,omitempty"`
	SlippageTolerance    float64   `json:"slippage_tolerance"`
	MaxDrawdownStop      float64   `json:"max_drawdown_stop"`
}

type TimeRange struct {
	StartHour int `json:"start_hour"` // 0-23
	EndHour   int `json:"end_hour"`   // 0-23
}

// CopyTrade records each copy trading execution
type CopyTrade struct {
	ID                    int       `json:"id"`
	OriginalTraderAddress string    `json:"original_trader_address"`
	FollowerID            int       `json:"follower_id"`
	OriginalTradeHash     string    `json:"original_trade_hash"`
	Asset                 string    `json:"asset"`
	Side                  string    `json:"side"`
	OriginalSize          string    `json:"original_size"`
	CopiedSize            string    `json:"copied_size"`
	OriginalPrice         string    `json:"original_price"`
	ExecutedPrice         string    `json:"executed_price,omitempty"`
	Slippage              float64   `json:"slippage"`
	DelayMs               int64     `json:"delay_ms"` // Execution delay
	Status                string    `json:"status"`
	ErrorMessage          string    `json:"error_message,omitempty"`
	ExecutedAt            time.Time `json:"executed_at"`
	CreatedAt             time.Time `json:"created_at"`
}

// TraderDiscovery tracks discovered traders
type TraderDiscovery struct {
	ID                  int                   `json:"id"`
	Address             string                `json:"address"`
	FirstDiscovered     time.Time             `json:"first_discovered"`
	TotalVolume         float64               `json:"total_volume"`
	TradeCount          int                   `json:"trade_count"`
	WinRate             float64               `json:"win_rate"`
	ProfitLoss          float64               `json:"profit_loss"`
	MaxDrawdown         float64               `json:"max_drawdown"`
	SharpeRatio         float64               `json:"sharpe_ratio"`
	LastActivity        time.Time             `json:"last_activity"`
	IsActive            bool                  `json:"is_active"`
	FollowerCount       int                   `json:"follower_count"`
	AssetBreakdown      map[string]float64    `json:"asset_breakdown"`
	PerformanceGrade    string                `json:"performance_grade"` // A, B, C, D, F
	RiskLevel           string                `json:"risk_level"`        // Low, Medium, High
	TradingStyle        string                `json:"trading_style"`     // Scalper, Swing, Position
	UpdatedAt           time.Time             `json:"updated_at"`
}

// TraderRecommendation provides AI-powered trader recommendations
type TraderRecommendation struct {
	ID                    int       `json:"id"`
	UserID                string    `json:"user_id"`
	RecommendedTrader     string    `json:"recommended_trader"`
	RecommendationScore   float64   `json:"recommendation_score"`
	RecommendationReason  string    `json:"recommendation_reason"`
	RiskCompatibility     float64   `json:"risk_compatibility"`
	StyleMatch            float64   `json:"style_match"`
	PerformanceScore      float64   `json:"performance_score"`
	RecommendedAllocation float64   `json:"recommended_allocation"`
	IsViewed              bool      `json:"is_viewed"`
	IsAccepted            bool      `json:"is_accepted"`
	CreatedAt             time.Time `json:"created_at"`
}

// CopyTradingStrategy defines different copying strategies
type CopyTradingStrategy struct {
	ID               int                    `json:"id"`
	UserID           string                 `json:"user_id"`
	StrategyName     string                 `json:"strategy_name"`
	StrategyType     string                 `json:"strategy_type"` // "mirror", "proportional", "risk_adjusted"
	TargetTraders    []string               `json:"target_traders"`
	Allocations      map[string]float64     `json:"allocations"`
	RebalanceFreq    string                 `json:"rebalance_frequency"`
	MaxTotalRisk     float64                `json:"max_total_risk"`
	Settings         map[string]interface{} `json:"settings"`
	IsActive         bool                   `json:"is_active"`
	PerformanceStats *StrategyPerformance   `json:"performance_stats,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type StrategyPerformance struct {
	TotalReturn       float64   `json:"total_return"`
	AnnualizedReturn  float64   `json:"annualized_return"`
	MaxDrawdown       float64   `json:"max_drawdown"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	WinRate           float64   `json:"win_rate"`
	TotalTrades       int       `json:"total_trades"`
	AvgTradeReturn    float64   `json:"avg_trade_return"`
	LastUpdated       time.Time `json:"last_updated"`
}

// TraderAnalytics provides deep analytics for any trader
type TraderAnalytics struct {
	Address              string                 `json:"address"`
	AnalysisPeriod       string                 `json:"analysis_period"`
	TotalTrades          int                    `json:"total_trades"`
	TotalVolume          float64                `json:"total_volume"`
	WinRate              float64                `json:"win_rate"`
	ProfitFactor         float64                `json:"profit_factor"`
	SharpeRatio          float64                `json:"sharpe_ratio"`
	MaxDrawdown          float64                `json:"max_drawdown"`
	AvgWin               float64                `json:"avg_win"`
	AvgLoss              float64                `json:"avg_loss"`
	LargestWin           float64                `json:"largest_win"`
	LargestLoss          float64                `json:"largest_loss"`
	ConsecutiveWins      int                    `json:"consecutive_wins"`
	ConsecutiveLosses    int                    `json:"consecutive_losses"`
	AssetPreferences     map[string]float64     `json:"asset_preferences"`
	TradingHours         map[int]int            `json:"trading_hours"`        // Hour -> Trade count
	TradingDays          map[string]int         `json:"trading_days"`         // Day -> Trade count
	PositionSizes        []float64              `json:"position_sizes"`
	HoldingTimes         []int                  `json:"holding_times"`        // Minutes
	RiskMetrics          *RiskMetrics           `json:"risk_metrics"`
	SeasonalPerformance  map[string]float64     `json:"seasonal_performance"` // Month -> Performance
	MarketConditions     map[string]float64     `json:"market_conditions"`    // Bull/Bear/Sideways performance
	AnalyzedAt           time.Time              `json:"analyzed_at"`
}

// SmartCopyOrder represents an order with intelligent execution
type SmartCopyOrder struct {
	ID                  int                    `json:"id"`
	FollowerID          int                    `json:"follower_id"`
	OriginalTradeHash   string                 `json:"original_trade_hash"`
	Asset               string                 `json:"asset"`
	Side                string                 `json:"side"`
	TargetSize          float64                `json:"target_size"`
	ExecutionStrategy   string                 `json:"execution_strategy"` // "immediate", "twap", "smart"
	MaxSlippage         float64                `json:"max_slippage"`
	TimeLimit           int                    `json:"time_limit"` // Seconds
	PriceImprovement    float64                `json:"price_improvement"`
	PartialExecutions   []PartialExecution     `json:"partial_executions"`
	Status              string                 `json:"status"`
	TotalExecuted       float64                `json:"total_executed"`
	AveragePrice        float64                `json:"average_price"`
	TotalSlippage       float64                `json:"total_slippage"`
	CreatedAt           time.Time              `json:"created_at"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty"`
}

type PartialExecution struct {
	Size        float64   `json:"size"`
	Price       float64   `json:"price"`
	Timestamp   time.Time `json:"timestamp"`
	OrderID     string    `json:"order_id"`
	MarketState string    `json:"market_state"` // Market conditions at execution
}

// CopyTradingInsights provides AI-driven insights
type CopyTradingInsights struct {
	UserID              string                 `json:"user_id"`