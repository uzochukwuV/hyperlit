package handlers

import (
	"context"
	"encoding/json"
	"hyperliquid-copy-trading/internal/database"
	"hyperliquid-copy-trading/internal/engine"
	"hyperliquid-copy-trading/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type APIHandler struct {
	copyEngine *engine.CopyEngine
	db         *database.PostgresDB
}

func NewAPIHandler(copyEngine *engine.CopyEngine, db *database.PostgresDB) *APIHandler {
	return &APIHandler{
		copyEngine: copyEngine,
		db:         db,
	}
}

// Response helpers
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func (h *APIHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *APIHandler) writeError(w http.ResponseWriter, status int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	h.writeJSON(w, status, response)
}

func (h *APIHandler) writeSuccess(w http.ResponseWriter, data interface{}, message string) {
	response := APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	h.writeJSON(w, http.StatusOK, response)
}

// Follower endpoints
func (h *APIHandler) CreateFollower(w http.ResponseWriter, r *http.Request) {
	var follower models.Follower
	if err := json.NewDecoder(r.Body).Decode(&follower); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate follower settings
	if errors := h.validateFollower(&follower); len(errors) > 0 {
		response := APIResponse{
			Success: false,
			Error:   "Validation failed",
			Data:    map[string]interface{}{"validation_errors": errors},
		}
		h.writeJSON(w, http.StatusBadRequest, response)
		return
	}

	// Set defaults
	follower.IsActive = true
	if follower.CopyPercentage == 0 {
		follower.CopyPercentage = 10 // Default 10%
	}
	if follower.MaxPositionSize == 0 {
		follower.MaxPositionSize = 1000 // Default $1000
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := h.copyEngine.AddFollower(ctx, &follower); err != nil {
		log.Error().Err(err).Msg("Failed to create follower")
		h.writeError(w, http.StatusInternalServerError, "Failed to create follower")
		return
	}

	log.Info().Int("follower_id", follower.ID).Str("leader", follower.LeaderAddress).Msg("Follower created")
	h.writeSuccess(w, follower, "Follower created successfully")
}

func (h *APIHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	followers, err := h.db.GetFollowers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get followers")
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve followers")
		return
	}

	h.writeSuccess(w, followers, "")
}

func (h *APIHandler) UpdateFollower(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid follower ID")
		return
	}

	var updates models.Follower
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	updates.ID = id

	// Validate updates
	if errors := h.validateFollower(&updates); len(errors) > 0 {
		response := APIResponse{
			Success: false,
			Error:   "Validation failed",
			Data:    map[string]interface{}{"validation_errors": errors},
		}
		h.writeJSON(w, http.StatusBadRequest, response)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := h.db.UpdateFollower(ctx, &updates); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to update follower")
		h.writeError(w, http.StatusInternalServerError, "Failed to update follower")
		return
	}

	log.Info().Int("follower_id", id).Msg("Follower updated")
	h.writeSuccess(w, updates, "Follower updated successfully")
}

func (h *APIHandler) DeleteFollower(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid follower ID")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := h.copyEngine.RemoveFollower(ctx, id); err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to delete follower")
		h.writeError(w, http.StatusInternalServerError, "Failed to delete follower")
		return
	}

	log.Info().Int("follower_id", id).Msg("Follower deleted")
	h.writeSuccess(w, nil, "Follower deleted successfully")
}

// Leader endpoints
func (h *APIHandler) GetLeaders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	leaders, err := h.db.GetActiveLeaders(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get leaders")
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve leaders")
		return
	}

	h.writeSuccess(w, leaders, "")
}

func (h *APIHandler) GetLeaderPerformance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Parse days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30 // Default to 30 days
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	performance, err := h.db.GetLeaderPerformance(ctx, address, days)
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("Failed to get leader performance")
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve leader performance")
		return
	}

	h.writeSuccess(w, performance, "")
}

