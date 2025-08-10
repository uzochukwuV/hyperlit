package engine

import (
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/models"
	"time"

	"github.com/rs/zerolog/log"
)

type RiskManager struct {
	config *config.Config
}

type RiskAssessment struct {
	Approved     bool    `json:"approved"`
	Reason       string  `json:"reason"`
	AdjustedSize float64 `json:"adjusted_size"`
	RiskScore    float64 `json:"risk_score"`
}

func NewRiskManager(cfg *config.Config) *RiskManager {
	return &RiskManager{
		config: cfg,
	}
}

func (rm *RiskManager) AssessRisk(follower *models.Follower, trade *models.Trade) *RiskAssessment {
	assessment := &RiskAssessment{
		Approved:     true,
		AdjustedSize: trade.Size,
		RiskScore:    0.0,
	}

	// Check if follower is active
	if !follower.IsActive {
		assessment.Approved = false
		assessment.Reason = "Follower is not active"
		return assessment
	}

	// Check maximum position size
	proposedSize := trade.Size * (follower.CopyPercentage / 100.0)
	if proposedSize > follower.MaxPositionSize {
		assessment.AdjustedSize = follower.MaxPositionSize
		assessment.RiskScore += 0.3
		log.Warn().
			Int("follower_id", follower.ID).
			Float64("proposed", proposedSize).
			Float64("max", follower.MaxPositionSize).
			Msg("Position size reduced due to max limit")
	}

	// Check for recent trading activity (prevent overtrading)
	if rm.isOvertrading(follower) {
		assessment.Approved = false
		assessment.Reason = "Overtrading detected - too many trades in short period"
		return assessment
	}

	// Asset-specific risk checks
	assetRisk := rm.assessAssetRisk(trade.Asset)
	assessment.RiskScore += assetRisk

	// Time-based risk (avoid trading during high volatility periods)
	timeRisk := rm.assessTimeRisk()
	assessment.RiskScore += timeRisk

	// Position concentration risk
	concentrationRisk := rm.assessConcentrationRisk(follower, trade)
	assessment.RiskScore += concentrationRisk

	// Apply risk-based position sizing
	if assessment.RiskScore > 0.5 {
		riskAdjustment := 1.0 - (assessment.RiskScore * 0.5)
		assessment.AdjustedSize *= riskAdjustment
		
		log.Info().
			Int("follower_id", follower.ID).
			Float64("risk_score", assessment.RiskScore).
			Float64("adjustment", riskAdjustment).
			Float64("adjusted_size", assessment.AdjustedSize).
			Msg("Position size adjusted for risk")
	}

	// Final risk check
	if assessment.RiskScore > 1.0 {
		assessment.Approved = false
		assessment.Reason = "Risk score too high"
		return assessment
	}

	// Minimum position size check
	if assessment.AdjustedSize < 0.001 {
		assessment.Approved = false
		assessment.Reason = "Position size too small after risk adjustments"
		return assessment
	}

	return assessment
}

func (rm *RiskManager) isOvertrading(follower *models.Follower) bool {
	// Check for maximum trades per hour/day from risk settings
	if follower.RiskSettings != nil {
		if maxTradesPerHour, exists := follower.RiskSettings["max_trades_per_hour"]; exists {
			if maxTrades, ok := maxTradesPerHour.(float64); ok {
				// In a real implementation, query recent trades from database
				// For now, return false as we don't have trade history accessible here
				_ = maxTrades
			}
		}
	}
	return false
}

func (rm *RiskManager) assessAssetRisk(asset string) float64 {
	// Different assets have different risk profiles
	riskScores := map[string]float64{
		"BTC":  0.1,
		"ETH":  0.15,
		"SOL":  0.25,
		"AVAX": 0.3,
		"DOGE": 0.4,
		"PEPE": 0.6,
	}

	if score, exists := riskScores[asset]; exists {
		return score
	}

	// Unknown assets get higher risk score
	return 0.5
}

func (rm *RiskManager) assessTimeRisk() float64 {
	now := time.Now().UTC()
	hour := now.Hour()

	// Higher risk during certain hours (market opens, etc.)
	highRiskHours := map[int]float64{
		0:  0.3, // Midnight UTC
		8:  0.2, // European market open
		13: 0.25, // US market open
		21: 0.2, // Asian market open
	}

	if risk, exists := highRiskHours[hour]; exists {
		return risk
	}

	return 0.1 // Base time risk
}

