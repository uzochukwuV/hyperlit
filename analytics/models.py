from typing import List, Optional, Dict, Any
from datetime import datetime
from pydantic import BaseModel, Field
from enum import Enum


class TradeDirection(str, Enum):
    BUY = "buy"
    SELL = "sell"


class TradeStatus(str, Enum):
    PENDING = "pending"
    FILLED = "filled"
    CANCELLED = "cancelled"
    REJECTED = "rejected"


class RiskLevel(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    EXTREME = "extreme"


class Trade(BaseModel):
    id: int
    leader_address: str
    follower_id: Optional[int] = None
    asset: str
    side: TradeDirection
    size: float
    price: float
    order_type: str
    is_leader_trade: bool
    executed_at: datetime
    hyperliquid_tx_id: str
    status: TradeStatus


class Position(BaseModel):
    id: int
    user_address: str
    asset: str
    side: str
    size: float
    entry_price: float
    current_price: float
    unrealized_pnl: float
    margin_used: float
    updated_at: datetime


class PerformanceMetrics(BaseModel):
    total_return_pct: float
    annualized_return_pct: float
    total_trades: int
    profitable_trades: int
    win_rate_pct: float
    avg_win_pct: float
    avg_loss_pct: float
    largest_win_pct: float
    largest_loss_pct: float
    profit_factor: float
    recovery_factor: float
    calmar_ratio: float


class RiskMetrics(BaseModel):
    max_drawdown_pct: float
    current_drawdown_pct: float
    volatility_pct: float
    downside_deviation_pct: float
    value_at_risk_95_pct: float
    conditional_var_95_pct: float
    sharpe_ratio: float
    sortino_ratio: float
    risk_level: RiskLevel
    risk_score: float = Field(ge=0.0, le=1.0)


class MarketMetrics(BaseModel):
    correlation_to_btc: float
    correlation_to_eth: float
    beta_to_market: float
    alpha: float
    tracking_error_pct: float
    information_ratio: float


class LeaderPerformanceAnalysis(BaseModel):
    leader_address: str
    analysis_period_days: int
    performance_metrics: PerformanceMetrics
    risk_metrics: RiskMetrics
    market_metrics: MarketMetrics
    trading_frequency: Dict[str, Any]
    asset_allocation: Dict[str, float]
    time_series_data: Dict[str, List[float]]
    predictions: Optional[Dict[str, Any]] = None
    analysis_timestamp: datetime


class FollowerOptimization(BaseModel):
    follower_id: int
    current_settings: Dict[str, Any]
    optimized_settings: Dict[str, Any]
    expected_improvement: Dict[str, float]
    risk_assessment: RiskMetrics
    recommended_leaders: List[str]
    portfolio_allocation: Dict[str, float]
    confidence_score: float = Field(ge=0.0, le=1.0)


class TradeRecommendation(BaseModel):
    asset: str
    action: TradeDirection
    confidence: float = Field(ge=0.0, le=1.0)
    expected_return_pct: float
    risk_pct: float
    time_horizon_hours: int
    reasoning: str
    leader_sources: List[str]
    generated_at: datetime


class MarketAnalysis(BaseModel):
    assets: List[str]
    timeframe_hours: int
    sentiment_scores: Dict[str, float]
    volatility_forecast: Dict[str, float]
    correlation_matrix: Dict[str, Dict[str, float]]
    market_regime: str
    risk_on_off_indicator: float
    fear_greed_index: float
    analysis_timestamp: datetime


class BacktestResult(BaseModel):
    backtest_id: str
    leader_address: str
    strategy_params: Dict[str, Any]
    start_date: datetime
    end_date: datetime
    initial_capital: float
    final_capital: float
    total_return_pct: float
    max_drawdown_pct: float
    sharpe_ratio: float
    total_trades: int
    win_rate_pct: float
    trade_history: List[Trade]
    equity_curve: List[Dict[str, Any]]
    monthly_returns: List[float]
    annual_statistics: Dict[str, float]


class AlertConfig(BaseModel):
    follower_id: int
    alert_type: str
    threshold: float
    comparison: str  # "greater_than", "less_than", "equal_to"
    is_active: bool
    created_at: datetime


class MLPrediction(BaseModel):
    leader_address: str
    prediction_type: str
    horizon_days: int
    predicted_return_pct: float
    confidence: float = Field(ge=0.0, le=1.0)
    probability_of_profit: float = Field(ge=0.0, le=1.0)
    expected_max_drawdown_pct: float
    feature_importance: Dict[str, float]
    model_version: str
    prediction_timestamp: datetime


class PortfolioAnalysis(BaseModel):
    portfolio_id: str
    leaders: List[str]
    current_allocation: Dict[str, float]
    optimized_allocation: Dict[str, float]
    expected_return_pct: float
    expected_volatility_pct: float
    sharpe_ratio: float
    diversification_ratio: float
    concentration_risk: float
    rebalancing_frequency: str
    next_rebalance_date: datetime


class TradingSession(BaseModel):
    session_id: str
    leader_address: str
    start_time: datetime
    end_time: Optional[datetime] = None
    session_pnl: float
    session_volume: float
    trades_count: int
    session_return_pct: float
    session_sharpe: float
    active_assets: List[str]


class LeaderRanking(BaseModel):
    rank: int
    leader_address: str
    score: float
    return_pct: float
    risk_adjusted_return: float
    consistency_score: float
    followers_count: int
    aum_estimate: float
    trend_direction: str  # "up", "down", "sideways"
    ranking_timestamp: datetime


class SystemHealth(BaseModel):
    database_status: bool
    ml_models_status: bool
    data_freshness_minutes: int
    last_update_timestamp: datetime
    active_backtests: int
    queue_size: int
    error_rate_pct: float
    avg_response_time_ms: float


class NotificationPreferences(BaseModel):
    user_id: str
    email_alerts: bool
    push_notifications: bool
    slack_webhook: Optional[str] = None
    discord_webhook: Optional[str] = None
    alert_frequency: str  # "immediate", "daily", "weekly"
    risk_threshold: float = Field(ge=0.0, le=1.0)


class AnalyticsConfig(BaseModel):
    model_update_frequency_hours: int = 24
    backtest_queue_size: int = 100
    max_concurrent_analyses: int = 10
    data_retention_days: int = 365
    cache_ttl_minutes: int = 15
    risk_free_rate_pct: float = 2.0
    benchmark_symbol: str = "BTC"
    ml_confidence_threshold: float = 0.7
