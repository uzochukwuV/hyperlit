-- Hyperliquid Copy Trading Database Schema
-- PostgreSQL initialization script

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "timescaledb" CASCADE;

-- Leaders table
CREATE TABLE IF NOT EXISTS leaders (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) UNIQUE NOT NULL,
    name VARCHAR(255),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    total_followers INTEGER DEFAULT 0,
    total_volume NUMERIC(20,8) DEFAULT 0,
    win_rate NUMERIC(5,4) DEFAULT 0,
    pnl_30d NUMERIC(20,8) DEFAULT 0,
    max_drawdown NUMERIC(10,4) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Followers table
CREATE TABLE IF NOT EXISTS followers (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    leader_address VARCHAR(42) NOT NULL REFERENCES leaders(address),
    api_wallet_address VARCHAR(42) NOT NULL,
    copy_percentage NUMERIC(5,4) NOT NULL CHECK (copy_percentage > 0 AND copy_percentage <= 100),
    max_position_size NUMERIC(20,8) NOT NULL CHECK (max_position_size > 0),
    stop_loss_percentage NUMERIC(5,4) CHECK (stop_loss_percentage > 0 AND stop_loss_percentage < 100),
    take_profit_percentage NUMERIC(5,4) CHECK (take_profit_percentage > 0),
    is_active BOOLEAN DEFAULT true,
    risk_settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, leader_address)
);

-- Trades table (hypertable for time-series optimization)
CREATE TABLE IF NOT EXISTS trades (
    id SERIAL PRIMARY KEY,
    leader_address VARCHAR(42) NOT NULL,
    follower_id INTEGER REFERENCES followers(id),
    asset VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    size NUMERIC(20,8) NOT NULL CHECK (size > 0),
    price NUMERIC(20,8) NOT NULL CHECK (price > 0),
    order_type VARCHAR(20) NOT NULL DEFAULT 'market',
    is_leader_trade BOOLEAN NOT NULL DEFAULT false,
    executed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    hyperliquid_tx_id VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'filled', 'cancelled', 'rejected', 'failed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Convert trades to hypertable for better time-series performance
SELECT create_hypertable('trades', 'executed_at', if_not_exists => TRUE);

-- Positions table
CREATE TABLE IF NOT EXISTS positions (
    id SERIAL PRIMARY KEY,
    user_address VARCHAR(42) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('long', 'short')),
    size NUMERIC(20,8) NOT NULL,
    entry_price NUMERIC(20,8) NOT NULL,
    current_price NUMERIC(20,8) NOT NULL,
    unrealized_pnl NUMERIC(20,8) NOT NULL DEFAULT 0,
    margin_used NUMERIC(20,8) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_address, asset)
);

-- Analytics cache table
CREATE TABLE IF NOT EXISTS analytics_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    result_type VARCHAR(100) NOT NULL,
    data JSONB NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Backtest results table
CREATE TABLE IF NOT EXISTS backtest_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    leader_address VARCHAR(42) NOT NULL,
    strategy_params JSONB NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    initial_capital NUMERIC(20,8) NOT NULL,
    final_capital NUMERIC(20,8),
    total_return_pct NUMERIC(10,4),
    max_drawdown_pct NUMERIC(10,4),
    sharpe_ratio NUMERIC(10,6),
    total_trades INTEGER,
    win_rate_pct NUMERIC(5,4),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    results JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB DEFAULT '{}',
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API keys table for secure storage
CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    key_name VARCHAR(100) NOT NULL,
    encrypted_private_key TEXT NOT NULL,
    wallet_address VARCHAR(42) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, key_name)
);

