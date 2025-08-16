package api

import (
	"context"
	"encoding/json"
	"fmt"
	"hyperliquid-copy-trading/internal/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// WebSocketClient handles real-time connections to Hyperliquid
type WebSocketClient struct {
	conn          *websocket.Conn
	api           *HyperliquidAPI
	subscriptions map[string]bool
	handlers      map[string][]func(interface{})
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(api *HyperliquidAPI) (*WebSocketClient, error) {
	wsURL := "wss://api.hyperliquid.xyz/ws"
	if api.config.Environment == "testnet" {
		wsURL = "wss://api-testnet.hyperliquid.xyz/ws"
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &WebSocketClient{
		conn:          conn,
		api:           api,
		subscriptions: make(map[string]bool),
		handlers:      make(map[string][]func(interface{})),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start message handling goroutine
	go client.handleMessages()

	return client, nil
}

// SubscribeUserFills subscribes to user fills for real-time trade replication
func (ws *WebSocketClient) SubscribeUserFills(userAddress string, handler func(models.EnhancedTradeEvent)) error {
	subscriptionKey := fmt.Sprintf("userFills:%s", userAddress)

	ws.mutex.Lock()
	if ws.subscriptions[subscriptionKey] {
		ws.mutex.Unlock()
		return fmt.Errorf("already subscribed to user fills for %s", userAddress)
	}
	ws.subscriptions[subscriptionKey] = true

	// Add handler
	if ws.handlers[subscriptionKey] == nil {
		ws.handlers[subscriptionKey] = make([]func(interface{}), 0)
	}
	ws.handlers[subscriptionKey] = append(ws.handlers[subscriptionKey], func(data interface{}) {
		if tradeEvent, ok := data.(models.EnhancedTradeEvent); ok {
			handler(tradeEvent)
		}
	})
	ws.mutex.Unlock()

	subscription := models.WebSocketMessage{
		Method: "subscribe",
		Subscription: map[string]interface{}{
			"type": "userFills",
			"user": userAddress,
		},
	}

	return ws.conn.WriteJSON(subscription)
}

// SubscribeL2Book subscribes to L2 order book updates
func (ws *WebSocketClient) SubscribeL2Book(coin string, handler func(models.L2Book)) error {
	subscriptionKey := fmt.Sprintf("l2Book:%s", coin)

	ws.mutex.Lock()
	if ws.subscriptions[subscriptionKey] {
		ws.mutex.Unlock()
		return fmt.Errorf("already subscribed to L2 book for %s", coin)
	}
	ws.subscriptions[subscriptionKey] = true

	// Add handler
	if ws.handlers[subscriptionKey] == nil {
		ws.handlers[subscriptionKey] = make([]func(interface{}), 0)
	}
	ws.handlers[subscriptionKey] = append(ws.handlers[subscriptionKey], func(data interface{}) {
		if l2Book, ok := data.(models.L2Book); ok {
			handler(l2Book)
		}
	})
	ws.mutex.Unlock()

	subscription := models.WebSocketMessage{
		Method: "subscribe",
		Subscription: map[string]interface{}{
			"type": "l2Book",
			"coin": coin,
		},
	}

	return ws.conn.WriteJSON(subscription)
}

// SubscribeAllMids subscribes to all mid prices
func (ws *WebSocketClient) SubscribeAllMids(handler func(map[string]string)) error {
	subscriptionKey := "allMids"

	ws.mutex.Lock()
	if ws.subscriptions[subscriptionKey] {
		ws.mutex.Unlock()
		return fmt.Errorf("already subscribed to all mids")
	}
	ws.subscriptions[subscriptionKey] = true

	// Add handler
	if ws.handlers[subscriptionKey] == nil {
		ws.handlers[subscriptionKey] = make([]func(interface{}), 0)
	}
	ws.handlers[subscriptionKey] = append(ws.handlers[subscriptionKey], func(data interface{}) {
		if mids, ok := data.(map[string]string); ok {
			handler(mids)
		}
	})
	ws.mutex.Unlock()

	subscription := models.WebSocketMessage{
		Method: "subscribe",
		Subscription: map[string]interface{}{
			"type": "allMids",
		},
	}

	return ws.conn.WriteJSON(subscription)
}

// Unsubscribe removes a subscription
func (ws *WebSocketClient) Unsubscribe(subscriptionType, identifier string) error {
	subscriptionKey := fmt.Sprintf("%s:%s", subscriptionType, identifier)
	if identifier == "" {
		subscriptionKey = subscriptionType
	}

	ws.mutex.Lock()
	delete(ws.subscriptions, subscriptionKey)
	delete(ws.handlers, subscriptionKey)
	ws.mutex.Unlock()

	unsubscribe := models.WebSocketMessage{
		Method: "unsubscribe",
		Subscription: map[string]interface{}{
			"type": subscriptionType,
		},
	}

	if identifier != "" {
		switch subscriptionType {
		case "userFills":
			unsubscribe.Subscription.(map[string]interface{})["user"] = identifier
		case "l2Book":
			unsubscribe.Subscription.(map[string]interface{})["coin"] = identifier
		}
	}

	return ws.conn.WriteJSON(unsubscribe)
}

// handleMessages processes incoming WebSocket messages
func (ws *WebSocketClient) handleMessages() {
	defer ws.Close()

	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
			var msg models.WebSocketMessage
			err := ws.conn.ReadJSON(&msg)
			if err != nil {
				log.Error().Err(err).Msg("WebSocket read error")
				return
			}

			ws.processMessage(msg)
		}
	}
}

// processMessage handles different types of WebSocket messages
func (ws *WebSocketClient) processMessage(msg models.WebSocketMessage) {
	if msg.Data == nil {
		return
	}

	// Convert data to JSON for easier parsing
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal WebSocket data")
		return
	}

	// Determine message type and route to appropriate handlers
	if subscription, ok := msg.Subscription.(map[string]interface{}); ok {
		msgType, typeOk := subscription["type"].(string)
		if !typeOk {
			return
		}

		switch msgType {
		case "userFills":
			ws.handleUserFills(dataBytes, subscription)
		case "l2Book":
			ws.handleL2Book(dataBytes, subscription)
		case "allMids":
			ws.handleAllMids(dataBytes)
		default:
			log.Debug().Str("type", msgType).Msg("Unknown WebSocket message type")
		}
	}
}

// handleUserFills processes user fill messages
func (ws *WebSocketClient) handleUserFills(dataBytes []byte, subscription map[string]interface{}) {
	userAddress, ok := subscription["user"].(string)
	if !ok {
		return
	}

	var tradeEvent models.EnhancedTradeEvent
	if err := json.Unmarshal(dataBytes, &tradeEvent); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal trade event")
		return
	}

	subscriptionKey := fmt.Sprintf("userFills:%s", userAddress)
	ws.mutex.RLock()
	handlers := ws.handlers[subscriptionKey]
	ws.mutex.RUnlock()

	for _, handler := range handlers {
		go handler(tradeEvent)
	}
}

