package database

import (
	"context"
	"hyperliquid-copy-trading/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type PostgresDB struct {
	pool *pgxpool.Pool
}

func NewPostgresDB(databaseURL string) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	log.Info().Msg("Connected to PostgreSQL database")

	return &PostgresDB{pool: pool}, nil
}

func (db *PostgresDB) Close() {
	db.pool.Close()
}

func (db *PostgresDB) CreateFollower(ctx context.Context, follower *models.Follower) error {
	query := `
		INSERT INTO followers (user_id, leader_address, api_wallet_address, copy_percentage, 
			max_position_size, stop_loss_percentage, take_profit_percentage, is_active, risk_settings)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := db.pool.QueryRow(ctx, query,
		follower.UserID,
		follower.LeaderAddress,
		follower.APIWalletAddress,
		follower.CopyPercentage,
		follower.MaxPositionSize,
		follower.StopLossPercentage,
		follower.TakeProfitPercentage,
		follower.IsActive,
		follower.RiskSettings,
	).Scan(&follower.ID, &follower.CreatedAt, &follower.UpdatedAt)

	return err
}

func (db *PostgresDB) GetFollowers(ctx context.Context) ([]models.Follower, error) {
	query := `
		SELECT id, user_id, leader_address, api_wallet_address, copy_percentage,
			max_position_size, stop_loss_percentage, take_profit_percentage, 
			is_active, risk_settings, created_at, updated_at
		FROM followers
		ORDER BY created_at DESC`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []models.Follower
	for rows.Next() {
		var f models.Follower
		err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.LeaderAddress,
			&f.APIWalletAddress,
			&f.CopyPercentage,
			&f.MaxPositionSize,
			&f.StopLossPercentage,
			&f.TakeProfitPercentage,
			&f.IsActive,
			&f.RiskSettings,
			&f.CreatedAt,
			&f.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		followers = append(followers, f)
	}

	return followers, nil
}

func (db *PostgresDB) GetFollowersByLeader(ctx context.Context, leaderAddress string) ([]models.Follower, error) {
	query := `
		SELECT id, user_id, leader_address, api_wallet_address, copy_percentage,
			max_position_size, stop_loss_percentage, take_profit_percentage, 
			is_active, risk_settings, created_at, updated_at
		FROM followers
		WHERE leader_address = $1 AND is_active = true`

	rows, err := db.pool.Query(ctx, query, leaderAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []models.Follower
	for rows.Next() {
		var f models.Follower
		err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.LeaderAddress,
			&f.APIWalletAddress,
			&f.CopyPercentage,
			&f.MaxPositionSize,
			&f.StopLossPercentage,
			&f.TakeProfitPercentage,
			&f.IsActive,
			&f.RiskSettings,
			&f.CreatedAt,
			&f.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		followers = append(followers, f)
	}

	return followers, nil
}

func (db *PostgresDB) UpdateFollower(ctx context.Context, follower *models.Follower) error {
	query := `
		UPDATE followers 
		SET copy_percentage = $1, max_position_size = $2, stop_loss_percentage = $3,
			take_profit_percentage = $4, is_active = $5, risk_settings = $6, updated_at = NOW()
		WHERE id = $7`

	_, err := db.pool.Exec(ctx, query,
		follower.CopyPercentage,
		follower.MaxPositionSize,
		follower.StopLossPercentage,
		follower.TakeProfitPercentage,
		follower.IsActive,
		follower.RiskSettings,
		follower.ID,
	)

	return err
}

func (db *PostgresDB) DeleteFollower(ctx context.Context, id int) error {
	query := `DELETE FROM followers WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, id)
	return err
}

func (db *PostgresDB) CreateTrade(ctx context.Context, trade *models.Trade) error {
	query := `
		INSERT INTO trades (leader_address, follower_id, asset, side, size, price, 
			order_type, is_leader_trade, executed_at, hyperliquid_tx_id, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`

	err := db.pool.QueryRow(ctx, query,
		trade.LeaderAddress,
		trade.FollowerID,
		trade.Asset,
		trade.Side,
		trade.Size,
		trade.Price,
		trade.OrderType,
		trade.IsLeaderTrade,
		trade.ExecutedAt,
		trade.HyperliquidTxID,
		trade.Status,
	).Scan(&trade.ID, &trade.CreatedAt)

	return err
}

