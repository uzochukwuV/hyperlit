package engine

import (
	"context"
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/api"
	"hyperliquid-copy-trading/internal/database"
	"hyperliquid-copy-trading/internal/models"
	"hyperliquid-copy-trading/internal/utils"
	"hyperliquid-copy-trading/internal/websocket"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type CopyEngine struct {
	config         *config.Config
	db             *database.PostgresDB
	wsManager      *websocket.Manager
	orderEngine    *OrderEngine
	riskManager    *RiskManager
	hyperliquidAPI *api.HyperliquidAPI
	activeLeaders  map[string]bool
	leadersMutex   sync.RWMutex
	shutdown       chan struct{}
	wg             sync.WaitGroup
}

func NewCopyEngine(cfg *config.Config, db *database.PostgresDB, wsManager *websocket.Manager) *CopyEngine {
	hyperliquidAPI, err := api.NewHyperliquidAPI(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Hyperliquid API")
	}

	return &CopyEngine{
		config:         cfg,
		db:             db,
		wsManager:      wsManager,
		orderEngine:    NewOrderEngine(cfg, hyperliquidAPI),
		riskManager:    NewRiskManager(cfg),
		hyperliquidAPI: hyperliquidAPI,
		activeLeaders:  make(map[string]bool),
		shutdown:       make(chan struct{}),
	}
}

func (ce *CopyEngine) Start() {
	log.Info().Msg("Starting Copy Trading Engine")

	ctx := context.Background()

	// Start WebSocket connection monitoring
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ce.wsManager.MonitorConnections(ctx)
	}()

	// Load existing followers and start monitoring their leaders
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ce.loadAndMonitorFollowers(ctx)
	}()

	// Start position monitoring
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ce.monitorPositions(ctx)
	}()

	// Start periodic cleanup and maintenance
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ce.performMaintenance(ctx)
	}()

	log.Info().Msg("Copy Trading Engine started successfully")
}

func (ce *CopyEngine) Stop() {
	log.Info().Msg("Stopping Copy Trading Engine")
	close(ce.shutdown)
	ce.wg.Wait()
	log.Info().Msg("Copy Trading Engine stopped")
}

// GetWSHealth returns WebSocket connection health status
func (ce *CopyEngine) GetWSHealth() map[string]interface{} {
	health := ce.wsManager.HealthCheck()
	activeConnections := ce.wsManager.GetActiveConnections()
	
	return map[string]interface{}{
		"healthy":            len(health) > 0,
		"active_connections": activeConnections,
		"connection_details": health,
	}
}

// GetActiveWebSocketConnections returns the number of active WebSocket connections
func (ce *CopyEngine) GetActiveWebSocketConnections() int {
	return ce.wsManager.GetActiveConnections()
}

// GetOrderQueueStatus returns order queue status from order engine
func (ce *CopyEngine) GetOrderQueueStatus() map[string]interface{} {
	return ce.orderEngine.GetQueueStatus()
}

// AddFollower adds a new follower and starts monitoring their leader
func (ce *CopyEngine) AddFollower(ctx context.Context, follower *models.Follower) error {
	// Create follower in database
	if err := ce.db.CreateFollower(ctx, follower); err != nil {
		return err
	}
	
	// Start monitoring the leader if not already monitored
	ce.leadersMutex.Lock()
	if !ce.activeLeaders[follower.LeaderAddress] {
		ce.activeLeaders[follower.LeaderAddress] = true
		ce.leadersMutex.Unlock()
		ce.startMonitoringLeader(follower.LeaderAddress)
	} else {
		ce.leadersMutex.Unlock()
	}
	
	log.Info().Int("follower_id", follower.ID).Str("leader", follower.LeaderAddress).Msg("Follower added")
	return nil
}

// RemoveFollower removes a follower and stops monitoring their leader if no other followers
func (ce *CopyEngine) RemoveFollower(ctx context.Context, followerID int) error {
	// Delete from database
	if err := ce.db.DeleteFollower(ctx, followerID); err != nil {
		return err
	}
	
	// Check if we need to stop monitoring any leaders
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ce.loadActiveFollowers(ctx)
	}()
	
	log.Info().Int("follower_id", followerID).Msg("Follower removed")
	return nil
}

func (ce *CopyEngine) loadAndMonitorFollowers(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ce.shutdown:
			return
		case <-ticker.C:
			ce.loadActiveFollowers(ctx)
		}
	}
}

func (ce *CopyEngine) loadActiveFollowers(ctx context.Context) {
	followers, err := ce.db.GetFollowers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load followers")
		return
	}

	leaderMap := make(map[string]bool)
	for _, follower := range followers {
		if follower.IsActive {
			leaderMap[follower.LeaderAddress] = true
		}
	}

	// Start monitoring new leaders
	for leaderAddress := range leaderMap {
		ce.leadersMutex.RLock()
		isMonitored := ce.activeLeaders[leaderAddress]
		ce.leadersMutex.RUnlock()

		if !isMonitored {
			ce.startMonitoringLeader(leaderAddress)
		}
	}

	// Stop monitoring leaders with no followers
	ce.leadersMutex.Lock()
	for leaderAddress := range ce.activeLeaders {
		if !leaderMap[leaderAddress] {
			ce.stopMonitoringLeader(leaderAddress)
		}
	}
	ce.leadersMutex.Unlock()
}