func (rm *RiskManager) assessConcentrationRisk(follower *models.Follower, trade *models.Trade) float64 {
	// In a real implementation, this would check existing positions
	// to prevent over-concentration in a single asset
	
	// For now, return base concentration risk
	return 0.1
}

func (rm *RiskManager) ValidateFollowerSettings(follower *models.Follower) []string {
	var errors []string

	// Validate copy percentage
	if follower.CopyPercentage <= 0 || follower.CopyPercentage > 100 {
		errors = append(errors, "Copy percentage must be between 0 and 100")
	}

	// Validate maximum position size
	if follower.MaxPositionSize <= 0 {
		errors = append(errors, "Maximum position size must be positive")
	}

	if follower.MaxPositionSize > rm.config.MaxPositionSize {
		errors = append(errors, "Maximum position size exceeds global limit")
	}

	// Validate stop loss if set
	if follower.StopLossPercentage != nil {
		if *follower.StopLossPercentage <= 0 || *follower.StopLossPercentage >= 100 {
			errors = append(errors, "Stop loss percentage must be between 0 and 100")
		}
	}

	// Validate take profit if set
	if follower.TakeProfitPercentage != nil {
		if *follower.TakeProfitPercentage <= 0 {
			errors = append(errors, "Take profit percentage must be positive")
		}
	}

	// Validate API wallet address
	if follower.APIWalletAddress == "" {
		errors = append(errors, "API wallet address is required")
	}

	// Validate leader address
	if follower.LeaderAddress == "" {
		errors = append(errors, "Leader address is required")
	}

	return errors
}

func (rm *RiskManager) ShouldTriggerStopLoss(follower *models.Follower, currentPnL float64, entryValue float64) bool {
	if follower.StopLossPercentage == nil {
		return false
	}

	pnlPercentage := (currentPnL / entryValue) * 100
	
	return pnlPercentage <= -*follower.StopLossPercentage
}

func (rm *RiskManager) ShouldTriggerTakeProfit(follower *models.Follower, currentPnL float64, entryValue float64) bool {
	if follower.TakeProfitPercentage == nil {
		return false
	}

	pnlPercentage := (currentPnL / entryValue) * 100
	
	return pnlPercentage >= *follower.TakeProfitPercentage
}

func (rm *RiskManager) CalculateMaxDrawdown(trades []models.Trade) float64 {
	if len(trades) == 0 {
		return 0
	}

	var runningPnL float64
	var maxPnL float64
	var maxDrawdown float64

	for _, trade := range trades {
		tradePnL := rm.calculateTradePnL(trade)
		runningPnL += tradePnL
		
		if runningPnL > maxPnL {
			maxPnL = runningPnL
		}
		
		drawdown := maxPnL - runningPnL
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

func (rm *RiskManager) calculateTradePnL(trade models.Trade) float64 {
	// Simplified PnL calculation
	if trade.Side == "buy" {
		return -trade.Size * trade.Price // Cost (negative)
	} else {
		return trade.Size * trade.Price // Revenue (positive)
	}
}

func (rm *RiskManager) GetRiskMetrics(follower *models.Follower, trades []models.Trade) map[string]interface{} {
	totalTrades := len(trades)
	if totalTrades == 0 {
		return map[string]interface{}{
			"total_trades":     0,
			"win_rate":        0.0,
			"max_drawdown":    0.0,
			"risk_score":      0.0,
		}
	}

	profitableTrades := 0
	var totalPnL float64

	for _, trade := range trades {
		tradePnL := rm.calculateTradePnL(trade)
		totalPnL += tradePnL
		if tradePnL > 0 {
			profitableTrades++
		}
	}

	winRate := float64(profitableTrades) / float64(totalTrades)
	maxDrawdown := rm.CalculateMaxDrawdown(trades)
	
	// Calculate overall risk score
	riskScore := 0.0
	if winRate < 0.3 {
		riskScore += 0.3
	}
	if maxDrawdown > 1000 { // Assuming USD
		riskScore += 0.4
	}
	if totalTrades > 100 { // High frequency trading
		riskScore += 0.2
	}

	return map[string]interface{}{
		"total_trades":     totalTrades,
		"profitable_trades": profitableTrades,
		"win_rate":        winRate,
		"total_pnl":       totalPnL,
		"max_drawdown":    maxDrawdown,
		"risk_score":      riskScore,
	}
}