// handleL2Book processes L2 book messages
func (ws *WebSocketClient) handleL2Book(dataBytes []byte, subscription map[string]interface{}) {
	coin, ok := subscription["coin"].(string)
	if !ok {
		return
	}

	var l2Book models.L2Book
	if err := json.Unmarshal(dataBytes, &l2Book); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal L2 book")
		return
	}

	subscriptionKey := fmt.Sprintf("l2Book:%s", coin)
	ws.mutex.RLock()
	handlers := ws.handlers[subscriptionKey]
	ws.mutex.RUnlock()

	for _, handler := range handlers {
		go handler(l2Book)
	}
}

// handleAllMids processes all mids messages
func (ws *WebSocketClient) handleAllMids(dataBytes []byte) {
	var mids map[string]string
	if err := json.Unmarshal(dataBytes, &mids); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal all mids")
		return
	}

	subscriptionKey := "allMids"
	ws.mutex.RLock()
	handlers := ws.handlers[subscriptionKey]
	ws.mutex.RUnlock()

	for _, handler := range handlers {
		go handler(mids)
	}
}

// Ping sends a ping message to keep connection alive
func (ws *WebSocketClient) Ping() error {
	return ws.conn.WriteMessage(websocket.PingMessage, []byte{})
}

// Close closes the WebSocket connection
func (ws *WebSocketClient) Close() error {
	ws.cancel()
	return ws.conn.Close()
}

// IsConnected checks if the WebSocket connection is still active
func (ws *WebSocketClient) IsConnected() bool {
	select {
	case <-ws.ctx.Done():
		return false
	default:
		return true
	}
}

// StartPingLoop starts a goroutine that sends periodic ping messages
func (ws *WebSocketClient) StartPingLoop(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ws.ctx.Done():
				return
			case <-ticker.C:
				if err := ws.Ping(); err != nil {
					log.Error().Err(err).Msg("Failed to send ping")
					return
				}
			}
		}
	}()
}
