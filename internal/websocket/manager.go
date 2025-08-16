package websocket

import (
	"context"
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/models"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	config        *config.Config
	clients       map[string]*Client
	clientsMutex  sync.RWMutex
	tradeChannels map[string]chan models.TradeEvent
	userChannels  map[string]chan models.UserEvent
	shutdown      chan struct{}
	wg            sync.WaitGroup
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:        cfg,
		clients:       make(map[string]*Client),
		tradeChannels: make(map[string]chan models.TradeEvent),
		userChannels:  make(map[string]chan models.UserEvent),
		shutdown:      make(chan struct{}),
	}
}

func (m *Manager) SubscribeToLeader(leaderAddress string) (chan models.TradeEvent, chan models.UserEvent, error) {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Check if we already have a client for this leader
	if _, exists := m.clients[leaderAddress]; exists {
		return m.tradeChannels[leaderAddress], m.userChannels[leaderAddress], nil
	}

	// Create new client
	client := NewClient(m.config, leaderAddress)
	
	// Create channels for this leader
	tradeChannel := make(chan models.TradeEvent, 1000)
	userChannel := make(chan models.UserEvent, 1000)
	
	m.clients[leaderAddress] = client
	m.tradeChannels[leaderAddress] = tradeChannel
	m.userChannels[leaderAddress] = userChannel

	// Start the client
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := client.Start(tradeChannel, userChannel); err != nil {
			log.Error().Err(err).Str("leader", leaderAddress).Msg("WebSocket client failed")
		}
	}()

	log.Info().Str("leader", leaderAddress).Msg("Subscribed to leader")
	return tradeChannel, userChannel, nil
}

func (m *Manager) UnsubscribeFromLeader(leaderAddress string) {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	if client, exists := m.clients[leaderAddress]; exists {
		client.Stop()
		delete(m.clients, leaderAddress)
		
		// Close channels
		if tradeChannel, exists := m.tradeChannels[leaderAddress]; exists {
			close(tradeChannel)
			delete(m.tradeChannels, leaderAddress)
		}
		
		if userChannel, exists := m.userChannels[leaderAddress]; exists {
			close(userChannel)
			delete(m.userChannels, leaderAddress)
		}

		log.Info().Str("leader", leaderAddress).Msg("Unsubscribed from leader")
	}
}

func (m *Manager) GetTradeStream(leaderAddress string) (chan models.TradeEvent, bool) {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	channel, exists := m.tradeChannels[leaderAddress]
	return channel, exists
}

func (m *Manager) GetUserStream(leaderAddress string) (chan models.UserEvent, bool) {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	channel, exists := m.userChannels[leaderAddress]
	return channel, exists
}

func (m *Manager) GetActiveConnections() int {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()
	return len(m.clients)
}

func (m *Manager) Close() {
	close(m.shutdown)

	m.clientsMutex.Lock()
	for leaderAddress, client := range m.clients {
		client.Stop()
		log.Info().Str("leader", leaderAddress).Msg("Stopping WebSocket client")
	}
	m.clientsMutex.Unlock()

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("All WebSocket clients stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Warn().Msg("Timeout waiting for WebSocket clients to stop")
	}
}

// HealthCheck returns the health status of all connections
func (m *Manager) HealthCheck() map[string]bool {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	health := make(map[string]bool)
	for leaderAddress, client := range m.clients {
		health[leaderAddress] = client.IsConnected()
	}
	return health
}

// RestartClient restarts a specific client connection
func (m *Manager) RestartClient(leaderAddress string) error {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	if client, exists := m.clients[leaderAddress]; exists {
		client.Stop()
		
		// Wait a bit before reconnecting
		time.Sleep(2 * time.Second)
		
		// Get existing channels
		tradeChannel := m.tradeChannels[leaderAddress]
		userChannel := m.userChannels[leaderAddress]
		
		// Restart the client
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			if err := client.Start(tradeChannel, userChannel); err != nil {
				log.Error().Err(err).Str("leader", leaderAddress).Msg("Failed to restart WebSocket client")
			}
		}()

		log.Info().Str("leader", leaderAddress).Msg("Restarted WebSocket client")
	}

	return nil
}

// MonitorConnections periodically checks connection health and reconnects if needed
func (m *Manager) MonitorConnections(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAndReconnect()
		}
	}
}

func (m *Manager) checkAndReconnect() {
	health := m.HealthCheck()
	
	for leaderAddress, isHealthy := range health {
		if !isHealthy {
			log.Warn().Str("leader", leaderAddress).Msg("Unhealthy connection detected, restarting")
			if err := m.RestartClient(leaderAddress); err != nil {
				log.Error().Err(err).Str("leader", leaderAddress).Msg("Failed to restart unhealthy connection")
			}
		}
	}
}
