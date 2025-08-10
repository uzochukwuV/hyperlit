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
