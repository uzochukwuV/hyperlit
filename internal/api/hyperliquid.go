package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hyperliquid-copy-trading/config"
	"hyperliquid-copy-trading/internal/models"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// NonceManager handles collision-free nonce generation
type NonceManager struct {
	nonces map[string]int64 // Map of wallet address to last used nonce
	mutex  sync.Mutex
}

func NewNonceManager() *NonceManager {
	return &NonceManager{
		nonces: make(map[string]int64),
	}
}

// GetNextNonce generates a unique nonce for the wallet within Hyperliquid's time window
func (nm *NonceManager) GetNextNonce(walletAddress string) int64 {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Use current Unix millisecond timestamp as base
	currentTime := time.Now().UnixMilli()
	lastNonce, exists := nm.nonces[walletAddress]

	// Ensure nonce is within valid window (T - 2 days to T + 1 day)
	minNonce := currentTime - 2*24*60*60*1000
	maxNonce := currentTime + 24*60*60*1000

	if !exists || lastNonce < currentTime {
		nm.nonces[walletAddress] = currentTime
		return currentTime
	}

	// Increment last nonce
	nextNonce := lastNonce + 1
	if nextNonce > maxNonce {
		nextNonce = currentTime
	}
	if nextNonce < minNonce {
		nextNonce = minNonce
	}

	nm.nonces[walletAddress] = nextNonce
	return nextNonce
}

type HyperliquidAPI struct {
	config       *config.Config
	httpClient   *http.Client
	signer       *Signer
	nonceManager *NonceManager
	perpMeta     *models.MetaInfo
	spotMeta     *models.SpotMetaInfo
	metaMutex    sync.RWMutex
}

func NewHyperliquidAPI(cfg *config.Config) (*HyperliquidAPI, error) {
	signer, err := NewSigner(cfg.APIWalletPrivateKeys["default"])
	if err != nil {
		return nil, err
	}

	api := &HyperliquidAPI{
		config:       cfg,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		signer:       signer,
		nonceManager: NewNonceManager(),
	}

	// Initialize metadata cache
	if err := api.refreshMetaData(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Failed to initialize metadata cache")
	}

	return api, nil
}

// refreshMetaData fetches and caches perp and spot metadata
func (api *HyperliquidAPI) refreshMetaData(ctx context.Context) error {
	perpMeta, err := api.GetMetaInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch perp metadata: %w", err)
	}

	spotMeta, err := api.GetSpotMetaInfo(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch spot metadata, continuing with perp only")
	}

	api.metaMutex.Lock()
	defer api.metaMutex.Unlock()
	api.perpMeta = perpMeta
	api.spotMeta = spotMeta
	return nil
}

func (api *HyperliquidAPI) GetMetaInfo(ctx context.Context) (*models.MetaInfo, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "meta",
	}

	var metaInfo models.MetaInfo
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &metaInfo)
	return &metaInfo, err
}

// GetSpotMetaInfo fetches spot metadata
func (api *HyperliquidAPI) GetSpotMetaInfo(ctx context.Context) (*models.SpotMetaInfo, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "spotMeta",
	}

	var spotMeta models.SpotMetaInfo
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &spotMeta)
	return &spotMeta, err
}

// GetSpotClearinghouseState fetches user's spot token balances
func (api *HyperliquidAPI) GetSpotClearinghouseState(ctx context.Context, userAddress string) (*models.SpotClearinghouseState, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "spotClearinghouseState",
		"user": userAddress,
	}

	var state models.SpotClearinghouseState
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &state)
	return &state, err
}

// GetL2Book fetches L2 order book snapshot for liquidity validation
func (api *HyperliquidAPI) GetL2Book(ctx context.Context, coin string) (*models.L2Book, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "l2Book",
		"coin": coin,
	}

	var book models.L2Book
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &book)
	return &book, err
}

// GetActiveAssetData fetches user's active asset data for margin checks
func (api *HyperliquidAPI) GetActiveAssetData(ctx context.Context, userAddress, coin string) (*models.ActiveAssetData, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "activeAssetData",
		"user": userAddress,
		"coin": coin,
	}

	var data models.ActiveAssetData
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &data)
	return &data, err
}