-- System metrics table
CREATE TABLE IF NOT EXISTS system_metrics (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(100) NOT NULL,
    metric_value NUMERIC(20,8) NOT NULL,
    metadata JSONB DEFAULT '{}',
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Convert system_metrics to hypertable
SELECT create_hypertable('system_metrics', 'recorded_at', if_not_exists => TRUE);

-- Indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_trades_leader_address ON trades(leader_address);
CREATE INDEX IF NOT EXISTS idx_trades_follower_id ON trades(follower_id);
CREATE INDEX IF NOT EXISTS idx_trades_executed_at ON trades(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_trades_asset ON trades(asset);
CREATE INDEX IF NOT EXISTS idx_trades_status ON trades(status);
CREATE INDEX IF NOT EXISTS idx_trades_leader_executed_at ON trades(leader_address, executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_followers_leader_address ON followers(leader_address);
CREATE INDEX IF NOT EXISTS idx_followers_user_id ON followers(user_id);
CREATE INDEX IF NOT EXISTS idx_followers_active ON followers(is_active) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_positions_user_address ON positions(user_address);
CREATE INDEX IF NOT EXISTS idx_positions_asset ON positions(asset);

CREATE INDEX IF NOT EXISTS idx_analytics_cache_type ON analytics_cache(result_type);
CREATE INDEX IF NOT EXISTS idx_analytics_cache_expires ON analytics_cache(expires_at);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;

CREATE INDEX IF NOT EXISTS idx_system_metrics_name ON system_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_system_metrics_recorded_at ON system_metrics(recorded_at DESC);

-- Views for common queries
CREATE OR REPLACE VIEW leader_performance_summary AS
SELECT 
    l.address,
    l.name,
    l.total_followers,
    l.total_volume,
    l.win_rate,
    l.pnl_30d,
    l.max_drawdown,
    COUNT(DISTINCT f.id) as active_followers,
    COALESCE(recent_trades.trade_count, 0) as trades_last_30d,
    COALESCE(recent_trades.volume_last_30d, 0) as volume_last_30d
FROM leaders l
LEFT JOIN followers f ON l.address = f.leader_address AND f.is_active = true
LEFT JOIN (
    SELECT 
        leader_address,
        COUNT(*) as trade_count,
        SUM(size * price) as volume_last_30d
    FROM trades 
    WHERE executed_at >= NOW() - INTERVAL '30 days' 
        AND is_leader_trade = true 
        AND status = 'filled'
    GROUP BY leader_address
) recent_trades ON l.address = recent_trades.leader_address
WHERE l.is_active = true
GROUP BY l.address, l.name, l.total_followers, l.total_volume, l.win_rate, 
         l.pnl_30d, l.max_drawdown, recent_trades.trade_count, recent_trades.volume_last_30d;

-- Follower performance view
CREATE OR REPLACE VIEW follower_performance_summary AS
SELECT 
    f.id,
    f.user_id,
    f.leader_address,
    f.copy_percentage,
    f.max_position_size,
    f.is_active,
    COUNT(t.id) as total_trades,
    COALESCE(SUM(CASE WHEN t.side = 'sell' THEN t.size * t.price ELSE -t.size * t.price END), 0) as total_pnl,
    COALESCE(SUM(CASE WHEN (CASE WHEN t.side = 'sell' THEN t.size * t.price ELSE -t.size * t.price END) > 0 THEN 1 ELSE 0 END), 0) as profitable_trades,
    CASE 
        WHEN COUNT(t.id) > 0 THEN 
            COALESCE(SUM(CASE WHEN (CASE WHEN t.side = 'sell' THEN t.size * t.price ELSE -t.size * t.price END) > 0 THEN 1 ELSE 0 END), 0)::float / COUNT(t.id)::float * 100
        ELSE 0 
    END as win_rate_pct
FROM followers f
LEFT JOIN trades t ON f.id = t.follower_id 
    AND t.is_leader_trade = false 
    AND t.status = 'filled'
    AND t.executed_at >= NOW() - INTERVAL '30 days'
GROUP BY f.id, f.user_id, f.leader_address, f.copy_percentage, f.max_position_size, f.is_active;

-- Functions for cleanup and maintenance
CREATE OR REPLACE FUNCTION cleanup_expired_analytics_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM analytics_cache WHERE expires_at <= NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to update leader statistics
CREATE OR REPLACE FUNCTION update_leader_stats(leader_addr VARCHAR(42))
RETURNS VOID AS $$
BEGIN
    UPDATE leaders SET
        total_followers = (
            SELECT COUNT(*) FROM followers 
            WHERE leader_address = leader_addr AND is_active = true
        ),
        total_volume = COALESCE((
            SELECT SUM(size * price) FROM trades 
            WHERE leader_address = leader_addr 
                AND is_leader_trade = true 
                AND status = 'filled'
                AND executed_at >= NOW() - INTERVAL '30 days'
        ), 0),
        updated_at = NOW()
    WHERE address = leader_addr;
END;
$$ LANGUAGE plpgsql;

-- Triggers for automatic updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_leaders_updated_at 
    BEFORE UPDATE ON leaders 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_followers_updated_at 
    BEFORE UPDATE ON followers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_positions_updated_at 
    BEFORE UPDATE ON positions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sample data for development (only if explicitly needed)
-- INSERT INTO leaders (address, name, description) VALUES 
-- ('0x742d35Cc6251B726C0532F58E7eF0C0F6b8b8E7f', 'Alpha Trader', 'Experienced BTC/ETH trader with 5+ years experience'),
-- ('0x8ba1f109551bD432803012645Hac136c36a4eb6', 'Momentum Master', 'Specializes in momentum trading across major altcoins'),
-- ('0x267be1C1D684F78cb4F6a176C4911b741E4Ffdc0', 'Risk Minimizer', 'Conservative trading approach with focus on capital preservation');

-- Continuous aggregates for time-series analysis
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_trade_summary
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', executed_at) AS day,
    leader_address,
    asset,
    COUNT(*) as trade_count,
    SUM(size * price) as volume,
    SUM(CASE WHEN side = 'sell' THEN size * price ELSE -size * price END) as net_pnl
FROM trades
WHERE is_leader_trade = true AND status = 'filled'
GROUP BY day, leader_address, asset
WITH NO DATA;

-- Refresh policy for continuous aggregates
SELECT add_continuous_aggregate_policy('daily_trade_summary',
    start_offset => INTERVAL '1 month',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- Data retention policy (keep detailed data for 1 year)
SELECT add_retention_policy('trades', INTERVAL '1 year');
SELECT add_retention_policy('system_metrics', INTERVAL '6 months');

-- === PERMISSIONLESS COPY TRADING TABLES ===

-- Permissionless followers table - allows copying any trader without registration
CREATE TABLE IF NOT EXISTS permissionless_followers (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    target_trader_address VARCHAR(42) NOT NULL,
    api_wallet_address VARCHAR(42) NOT NULL,
    copy_percentage NUMERIC(5,4) NOT NULL CHECK (copy_percentage > 0 AND copy_percentage <= 100),
    max_position_size NUMERIC(20,8) NOT NULL CHECK (max_position_size > 0),
    min_trade_size NUMERIC(20,8) NOT NULL DEFAULT 0,
    asset_whitelist TEXT[], -- Array of allowed assets
    asset_blacklist TEXT[], -- Array of blocked assets
    auto_discovery_enabled BOOLEAN DEFAULT false,
    copy_filters JSONB DEFAULT '{}', -- Complex filtering rules
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, target_trader_address)
);

-- Copy trades table - records each copy trading execution
CREATE TABLE IF NOT EXISTS copy_trades (
    id SERIAL PRIMARY KEY,
    original_trader_address VARCHAR(42) NOT NULL,
    follower_id INTEGER NOT NULL REFERENCES permissionless_followers(id),
    original_trade_hash VARCHAR(100) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    original_size VARCHAR(50) NOT NULL,
    copied_size VARCHAR(50) NOT NULL,
    original_price VARCHAR(50) NOT NULL,
    executed_price VARCHAR(50),
    slippage NUMERIC(10,6) DEFAULT 0,
    delay_ms BIGINT DEFAULT 0, -- Execution delay in milliseconds
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    executed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Convert copy_trades to hypertable for time-series optimization
SELECT create_hypertable('copy_trades', 'executed_at', if_not_exists => TRUE);

-- Trader discovery table - tracks discovered traders and their performance
CREATE TABLE IF NOT EXISTS trader_discovery (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) UNIQUE NOT NULL,
    first_discovered TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    total_volume NUMERIC(20,8) DEFAULT 0,
    trade_count INTEGER DEFAULT 0,
    win_rate NUMERIC(5,4) DEFAULT 0,
    profit_loss NUMERIC(20,8) DEFAULT 0,
    max_drawdown NUMERIC(10,4) DEFAULT 0,
    sharpe_ratio NUMERIC(10,6) DEFAULT 0,
    last_activity TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    follower_count INTEGER DEFAULT 0,
    asset_breakdown JSONB DEFAULT '{}', -- Asset allocation breakdown
    performance_grade VARCHAR(2) DEFAULT 'C', -- A, B, C, D, F
    risk_level VARCHAR(10) DEFAULT 'medium', -- low, medium, high
    trading_style VARCHAR(20) DEFAULT 'unknown', -- scalper, swing, position
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Trader recommendations table - AI-powered trader suggestions
CREATE TABLE IF NOT EXISTS trader_recommendations (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    recommended_trader VARCHAR(42) NOT NULL,
    recommendation_score NUMERIC(5,4) NOT NULL, -- 0-100 score
    recommendation_reason TEXT,
    risk_compatibility NUMERIC(5,4) DEFAULT 0,
    style_match NUMERIC(5,4) DEFAULT 0,
    performance_score NUMERIC(5,4) DEFAULT 0,
    recommended_allocation NUMERIC(5,4) DEFAULT 0,
    is_viewed BOOLEAN DEFAULT false,
    is_accepted BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Copy trading strategies table - different copying approaches
CREATE TABLE IF NOT EXISTS copy_trading_strategies (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    strategy_name VARCHAR(100) NOT NULL,
    strategy_type VARCHAR(20) NOT NULL, -- mirror, proportional, risk_adjusted
    target_traders TEXT[] NOT NULL, -- Array of trader addresses
    allocations JSONB DEFAULT '{}', -- Allocation percentages per trader
    rebalance_frequency VARCHAR(20) DEFAULT 'daily',
    max_total_risk NUMERIC(5,4) DEFAULT 100,
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    performance_stats JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Smart copy orders table - intelligent order execution
CREATE TABLE IF NOT EXISTS smart_copy_orders (
    id SERIAL PRIMARY KEY,
    follower_id INTEGER NOT NULL REFERENCES permissionless_followers(id),
    original_trade_hash VARCHAR(100) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    target_size NUMERIC(20,8) NOT NULL,
    execution_strategy VARCHAR(20) DEFAULT 'immediate', -- immediate, twap, smart
    max_slippage NUMERIC(10,6) DEFAULT 0.5,
    time_limit_seconds INTEGER DEFAULT 300,
    price_improvement NUMERIC(10,6) DEFAULT 0,
    partial_executions JSONB DEFAULT '[]',
    status VARCHAR(20) DEFAULT 'pending',
    total_executed NUMERIC(20,8) DEFAULT 0,
    average_price NUMERIC(20,8) DEFAULT 0,
    total_slippage NUMERIC(10,6) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Trading insights table - AI-driven analytics
CREATE TABLE IF NOT EXISTS copy_trading_insights (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    analysis_period VARCHAR(20) NOT NULL,
    total_copied_trades INTEGER DEFAULT 0,
    successful_trades INTEGER DEFAULT 0,
    total_return NUMERIC(10,4) DEFAULT 0,
    best_performing_leader VARCHAR(42),
    worst_performing_leader VARCHAR(42),
    optimal_allocations JSONB DEFAULT '{}',
    risk_adjusted_return NUMERIC(10,4) DEFAULT 0,
    recommended_actions TEXT[],
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for permissionless copy trading tables
CREATE INDEX IF NOT EXISTS idx_permissionless_followers_trader ON permissionless_followers(target_trader_address);
CREATE INDEX IF NOT EXISTS idx_permissionless_followers_user ON permissionless_followers(user_id);
CREATE INDEX IF NOT EXISTS idx_permissionless_followers_active ON permissionless_followers(is_active) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_copy_trades_trader ON copy_trades(original_trader_address);
CREATE INDEX IF NOT EXISTS idx_copy_trades_follower ON copy_trades(follower_id);
CREATE INDEX IF NOT EXISTS idx_copy_trades_executed_at ON copy_trades(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_copy_trades_status ON copy_trades(status);
CREATE INDEX IF NOT EXISTS idx_copy_trades_hash ON copy_trades(original_trade_hash);

CREATE INDEX IF NOT EXISTS idx_trader_discovery_address ON trader_discovery(address);
CREATE INDEX IF NOT EXISTS idx_trader_discovery_active ON trader_discovery(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_trader_discovery_performance ON trader_discovery(sharpe_ratio DESC, total_volume DESC);
CREATE INDEX IF NOT EXISTS idx_trader_discovery_grade ON trader_discovery(performance_grade);

CREATE INDEX IF NOT EXISTS idx_trader_recommendations_user ON trader_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_trader_recommendations_score ON trader_recommendations(recommendation_score DESC);
CREATE INDEX IF NOT EXISTS idx_trader_recommendations_unviewed ON trader_recommendations(user_id, is_viewed) WHERE is_viewed = false;

CREATE INDEX IF NOT EXISTS idx_copy_strategies_user ON copy_trading_strategies(user_id);
CREATE INDEX IF NOT EXISTS idx_copy_strategies_active ON copy_trading_strategies(is_active) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_smart_orders_follower ON smart_copy_orders(follower_id);
CREATE INDEX IF NOT EXISTS idx_smart_orders_status ON smart_copy_orders(status);
CREATE INDEX IF NOT EXISTS idx_smart_orders_created_at ON smart_copy_orders(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_insights_user ON copy_trading_insights(user_id);
CREATE INDEX IF NOT EXISTS idx_insights_generated_at ON copy_trading_insights(generated_at DESC);

-- Triggers for permissionless tables
CREATE TRIGGER update_permissionless_followers_updated_at 
    BEFORE UPDATE ON permissionless_followers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_trader_discovery_updated_at 
    BEFORE UPDATE ON trader_discovery 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_copy_strategies_updated_at 
    BEFORE UPDATE ON copy_trading_strategies 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Views for permissionless copy trading
CREATE OR REPLACE VIEW permissionless_trader_leaderboard AS
SELECT 
    td.address,
    td.total_volume,
    td.trade_count,
    td.win_rate,
    td.profit_loss,
    td.max_drawdown,
    td.sharpe_ratio,
    td.last_activity,
    td.follower_count,
    td.performance_grade,
    td.risk_level,
    td.trading_style,
    COALESCE(recent_activity.recent_trades, 0) as trades_last_7d,
    COALESCE(recent_activity.recent_volume, 0) as volume_last_7d
FROM trader_discovery td
LEFT JOIN (
    SELECT 
        original_trader_address,
        COUNT(*) as recent_trades,
        SUM(CAST(original_size AS NUMERIC) * CAST(original_price AS NUMERIC)) as recent_volume
    FROM copy_trades 
    WHERE executed_at >= NOW() - INTERVAL '7 days'
    GROUP BY original_trader_address
) recent_activity ON td.address = recent_activity.original_trader_address
WHERE td.is_active = true
ORDER BY td.sharpe_ratio DESC, td.total_volume DESC;

-- Continuous aggregates for copy trading analytics
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_copy_trade_summary
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', executed_at) AS day,
    original_trader_address,
    COUNT(*) as copy_trade_count,
    COUNT(DISTINCT follower_id) as unique_followers,
    SUM(CAST(copied_size AS NUMERIC) * CAST(executed_price AS NUMERIC)) as total_copy_volume,
    AVG(slippage) as avg_slippage,
    AVG(delay_ms) as avg_delay_ms,
    COUNT(CASE WHEN status = 'executed' THEN 1 END) as successful_copies,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_copies
FROM copy_trades
GROUP BY day, original_trader_address
WITH NO DATA;

-- Refresh policy for copy trading aggregates
SELECT add_continuous_aggregate_policy('daily_copy_trade_summary',
    start_offset => INTERVAL '1 month',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- Data retention for copy trading tables
SELECT add_retention_policy('copy_trades', INTERVAL '2 years');

-- Functions for permissionless copy trading
CREATE OR REPLACE FUNCTION update_trader_follower_count(trader_addr VARCHAR(42))
RETURNS VOID AS $$
BEGIN
    UPDATE trader_discovery SET
        follower_count = (
            SELECT COUNT(*) FROM permissionless_followers 
            WHERE target_trader_address = trader_addr AND is_active = true
        ),
        updated_at = NOW()
    WHERE address = trader_addr;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate trader performance metrics
CREATE OR REPLACE FUNCTION calculate_trader_performance(trader_addr VARCHAR(42), days_back INTEGER DEFAULT 30)
RETURNS TABLE (
    total_trades INTEGER,
    win_rate NUMERIC,
    total_volume NUMERIC,
    profit_loss NUMERIC,
    sharpe_ratio NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*)::INTEGER as total_trades,
        COALESCE(
            (COUNT(CASE WHEN CAST(copied_size AS NUMERIC) * CAST(executed_price AS NUMERIC) > 0 THEN 1 END)::NUMERIC / 
             NULLIF(COUNT(*)::NUMERIC, 0)) * 100, 0
        ) as win_rate,
        COALESCE(SUM(CAST(copied_size AS NUMERIC) * CAST(executed_price AS NUMERIC)), 0) as total_volume,
        COALESCE(SUM(
            CASE WHEN side = 'B' THEN -CAST(copied_size AS NUMERIC) * CAST(executed_price AS NUMERIC)
                 ELSE CAST(copied_size AS NUMERIC) * CAST(executed_price AS NUMERIC)
            END
        ), 0) as profit_loss,
        COALESCE(
            CASE WHEN STDDEV(CAST(copied_size AS NUMERIC) * CAST(executed_price AS NUMERIC)) > 0 
                 THEN AVG(CAST(copied_size AS NUMERIC) * CAST(executed_price AS NUMERIC)) / 
                      STDDEV(CAST(copied_size AS NUMERIC) * CAST(executed_price AS NUMERIC))
                 ELSE 0 
            END, 0
        ) as sharpe_ratio
    FROM copy_trades
    WHERE original_trader_address = trader_addr 
        AND executed_at >= NOW() - (days_back || ' days')::INTERVAL
        AND status = 'executed';
END;
$$ LANGUAGE plpgsql;

-- Security: Row Level Security (if needed)
-- ALTER TABLE followers ENABLE ROW LEVEL SECURITY;
-- CREATE POLICY followers_policy ON followers FOR ALL TO authenticated_users USING (user_id = current_setting('app.current_user'));

COMMIT;
