import asyncio
import numpy as np
import pandas as pd
from typing import Dict, List, Optional, Any, Tuple
from datetime import datetime, timedelta
import logging
import uuid

from database import get_db
from models import (
    LeaderPerformanceAnalysis,
    PerformanceMetrics,
    RiskMetrics,
    MarketMetrics,
    FollowerOptimization,
    TradeRecommendation,
    RiskLevel
)

logger = logging.getLogger(__name__)

class AnalyticsEngine:
    def __init__(self):
        self.db = get_db()
        self.active_backtests = {}
        self.cache = {}

    async def analyze_leader_performance(
        self,
        leader_address: str,
        days: int = 30,
        include_predictions: bool = False
    ) -> Optional[LeaderPerformanceAnalysis]:
        """
        Comprehensive performance analysis for a leader
        """
        try:
            logger.info(f"Starting performance analysis for {leader_address}")
            
            # Get basic performance metrics
            metrics = await self.db.get_leader_performance_metrics(leader_address, days)
            if not metrics:
                return None

            # Get trade data for detailed analysis
            trades = await self.db.get_leader_trades(leader_address, days)
            if not trades:
                return None

            # Calculate performance metrics
            performance_metrics = await self._calculate_performance_metrics(trades, days)
            
            # Calculate risk metrics
            risk_metrics = await self._calculate_risk_metrics(trades)
            
            # Calculate market metrics (correlation, beta, etc.)
            market_metrics = await self._calculate_market_metrics(trades, days)
            
            # Get asset allocation
            asset_allocation = await self.db.get_asset_allocation(leader_address, days)
            
            # Get time series data
            time_series_data = await self.db.get_time_series_data(leader_address, days)
            
            # Calculate trading frequency patterns
            trading_frequency = await self._analyze_trading_frequency(trades)
            
            # ML predictions if requested
            predictions = None
            if include_predictions:
                try:
                    from ml_models import MLPredictor
                    ml_predictor = MLPredictor()
                    predictions = await ml_predictor.predict_leader_performance(
                        leader_address, horizon_days=7
                    )
                except Exception as e:
                    logger.warning(f"Failed to generate predictions: {e}")

            analysis = LeaderPerformanceAnalysis(
                leader_address=leader_address,
                analysis_period_days=days,
                performance_metrics=performance_metrics,
                risk_metrics=risk_metrics,
                market_metrics=market_metrics,
                trading_frequency=trading_frequency,
                asset_allocation=asset_allocation,
                time_series_data=time_series_data,
                predictions=predictions,
                analysis_timestamp=datetime.utcnow()
            )

            return analysis

        except Exception as e:
            logger.error(f"Error analyzing leader performance: {e}")
            return None

    async def _calculate_performance_metrics(
        self,
        trades: List[Dict],
        days: int
    ) -> PerformanceMetrics:
        """Calculate detailed performance metrics"""
        
        df = pd.DataFrame(trades)
        if df.empty:
            return self._empty_performance_metrics()

        # Calculate trade PnL
        df['pnl'] = df.apply(lambda row: 
            row['size'] * row['price'] if row['side'] == 'sell'
            else -row['size'] * row['price'], axis=1
        )

        total_pnl = df['pnl'].sum()
        total_trades = len(df)
        profitable_trades = len(df[df['pnl'] > 0])
        
        win_rate = (profitable_trades / total_trades * 100) if total_trades > 0 else 0
        
        winning_trades = df[df['pnl'] > 0]['pnl']
        losing_trades = df[df['pnl'] < 0]['pnl']
        
        avg_win = winning_trades.mean() if len(winning_trades) > 0 else 0
        avg_loss = abs(losing_trades.mean()) if len(losing_trades) > 0 else 0
        
        largest_win = winning_trades.max() if len(winning_trades) > 0 else 0
        largest_loss = abs(losing_trades.min()) if len(losing_trades) > 0 else 0
        
        # Calculate profit factor
        total_wins = winning_trades.sum() if len(winning_trades) > 0 else 0
        total_losses = abs(losing_trades.sum()) if len(losing_trades) > 0 else 0
        profit_factor = total_wins / total_losses if total_losses > 0 else 0
        
        # Calculate annualized return (assuming starting capital)
        starting_capital = 10000  # Default assumption
        total_return_pct = (total_pnl / starting_capital) * 100
        annualized_return = total_return_pct * (365 / days) if days > 0 else 0
        
        # Recovery factor (total return / max drawdown)
        max_drawdown_pct = await self._calculate_max_drawdown_pct(df)
        recovery_factor = total_return_pct / max_drawdown_pct if max_drawdown_pct > 0 else 0
        
        # Calmar ratio (annualized return / max drawdown)
        calmar_ratio = annualized_return / max_drawdown_pct if max_drawdown_pct > 0 else 0

        return PerformanceMetrics(
            total_return_pct=total_return_pct,
            annualized_return_pct=annualized_return,
            total_trades=total_trades,
            profitable_trades=profitable_trades,
            win_rate_pct=win_rate,
            avg_win_pct=(avg_win / starting_capital) * 100,
            avg_loss_pct=(avg_loss / starting_capital) * 100,
            largest_win_pct=(largest_win / starting_capital) * 100,
            largest_loss_pct=(largest_loss / starting_capital) * 100,
            profit_factor=profit_factor,
            recovery_factor=recovery_factor,
            calmar_ratio=calmar_ratio
        )

    async def _calculate_risk_metrics(self, trades: List[Dict]) -> RiskMetrics:
        """Calculate comprehensive risk metrics"""
        
        df = pd.DataFrame(trades)
        if df.empty:
            return self._empty_risk_metrics()

        # Calculate returns
        df['pnl'] = df.apply(lambda row: 
            row['size'] * row['price'] if row['side'] == 'sell'
            else -row['size'] * row['price'], axis=1
        )
        
        returns = df['pnl'].values
        
        # Portfolio volatility (annualized)
        volatility_daily = np.std(returns)
        volatility_annualized = volatility_daily * np.sqrt(252) / 10000 * 100  # Convert to %
        
        # Downside deviation
        negative_returns = returns[returns < 0]
        downside_deviation = np.std(negative_returns) if len(negative_returns) > 0 else 0
        downside_deviation_pct = downside_deviation / 10000 * 100
        
        # Value at Risk (95% confidence)
        var_95 = np.percentile(returns, 5) / 10000 * 100 if len(returns) > 0 else 0
        
        # Conditional Value at Risk (Expected Shortfall)
        cvar_95 = np.mean(returns[returns <= np.percentile(returns, 5)]) / 10000 * 100 if len(returns) > 0 else 0
        
        # Sharpe Ratio
        avg_return = np.mean(returns)
        sharpe_ratio = avg_return / volatility_daily if volatility_daily > 0 else 0
        
        # Sortino Ratio
        sortino_ratio = avg_return / downside_deviation if downside_deviation > 0 else 0
        
        # Max drawdown
        max_drawdown_pct = await self._calculate_max_drawdown_pct(df)
        
        # Current drawdown (simplified - would need real-time data)
        current_drawdown_pct = 0  # Would calculate from current positions
        
        # Risk level and score
        risk_level, risk_score = self._assess_risk_level(
            volatility_annualized, max_drawdown_pct, var_95
        )

        return RiskMetrics(
            max_drawdown_pct=max_drawdown_pct,
            current_drawdown_pct=current_drawdown_pct,
            volatility_pct=volatility_annualized,
            downside_deviation_pct=downside_deviation_pct,
            value_at_risk_95_pct=var_95,
            conditional_var_95_pct=cvar_95,
            sharpe_ratio=sharpe_ratio,
            sortino_ratio=sortino_ratio,
            risk_level=risk_level,
            risk_score=risk_score
        )

    async def _calculate_market_metrics(
        self,
        trades: List[Dict],
        days: int
    ) -> MarketMetrics:
        """Calculate market correlation and beta metrics"""
        
        # This would require market data integration
        # For now, return placeholder values
        return MarketMetrics(
            correlation_to_btc=0.0,
            correlation_to_eth=0.0,
            beta_to_market=1.0,
            alpha=0.0,
            tracking_error_pct=0.0,
            information_ratio=0.0
        )

    async def _analyze_trading_frequency(self, trades: List[Dict]) -> Dict[str, Any]:
        """Analyze trading patterns and frequency"""
        
        df = pd.DataFrame(trades)
        if df.empty:
            return {}

        df['executed_at'] = pd.to_datetime(df['executed_at'])
        df['hour'] = df['executed_at'].dt.hour
        df['day_of_week'] = df['executed_at'].dt.day_name()
        
        return {
            'trades_per_day': len(df) / df['executed_at'].dt.date.nunique(),
            'most_active_hours': df['hour'].value_counts().head(3).to_dict(),
            'most_active_days': df['day_of_week'].value_counts().head(3).to_dict(),
            'avg_time_between_trades_minutes': self._calculate_avg_time_between_trades(df),
            'trading_intensity_score': self._calculate_trading_intensity(df)
        }

    async def _calculate_max_drawdown_pct(self, df: pd.DataFrame) -> float:
        """Calculate maximum drawdown percentage"""
        
        if df.empty:
            return 0.0
            
        df_sorted = df.sort_values('executed_at')
        df_sorted['cumulative_pnl'] = df_sorted['pnl'].cumsum()
        df_sorted['running_max'] = df_sorted['cumulative_pnl'].expanding().max()
        df_sorted['drawdown'] = df_sorted['cumulative_pnl'] - df_sorted['running_max']
        
        max_drawdown = abs(df_sorted['drawdown'].min())
        return (max_drawdown / 10000) * 100  # Convert to percentage

    def _assess_risk_level(
        self,
        volatility: float,
        max_drawdown: float,
        var_95: float
    ) -> Tuple[RiskLevel, float]:
        """Assess overall risk level and score"""
        
        # Risk scoring based on multiple factors
        risk_score = 0.0
        
        # Volatility component (0-0.3)
        if volatility > 50:
            risk_score += 0.3
        elif volatility > 30:
            risk_score += 0.2
        elif volatility > 15:
            risk_score += 0.1
        
        # Max drawdown component (0-0.4)
        if max_drawdown > 30:
            risk_score += 0.4
        elif max_drawdown > 20:
            risk_score += 0.3
        elif max_drawdown > 10:
            risk_score += 0.2
        elif max_drawdown > 5:
            risk_score += 0.1
        
        # VaR component (0-0.3)
        if abs(var_95) > 10:
            risk_score += 0.3
        elif abs(var_95) > 5:
            risk_score += 0.2
        elif abs(var_95) > 2:
            risk_score += 0.1
        
        # Determine risk level
        if risk_score >= 0.7:
            risk_level = RiskLevel.EXTREME
        elif risk_score >= 0.5:
            risk_level = RiskLevel.HIGH
        elif risk_score >= 0.3:
            risk_level = RiskLevel.MEDIUM
        else:
            risk_level = RiskLevel.LOW
        
        return risk_level, min(risk_score, 1.0)

    def _calculate_avg_time_between_trades(self, df: pd.DataFrame) -> float:
        """Calculate average time between trades in minutes"""
        if len(df) < 2:
            return 0.0
            
        df_sorted = df.sort_values('executed_at')
        time_diffs = df_sorted['executed_at'].diff().dropna()
        avg_diff = time_diffs.mean()
        
        return avg_diff.total_seconds() / 60  # Convert to minutes

    def _calculate_trading_intensity(self, df: pd.DataFrame) -> float:
        """Calculate trading intensity score (0-1)"""
        if df.empty:
            return 0.0
            
        # Based on trades per day and trade size variance
        trades_per_day = len(df) / df['executed_at'].dt.date.nunique()
        size_variance = df['size'].var()
        
        # Normalize to 0-1 scale
        intensity = min(trades_per_day / 50, 1.0)  # Max intensity at 50 trades/day
        
        return intensity

    def _empty_performance_metrics(self) -> PerformanceMetrics:
        """Return empty performance metrics"""
        return PerformanceMetrics(
            total_return_pct=0.0,
            annualized_return_pct=0.0,
            total_trades=0,
            profitable_trades=0,
            win_rate_pct=0.0,
            avg_win_pct=0.0,
            avg_loss_pct=0.0,
            largest_win_pct=0.0,
            largest_loss_pct=0.0,
            profit_factor=0.0,
            recovery_factor=0.0,
            calmar_ratio=0.0
        )

    def _empty_risk_metrics(self) -> RiskMetrics:
        """Return empty risk metrics"""
        return RiskMetrics(
            max_drawdown_pct=0.0,
            current_drawdown_pct=0.0,
            volatility_pct=0.0,
            downside_deviation_pct=0.0,
            value_at_risk_95_pct=0.0,
            conditional_var_95_pct=0.0,
            sharpe_ratio=0.0,
            sortino_ratio=0.0,
            risk_level=RiskLevel.LOW,
            risk_score=0.0
        )

    async def optimize_follower_strategy(
        self,
        follower_id: int,
        risk_tolerance: float,
        max_drawdown_pct: float,
        target_return_pct: float
    ) -> FollowerOptimization:
        """Optimize follower strategy based on preferences"""
        
        # Get follower's current trades and performance
        trades = await self.db.get_follower_trades(follower_id, days=90)
        
        # Calculate current performance
        current_metrics = await self._calculate_performance_metrics(trades, 90)
        current_risk = await self._calculate_risk_metrics(trades)
        
        # Generate optimized settings based on risk tolerance
        optimized_settings = self._generate_optimized_settings(
            risk_tolerance, max_drawdown_pct, target_return_pct
        )
        
        # Calculate expected improvement
        expected_improvement = self._calculate_expected_improvement(
            current_metrics, optimized_settings
        )
        
        # Recommend leaders based on criteria
        recommended_leaders = await self._recommend_leaders_for_follower(
            risk_tolerance, target_return_pct
        )
        
        # Calculate portfolio allocation
        portfolio_allocation = self._calculate_portfolio_allocation(
            recommended_leaders, risk_tolerance
        )

        return FollowerOptimization(
            follower_id=follower_id,
            current_settings={
                "performance": current_metrics.dict(),
                "risk": current_risk.dict()
            },
            optimized_settings=optimized_settings,
            expected_improvement=expected_improvement,
            risk_assessment=current_risk,
            recommended_leaders=recommended_leaders,
            portfolio_allocation=portfolio_allocation,
            confidence_score=0.75  # Would be based on model confidence
        )

    def _generate_optimized_settings(
        self,
        risk_tolerance: float,
        max_drawdown_pct: float,
        target_return_pct: float
    ) -> Dict[str, Any]:
        """Generate optimized settings based on risk preferences"""
        
        # Copy percentage based on risk tolerance
        copy_percentage = min(risk_tolerance * 20, 15)  # Max 15%
        
        # Position sizing based on drawdown limit
        max_position_size = 1000 * (1 + risk_tolerance)
        
        # Stop loss based on max drawdown tolerance
        stop_loss_pct = max_drawdown_pct * 0.5  # Half of max drawdown tolerance
        
        # Take profit based on target return
        take_profit_pct = target_return_pct * 0.3  # 30% of target return per trade
        
        return {
            "copy_percentage": copy_percentage,
            "max_position_size": max_position_size,
            "stop_loss_percentage": stop_loss_pct,
            "take_profit_percentage": take_profit_pct,
            "risk_settings": {
                "max_trades_per_day": int(5 * (1 + risk_tolerance)),
                "max_consecutive_losses": int(3 * (2 - risk_tolerance)),
                "correlation_limit": 0.8 - (risk_tolerance * 0.2)
            }
        }

    def _calculate_expected_improvement(
        self,
        current_metrics: PerformanceMetrics,
        optimized_settings: Dict[str, Any]
    ) -> Dict[str, float]:
        """Calculate expected improvement from optimization"""
        
        # Simplified improvement calculation
        # In reality, this would use historical backtesting
        
        return {
            "return_improvement_pct": 2.5,
            "risk_reduction_pct": 15.0,
            "sharpe_improvement": 0.3,
            "drawdown_reduction_pct": 20.0
        }

    async def _recommend_leaders_for_follower(
        self,
        risk_tolerance: float,
        target_return_pct: float
    ) -> List[str]:
        """Recommend leaders based on follower's risk and return preferences"""
        
        # This would query top performers matching criteria
        # For now, return placeholder recommendations
        
        return [
            "0x1234567890123456789012345678901234567890",
            "0x2345678901234567890123456789012345678901",
            "0x3456789012345678901234567890123456789012"
        ]

    def _calculate_portfolio_allocation(
        self,
        leaders: List[str],
        risk_tolerance: float
    ) -> Dict[str, float]:
        """Calculate optimal portfolio allocation across leaders"""
        
        # Equal weight allocation for simplicity
        # In reality, would use portfolio optimization
        if not leaders:
            return {}
            
        weight_per_leader = 100 / len(leaders)
        
        return {leader: weight_per_leader for leader in leaders}

    async def calculate_risk_metrics(
        self,
        leader_address: str,
        days: int = 90
    ) -> Optional[RiskMetrics]:
        """Calculate comprehensive risk metrics for a leader"""
        
        trades = await self.db.get_leader_trades(leader_address, days)
        if not trades:
            return None
            
        return await self._calculate_risk_metrics(trades)

    async def analyze_market_sentiment(
        self,
        assets: List[str],
        timeframe_hours: int
    ) -> Dict[str, Any]:
        """Analyze market sentiment for specified assets"""
        
        # Placeholder implementation
        # Would integrate with sentiment analysis APIs
        
        sentiment_scores = {}
        for asset in assets:
            # Random sentiment for demo
            import random
            sentiment_scores[asset] = random.uniform(-1.0, 1.0)
        
        return {
            "sentiment_scores": sentiment_scores,
            "overall_sentiment": sum(sentiment_scores.values()) / len(sentiment_scores),
            "market_fear_greed": random.uniform(0, 100),
            "volatility_outlook": "moderate"
        }

    async def generate_trade_recommendations(
        self,
        follower_id: int,
        max_recommendations: int = 5
    ) -> List[TradeRecommendation]:
        """Generate personalized trade recommendations"""
        
        # Placeholder implementation
        recommendations = []
        
        assets = ["BTC", "ETH", "SOL"]
        actions = ["buy", "sell"]
        
        import random
        
        for i in range(min(max_recommendations, 3)):
            rec = TradeRecommendation(
                asset=random.choice(assets),
                action=random.choice(actions),
                confidence=random.uniform(0.6, 0.9),
                expected_return_pct=random.uniform(2.0, 8.0),
                risk_pct=random.uniform(1.0, 5.0),
                time_horizon_hours=random.randint(6, 48),
                reasoning=f"Based on technical analysis and leader patterns",
                leader_sources=["0x123..."],
                generated_at=datetime.utcnow()
            )
            recommendations.append(rec)
        
        return recommendations

    async def start_backtest(
        self,
        leader_address: str,
        start_date: datetime,
        end_date: datetime,
        initial_capital: float,
        copy_percentage: float
    ) -> str:
        """Start a backtesting job"""
        
        backtest_id = str(uuid.uuid4())
        
        # Store backtest configuration
        self.active_backtests[backtest_id] = {
            "status": "queued",
            "leader_address": leader_address,
            "start_date": start_date,
            "end_date": end_date,
            "initial_capital": initial_capital,
            "copy_percentage": copy_percentage,
            "created_at": datetime.utcnow(),
            "progress": 0
        }
        
        return backtest_id

    async def execute_backtest(self, backtest_id: str):
        """Execute backtest in background"""
        
        try:
            if backtest_id not in self.active_backtests:
                return
                
            config = self.active_backtests[backtest_id]
            config["status"] = "running"
            config["started_at"] = datetime.utcnow()
            
            # Simulate backtest execution
            await asyncio.sleep(5)  # Simulate processing time
            
            # Generate mock results
            config["status"] = "completed"
            config["completed_at"] = datetime.utcnow()
            config["results"] = {
                "total_return_pct": 15.7,
                "max_drawdown_pct": 8.3,
                "sharpe_ratio": 1.42,
                "total_trades": 45,
                "win_rate_pct": 67.5
            }
            
        except Exception as e:
            logger.error(f"Backtest execution failed: {e}")
            if backtest_id in self.active_backtests:
                self.active_backtests[backtest_id]["status"] = "failed"
                self.active_backtests[backtest_id]["error"] = str(e)

    async def get_backtest_status(self, backtest_id: str) -> Optional[Dict[str, Any]]:
        """Get backtest status and results"""
        
        return self.active_backtests.get(backtest_id)

    async def compare_leaders(
        self,
        addresses: List[str],
        days: int,
        metrics: List[str]
    ) -> Dict[str, Any]:
        """Compare multiple leaders across specified metrics"""
        
        comparison_data = {}
        
        for address in addresses:
            trades = await self.db.get_leader_trades(address, days)
            if trades:
                performance = await self._calculate_performance_metrics(trades, days)
                risk = await self._calculate_risk_metrics(trades)
                
                comparison_data[address] = {
                    "return": performance.total_return_pct,
                    "sharpe": risk.sharpe_ratio,
                    "max_drawdown": risk.max_drawdown_pct,
                    "win_rate": performance.win_rate_pct,
                    "total_trades": performance.total_trades,
                    "volatility": risk.volatility_pct
                }
        
        return comparison_data

    async def get_trending_leaders(
        self,
        timeframe: str,
        min_followers: int,
        limit: int
    ) -> List[Dict[str, Any]]:
        """Get trending/top performing leaders"""
        
        # Convert timeframe to days
        timeframe_map = {"1d": 1, "3d": 3, "7d": 7, "30d": 30}
        days = timeframe_map.get(timeframe, 7)
        
        # This would query the database for top performers
        # For now, return mock data
        
        trending = []
        for i in range(limit):
            trending.append({
                "address": f"0x{i:040d}",
                "return_pct": 25.5 - (i * 2),
                "sharpe_ratio": 2.1 - (i * 0.1),
                "followers_count": min_followers + i * 5,
                "rank": i + 1
            })
        
        return trending

    async def health_check(self) -> Dict[str, bool]:
        """Check health of analytics service components"""
        
        health = {
            "database": False,
            "ml_models": False,
            "cache": True
        }
        
        try:
            # Check database
            db_health = await self.db.health_check()
            health["database"] = db_health["status"] == "healthy"
            
            # Check ML models
            try:
                from ml_models import MLPredictor
                ml_predictor = MLPredictor()
                health["ml_models"] = await ml_predictor.health_check()
            except:
                health["ml_models"] = False
        
        except Exception as e:
            logger.error(f"Health check error: {e}")
        
        return health

    async def analyze_portfolio(
        self,
        addresses: List[str],
        weights: Optional[List[float]],
        rebalance_frequency: str
    ) -> Dict[str, Any]:
        """Analyze a portfolio of multiple leaders"""
        
        if not weights:
            weights = [1/len(addresses)] * len(addresses)
        
        # Normalize weights
        total_weight = sum(weights)
        weights = [w/total_weight for w in weights]
        
        portfolio_metrics = {
            "expected_return_pct": 0.0,
            "expected_volatility_pct": 0.0,
            "sharpe_ratio": 0.0,
            "max_drawdown_pct": 0.0
        }
        
        # Calculate weighted portfolio metrics
        for i, address in enumerate(addresses):
            trades = await self.db.get_leader_trades(address, 30)
            if trades:
                performance = await self._calculate_performance_metrics(trades, 30)
                risk = await self._calculate_risk_metrics(trades)
                
                weight = weights[i]
                portfolio_metrics["expected_return_pct"] += performance.total_return_pct * weight
                portfolio_metrics["expected_volatility_pct"] += risk.volatility_pct * weight
                portfolio_metrics["max_drawdown_pct"] += risk.max_drawdown_pct * weight
        
        # Calculate portfolio Sharpe ratio
        if portfolio_metrics["expected_volatility_pct"] > 0:
            portfolio_metrics["sharpe_ratio"] = (
                portfolio_metrics["expected_return_pct"] / 
                portfolio_metrics["expected_volatility_pct"]
            )
        
        return {
            "portfolio_metrics": portfolio_metrics,
            "allocation": dict(zip(addresses, weights)),
            "rebalance_frequency": rebalance_frequency,
            "diversification_benefit": self._calculate_diversification_benefit(addresses, weights)
        }

    def _calculate_diversification_benefit(
        self,
        addresses: List[str],
        weights: List[float]
    ) -> float:
        """Calculate diversification benefit of portfolio"""
        
        # Simplified calculation
        # In reality, would calculate based on correlation matrix
        
        if len(addresses) <= 1:
            return 0.0
        
        # More assets = more diversification benefit
        return min(len(addresses) * 0.15, 0.8)
