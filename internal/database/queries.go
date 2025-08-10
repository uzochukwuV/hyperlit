package database

import (
	"context"
	"hyperliquid-copy-trading/internal/models"
	"time"
)

// GetLeaderPerformance calculates performance metrics for a leader
func (db *PostgresDB) GetLeaderPerformance(ctx context.Context, leaderAddress string, days int) (*models.PnLAnalytics, error) {
	query := `
		WITH trade_pnl AS (
			SELECT 
				t.id,
				t.side,
				t.size,
				t.price,
				t.executed_at,
				CASE 
					WHEN t.side = 'sell' THEN t.size * t.price
					ELSE -t.size * t.price
				END as pnl_contribution
			FROM trades t
			WHERE t.leader_address = $1 
				AND t.is_leader_trade = true
				AND t.executed_at >= NOW() - INTERVAL '%d days'
				AND t.status = 'filled'
			ORDER BY t.executed_at
		),
		daily_pnl AS (
			SELECT 
				DATE(executed_at) as trade_date,
				SUM(pnl_contribution) as daily_pnl
			FROM trade_pnl
			GROUP BY DATE(executed_at)
			ORDER BY trade_date
		),
		performance_metrics AS (
			SELECT 
				COUNT(*) as total_trades,
				SUM(CASE WHEN pnl_contribution > 0 THEN 1 ELSE 0 END) as profitable_trades,
				SUM(pnl_contribution) as total_pnl,
				AVG(pnl_contribution) as avg_pnl,
				STDDEV(pnl_contribution) as pnl_stddev
			FROM trade_pnl
		)
		SELECT 
			pm.total_trades,
			pm.profitable_trades,
			pm.total_pnl,
			CASE 
				WHEN pm.total_trades > 0 THEN pm.profitable_trades::float / pm.total_trades::float
				ELSE 0
			END as win_rate,
			CASE 
				WHEN pm.pnl_stddev > 0 AND pm.pnl_stddev IS NOT NULL THEN pm.avg_pnl / pm.pnl_stddev
				ELSE 0
			END as sharpe_ratio,
			COALESCE(array_agg(dp.daily_pnl ORDER BY dp.trade_date), ARRAY[]::numeric[]) as daily_pnl_array
		FROM performance_metrics pm
		LEFT JOIN daily_pnl dp ON true
		GROUP BY pm.total_trades, pm.profitable_trades, pm.total_pnl, pm.avg_pnl, pm.pnl_stddev`

	var analytics models.PnLAnalytics
	var dailyPnLArray []float64

	row := db.pool.QueryRow(ctx, query, leaderAddress, days)
	
	err := row.Scan(
		&analytics.TotalTrades,
		&analytics.ProfitableTrades,
		&analytics.TotalPnL,
		&analytics.WinRate,
		&analytics.SharpeRatio,
		&dailyPnLArray,
	)
	
	if err != nil {
		return nil, err
	}

	analytics.DailyPnL = dailyPnLArray

	// Calculate max drawdown
	maxDrawdown, err := db.calculateMaxDrawdown(ctx, leaderAddress, days)
	if err == nil {
		analytics.MaxDrawdown = maxDrawdown
	}

	return &analytics, nil
}

func (db *PostgresDB) calculateMaxDrawdown(ctx context.Context, leaderAddress string, days int) (float64, error) {
	query := `
		WITH cumulative_pnl AS (
			SELECT 
				executed_at,
				SUM(CASE 
					WHEN side = 'sell' THEN size * price
					ELSE -size * price
				END) OVER (ORDER BY executed_at) as running_pnl
			FROM trades
			WHERE leader_address = $1 
				AND is_leader_trade = true
				AND executed_at >= NOW() - INTERVAL '%d days'
				AND status = 'filled'
			ORDER BY executed_at
		),
		running_max AS (
			SELECT 
				executed_at,
				running_pnl,
				MAX(running_pnl) OVER (ORDER BY executed_at ROWS UNBOUNDED PRECEDING) as running_max_pnl
			FROM cumulative_pnl
		)
		SELECT 
			COALESCE(MIN(running_pnl - running_max_pnl), 0) as max_drawdown
		FROM running_max`

	var maxDrawdown float64
	err := db.pool.QueryRow(ctx, query, leaderAddress, days).Scan(&maxDrawdown)
	return maxDrawdown, err
}

