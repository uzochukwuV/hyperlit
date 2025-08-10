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
	"time"

	"github.com/rs/zerolog/log"
)

type HyperliquidAPI struct {
	config     *config.Config
	httpClient *http.Client
	signer     *Signer
}

func NewHyperliquidAPI(cfg *config.Config) (*HyperliquidAPI, error) {
	signer, err := NewSigner(cfg.APIWalletPrivateKeys["default"])
	if err != nil {
		return nil, err
	}

	return &HyperliquidAPI{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		signer: signer,
	}, nil
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

func (api *HyperliquidAPI) PlaceOrder(ctx context.Context, order *models.OrderRequest, apiWalletAddress string) (*models.HyperliquidAPIResponse, error) {
	apiURL := api.config.HyperliquidAPIURL
	if api.config.Environment == "testnet" {
		apiURL = api.config.HyperliquidTestnetURL
	}

	// Convert asset to proper format
	assetID, err := api.getAssetID(order.Asset)
	if err != nil {
		return nil, err
	}

	// Prepare order data
	orderAction := map[string]interface{}{
		"type":   "order",
		"orders": []map[string]interface{}{
			{
				"a":         assetID,
				"b":         order.IsBuy,
				"p":         api.formatPrice(order.Price),
				"s":         api.formatSize(order.Size),
				"r":         false, // reduceOnly
				"t":         api.getOrderTypeCode(order.OrderType),
				"c":         nil, // cloid (client order ID)
			},
		},
		"grouping": "na",
	}

	// Sign the action
	signature, err := api.signer.SignAction(orderAction, apiWalletAddress, order.Nonce)
	if err != nil {
		return nil, err
	}

	// Prepare request body
	reqBody := map[string]interface{}{
		"action":    orderAction,
		"nonce":     order.Nonce,
		"signature": signature,
		"vaultAddress": nil,
	}

	var response models.HyperliquidAPIResponse
	err = api.makeRequest(ctx, apiURL+"/exchange", reqBody, &response)
	return &response, err
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
				"a":   assetID,
				"o":   oid,
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

func (api *HyperliquidAPI) BatchOrders(ctx context.Context, orders []*models.OrderRequest, apiWalletAddress string, nonce int64) (*models.HyperliquidAPIResponse, error) {
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
			"t": api.getOrderTypeCode(order.OrderType),
			"c": nil,
		}
		orderList = append(orderList, orderData)
	}

	batchAction := map[string]interface{}{
		"type":     "order",
		"orders":   orderList,
		"grouping": "na",
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

	var response models.HyperliquidAPIResponse
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

func (api *HyperliquidAPI) getAssetID(asset string) (int, error) {
	// This is a simplified implementation
	// In production, you'd get this from the meta endpoint
	switch asset {
	case "BTC":
		return 0, nil
	case "ETH":
		return 1, nil
	case "SOL":
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported asset: %s", asset)
	}
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

func (api *HyperliquidAPI) getOrderTypeCode(orderType string) map[string]interface{} {
	switch orderType {
	case "market":
		return map[string]interface{}{"limit": map[string]interface{}{"tif": "Ioc"}}
	case "limit":
		return map[string]interface{}{"limit": map[string]interface{}{"tif": "Gtc"}}
	default:
		return map[string]interface{}{"limit": map[string]interface{}{"tif": "Gtc"}}
	}
}