// GetUserFees fetches user's fee schedule
func (api *HyperliquidAPI) GetUserFees(ctx context.Context, userAddress string) (*models.UserFees, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "userFees",
		"user": userAddress,
	}

	var fees models.UserFees
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &fees)
	return &fees, err
}

// GetPortfolio fetches user's portfolio data
func (api *HyperliquidAPI) GetPortfolio(ctx context.Context, userAddress string) (*models.Portfolio, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "portfolio",
		"user": userAddress,
	}

	var portfolio models.Portfolio
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &portfolio)
	return &portfolio, err
}

// GetUserFills fetches user's trade fills
func (api *HyperliquidAPI) GetUserFills(ctx context.Context, userAddress string) ([]models.EnhancedTradeEvent, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "userFills",
		"user": userAddress,
	}

	var fills []models.EnhancedTradeEvent
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &fills)
	return fills, err
}

// GetPerpsAtOpenInterestCap fetches assets at open interest cap
func (api *HyperliquidAPI) GetPerpsAtOpenInterestCap(ctx context.Context) ([]string, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "perpsAtOpenInterestCap",
	}

	var cappedAssets []string
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &cappedAssets)
	return cappedAssets, err
}

func (api *HyperliquidAPI) GetUserState(ctx context.Context, userAddress string) (*models.UserState, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "clearinghouseState",
		"user": userAddress,
	}

	var userState models.UserState
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &userState)
	return &userState, err
}

func (api *HyperliquidAPI) GetAllMids(ctx context.Context) (map[string]string, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "allMids",
	}

	var mids map[string]string
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &mids)
	return mids, err
}

// ValidateOrder checks liquidity and margin/balance before placing order
func (api *HyperliquidAPI) ValidateOrder(ctx context.Context, order *models.EnhancedOrderRequest, userAddress string, isPerp bool) error {
	// Check liquidity
	l2Book, err := api.GetL2Book(ctx, order.Asset)
	if err != nil {
		return fmt.Errorf("failed to get order book for %s: %w", order.Asset, err)
	}

	price, _ := strconv.ParseFloat(api.formatPrice(order.Price), 64)
	size := order.Size
	side := "asks"
	if order.IsBuy {
		side = "bids"
	}

	availableSize := 0.0
	if levels, ok := l2Book.Levels[side]; ok {
		for _, level := range levels {
			levelPx, _ := strconv.ParseFloat(level.Px, 64)
			if (order.IsBuy && levelPx <= price) || (!order.IsBuy && levelPx >= price) {
				levelSz, _ := strconv.ParseFloat(level.Sz, 64)
				availableSize += levelSz
			}
		}
	}

	if availableSize < size {
		return fmt.Errorf("insufficient liquidity for %s: need %f, available %f", order.Asset, size, availableSize)
	}

	// Check margin (perps) or balance (spot)
	if isPerp {
		assetData, err := api.GetActiveAssetData(ctx, userAddress, order.Asset)
		if err != nil {
			return fmt.Errorf("failed to get asset data: %w", err)
		}
		if len(assetData.MaxTradeSzs) > 0 {
			maxSz, _ := strconv.ParseFloat(assetData.MaxTradeSzs[0], 64)
			if size > maxSz {
				return fmt.Errorf("order size %f exceeds maxTradeSzs %f", size, maxSz)
			}
		}
	} else {
		balances, err := api.GetSpotClearinghouseState(ctx, userAddress)
		if err != nil {
			return fmt.Errorf("failed to get spot balances: %w", err)
		}

		// For spot, check if user has sufficient balance
		coin := order.Asset
		for _, balance := range balances.Balances {
			if balance.Coin == coin {
				total, _ := strconv.ParseFloat(balance.Total, 64)
				if total < size {
					return fmt.Errorf("insufficient balance for %s: need %f, available %f", coin, size, total)
				}
				break
			}
		}
	}

	return nil
}