// Trade endpoints
func (h *APIHandler) GetTrades(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	trades, err := h.db.GetTrades(ctx, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get trades")
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve trades")
		return
	}

	response := map[string]interface{}{
		"trades": trades,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(trades),
		},
	}

	h.writeSuccess(w, response, "")
}

// Position endpoints
func (h *APIHandler) GetPositions(w http.ResponseWriter, r *http.Request) {
	userAddress := r.URL.Query().Get("user_address")
	if userAddress == "" {
		h.writeError(w, http.StatusBadRequest, "user_address parameter is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	positions, err := h.db.GetPositions(ctx, userAddress)
	if err != nil {
		log.Error().Err(err).Str("user", userAddress).Msg("Failed to get positions")
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve positions")
		return
	}

	h.writeSuccess(w, positions, "")
}

// Analytics endpoints
func (h *APIHandler) GetPnLAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	followerIDStr := vars["follower_id"]
	
	followerID, err := strconv.Atoi(followerIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid follower ID")
		return
	}

	// Parse days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30 // Default to 30 days
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	analytics, err := h.db.GetFollowerPnL(ctx, followerID, days)
	if err != nil {
		log.Error().Err(err).Int("follower_id", followerID).Msg("Failed to get PnL analytics")
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve PnL analytics")
		return
	}

	h.writeSuccess(w, analytics, "")
}

// Health check endpoint
func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check WebSocket connections health
	wsHealth := h.copyEngine.GetWSHealth()
	
	// Check database connection
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	_, dbErr := h.db.GetFollowers(ctx)
	dbHealthy := dbErr == nil

	health := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now().Unix(),
		"services": map[string]interface{}{
			"database":   dbHealthy,
			"websocket":  wsHealth,
		},
	}

	if !dbHealthy {
		health["status"] = "unhealthy"
		health["errors"] = []string{"Database connection failed"}
	}

	status := http.StatusOK
	if health["status"] == "unhealthy" {
		status = http.StatusServiceUnavailable
	}

	h.writeJSON(w, status, health)
}

// System status endpoint
func (h *APIHandler) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	wsConnections := h.copyEngine.GetActiveWebSocketConnections()
	orderQueueStatus := h.copyEngine.GetOrderQueueStatus()

	status := map[string]interface{}{
		"timestamp":           time.Now().Unix(),
		"websocket_connections": wsConnections,
		"order_queue":         orderQueueStatus,
		"uptime_seconds":      time.Since(time.Now()).Seconds(), // Would track actual uptime
	}

	h.writeSuccess(w, status, "")
}

// Validation helpers
func (h *APIHandler) validateFollower(follower *models.Follower) []string {
	var errors []string

	if follower.UserID == "" {
		errors = append(errors, "user_id is required")
	}

	if follower.LeaderAddress == "" {
		errors = append(errors, "leader_address is required")
	}

	if follower.APIWalletAddress == "" {
		errors = append(errors, "api_wallet_address is required")
	}

	if follower.CopyPercentage <= 0 || follower.CopyPercentage > 100 {
		errors = append(errors, "copy_percentage must be between 0 and 100")
	}

	if follower.MaxPositionSize <= 0 {
		errors = append(errors, "max_position_size must be positive")
	}

	if follower.StopLossPercentage != nil && (*follower.StopLossPercentage <= 0 || *follower.StopLossPercentage >= 100) {
		errors = append(errors, "stop_loss_percentage must be between 0 and 100")
	}

	if follower.TakeProfitPercentage != nil && *follower.TakeProfitPercentage <= 0 {
		errors = append(errors, "take_profit_percentage must be positive")
	}

	return errors
}

// CORS middleware
func (h *APIHandler) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func (h *APIHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap ResponseWriter to capture status code
		wrapped := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.statusCode).
			Dur("duration", time.Since(start)).
			Str("remote_addr", r.RemoteAddr).
			Msg("HTTP request")
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
