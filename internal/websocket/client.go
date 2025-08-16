package websocket

import (
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/models"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Client struct {
	config        *config.Config
	leaderAddress string
	conn          *websocket.Conn
	connected     bool
	connMutex     sync.RWMutex
	shutdown      chan struct{}
	wg            sync.WaitGroup
}

func NewClient(cfg *config.Config, leaderAddress string) *Client {
	return &Client{
		config:        cfg,
		leaderAddress: leaderAddress,
		shutdown:      make(chan struct{}),
	}
}

func (c *Client) Start(tradeChannel chan models.TradeEvent, userChannel chan models.UserEvent) error {
	// Determine WebSocket URL based on environment
	wsURL := c.config.HyperliquidWSURL
	if c.config.Environment == "testnet" {
		wsURL = c.config.HyperliquidTestnetWSURL
	}

	// Connect to WebSocket
	u, err := url.Parse(wsURL)
	if err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.connMutex.Lock()
	c.conn = conn
	c.connected = true
	c.connMutex.Unlock()

	log.Info().Str("leader", c.leaderAddress).Str("url", wsURL).Msg("WebSocket connected")

	// Subscribe to user events
	if err := c.subscribeToUserEvents(); err != nil {
		log.Error().Err(err).Msg("Failed to subscribe to user events")
		return err
	}

	// Subscribe to trades
	if err := c.subscribeToTrades(); err != nil {
		log.Error().Err(err).Msg("Failed to subscribe to trades")
		return err
	}

	// Start message handling
	c.wg.Add(2)
	go c.readMessages(tradeChannel, userChannel)
	go c.pingLoop()

	return nil
}

func (c *Client) Stop() {
	c.connMutex.Lock()
	if c.connected {
		c.connected = false
		close(c.shutdown)
		if c.conn != nil {
			c.conn.Close()
		}
	}
	c.connMutex.Unlock()

	c.wg.Wait()
}

func (c *Client) IsConnected() bool {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return c.connected
}

func (c *Client) subscribeToUserEvents() error {
	subscription := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "userEvents",
			"user": c.leaderAddress,
		},
	}

	return c.sendMessage(subscription)
}

func (c *Client) subscribeToTrades() error {
	// Subscribe to all coin trades to catch leader trades
	// In production, you might want to be more selective
	subscription := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "allTrades",
		},
	}

	return c.sendMessage(subscription)
}

func (c *Client) sendMessage(msg interface{}) error {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()

	if !c.connected || c.conn == nil {
		return websocket.ErrCloseSent
	}

	return c.conn.WriteJSON(msg)
}

func (c *Client) readMessages(tradeChannel chan models.TradeEvent, userChannel chan models.UserEvent) {
	defer c.wg.Done()

	for {
		select {
		case <-c.shutdown:
			return
		default:
			c.connMutex.RLock()
			conn := c.conn
			connected := c.connected
			c.connMutex.RUnlock()

			if !connected || conn == nil {
				return
			}

			// Set read deadline
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			var message map[string]interface{}
			if err := conn.ReadJSON(&message); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Error().Err(err).Str("leader", c.leaderAddress).Msg("WebSocket read error")
				}
				return
			}

			c.processMessage(message, tradeChannel, userChannel)
		}
	}
}

func (c *Client) processMessage(message map[string]interface{}, tradeChannel chan models.TradeEvent, userChannel chan models.UserEvent) {
	channel, ok := message["channel"].(string)
	if !ok {
		return
	}

	data, ok := message["data"].(map[string]interface{})
	if !ok {
		return
	}

	switch channel {
	case "trades":
		c.processTrades(data, tradeChannel)
	case "userEvents":
		c.processUserEvents(data, userChannel)
	case "allTrades":
		c.processAllTrades(data, tradeChannel)
	}
}

func (c *Client) processTrades(data map[string]interface{}, tradeChannel chan models.TradeEvent) {
	trades, ok := data["trades"].([]interface{})
	if !ok {
		return
	}

	for _, tradeData := range trades {
		trade, ok := tradeData.(map[string]interface{})
		if !ok {
			continue
		}

		tradeEvent := c.parseTradeEvent(trade)
		if tradeEvent != nil && tradeEvent.User == c.leaderAddress {
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

func (c *Client) processAllTrades(data map[string]interface{}, tradeChannel chan models.TradeEvent) {
	trades, ok := data["trades"].([]interface{})
	if !ok {
		return
	}

	for _, tradeData := range trades {
		trade, ok := tradeData.(map[string]interface{})
		if !ok {
			continue
		}

		tradeEvent := c.parseTradeEvent(trade)
		if tradeEvent != nil && tradeEvent.User == c.leaderAddress {
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

func (c *Client) processUserEvents(data map[string]interface{}, userChannel chan models.UserEvent) {
	events, ok := data["events"].([]interface{})
	if !ok {
		return
	}

	for _, eventData := range events {
		event, ok := eventData.(map[string]interface{})
		if !ok {
			continue
		}

		eventType, ok := event["type"].(string)
		if !ok {
			continue
		}

		userEvent := models.UserEvent{
			Type: eventType,
			Data: event,
		}

		select {
		case userChannel <- userEvent:
		case <-c.shutdown:
			return
		default:
			log.Warn().Msg("User channel full, dropping user event")
		}
	}
}

func (c *Client) parseTradeEvent(trade map[string]interface{}) *models.TradeEvent {
	event := &models.TradeEvent{}

	if coin, ok := trade["coin"].(string); ok {
		event.Coin = coin
	}

	if side, ok := trade["side"].(string); ok {
		event.Side = side
	}

	if px, ok := trade["px"].(string); ok {
		event.Px = px
	}

	if sz, ok := trade["sz"].(string); ok {
		event.Sz = sz
	}

	if hash, ok := trade["hash"].(string); ok {
		event.Hash = hash
	}

	if timeFloat, ok := trade["time"].(float64); ok {
		event.Time = int64(timeFloat)
	}

	if startPos, ok := trade["startPos"].(string); ok {
		event.StartPos = startPos
	}

	if tidFloat, ok := trade["tid"].(float64); ok {
		event.Tid = int64(tidFloat)
	}

	if fee, ok := trade["fee"].(string); ok {
		event.Fee = fee
	}

	if user, ok := trade["user"].(string); ok {
		event.User = user
	}

	return event
}

func (c *Client) pingLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.shutdown:
			return
		case <-ticker.C:
			c.connMutex.RLock()
			conn := c.conn
			connected := c.connected
			c.connMutex.RUnlock()

			if !connected || conn == nil {
				return
			}

			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error().Err(err).Str("leader", c.leaderAddress).Msg("Failed to send ping")
				return
			}
		}
	}
}
