package models

import (
	"time"
)

type Leader struct {
	ID             int       `json:"id"`
	Address        string    `json:"address"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	IsActive       bool      `json:"is_active"`
	TotalFollowers int       `json:"total_followers"`
	TotalVolume    float64   `json:"total_volume"`
	WinRate        float64   `json:"win_rate"`
	PnL30d         float64   `json:"pnl_30d"`
	MaxDrawdown    float64   `json:"max_drawdown"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Follower struct {
	ID                   int                    `json:"id"`
	UserID               string                 `json:"user_id"`
	LeaderAddress        string                 `json:"leader_address"`
	APIWalletAddress     string                 `json:"api_wallet_address"`
	CopyPercentage       float64                `json:"copy_percentage"`
	MaxPositionSize      float64                `json:"max_position_size"`
	StopLossPercentage   *float64               `json:"stop_loss_percentage,omitempty"`
	TakeProfitPercentage *float64               `json:"take_profit_percentage,omitempty"`
	IsActive             bool                   `json:"is_active"`
	RiskSettings         map[string]interface{} `json:"risk_settings"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

type Trade struct {
	ID              int       `json:"id"`
	LeaderAddress   string    `json:"leader_address"`
	FollowerID      *int      `json:"follower_id,omitempty"`
	Asset           string    `json:"asset"`
	Side            string    `json:"side"` // "buy" or "sell"
	Size            float64   `json:"size"`
	Price           float64   `json:"price"`
	OrderType       string    `json:"order_type"`
	IsLeaderTrade   bool      `json:"is_leader_trade"`
	ExecutedAt      time.Time `json:"executed_at"`
	HyperliquidTxID string    `json:"hyperliquid_tx_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type Position struct {
	ID            int       `json:"id"`
	UserAddress   string    `json:"user_address"`
	Asset         string    `json:"asset"`
	Side          string    `json:"side"`
	Size          float64   `json:"size"`
	EntryPrice    float64   `json:"entry_price"`
	CurrentPrice  float64   `json:"current_price"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	MarginUsed    float64   `json:"margin_used"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OrderRequest struct {
	Asset     string   `json:"asset"`
	IsBuy     bool     `json:"is_buy"`
	Size      float64  `json:"size"`
	Price     *float64 `json:"price,omitempty"`
	OrderType string   `json:"order_type"`
	Nonce     int64    `json:"nonce"`
}

type WebSocketMessage struct {
	Method       string      `json:"method"`
	Subscription interface{} `json:"subscription,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

type UserEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type TradeEvent struct {
	Coin     string `json:"coin"`
	Side     string `json:"side"`
	Px       string `json:"px"`
	Sz       string `json:"sz"`
	Hash     string `json:"hash"`
	Time     int64  `json:"time"`
	StartPos string `json:"startPos"`
	Tid      int64  `json:"tid"`
	Fee      string `json:"fee"`
	User     string `json:"user"`
}

type HyperliquidAPIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type MetaInfo struct {
	Universe []AssetInfo `json:"universe"`
}

type AssetInfo struct {
	Name         string `json:"name"`
	SzDecimals   int    `json:"szDecimals"`
	MaxLeverage  int    `json:"maxLeverage"`
	OnlyIsolated bool   `json:"onlyIsolated"`
}

type UserState struct {
	AssetPositions             []AssetPosition `json:"assetPositions"`
	MarginSummary              MarginSummary   `json:"marginSummary"`
	CrossMaintenanceMarginUsed string          `json:"crossMaintenanceMarginUsed"`
}

type AssetPosition struct {
	Position Position `json:"position"`
	Type     string   `json:"type"`
}

type MarginSummary struct {
	AccountValue string `json:"accountValue"`
	TotalNtlPos  string `json:"totalNtlPos"`
	TotalRawUsd  string `json:"totalRawUsd"`
}

type PnLAnalytics struct {
	TotalPnL         float64   `json:"total_pnl"`
	DailyPnL         []float64 `json:"daily_pnl"`
	WinRate          float64   `json:"win_rate"`
	MaxDrawdown      float64   `json:"max_drawdown"`
	SharpeRatio      float64   `json:"sharpe_ratio"`
	TotalTrades      int       `json:"total_trades"`
	ProfitableTrades int       `json:"profitable_trades"`
}

// Enhanced order request with TIF and client order ID support
type EnhancedOrderRequest struct {
	Asset     string   `json:"asset"`
	IsBuy     bool     `json:"is_buy"`
	Size      float64  `json:"size"`
	Price     *float64 `json:"price,omitempty"`
	OrderType string   `json:"order_type"`
	Tif       string   `json:"tif,omitempty"`       // Time in force: "Gtc", "Ioc", "Alo"
	ClOid     *string  `json:"clOid,omitempty"`    // Client order ID
}

// Order response models
type OrderResponse struct {
	Status string           `json:"status"`
	Data   OrderResponseData `json:"data"`
}

type OrderResponseData struct {
	Statuses []OrderStatus `json:"statuses"`
}

type OrderStatus struct {
	Error   string            `json:"error,omitempty"`
	Resting *OrderRestingInfo `json:"resting,omitempty"`
	Filled  *OrderFillInfo    `json:"filled,omitempty"`
}

type OrderRestingInfo struct {
	Oid int64 `json:"oid"`
}

type OrderFillInfo struct {
	TotalSz string `json:"totalSz"`
	AvgPx   string `json:"avgPx"`
}

// Enhanced trade event with more fields
type EnhancedTradeEvent struct {
	Coin      string `json:"coin"`
	Side      string `json:"side"`
	Px        string `json:"px"`
	Sz        string `json:"sz"`
	Hash      string `json:"hash"`
	Time      int64  `json:"time"`
	StartPos  string `json:"startPos"`
	Tid       int64  `json:"tid"`
	Fee       string `json:"fee"`
	User      string `json:"user"`
	ClosedPnl string `json:"closedPnl,omitempty"`
	Dir       string `json:"dir,omitempty"`
}

// Asset data for risk management
type ActiveAssetData struct {
	MaxTradeSzs []string `json:"maxTradeSzs"`
}

// User fee schedule
type UserFees struct {
	UserCrossRate string `json:"userCrossRate"`
	UserIsolatedRate string `json:"userIsolatedRate"`
}

// Portfolio data
type Portfolio struct {
	TotalNtlPos string `json:"totalNtlPos"`
	MarginUsed  string `json:"marginUsed"`
}

// Enhanced models for spot market support
type SpotMetaInfo struct {
	Tokens   []TokenInfo `json:"tokens"`
	Universe []SpotPair  `json:"universe"`
}

type TokenInfo struct {
	Name        string `json:"name"`
	SzDecimals  int    `json:"szDecimals"`
	WeiDecimals int    `json:"weiDecimals"`
	Index       int    `json:"index"`
	TokenID     string `json:"tokenId"`
	IsCanonical bool   `json:"isCanonical"`
}

type SpotPair struct {
	Name        string `json:"name"`
	Tokens      [2]int `json:"tokens"`
	Index       int    `json:"index"`
	IsCanonical bool   `json:"isCanonical"`
}

type SpotClearinghouseState struct {
	Balances []SpotBalance `json:"balances"`
}

type SpotBalance struct {
	Coin     string `json:"coin"`
	Token    int    `json:"token"`
	Hold     string `json:"hold"`
	Total    string `json:"total"`
	EntryNtl string `json:"entryNtl"`
}

type SpotMetaAndAssetCtxs struct {
	Meta     SpotMetaInfo       `json:"meta"`
	Contexts []SpotAssetContext `json:"contexts"`
}

type SpotAssetContext struct {
	DayNtlVlm string `json:"dayNtlVlm"`
	MarkPx    string `json:"markPx"`
	MidPx     string `json:"midPx"`
	PrevDayPx string `json:"prevDayPx"`
}

// Enhanced AssetInfo with delisted and open interest cap support
type EnhancedAssetInfo struct {
	Name         string `json:"name"`
	SzDecimals   int    `json:"szDecimals"`
	MaxLeverage  int    `json:"maxLeverage"`
	OnlyIsolated bool   `json:"onlyIsolated"`
	IsDelisted   bool   `json:"isDelisted,omitempty"`
}

// L2 Order Book models
type L2Book struct {
	Coin   string                  `json:"coin"`
	Time   int64                   `json:"time"`
	Levels map[string][]PriceLevel `json:"levels"` // "bids" and "asks"
}

type PriceLevel struct {
	Px string `json:"px"`
	Sz string `json:"sz"`
	N  int    `json:"n"`
}

type Leverage struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

// User funding for funding cost tracking
type UserFunding struct {
	Delta FundingDelta `json:"delta"`
	Hash  string       `json:"hash"`
	Time  int64        `json:"time"`
}

type FundingDelta struct {
	Coin        string `json:"coin"`
	FundingRate string `json:"fundingRate"`
	Szi         string `json:"szi"`
	Type        string `json:"type"`
	Usdc        string `json:"usdc"`
}

// Risk assessment models
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type RiskMetrics struct {
	Volatility       float64   `json:"volatility_pct"`
	MaxDrawdown      float64   `json:"max_drawdown_pct"`
	VaR95            float64   `json:"var_95_pct"`
	RiskLevel        RiskLevel `json:"risk_level"`
	RiskScore        float64   `json:"risk_score"`
	AvgTimeBetween   float64   `json:"avg_time_between_trades_minutes"`
	TradingIntensity float64   `json:"trading_intensity"`
}

type PerformanceMetrics struct {
	TotalReturn      float64 `json:"total_return_pct"`
	AnnualizedReturn float64 `json:"annualized_return_pct"`
	SharpeRatio      float64 `json:"sharpe_ratio"`
	WinRate          float64 `json:"win_rate"`
	AvgWin           float64 `json:"avg_win"`
	AvgLoss          float64 `json:"avg_loss"`
	ProfitFactor     float64 `json:"profit_factor"`
	TotalTrades      int     `json:"total_trades"`
}

type MarketMetrics struct {
	Beta        float64 `json:"beta"`
	Correlation float64 `json:"correlation"`
	Alpha       float64 `json:"alpha"`
}

// Leader performance analysis
type LeaderPerformanceAnalysis struct {
	LeaderAddress      string             `json:"leader_address"`
	AnalysisPeriodDays int                `json:"analysis_period_days"`
	PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
	RiskMetrics        RiskMetrics        `json:"risk_metrics"`
	MarketMetrics      MarketMetrics      `json:"market_metrics"`
	TradingFrequency   map[string]float64 `json:"trading_frequency"`
	AssetAllocation    map[string]float64 `json:"asset_allocation"`
	TimeSeriesData     []TimeSeriesPoint  `json:"time_series_data"`
	Predictions        interface{}        `json:"predictions,omitempty"`
	AnalysisTimestamp  time.Time          `json:"analysis_timestamp"`
}

type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Type      string    `json:"type"` // "pnl", "equity", "drawdown"
}

// Follower optimization
type FollowerOptimization struct {
	FollowerID            int                    `json:"follower_id"`
	CurrentSettings       map[string]interface{} `json:"current_settings"`
	OptimizedSettings     map[string]interface{} `json:"optimized_settings"`
	ExpectedImprovement   float64                `json:"expected_improvement_pct"`
	RecommendedLeaders    []string               `json:"recommended_leaders"`
	PortfolioAllocation   map[string]float64     `json:"portfolio_allocation"`
	RiskAssessment        RiskMetrics            `json:"risk_assessment"`
	OptimizationTimestamp time.Time              `json:"optimization_timestamp"`
}

// Trade recommendation
type TradeRecommendation struct {
	ID              string    `json:"id"`
	FollowerID      int       `json:"follower_id"`
	Asset           string    `json:"asset"`
	Side            string    `json:"side"`
	RecommendedSize float64   `json:"recommended_size"`
	ConfidenceScore float64   `json:"confidence_score"`
	Reasoning       string    `json:"reasoning"`
	ExpectedReturn  float64   `json:"expected_return_pct"`
	RiskLevel       RiskLevel `json:"risk_level"`
	ValidUntil      time.Time `json:"valid_until"`
	CreatedAt       time.Time `json:"created_at"`
}
