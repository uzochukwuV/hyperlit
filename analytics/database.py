import os
import asyncpg
import asyncio
from typing import Optional, List, Dict, Any
from datetime import datetime, timedelta
import logging
from contextlib import asynccontextmanager

logger = logging.getLogger(__name__)

class AnalyticsDatabase:
    def __init__(self):
        self.pool: Optional[asyncpg.Pool] = None
        self.database_url = os.getenv(
            "DATABASE_URL",
            "postgresql://postgres:password@localhost:5432/copytrading"
        )

    async def init_pool(self):
        """Initialize the database connection pool"""
        try:
            self.pool = await asyncpg.create_pool(
                self.database_url,
                min_size=5,
                max_size=20,
                command_timeout=60,
                server_settings={
                    'jit': 'off',  # Disable JIT compilation for better performance on small queries
                    'application_name': 'hyperliquid_copy_trading_analytics'
                }
            )
            logger.info("Database connection pool initialized")
            
            # Test the connection
            async with self.pool.acquire() as conn:
                await conn.fetchval("SELECT 1")
            
        except Exception as e:
            logger.error(f"Failed to initialize database pool: {e}")
            raise

    async def close_pool(self):
        """Close the database connection pool"""
        if self.pool:
            await self.pool.close()
            logger.info("Database connection pool closed")

    @asynccontextmanager
    async def get_connection(self):
        """Get a database connection from the pool"""
        if not self.pool:
            raise RuntimeError("Database pool not initialized")
        
        async with self.pool.acquire() as connection:
            yield connection

    async def get_leader_trades(
        self,
        leader_address: str,
        days: int = 30,
        limit: Optional[int] = None
    ) -> List[Dict[str, Any]]:
        """Get trades for a specific leader"""
        try:
            async with self.get_connection() as conn:
                query = """
                    SELECT 
                        id, leader_address, asset, side, size, price,
                        order_type, executed_at, hyperliquid_tx_id, status,
                        created_at
                    FROM trades
                    WHERE leader_address = $1 
                        AND is_leader_trade = true
                        AND executed_at >= NOW() - INTERVAL '%s days'
                        AND status = 'filled'
                    ORDER BY executed_at DESC
                """
                
                if limit:
                    query += f" LIMIT {limit}"
                
                rows = await conn.fetch(query, leader_address, days)
                return [dict(row) for row in rows]
                
        except Exception as e:
            logger.error(f"Error fetching leader trades: {e}")
            return []

    async def get_follower_trades(
        self,
        follower_id: int,
        days: int = 30
    ) -> List[Dict[str, Any]]:
        """Get trades for a specific follower"""
        try:
            async with self.get_connection() as conn:
                query = """
                    SELECT 
                        t.id, t.leader_address, t.asset, t.side, t.size, t.price,
                        t.order_type, t.executed_at, t.hyperliquid_tx_id, t.status,
                        t.created_at, f.copy_percentage, f.max_position_size
                    FROM trades t
                    JOIN followers f ON t.follower_id = f.id
                    WHERE t.follower_id = $1 
                        AND t.is_leader_trade = false
                        AND t.executed_at >= NOW() - INTERVAL '%s days'
                        AND t.status = 'filled'
                    ORDER BY t.executed_at DESC
                """
                
                rows = await conn.fetch(query, follower_id, days)
                return [dict(row) for row in rows]
                
        except Exception as e:
            logger.error(f"Error fetching follower trades: {e}")
            return []

    async def get_leader_performance_metrics(
        self,
        leader_address: str,
        days: int = 30
    ) -> Optional[Dict[str, Any]]:
        """Calculate comprehensive performance metrics for a leader"""
        try:
            async with self.get_connection() as conn:
                query = """
                    WITH daily_pnl AS (
                        SELECT 
                            DATE(executed_at) as trade_date,
                            SUM(CASE 
                                WHEN side = 'sell' THEN size * price
                                ELSE -size * price
                            END) as daily_pnl,
                            COUNT(*) as daily_trades,
                            SUM(size * price) as daily_volume
                        FROM trades
                        WHERE leader_address = $1 
                            AND is_leader_trade = true
                            AND executed_at >= NOW() - INTERVAL '%s days'
                            AND status = 'filled'
                        GROUP BY DATE(executed_at)
                        ORDER BY trade_date
                    ),
                    cumulative_metrics AS (
                        SELECT 
                            trade_date,
                            daily_pnl,
                            daily_trades,
                            daily_volume,
                            SUM(daily_pnl) OVER (ORDER BY trade_date) as cumulative_pnl,
                            MAX(SUM(daily_pnl) OVER (ORDER BY trade_date)) OVER (ORDER BY trade_date ROWS UNBOUNDED PRECEDING) as running_max
                        FROM daily_pnl
                    ),
                    performance_stats AS (
                        SELECT 
                            COUNT(*) as trading_days,
                            SUM(daily_trades) as total_trades,
                            SUM(daily_volume) as total_volume,
                            SUM(daily_pnl) as total_pnl,
                            AVG(daily_pnl) as avg_daily_pnl,
                            STDDEV(daily_pnl) as daily_pnl_stddev,
                            SUM(CASE WHEN daily_pnl > 0 THEN 1 ELSE 0 END) as profitable_days,
                            MIN(cumulative_pnl - running_max) as max_drawdown,
                            MAX(daily_pnl) as best_day,
                            MIN(daily_pnl) as worst_day
                        FROM cumulative_metrics
                    )
                    SELECT 
                        trading_days,
                        total_trades,
                        total_volume,
                        total_pnl,
                        avg_daily_pnl,
                        daily_pnl_stddev,
                        profitable_days,
                        CASE WHEN trading_days > 0 THEN profitable_days::float / trading_days::float ELSE 0 END as win_rate,
                        CASE WHEN daily_pnl_stddev > 0 THEN avg_daily_pnl / daily_pnl_stddev ELSE 0 END as sharpe_ratio,
                        COALESCE(max_drawdown, 0) as max_drawdown,
                        best_day,
                        worst_day,
                        CASE WHEN total_volume > 0 THEN total_pnl / total_volume * 100 ELSE 0 END as return_on_volume_pct
                    FROM performance_stats
                """
                
                row = await conn.fetchrow(query, leader_address, days)
                return dict(row) if row else None
                
        except Exception as e:
            logger.error(f"Error calculating leader performance metrics: {e}")
            return None

    async def get_asset_allocation(
        self,
        leader_address: str,
        days: int = 30
    ) -> Dict[str, float]:
        """Get asset allocation breakdown for a leader"""
        try:
            async with self.get_connection() as conn:
                query = """
                    SELECT 
                        asset,
                        SUM(size * price) as total_volume,
                        COUNT(*) as trade_count
                    FROM trades
                    WHERE leader_address = $1 
                        AND is_leader_trade = true
                        AND executed_at >= NOW() - INTERVAL '%s days'
                        AND status = 'filled'
                    GROUP BY asset
                    ORDER BY total_volume DESC
                """
                
                rows = await conn.fetch(query, leader_address, days)
                
                total_volume = sum(row['total_volume'] for row in rows)
                if total_volume == 0:
                    return {}
                
                return {
                    row['asset']: (row['total_volume'] / total_volume) * 100
                    for row in rows
                }
                
        except Exception as e:
            logger.error(f"Error calculating asset allocation: {e}")
            return {}

    async def get_time_series_data(
        self,
        leader_address: str,
        days: int = 30
    ) -> Dict[str, List[float]]:
        """Get time series data for charting"""
        try:
            async with self.get_connection() as conn:
                query = """
                    WITH daily_metrics AS (
                        SELECT 
                            DATE(executed_at) as trade_date,
                            SUM(CASE 
                                WHEN side = 'sell' THEN size * price
                                ELSE -size * price
                            END) as daily_pnl,
                            COUNT(*) as daily_trades,
                            SUM(size * price) as daily_volume
                        FROM trades
                        WHERE leader_address = $1 
                            AND is_leader_trade = true
                            AND executed_at >= NOW() - INTERVAL '%s days'
                            AND status = 'filled'
                        GROUP BY DATE(executed_at)
                        ORDER BY trade_date
                    )
                    SELECT 
                        trade_date,
                        daily_pnl,
                        daily_trades,
                        daily_volume,
                        SUM(daily_pnl) OVER (ORDER BY trade_date) as cumulative_pnl
                    FROM daily_metrics
                    ORDER BY trade_date
                """
                
                rows = await conn.fetch(query, leader_address, days)
                
                return {
                    'dates': [row['trade_date'].isoformat() for row in rows],
                    'daily_pnl': [float(row['daily_pnl']) for row in rows],
                    'cumulative_pnl': [float(row['cumulative_pnl']) for row in rows],
                    'daily_trades': [int(row['daily_trades']) for row in rows],
                    'daily_volume': [float(row['daily_volume']) for row in rows]
                }
                
        except Exception as e:
            logger.error(f"Error fetching time series data: {e}")
            return {}

    async def get_followers_by_leader(
        self,
        leader_address: str
    ) -> List[Dict[str, Any]]:
        """Get all followers for a specific leader"""
        try:
            async with self.get_connection() as conn:
                query = """
                    SELECT 
                        id, user_id, leader_address, api_wallet_address,
                        copy_percentage, max_position_size, stop_loss_percentage,
                        take_profit_percentage, is_active, risk_settings,
                        created_at, updated_at
                    FROM followers
                    WHERE leader_address = $1
                    ORDER BY created_at DESC
                """
                
                rows = await conn.fetch(query, leader_address)
                return [dict(row) for row in rows]
                
        except Exception as e:
            logger.error(f"Error fetching followers: {e}")
            return []

    async def store_analytics_result(
        self,
        result_type: str,
        data: Dict[str, Any],
        expiry_hours: int = 24
    ) -> str:
        """Store analytics result with expiration"""
        try:
            import json
            import uuid
            
            result_id = str(uuid.uuid4())
            expiry_time = datetime.utcnow() + timedelta(hours=expiry_hours)
            
            async with self.get_connection() as conn:
                await conn.execute("""
                    INSERT INTO analytics_cache (id, result_type, data, expires_at, created_at)
                    VALUES ($1, $2, $3, $4, NOW())
                    ON CONFLICT (id) DO UPDATE SET
                        data = EXCLUDED.data,
                        expires_at = EXCLUDED.expires_at,
                        created_at = EXCLUDED.created_at
                """, result_id, result_type, json.dumps(data), expiry_time)
            
            return result_id
            
        except Exception as e:
            logger.error(f"Error storing analytics result: {e}")
            raise

    async def get_analytics_result(
        self,
        result_id: str
    ) -> Optional[Dict[str, Any]]:
        """Retrieve cached analytics result"""
        try:
            import json
            
            async with self.get_connection() as conn:
                row = await conn.fetchrow("""
                    SELECT data, expires_at
                    FROM analytics_cache
                    WHERE id = $1 AND expires_at > NOW()
                """, result_id)
                
                if row:
                    return json.loads(row['data'])
                return None
                
        except Exception as e:
            logger.error(f"Error retrieving analytics result: {e}")
            return None

    async def cleanup_expired_cache(self):
        """Clean up expired analytics cache entries"""
        try:
            async with self.get_connection() as conn:
                result = await conn.execute("""
                    DELETE FROM analytics_cache 
                    WHERE expires_at <= NOW()
                """)
                
                rows_deleted = int(result.split()[-1])
                if rows_deleted > 0:
                    logger.info(f"Cleaned up {rows_deleted} expired cache entries")
                    
        except Exception as e:
            logger.error(f"Error cleaning up cache: {e}")

    async def health_check(self) -> Dict[str, Any]:
        """Check database health and return metrics"""
        try:
            async with self.get_connection() as conn:
                # Test basic connectivity
                await conn.fetchval("SELECT 1")
                
                # Get connection pool stats
                pool_stats = {
                    'size': self.pool.get_size(),
                    'min_size': self.pool.get_min_size(),
                    'max_size': self.pool.get_max_size(),
                    'idle_size': self.pool.get_idle_size()
                }
                
                # Get recent activity
                recent_trades = await conn.fetchval("""
                    SELECT COUNT(*)
                    FROM trades
                    WHERE created_at >= NOW() - INTERVAL '1 hour'
                """)
                
                return {
                    'status': 'healthy',
                    'pool_stats': pool_stats,
                    'recent_trades': recent_trades,
                    'timestamp': datetime.utcnow()
                }
                
        except Exception as e:
            logger.error(f"Database health check failed: {e}")
            return {
                'status': 'unhealthy',
                'error': str(e),
                'timestamp': datetime.utcnow()
            }


# Global database instance
db = AnalyticsDatabase()

async def init_db():
    """Initialize database connection"""
    await db.init_pool()

async def close_db():
    """Close database connection"""
    await db.close_pool()

def get_db() -> AnalyticsDatabase:
    """Get database instance"""
    return db