// GetActiveLeaders returns all leaders with active followers
func (db *PostgresDB) GetActiveLeaders(ctx context.Context) ([]models.Leader, error) {
	query := `
		SELECT DISTINCT 
			f.leader_address,
			COUNT(f.id) as follower_count,
			COALESCE(SUM(CASE 
				WHEN t.side = 'sell' THEN t.size * t.price
				ELSE -t.size * t.price
			END), 0) as total_volume
		FROM followers f
		LEFT JOIN trades t ON f.leader_address = t.leader_address 
			AND t.is_leader_trade = true
			AND t.executed_at >= NOW() - INTERVAL '30 days'
		WHERE f.is_active = true
		GROUP BY f.leader_address
		ORDER BY follower_count DESC`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaders []models.Leader
	for rows.Next() {
		var leader models.Leader
		err := rows.Scan(
			&leader.Address,
			&leader.TotalFollowers,
			&leader.TotalVolume,
		)
		if err != nil {
			return nil, err
		}

		// Get performance metrics
		analytics, err := db.GetLeaderPerformance(ctx, leader.Address, 30)
		if err == nil {
			leader.WinRate = analytics.WinRate
			leader.PnL30d = analytics.TotalPnL
			leader.MaxDrawdown = analytics.MaxDrawdown
		}

		leader.IsActive = true
		leader.UpdatedAt = time.Now()
		leaders = append(leaders, leader)
	}

	return leaders, nil
}

// GetFollowerPnL calculates PnL for a specific follower
func (db *PostgresDB) GetFollowerPnL(ctx context.Context, followerID int, days int) (*models.PnLAnalytics, error) {
	query := `
		WITH trade_pnl AS (
			SELECT 
				t.id,
				t.side,
				t.size,
				t.price,
				t.executed_at,
				CASE 
					WHEN t.side = 'sell' THEN t.size * t.price
					ELSE -t.size * t.price
				END as pnl_contribution
			FROM trades t
			WHERE t.follower_id = $1 
				AND t.is_leader_trade = false
				AND t.executed_at >= NOW() - INTERVAL '%d days'
				AND t.status = 'filled'
			ORDER BY t.executed_at
		),
		daily_pnl AS (
			SELECT 
				DATE(executed_at) as trade_date,
				SUM(pnl_contribution) as daily_pnl
			FROM trade_pnl
			GROUP BY DATE(executed_at)
			ORDER BY trade_date
		),
		performance_metrics AS (
			SELECT 
				COUNT(*) as total_trades,
				SUM(CASE WHEN pnl_contribution > 0 THEN 1 ELSE 0 END) as profitable_trades,
				SUM(pnl_contribution) as total_pnl,
				AVG(pnl_contribution) as avg_pnl,
				STDDEV(pnl_contribution) as pnl_stddev
			FROM trade_pnl
		)
		SELECT 
			pm.total_trades,
			pm.profitable_trades,
			pm.total_pnl,
			CASE 
				WHEN pm.total_trades > 0 THEN pm.profitable_trades::float / pm.total_trades::float
				ELSE 0
			END as win_rate,
			CASE 
				WHEN pm.pnl_stddev > 0 AND pm.pnl_stddev IS NOT NULL THEN pm.avg_pnl / pm.pnl_stddev
				ELSE 0
			END as sharpe_ratio,
			COALESCE(array_agg(dp.daily_pnl ORDER BY dp.trade_date), ARRAY[]::numeric[]) as daily_pnl_array
		FROM performance_metrics pm
		LEFT JOIN daily_pnl dp ON true
		GROUP BY pm.total_trades, pm.profitable_trades, pm.total_pnl, pm.avg_pnl, pm.pnl_stddev`

	var analytics models.PnLAnalytics
	var dailyPnLArray []float64

	row := db.pool.QueryRow(ctx, query, followerID, days)
	
	err := row.Scan(
		&analytics.TotalTrades,
		&analytics.ProfitableTrades,
		&analytics.TotalPnL,
		&analytics.WinRate,
		&analytics.SharpeRatio,
		&dailyPnLArray,
	)
	
	if err != nil {
		return nil, err
	}

	analytics.DailyPnL = dailyPnLArray

	// Calculate max drawdown for follower
	maxDrawdown, err := db.calculateFollowerMaxDrawdown(ctx, followerID, days)
	if err == nil {
		analytics.MaxDrawdown = maxDrawdown
	}

	return &analytics, nil
}

func (db *PostgresDB) calculateFollowerMaxDrawdown(ctx context.Context, followerID int, days int) (float64, error) {
	query := `
		WITH cumulative_pnl AS (
			SELECT 
				executed_at,
				SUM(CASE 
					WHEN side = 'sell' THEN size * price
					ELSE -size * price
				END) OVER (ORDER BY executed_at) as running_pnl
			FROM trades
			WHERE follower_id = $1 
				AND is_leader_trade = false
				AND executed_at >= NOW() - INTERVAL '%d days'
				AND status = 'filled'
			ORDER BY executed_at
		),
		running_max AS (
			SELECT 
				executed_at,
				running_pnl,
				MAX(running_pnl) OVER (ORDER BY executed_at ROWS UNBOUNDED PRECEDING) as running_max_pnl
			FROM cumulative_pnl
		)
		SELECT 
			COALESCE(MIN(running_pnl - running_max_pnl), 0) as max_drawdown
		FROM running_max`

	var maxDrawdown float64
	err := db.pool.QueryRow(ctx, query, followerID, days).Scan(&maxDrawdown)
	return maxDrawdown, err
}