func (db *PostgresDB) GetTrades(ctx context.Context, limit, offset int) ([]models.Trade, error) {
	query := `
		SELECT id, leader_address, follower_id, asset, side, size, price, 
			order_type, is_leader_trade, executed_at, hyperliquid_tx_id, status, created_at
		FROM trades
		ORDER BY executed_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		err := rows.Scan(
			&t.ID,
			&t.LeaderAddress,
			&t.FollowerID,
			&t.Asset,
			&t.Side,
			&t.Size,
			&t.Price,
			&t.OrderType,
			&t.IsLeaderTrade,
			&t.ExecutedAt,
			&t.HyperliquidTxID,
			&t.Status,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}

	return trades, nil
}

func (db *PostgresDB) GetTradesByFollower(ctx context.Context, followerID int) ([]models.Trade, error) {
	query := `
		SELECT id, leader_address, follower_id, asset, side, size, price, 
			order_type, is_leader_trade, executed_at, hyperliquid_tx_id, status, created_at
		FROM trades
		WHERE follower_id = $1
		ORDER BY executed_at DESC`

	rows, err := db.pool.Query(ctx, query, followerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		err := rows.Scan(
			&t.ID,
			&t.LeaderAddress,
			&t.FollowerID,
			&t.Asset,
			&t.Side,
			&t.Size,
			&t.Price,
			&t.OrderType,
			&t.IsLeaderTrade,
			&t.ExecutedAt,
			&t.HyperliquidTxID,
			&t.Status,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}

	return trades, nil
}

func (db *PostgresDB) UpsertPosition(ctx context.Context, position *models.Position) error {
	query := `
		INSERT INTO positions (user_address, asset, side, size, entry_price, current_price, 
			unrealized_pnl, margin_used, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (user_address, asset)
		DO UPDATE SET
			side = EXCLUDED.side,
			size = EXCLUDED.size,
			entry_price = EXCLUDED.entry_price,
			current_price = EXCLUDED.current_price,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			margin_used = EXCLUDED.margin_used,
			updated_at = NOW()
		RETURNING id`

	err := db.pool.QueryRow(ctx, query,
		position.UserAddress,
		position.Asset,
		position.Side,
		position.Size,
		position.EntryPrice,
		position.CurrentPrice,
		position.UnrealizedPnL,
		position.MarginUsed,
	).Scan(&position.ID)

	return err
}

func (db *PostgresDB) GetPositions(ctx context.Context, userAddress string) ([]models.Position, error) {
	query := `
		SELECT id, user_address, asset, side, size, entry_price, current_price,
			unrealized_pnl, margin_used, updated_at
		FROM positions
		WHERE user_address = $1`

	rows, err := db.pool.Query(ctx, query, userAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []models.Position
	for rows.Next() {
		var p models.Position
		err := rows.Scan(
			&p.ID,
			&p.UserAddress,
			&p.Asset,
			&p.Side,
			&p.Size,
			&p.EntryPrice,
			&p.CurrentPrice,
			&p.UnrealizedPnL,
			&p.MarginUsed,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}

	return positions, nil
}

// === PERMISSIONLESS COPY TRADING DATABASE METHODS ===

func (db *PostgresDB) CreatePermissionlessFollower(ctx context.Context, follower *models.PermissionlessFollower) error {
	query := `
		INSERT INTO permissionless_followers (user_id, target_trader_address, api_wallet_address, 
			copy_percentage, max_position_size, min_trade_size, asset_whitelist, asset_blacklist,
			auto_discovery_enabled, copy_filters, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	err := db.pool.QueryRow(ctx, query,
		follower.UserID,
		follower.TargetTraderAddress,
		follower.APIWalletAddress,
		follower.CopyPercentage,
		follower.MaxPositionSize,
		follower.MinTradeSize,
		follower.AssetWhitelist,
		follower.AssetBlacklist,
		follower.AutoDiscoveryEnabled,
		follower.CopyFilters,
		follower.IsActive,
	).Scan(&follower.ID, &follower.CreatedAt, &follower.UpdatedAt)

	return err
}

func (db *PostgresDB) GetPermissionlessFollowersByTrader(ctx context.Context, traderAddress string) ([]*models.PermissionlessFollower, error) {
	query := `
		SELECT id, user_id, target_trader_address, api_wallet_address, copy_percentage,
			max_position_size, min_trade_size, asset_whitelist, asset_blacklist,
			auto_discovery_enabled, copy_filters, is_active, created_at, updated_at
		FROM permissionless_followers
		WHERE target_trader_address = $1 AND is_active = true`

	rows, err := db.pool.Query(ctx, query, traderAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []*models.PermissionlessFollower
	for rows.Next() {
		var f models.PermissionlessFollower
		err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.TargetTraderAddress,
			&f.APIWalletAddress,
			&f.CopyPercentage,
			&f.MaxPositionSize,
			&f.MinTradeSize,
			&f.AssetWhitelist,
			&f.AssetBlacklist,
			&f.AutoDiscoveryEnabled,
			&f.CopyFilters,
			&f.IsActive,
			&f.CreatedAt,
			&f.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		followers = append(followers, &f)
	}

	return followers, nil
}

func (db *PostgresDB) CreateCopyTrade(ctx context.Context, copyTrade *models.CopyTrade) error {
	query := `
		INSERT INTO copy_trades (original_trader_address, follower_id, original_trade_hash,
			asset, side, original_size, copied_size, original_price, executed_price,
			slippage, delay_ms, status, error_message, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at`

	err := db.pool.QueryRow(ctx, query,
		copyTrade.OriginalTraderAddress,
		copyTrade.FollowerID,
		copyTrade.OriginalTradeHash,
		copyTrade.Asset,
		copyTrade.Side,
		copyTrade.OriginalSize,
		copyTrade.CopiedSize,
		copyTrade.OriginalPrice,
		copyTrade.ExecutedPrice,
		copyTrade.Slippage,
		copyTrade.DelayMs,
		copyTrade.Status,
		copyTrade.ErrorMessage,
		copyTrade.ExecutedAt,
	).Scan(&copyTrade.ID, &copyTrade.CreatedAt)

	return err
}

func (db *PostgresDB) GetCopyTradesByFollower(ctx context.Context, followerID int) ([]*models.CopyTrade, error) {
	query := `
		SELECT id, original_trader_address, follower_id, original_trade_hash, asset, side,
			original_size, copied_size, original_price, executed_price, slippage, delay_ms,
			status, error_message, executed_at, created_at
		FROM copy_trades
		WHERE follower_id = $1
		ORDER BY executed_at DESC`

	rows, err := db.pool.Query(ctx, query, followerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var copyTrades []*models.CopyTrade
	for rows.Next() {
		var ct models.CopyTrade
		err := rows.Scan(
			&ct.ID,
			&ct.OriginalTraderAddress,
			&ct.FollowerID,
			&ct.OriginalTradeHash,
			&ct.Asset,
			&ct.Side,
			&ct.OriginalSize,
			&ct.CopiedSize,
			&ct.OriginalPrice,
			&ct.ExecutedPrice,
			&ct.Slippage,
			&ct.DelayMs,
			&ct.Status,
			&ct.ErrorMessage,
			&ct.ExecutedAt,
			&ct.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		copyTrades = append(copyTrades, &ct)
	}

	return copyTrades, nil
}

func (db *PostgresDB) CreateTraderDiscovery(ctx context.Context, discovery *models.TraderDiscovery) error {
	query := `
		INSERT INTO trader_discovery (address, first_discovered, total_volume, trade_count,
			win_rate, profit_loss, max_drawdown, sharpe_ratio, last_activity, is_active,
			follower_count, asset_breakdown, performance_grade, risk_level, trading_style)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (address) DO UPDATE SET
			total_volume = EXCLUDED.total_volume,
			trade_count = EXCLUDED.trade_count,
			win_rate = EXCLUDED.win_rate,
			profit_loss = EXCLUDED.profit_loss,
			max_drawdown = EXCLUDED.max_drawdown,
			sharpe_ratio = EXCLUDED.sharpe_ratio,
			last_activity = EXCLUDED.last_activity,
			is_active = EXCLUDED.is_active,
			follower_count = EXCLUDED.follower_count,
			asset_breakdown = EXCLUDED.asset_breakdown,
			performance_grade = EXCLUDED.performance_grade,
			risk_level = EXCLUDED.risk_level,
			trading_style = EXCLUDED.trading_style,
			updated_at = NOW()
		RETURNING id, updated_at`

	err := db.pool.QueryRow(ctx, query,
		discovery.Address,
		discovery.FirstDiscovered,
		discovery.TotalVolume,
		discovery.TradeCount,
		discovery.WinRate,
		discovery.ProfitLoss,
		discovery.MaxDrawdown,
		discovery.SharpeRatio,
		discovery.LastActivity,
		discovery.IsActive,
		discovery.FollowerCount,
		discovery.AssetBreakdown,
		discovery.PerformanceGrade,
		discovery.RiskLevel,
		discovery.TradingStyle,
	).Scan(&discovery.ID, &discovery.UpdatedAt)

	return err
}

func (db *PostgresDB) GetTopTraders(ctx context.Context, limit int) ([]*models.TraderDiscovery, error) {
	query := `
		SELECT id, address, first_discovered, total_volume, trade_count, win_rate,
			profit_loss, max_drawdown, sharpe_ratio, last_activity, is_active,
			follower_count, asset_breakdown, performance_grade, risk_level, trading_style, updated_at
		FROM trader_discovery
		WHERE is_active = true AND trade_count > 10
		ORDER BY sharpe_ratio DESC, total_volume DESC
		LIMIT $1`

	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var traders []*models.TraderDiscovery
	for rows.Next() {
		var td models.TraderDiscovery
		err := rows.Scan(
			&td.ID,
			&td.Address,
			&td.FirstDiscovered,
			&td.TotalVolume,
			&td.TradeCount,
			&td.WinRate,
			&td.ProfitLoss,
			&td.MaxDrawdown,
			&td.SharpeRatio,
			&td.LastActivity,
			&td.IsActive,
			&td.FollowerCount,
			&td.AssetBreakdown,
			&td.PerformanceGrade,
			&td.RiskLevel,
			&td.TradingStyle,
			&td.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		traders = append(traders, &td)
	}

	return traders, nil
}
