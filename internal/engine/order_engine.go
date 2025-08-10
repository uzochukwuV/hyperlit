package engine

import (
	"context"
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/api"
	"hyperliquid-copy-trading/internal/models"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type OrderEngine struct {
	config         *config.Config
	hyperliquidAPI *api.HyperliquidAPI
	orderQueue     chan *OrderBatch
	shutdown       chan struct{}
	wg             sync.WaitGroup
}

type OrderBatch struct {
	Orders    []*models.OrderRequest
	Followers []models.Follower
	Timestamp time.Time
}

func NewOrderEngine(cfg *config.Config, api *api.HyperliquidAPI) *OrderEngine {
	engine := &OrderEngine{
		config:         cfg,
		hyperliquidAPI: api,
		orderQueue:     make(chan *OrderBatch, 1000),
		shutdown:       make(chan struct{}),
	}

	// Start order processing worker
	engine.wg.Add(1)
	go engine.processOrders()

	return engine
}

func (oe *OrderEngine) ExecuteBatch(ctx context.Context, orders []*models.OrderRequest, followers []models.Follower) {
	batch := &OrderBatch{
		Orders:    orders,
		Followers: followers,
		Timestamp: time.Now(),
	}

	select {
	case oe.orderQueue <- batch:
		log.Debug().Int("orders", len(orders)).Msg("Order batch queued")
	default:
		log.Warn().Msg("Order queue full, dropping batch")
	}
}

func (oe *OrderEngine) processOrders() {
	defer oe.wg.Done()

	for {
		select {
		case <-oe.shutdown:
			return
		case batch := <-oe.orderQueue:
			oe.processBatch(batch)
		}
	}
}

func (oe *OrderEngine) processBatch(batch *OrderBatch) {
	ctx := context.Background()
	
	// Group orders by API wallet
	walletGroups := make(map[string]*WalletBatch)
	
	for i, order := range batch.Orders {
		if i >= len(batch.Followers) {
			continue
		}
		
		follower := batch.Followers[i]
		walletAddr := follower.APIWalletAddress
		
		if walletGroups[walletAddr] == nil {
			walletGroups[walletAddr] = &WalletBatch{
				APIWalletAddress: walletAddr,
				Orders:          []*models.OrderRequest{},
				Followers:       []models.Follower{},
			}
		}
		
		walletGroups[walletAddr].Orders = append(walletGroups[walletAddr].Orders, order)
		walletGroups[walletAddr].Followers = append(walletGroups[walletAddr].Followers, follower)
	}

	// Process each wallet batch
	for _, walletBatch := range walletGroups {
		oe.wg.Add(1)
		go func(wb *WalletBatch) {
			defer oe.wg.Done()
			oe.processWalletBatch(ctx, wb)
		}(walletBatch)
	}
}

type WalletBatch struct {
	APIWalletAddress string
	Orders          []*models.OrderRequest
	Followers       []models.Follower
}

func (oe *OrderEngine) processWalletBatch(ctx context.Context, batch *WalletBatch) {
	if len(batch.Orders) == 0 {
		return
	}

	// Generate unique nonce for this batch
	nonce := time.Now().UnixMilli()

	// Set nonce for all orders in batch
	for _, order := range batch.Orders {
		order.Nonce = nonce
	}

	log.Info().
		Str("wallet", batch.APIWalletAddress).
		Int("orders", len(batch.Orders)).
		Int64("nonce", nonce).
		Msg("Processing wallet batch")

	// Execute batch order
	response, err := oe.hyperliquidAPI.BatchOrders(ctx, batch.Orders, batch.APIWalletAddress, nonce)
	if err != nil {
		log.Error().
			Err(err).
			Str("wallet", batch.APIWalletAddress).
			Msg("Failed to execute batch orders")
		
		// Mark all orders as failed
		oe.markOrdersStatus(batch, "failed", err.Error())
		return
	}

	// Process response
	if response.Status == "ok" {
		log.Info().
			Str("wallet", batch.APIWalletAddress).
			Int("orders", len(batch.Orders)).
			Msg("Batch orders executed successfully")
		
		oe.markOrdersStatus(batch, "submitted", "")
		
		// Start monitoring order status
		oe.wg.Add(1)
		go func() {
			defer oe.wg.Done()
			oe.monitorBatchStatus(ctx, batch, response)
		}()
	} else {
		log.Error().
			Str("wallet", batch.APIWalletAddress).
			Interface("response", response).
			Msg("Batch order execution failed")
		
		oe.markOrdersStatus(batch, "failed", "API returned error status")
	}
}

func (oe *OrderEngine) markOrdersStatus(batch *WalletBatch, status string, errorMsg string) {
	// In a real implementation, you would update the database here
	for i, order := range batch.Orders {
		if i < len(batch.Followers) {
			follower := batch.Followers[i]
			log.Debug().
				Int("follower_id", follower.ID).
				Str("status", status).
				Str("asset", order.Asset).
				Float64("size", order.Size).
				Str("error", errorMsg).
				Msg("Order status updated")
		}
	}
}

func (oe *OrderEngine) monitorBatchStatus(ctx context.Context, batch *WalletBatch, response *models.HyperliquidAPIResponse) {
	// Extract order IDs from response
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		log.Error().Msg("Invalid response data format")
		return
	}

	statuses, ok := data["statuses"].([]interface{})
	if !ok {
		log.Error().Msg("No order statuses in response")
		return
	}

	// Monitor each order
	for i, statusData := range statuses {
		if i >= len(batch.Orders) || i >= len(batch.Followers) {
			continue
		}

		statusMap, ok := statusData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract order ID if available
		if resting, ok := statusMap["resting"].(map[string]interface{}); ok {
			if oidFloat, exists := resting["oid"].(float64); exists {
				oid := int64(oidFloat)
				follower := batch.Followers[i]
				
				oe.wg.Add(1)
				go func(orderID int64, f models.Follower) {
					defer oe.wg.Done()
					oe.monitorOrderStatus(ctx, orderID, f.APIWalletAddress, f.ID)
				}(oid, follower)
			}
		}
	}
}