// PlaceOrder with enhanced validation and error handling
func (api *HyperliquidAPI) PlaceOrder(ctx context.Context, order *models.EnhancedOrderRequest, apiWalletAddress string) (*models.OrderResponse, error) {
	// Validate order first
	if err := api.ValidateOrder(ctx, order, apiWalletAddress, true); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// Check fees
	fees, err := api.GetUserFees(ctx, apiWalletAddress)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get user fees")
	} else {
		// Estimate fee and ensure sufficient balance
		feeRate, _ := strconv.ParseFloat(fees.UserCrossRate, 64)
		notional := *order.Price * order.Size
		estimatedFee := notional * feeRate

		userState, err := api.GetUserState(ctx, apiWalletAddress)
		if err == nil {
			available, _ := strconv.ParseFloat(userState.MarginSummary.AccountValue, 64)
			if available < estimatedFee {
				return nil, fmt.Errorf("insufficient funds for fees: need %f, available %f", estimatedFee, available)
			}
		}
	}

	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	// Convert asset to proper format
	assetID, err := api.getAssetID(order.Asset)
	if err != nil {
		return nil, err
	}

	// Generate nonce automatically
	nonce := api.nonceManager.GetNextNonce(apiWalletAddress)

	// Prepare order data with enhanced options
	orderData := map[string]interface{}{
		"a": assetID,
		"b": order.IsBuy,
		"p": api.formatPrice(order.Price),
		"s": api.formatSize(order.Size),
		"r": false, // reduceOnly
		"t": api.getOrderTypeCode(order.OrderType, order.Tif),
	}

	if order.ClOid != nil {
		orderData["c"] = *order.ClOid
	} else {
		orderData["c"] = nil
	}

	orderAction := map[string]interface{}{
		"type":     "order",
		"orders":   []map[string]interface{}{orderData},
		"grouping": "na",
	}

	// Sign the action
	signature, err := api.signer.SignAction(orderAction, apiWalletAddress, nonce)
	if err != nil {
		return nil, err
	}

	// Prepare request body
	reqBody := map[string]interface{}{
		"action":       orderAction,
		"nonce":        nonce,
		"signature":    signature,
		"vaultAddress": nil,
	}

	var response models.OrderResponse
	err = api.makeRequest(ctx, apiURL+"/exchange", reqBody, &response)
	if err != nil {
		return nil, err
	}

	// Enhanced error handling
	if response.Status != "success" {
		for _, status := range response.Data.Statuses {
			if status.Error != "" {
				return nil, fmt.Errorf("order rejected: %s", status.Error)
			}
		}
		return nil, fmt.Errorf("order failed with status: %s", response.Status)
	}

	return &response, nil
}


func (api *HyperliquidAPI) CancelOrder(ctx context.Context, asset string, oid int64, apiWalletAddress string, nonce int64) (*models.HyperliquidAPIResponse, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	assetID, err := api.getAssetID(asset)
	if err != nil {
		return nil, err
	}

	cancelAction := map[string]interface{}{
		"type": "cancel",
		"cancels": []map[string]interface{}{
			{
				"a": assetID,
				"o": oid,
			},
		},
	}

	signature, err := api.signer.SignAction(cancelAction, apiWalletAddress, nonce)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"action":       cancelAction,
		"nonce":        nonce,
		"signature":    signature,
		"vaultAddress": nil,
	}

	var response models.HyperliquidAPIResponse
	err = api.makeRequest(ctx, apiURL+"/exchange", reqBody, &response)
	return &response, err
}

// BatchOrders with IOC/GTC and ALO separation
func (api *HyperliquidAPI) BatchOrders(ctx context.Context, orders []*models.EnhancedOrderRequest, apiWalletAddress string) (*models.OrderResponse, error) {
	// Separate IOC/GTC and ALO orders
	var iocOrders, aloOrders []*models.EnhancedOrderRequest
	for _, order := range orders {
		if order.Tif == "Alo" {
			aloOrders = append(aloOrders, order)
		} else {
			iocOrders = append(iocOrders, order)
		}
	}

	// Process IOC/GTC batch first
	if len(iocOrders) > 0 {
		nonce := api.nonceManager.GetNextNonce(apiWalletAddress)
		response, err := api.batchOrders(ctx, iocOrders, apiWalletAddress, nonce, "na")
		if err != nil {
			return nil, fmt.Errorf("IOC/GTC batch failed: %w", err)
		}
		if response.Status != "success" {
			return response, fmt.Errorf("IOC/GTC batch failed with status: %s", response.Status)
		}
	}

	// Process ALO batch
	if len(aloOrders) > 0 {
		nonce := api.nonceManager.GetNextNonce(apiWalletAddress)
		response, err := api.batchOrders(ctx, aloOrders, apiWalletAddress, nonce, "alo")
		if err != nil {
			return nil, fmt.Errorf("ALO batch failed: %w", err)
		}
		if response.Status != "success" {
			return response, fmt.Errorf("ALO batch failed with status: %s", response.Status)
		}
	}

	return &models.OrderResponse{Status: "success"}, nil
}

