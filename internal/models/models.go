package models

import (
	"time"
)

type Leader struct {
	ID              int       `json:"id"`
	Address         string    `json:"address"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	IsActive        bool      `json:"is_active"`
	TotalFollowers  int       `json:"total_followers"`
	TotalVolume     float64   `json:"total_volume"`
	WinRate         float64   `json:"win_rate"`
	PnL30d          float64   `json:"pnl_30d"`
	MaxDrawdown     float64   `json:"max_drawdown"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Follower struct {
	ID                  int                    `json:"id"`
	UserID              string                 `json:"user_id"`
	LeaderAddress       string                 `json:"leader_address"`
	APIWalletAddress    string                 `json:"api_wallet_address"`
	CopyPercentage      float64                `json:"copy_percentage"`
	MaxPositionSize     float64                `json:"max_position_size"`
	StopLossPercentage  *float64               `json:"stop_loss_percentage,omitempty"`
	TakeProfitPercentage *float64              `json:"take_profit_percentage,omitempty"`
	IsActive            bool                   `json:"is_active"`
	RiskSettings        map[string]interface{} `json:"risk_settings"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type Trade struct {
	ID               int       `json:"id"`
	LeaderAddress    string    `json:"leader_address"`
	FollowerID       *int      `json:"follower_id,omitempty"`
	Asset            string    `json:"asset"`
	Side             string    `json:"side"` // "buy" or "sell"
	Size             float64   `json:"size"`
	Price            float64   `json:"price"`
	OrderType        string    `json:"order_type"`
	IsLeaderTrade    bool      `json:"is_leader_trade"`
	ExecutedAt       time.Time `json:"executed_at"`
	HyperliquidTxID  string    `json:"hyperliquid_tx_id"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

type Position struct {
	ID            int     `json:"id"`
	UserAddress   string  `json:"user_address"`
	Asset         string  `json:"asset"`
	Side          string  `json:"side"`
	Size          float64 `json:"size"`
	EntryPrice    float64 `json:"entry_price"`
	CurrentPrice  float64 `json:"current_price"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
	MarginUsed    float64 `json:"margin_used"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OrderRequest struct {
	Asset     string  `json:"asset"`
	IsBuy     bool    `json:"is_buy"`
	Size      float64 `json:"size"`
	Price     *float64 `json:"price,omitempty"`
	OrderType string  `json:"order_type"`
	Nonce     int64   `json:"nonce"`
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
	Coin      string  `json:"coin"`
	Side      string  `json:"side"`
	Px        string  `json:"px"`
	Sz        string  `json:"sz"`
	Hash      string  `json:"hash"`
	Time      int64   `json:"time"`
	StartPos  string  `json:"startPos"`
	Tid       int64   `json:"tid"`
	Fee       string  `json:"fee"`
	User      string  `json:"user"`
}

type HyperliquidAPIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type MetaInfo struct {
	Universe []AssetInfo `json:"universe"`
}

type AssetInfo struct {
	Name      string `json:"name"`
	SzDecimals int   `json:"szDecimals"`
	MaxLeverage int  `json:"maxLeverage"`
	OnlyIsolated bool `json:"onlyIsolated"`
}

type UserState struct {
	AssetPositions []AssetPosition `json:"assetPositions"`
	MarginSummary  MarginSummary   `json:"marginSummary"`
	CrossMaintenanceMarginUsed string `json:"crossMaintenanceMarginUsed"`
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
	TotalPnL        float64   `json:"total_pnl"`
	DailyPnL        []float64 `json:"daily_pnl"`
	WinRate         float64   `json:"win_rate"`
	MaxDrawdown     float64   `json:"max_drawdown"`
	SharpeRatio     float64   `json:"sharpe_ratio"`
	TotalTrades     int       `json:"total_trades"`
	ProfitableTrades int      `json:"profitable_trades"`
}
