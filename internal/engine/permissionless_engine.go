package engine

import (
	"context"
	"fmt"
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/api"
	"hyperliquid-copy-trading/internal/database"
	"hyperliquid-copy-trading/internal/models"
	"hyperliquid-copy-trading/internal/utils"
	"hyperliquid-copy-trading/internal/websocket"
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

type PerformanceTracker struct {
	// Implementation for performance analysis
}

func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{}
}

func (pt *PerformanceTracker) AnalyzeTraderPerformance(fills []models.EnhancedTradeEvent) (*models.PnLAnalytics, error) {
	// Implement performance analysis logic
	analytics := &models.PnLAnalytics{
		TotalTrades: len(fills),
	}
	
	// Calculate basic metrics
	var totalPnL float64
	var profitableTrades int
	
	for _, fill := range fills {
		// Parse closed PnL if available
		if fill.ClosedPnl != "" {
			pnl, err := utils.ParseFloat(fill.ClosedPnl)
			if err == nil {
				totalPnL += pnl
				if pnl > 0 {
					profitableTrades++
				}
			}
		}
	}
	
	analytics.TotalPnL = totalPnL
	analytics.ProfitableTrades = profitableTrades
	if len(fills) > 0 {
		analytics.WinRate = float64(profitableTrades) / float64(len(fills))
	}
	
	return analytics, nil
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