func (api *HyperliquidAPI) batchOrders(ctx context.Context, orders []*models.EnhancedOrderRequest, apiWalletAddress string, nonce int64, grouping string) (*models.OrderResponse, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	var orderList []map[string]interface{}
	for _, order := range orders {
		assetID, err := api.getAssetID(order.Asset)
		if err != nil {
			return nil, err
		}

		orderData := map[string]interface{}{
			"a": assetID,
			"b": order.IsBuy,
			"p": api.formatPrice(order.Price),
			"s": api.formatSize(order.Size),
			"r": false,
			"t": api.getOrderTypeCode(order.OrderType, order.Tif),
		}

		if order.ClOid != nil {
			orderData["c"] = *order.ClOid
		} else {
			orderData["c"] = nil
		}

		orderList = append(orderList, orderData)
	}

	batchAction := map[string]interface{}{
		"type":     "order",
		"orders":   orderList,
		"grouping": grouping,
	}

	signature, err := api.signer.SignAction(batchAction, apiWalletAddress, nonce)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"action":       batchAction,
		"nonce":        nonce,
		"signature":    signature,
		"vaultAddress": nil,
	}

	var response models.OrderResponse
	err = api.makeRequest(ctx, apiURL+"/exchange", reqBody, &response)
	return &response, err
}

func (api *HyperliquidAPI) GetOrderStatus(ctx context.Context, userAddress string, oid int64) (map[string]interface{}, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	reqBody := map[string]interface{}{
		"type": "orderStatus",
		"user": userAddress,
		"oid":  oid,
	}

	var status map[string]interface{}
	err := api.makeRequest(ctx, apiURL+"/info", reqBody, &status)
	return status, err
}

func (api *HyperliquidAPI) makeRequest(ctx context.Context, url string, reqBody interface{}, response interface{}) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Debug().Str("url", url).RawJSON("response", body).Msg("API response")

	return json.Unmarshal(body, response)
}

// getAssetID returns the asset ID for a given asset name (perp or spot) with dynamic lookup
func (api *HyperliquidAPI) getAssetID(asset string) (int, error) {
	api.metaMutex.RLock()
	defer api.metaMutex.RUnlock()

	// Check perpetuals first
	if api.perpMeta != nil {
		for i, assetInfo := range api.perpMeta.Universe {
			if assetInfo.Name == asset {
				// Check if asset is delisted (if enhanced asset info is available)
				if enhancedInfo, ok := interface{}(assetInfo).(models.EnhancedAssetInfo); ok {
					if enhancedInfo.IsDelisted {
						return 0, fmt.Errorf("asset %s is delisted", asset)
					}
				}
				return i, nil
			}
		}
	}

	// Check spot markets
	if api.spotMeta != nil {
		for _, pair := range api.spotMeta.Universe {
			if pair.Name == asset {
				return 10000 + pair.Index, nil
			}
		}
	}

	// Check if asset is at open interest cap
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cappedAssets, err := api.GetPerpsAtOpenInterestCap(ctx)
	if err == nil {
		for _, cappedAsset := range cappedAssets {
			if cappedAsset == asset {
				return 0, fmt.Errorf("asset %s is at open interest cap", asset)
			}
		}
	}

	return 0, fmt.Errorf("unsupported asset: %s", asset)
}

func (api *HyperliquidAPI) formatPrice(price *float64) string {
	if price == nil {
		return ""
	}
	return strconv.FormatFloat(*price, 'f', -1, 64)
}

func (api *HyperliquidAPI) formatSize(size float64) string {
	return strconv.FormatFloat(size, 'f', -1, 64)
}

// getOrderTypeCode with enhanced TIF support
func (api *HyperliquidAPI) getOrderTypeCode(orderType string, tif string) map[string]interface{} {
	if tif == "" {
		switch orderType {
		case "market":
			tif = "Ioc"
		case "limit":
			tif = "Gtc"
		default:
			tif = "Gtc"
		}
	}

	return map[string]interface{}{
		"limit": map[string]interface{}{
			"tif": tif,
		},
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
