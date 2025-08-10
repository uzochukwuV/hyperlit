package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// RoundToDecimals rounds a float64 to the specified number of decimal places
func RoundToDecimals(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// FormatPrice formats a price for display
func FormatPrice(price float64) string {
	return strconv.FormatFloat(price, 'f', 4, 64)
}

// FormatSize formats a size for display
func FormatSize(size float64) string {
	return strconv.FormatFloat(size, 'f', 6, 64)
}

// ParseFloat safely parses a string to float64
func ParseFloat(str string) (float64, error) {
	str = strings.TrimSpace(str)
	if str == "" {
		return 0, nil
	}
	return strconv.ParseFloat(str, 64)
}

// CalculatePercentageChange calculates percentage change between two values
func CalculatePercentageChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		return 0
	}
	return ((newValue - oldValue) / oldValue) * 100
}

// CalculatePnL calculates profit/loss for a trade
func CalculatePnL(side string, entryPrice, exitPrice, size float64) float64 {
	if side == "buy" {
		return (exitPrice - entryPrice) * size
	} else {
		return (entryPrice - exitPrice) * size
	}
}

// CalculatePositionValue calculates the notional value of a position
func CalculatePositionValue(price, size float64) float64 {
	return price * size
}

// ValidateAddress checks if an Ethereum address is valid format
func ValidateAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	// Check if remaining characters are valid hex
	_, err := hex.DecodeString(address[2:])
	return err == nil
}

// GenerateNonce generates a unique nonce for Hyperliquid
func GenerateNonce() int64 {
	return time.Now().UnixMilli()
}

// GenerateRandomID generates a random ID
func GenerateRandomID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// TimeToUnixMilli converts time to unix milliseconds
func TimeToUnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

// UnixMilliToTime converts unix milliseconds to time
func UnixMilliToTime(unixMilli int64) time.Time {
	return time.Unix(unixMilli/1000, (unixMilli%1000)*1000000)
}

// CalculateWinRate calculates win rate from profitable and total trades
func CalculateWinRate(profitableTrades, totalTrades int) float64 {
	if totalTrades == 0 {
		return 0
	}
	return (float64(profitableTrades) / float64(totalTrades)) * 100
}

// CalculateSharpeRatio calculates Sharpe ratio from returns
func CalculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Calculate mean return
	var sum float64
	for _, ret := range returns {
		sum += ret
	}
	meanReturn := sum / float64(len(returns))

	// Calculate standard deviation
	var variance float64
	for _, ret := range returns {
		variance += math.Pow(ret-meanReturn, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))

	if stdDev == 0 {
		return 0
	}

	return (meanReturn - riskFreeRate) / stdDev
}

// CalculateMaxDrawdown calculates maximum drawdown from equity curve
func CalculateMaxDrawdown(equityCurve []float64) float64 {
	if len(equityCurve) == 0 {
		return 0
	}

	var maxDrawdown float64
	peak := equityCurve[0]

	for _, value := range equityCurve {
		if value > peak {
			peak = value
		}
		drawdown := (peak - value) / peak * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// TruncateString truncates a string to the specified length
func TruncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length-3] + "..."
}

// FormatPercentage formats a percentage for display
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}

// Min returns the minimum of two float64 values
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two float64 values
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Abs returns the absolute value of a float64
func Abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// FormatCurrency formats a value as currency
func FormatCurrency(value float64, symbol string) string {
	return fmt.Sprintf("%s%.2f", symbol, value)
}

// CalculateCompoundAnnualGrowthRate calculates CAGR
func CalculateCompoundAnnualGrowthRate(startValue, endValue float64, periods float64) float64 {
	if startValue == 0 || periods == 0 {
		return 0
	}
	return (math.Pow(endValue/startValue, 1/periods) - 1) * 100
}

// IsValidTimeframe checks if a timeframe string is valid
func IsValidTimeframe(timeframe string) bool {
	validTimeframes := []string{"1m", "5m", "15m", "1h", "4h", "1d", "1w", "1M"}
	for _, tf := range validTimeframes {
		if tf == timeframe {
			return true
		}
	}
	return false
}

// SafeDivide performs division with zero check
func SafeDivide(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// CalculateVolatility calculates volatility from price returns
func CalculateVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Calculate mean
	var sum float64
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	// Calculate variance
	var variance float64
	for _, ret := range returns {
		variance += math.Pow(ret-mean, 2)
	}
	variance /= float64(len(returns) - 1)

	return math.Sqrt(variance) * 100 // Convert to percentage
}

// GetAssetIDFromName converts asset name to ID for Hyperliquid
func GetAssetIDFromName(assetName string) (int, error) {
	assetMap := map[string]int{
		"BTC":  0,
		"ETH":  1,
		"SOL":  2,
		"AVAX": 3,
		"DOGE": 4,
		"ATOM": 5,
		"NEAR": 6,
		"FTM":  7,
		"GMX":  8,
		"ARB":  9,
	}

	if id, exists := assetMap[strings.ToUpper(assetName)]; exists {
		return id, nil
	}

	return 0, fmt.Errorf("unknown asset: %s", assetName)
}

// ValidateOrderSize checks if order size meets minimum requirements
func ValidateOrderSize(size float64, minSize float64) bool {
	return size >= minSize
}

// CalculateMarginRequired calculates margin required for a position
func CalculateMarginRequired(price, size, leverage float64) float64 {
	notionalValue := price * size
	return notionalValue / leverage
}

// FormatTradeSize formats trade size with appropriate precision
func FormatTradeSize(size float64, asset string) string {
	// Different assets have different precision requirements
	switch strings.ToUpper(asset) {
	case "BTC":
		return strconv.FormatFloat(size, 'f', 6, 64)
	case "ETH":
		return strconv.FormatFloat(size, 'f', 5, 64)
	default:
		return strconv.FormatFloat(size, 'f', 4, 64)
	}
}