func (ce *CopyEngine) startMonitoringLeader(leaderAddress string) {
	ce.leadersMutex.Lock()
	ce.activeLeaders[leaderAddress] = true
	ce.leadersMutex.Unlock()

	tradeChannel, userChannel, err := ce.wsManager.SubscribeToLeader(leaderAddress)
	if err != nil {
		log.Error().Err(err).Str("leader", leaderAddress).Msg("Failed to subscribe to leader")
		return
	}

	// Start monitoring trade events
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ce.monitorLeaderTrades(leaderAddress, tradeChannel)
	}()

	// Start monitoring user events
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ce.monitorLeaderUserEvents(leaderAddress, userChannel)
	}()

	log.Info().Str("leader", leaderAddress).Msg("Started monitoring leader")
}

func (ce *CopyEngine) stopMonitoringLeader(leaderAddress string) {
	delete(ce.activeLeaders, leaderAddress)
	ce.wsManager.UnsubscribeFromLeader(leaderAddress)
	log.Info().Str("leader", leaderAddress).Msg("Stopped monitoring leader")
}

func (ce *CopyEngine) monitorLeaderTrades(leaderAddress string, tradeChannel chan models.TradeEvent) {
	for {
		select {
		case <-ce.shutdown:
			return
		case tradeEvent, ok := <-tradeChannel:
			if !ok {
				return
			}
			ce.processLeaderTrade(leaderAddress, tradeEvent)
		}
	}
}

func (ce *CopyEngine) monitorLeaderUserEvents(leaderAddress string, userChannel chan models.UserEvent) {
	for {
		select {
		case <-ce.shutdown:
			return
		case userEvent, ok := <-userChannel:
			if !ok {
				return
			}
			ce.processLeaderUserEvent(leaderAddress, userEvent)
		}
	}
}

func (ce *CopyEngine) processLeaderTrade(leaderAddress string, tradeEvent models.TradeEvent) {
	log.Info().
		Str("leader", leaderAddress).
		Str("coin", tradeEvent.Coin).
		Str("side", tradeEvent.Side).
		Str("size", tradeEvent.Sz).
		Str("price", tradeEvent.Px).
		Msg("Leader trade detected")

	ctx := context.Background()

	// Store leader trade
	price, _ := strconv.ParseFloat(tradeEvent.Px, 64)
	size, _ := strconv.ParseFloat(tradeEvent.Sz, 64)

	leaderTrade := &models.Trade{
		LeaderAddress:   leaderAddress,
		Asset:           tradeEvent.Coin,
		Side:            tradeEvent.Side,
		Size:            size,
		Price:           price,
		OrderType:       "market",
		IsLeaderTrade:   true,
		ExecutedAt:      time.Unix(tradeEvent.Time/1000, 0),
		HyperliquidTxID: tradeEvent.Hash,
		Status:          "filled",
	}

	if err := ce.db.CreateTrade(ctx, leaderTrade); err != nil {
		log.Error().Err(err).Msg("Failed to store leader trade")
	}

	// Get followers for this leader
	followers, err := ce.db.GetFollowersByLeader(ctx, leaderAddress)
	if err != nil {
		log.Error().Err(err).Str("leader", leaderAddress).Msg("Failed to get followers")
		return
	}

	// Process followers in batches
	ce.processFollowersInBatches(ctx, followers, tradeEvent, leaderTrade)
}

func (ce *CopyEngine) processFollowersInBatches(ctx context.Context, followers []models.Follower, tradeEvent models.TradeEvent, leaderTrade *models.Trade) {
	batchSize := ce.config.MaxOrderBatchSize

	for i := 0; i < len(followers); i += batchSize {
		end := i + batchSize
		if end > len(followers) {
			end = len(followers)
		}

		batch := followers[i:end]
		ce.wg.Add(1)
		go func(followerBatch []models.Follower) {
			defer ce.wg.Done()
			ce.processBatch(ctx, followerBatch, tradeEvent, leaderTrade)
		}(batch)

		// Add small delay between batches to respect rate limits
		time.Sleep(ce.config.OrderBatchInterval)
	}
}