func (oe *OrderEngine) monitorOrderStatus(ctx context.Context, orderID int64, walletAddress string, followerID int) {
	maxAttempts := 30 // 5 minutes with 10-second intervals
	attempt := 0

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-oe.shutdown:
			return
		case <-ticker.C:
			attempt++
			
			status, err := oe.hyperliquidAPI.GetOrderStatus(ctx, walletAddress, orderID)
			if err != nil {
				log.Error().
					Err(err).
					Int64("oid", orderID).
					Int("follower_id", followerID).
					Msg("Failed to get order status")
				
				if attempt >= maxAttempts {
					return
				}
				continue
			}

			// Check if order is filled
			if orderStatus, ok := status["status"].(string); ok {
				switch orderStatus {
				case "filled":
					log.Info().
						Int64("oid", orderID).
						Int("follower_id", followerID).
						Msg("Order filled successfully")
					
					// Update trade status in database
					oe.updateTradeStatus(followerID, orderID, "filled", status)
					return
					
				case "cancelled", "rejected":
					log.Warn().
						Int64("oid", orderID).
						Int("follower_id", followerID).
						Str("status", orderStatus).
						Msg("Order cancelled or rejected")
					
					oe.updateTradeStatus(followerID, orderID, orderStatus, status)
					return
				}
			}

			if attempt >= maxAttempts {
				log.Warn().
					Int64("oid", orderID).
					Int("follower_id", followerID).
					Msg("Order status monitoring timeout")
				return
			}
		}
	}
}

func (oe *OrderEngine) updateTradeStatus(followerID int, orderID int64, status string, orderData map[string]interface{}) {
	// In a real implementation, update the database here
	log.Debug().
		Int("follower_id", followerID).
		Int64("oid", orderID).
		Str("status", status).
		Interface("order_data", orderData).
		Msg("Trade status updated")
}

func (oe *OrderEngine) ExecuteSingleOrder(ctx context.Context, order *models.OrderRequest, follower *models.Follower) error {
	response, err := oe.hyperliquidAPI.PlaceOrder(ctx, order, follower.APIWalletAddress)
	if err != nil {
		return err
	}

	if response.Status != "ok" {
		return fmt.Errorf("order failed: %v", response.Data)
	}

	log.Info().
		Int("follower_id", follower.ID).
		Str("asset", order.Asset).
		Float64("size", order.Size).
		Msg("Single order executed successfully")

	return nil
}

func (oe *OrderEngine) CancelOrder(ctx context.Context, asset string, orderID int64, walletAddress string) error {
	nonce := time.Now().UnixMilli()
	response, err := oe.hyperliquidAPI.CancelOrder(ctx, asset, orderID, walletAddress, nonce)
	if err != nil {
		return err
	}

	if response.Status != "ok" {
		return fmt.Errorf("cancel failed: %v", response.Data)
	}

	return nil
}

func (oe *OrderEngine) Stop() {
	close(oe.shutdown)
	oe.wg.Wait()
}

// GetQueueStatus returns current queue statistics
func (oe *OrderEngine) GetQueueStatus() map[string]interface{} {
	return map[string]interface{}{
		"queue_length": len(oe.orderQueue),
		"queue_capacity": cap(oe.orderQueue),
	}
}

// FlushQueue processes all remaining orders in queue
func (oe *OrderEngine) FlushQueue(ctx context.Context) {
	log.Info().Msg("Flushing order queue")
	
	timeout := time.After(30 * time.Second)
	
	for {
		select {
		case <-timeout:
			log.Warn().Msg("Queue flush timeout")
			return
		case batch := <-oe.orderQueue:
			oe.processBatch(batch)
		default:
			log.Info().Msg("Order queue flushed")
			return
		}
	}
}