func (ce *CopyEngine) processBatch(ctx context.Context, followers []models.Follower, tradeEvent models.TradeEvent, leaderTrade *models.Trade) {
	var orders []*models.OrderRequest

	for _, follower := range followers {
		// Risk assessment
		riskAssessment := ce.riskManager.AssessRisk(&follower, leaderTrade)
		if !riskAssessment.Approved {
			log.Warn().
				Int("follower_id", follower.ID).
				Str("reason", riskAssessment.Reason).
				Msg("Trade rejected by risk management")
			continue
		}

		// Calculate position size
		positionSize := ce.calculatePositionSize(&follower, leaderTrade, riskAssessment.AdjustedSize)

		if positionSize <= 0 {
			continue
		}

		// Create order
		order := &models.OrderRequest{
			Asset:     leaderTrade.Asset,
			IsBuy:     leaderTrade.Side == "buy",
			Size:      positionSize,
			OrderType: "market",
			Nonce:     time.Now().UnixMilli() + int64(follower.ID), // Ensure unique nonce
		}

		orders = append(orders, order)

		// Store follower trade
		followerTrade := &models.Trade{
			LeaderAddress: leaderTrade.LeaderAddress,
			FollowerID:    &follower.ID,
			Asset:         leaderTrade.Asset,
			Side:          leaderTrade.Side,
			Size:          positionSize,
			Price:         leaderTrade.Price,
			OrderType:     "market",
			IsLeaderTrade: false,
			ExecutedAt:    time.Now(),
			Status:        "pending",
		}

		if err := ce.db.CreateTrade(ctx, followerTrade); err != nil {
			log.Error().Err(err).Int("follower_id", follower.ID).Msg("Failed to store follower trade")
		}
	}

	// Execute batch orders
	if len(orders) > 0 {
		ce.orderEngine.ExecuteBatch(ctx, orders, followers)
	}
}

func (ce *CopyEngine) calculatePositionSize(follower *models.Follower, leaderTrade *models.Trade, adjustedSize float64) float64 {
	// Base size calculation using copy percentage
	baseSize := leaderTrade.Size * (follower.CopyPercentage / 100.0)

	// Apply risk adjustment
	if adjustedSize < baseSize {
		baseSize = adjustedSize
	}

	// Apply maximum position size limit
	if baseSize > follower.MaxPositionSize {
		baseSize = follower.MaxPositionSize
	}

	// Round to appropriate precision (would get this from meta info in production)
	return utils.RoundToDecimals(baseSize, 3)
}

func (ce *CopyEngine) processLeaderUserEvent(leaderAddress string, userEvent models.UserEvent) {
	log.Debug().
		Str("leader", leaderAddress).
		Str("type", userEvent.Type).
		Msg("Leader user event")

	// Handle different user event types
	switch userEvent.Type {
	case "order":
		// Handle order events (filled, cancelled, etc.)
	case "liquidation":
		// Handle liquidation events
	case "funding":
		// Handle funding payments
	}
}

func (ce *CopyEngine) monitorPositions(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ce.shutdown:
			return
		case <-ticker.C:
			ce.updatePositions(ctx)
		}
	}
}

func (ce *CopyEngine) updatePositions(ctx context.Context) {
	// Get all unique user addresses from followers
	followers, err := ce.db.GetFollowers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get followers for position update")
		return
	}

	userAddresses := make(map[string]bool)
	for _, follower := range followers {
		if follower.IsActive {
			userAddresses[follower.APIWalletAddress] = true
		}
	}

	// Update positions for each user
	for userAddress := range userAddresses {
		ce.wg.Add(1)
		go func(addr string) {
			defer ce.wg.Done()
			ce.updateUserPositions(ctx, addr)
		}(userAddress)
	}
}

func (ce *CopyEngine) updateUserPositions(ctx context.Context, userAddress string) {
	userState, err := ce.hyperliquidAPI.GetUserState(ctx, userAddress)
	if err != nil {
		log.Error().Err(err).Str("user", userAddress).Msg("Failed to get user state")
		return
	}

	for _, assetPos := range userState.AssetPositions {
		position := &models.Position{
			UserAddress:   userAddress,
			Asset:         assetPos.Position.Asset,
			Side:          assetPos.Position.Side,
			Size:          assetPos.Position.Size,
			EntryPrice:    assetPos.Position.EntryPrice,
			CurrentPrice:  assetPos.Position.CurrentPrice,
			UnrealizedPnL: assetPos.Position.UnrealizedPnL,
			MarginUsed:    assetPos.Position.MarginUsed,
		}

		if err := ce.db.UpsertPosition(ctx, position); err != nil {
			log.Error().Err(err).Str("user", userAddress).Msg("Failed to update position")
		}
	}
}

func (ce *CopyEngine) performMaintenance(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ce.shutdown:
			return
		case <-ticker.C:
			ce.runMaintenance(ctx)
		}
	}
}

func (ce *CopyEngine) runMaintenance(ctx context.Context) {
	log.Debug().Msg("Running maintenance tasks")

	// Check WebSocket health
	health := ce.wsManager.HealthCheck()
	unhealthyCount := 0
	for _, isHealthy := range health {
		if !isHealthy {
			unhealthyCount++
		}
	}

	if unhealthyCount > 0 {
		log.Warn().Int("unhealthy_connections", unhealthyCount).Msg("Found unhealthy WebSocket connections")
	}

	// Log active monitoring stats
	ce.leadersMutex.RLock()
	activeLeaderCount := len(ce.activeLeaders)
	ce.leadersMutex.RUnlock()

	log.Info().
		Int("active_leaders", activeLeaderCount).
		Int("websocket_connections", ce.wsManager.GetActiveConnections()).
		Msg("Copy engine status")
}

