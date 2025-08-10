from typing import List, Optional, Dict, Any
from datetime import datetime, timedelta
import logging

from fastapi import APIRouter, HTTPException, Query, BackgroundTasks
from pydantic import BaseModel, Field

from analytics_engine import AnalyticsEngine
from ml_models import MLPredictor
from models import (
    LeaderPerformanceAnalysis,
    FollowerOptimization,
    RiskMetrics,
    MarketAnalysis,
    TradeRecommendation
)

logger = logging.getLogger(__name__)

router = APIRouter()
analytics_engine = AnalyticsEngine()
ml_predictor = MLPredictor()


class AnalyticsRequest(BaseModel):
    leader_address: str
    timeframe_days: int = Field(default=30, ge=1, le=365)


class OptimizationRequest(BaseModel):
    follower_id: int
    risk_tolerance: float = Field(default=0.5, ge=0.0, le=1.0)
    max_drawdown_pct: float = Field(default=10.0, ge=1.0, le=50.0)
    target_return_pct: float = Field(default=15.0, ge=1.0, le=100.0)


class PortfolioAnalysisRequest(BaseModel):
    addresses: List[str]
    weights: Optional[List[float]] = None
    rebalance_frequency: str = "weekly"


@router.get("/analytics/leader/{leader_address}/performance", 
            response_model=LeaderPerformanceAnalysis)
async def analyze_leader_performance(
    leader_address: str,
    days: int = Query(default=30, ge=1, le=365),
    include_predictions: bool = Query(default=False)
) -> LeaderPerformanceAnalysis:
    """
    Comprehensive performance analysis for a leader trader
    """
    try:
        logger.info(f"Analyzing leader performance for {leader_address} over {days} days")
        
        # Get comprehensive analytics
        analysis = await analytics_engine.analyze_leader_performance(
            leader_address, days, include_predictions
        )
        
        if not analysis:
            raise HTTPException(
                status_code=404, 
                detail=f"No data found for leader {leader_address}"
            )
        
        return analysis
    
    except Exception as e:
        logger.error(f"Error analyzing leader performance: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/analytics/leader/{leader_address}/risk-metrics", 
            response_model=RiskMetrics)
async def get_leader_risk_metrics(
    leader_address: str,
    days: int = Query(default=90, ge=30, le=365)
) -> RiskMetrics:
    """
    Calculate comprehensive risk metrics for a leader
    """
    try:
        risk_metrics = await analytics_engine.calculate_risk_metrics(
            leader_address, days
        )
        
        if not risk_metrics:
            raise HTTPException(
                status_code=404,
                detail=f"No risk data available for leader {leader_address}"
            )
        
        return risk_metrics
    
    except Exception as e:
        logger.error(f"Error calculating risk metrics: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/analytics/follower/optimize")
async def optimize_follower_strategy(
    request: OptimizationRequest
) -> FollowerOptimization:
    """
    Optimize follower strategy based on risk preferences and performance goals
    """
    try:
        logger.info(f"Optimizing strategy for follower {request.follower_id}")
        
        optimization = await analytics_engine.optimize_follower_strategy(
            follower_id=request.follower_id,
            risk_tolerance=request.risk_tolerance,
            max_drawdown_pct=request.max_drawdown_pct,
            target_return_pct=request.target_return_pct
        )
        
        return optimization
    
    except Exception as e:
        logger.error(f"Error optimizing follower strategy: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/analytics/market/sentiment")
async def get_market_sentiment(
    assets: List[str] = Query(default=["BTC", "ETH", "SOL"]),
    timeframe_hours: int = Query(default=24, ge=1, le=168)
) -> Dict[str, Any]:
    """
    Analyze market sentiment for specified assets
    """
    try:
        sentiment = await analytics_engine.analyze_market_sentiment(
            assets, timeframe_hours
        )
        
        return {
            "timestamp": datetime.utcnow(),
            "timeframe_hours": timeframe_hours,
            "sentiment_analysis": sentiment
        }
    
    except Exception as e:
        logger.error(f"Error analyzing market sentiment: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/analytics/portfolio/analysis")
async def analyze_portfolio(
    request: PortfolioAnalysisRequest
) -> Dict[str, Any]:
    """
    Analyze a portfolio of multiple leaders with optimal allocation
    """
    try:
        if request.weights and len(request.weights) != len(request.addresses):
            raise HTTPException(
                status_code=400,
                detail="Weights length must match addresses length"
            )
        
        analysis = await analytics_engine.analyze_portfolio(
            addresses=request.addresses,
            weights=request.weights,
            rebalance_frequency=request.rebalance_frequency
        )
        
        return analysis
    
    except Exception as e:
        logger.error(f"Error analyzing portfolio: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/ml/predictions/{leader_address}")
async def get_ml_predictions(
    leader_address: str,
    horizon_days: int = Query(default=7, ge=1, le=30),
    confidence_threshold: float = Query(default=0.7, ge=0.5, le=0.99)
) -> Dict[str, Any]:
    """
    Get ML-based predictions for leader performance
    """
    try:
        predictions = await ml_predictor.predict_leader_performance(
            leader_address=leader_address,
            horizon_days=horizon_days,
            confidence_threshold=confidence_threshold
        )
        
        if not predictions:
            raise HTTPException(
                status_code=404,
                detail=f"Insufficient data for predictions on {leader_address}"
            )
        
        return {
            "leader_address": leader_address,
            "horizon_days": horizon_days,
            "confidence_threshold": confidence_threshold,
            "predictions": predictions,
            "generated_at": datetime.utcnow()
        }
    
    except Exception as e:
        logger.error(f"Error generating ML predictions: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/analytics/recommendations/{follower_id}")
async def get_trade_recommendations(
    follower_id: int,
    max_recommendations: int = Query(default=5, ge=1, le=20)
) -> List[TradeRecommendation]:
    """
    Get personalized trade recommendations for a follower
    """
    try:
        recommendations = await analytics_engine.generate_trade_recommendations(
            follower_id=follower_id,
            max_recommendations=max_recommendations
        )
        
        return recommendations
    
    except Exception as e:
        logger.error(f"Error generating trade recommendations: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/analytics/backtest")
async def run_backtest(
    background_tasks: BackgroundTasks,
    leader_address: str,
    start_date: datetime,
    end_date: datetime,
    initial_capital: float = 10000.0,
    copy_percentage: float = 10.0
) -> Dict[str, Any]:
    """
    Run backtesting for a copy trading strategy
    """
    try:
        if end_date <= start_date:
            raise HTTPException(
                status_code=400,
                detail="End date must be after start date"
            )
        
        if (end_date - start_date).days > 365:
            raise HTTPException(
                status_code=400,
                detail="Backtest period cannot exceed 365 days"
            )
        
        # Run backtest in background
        backtest_id = await analytics_engine.start_backtest(
            leader_address=leader_address,
            start_date=start_date,
            end_date=end_date,
            initial_capital=initial_capital,
            copy_percentage=copy_percentage
        )
        
        background_tasks.add_task(
            analytics_engine.execute_backtest,
            backtest_id
        )
        
        return {
            "backtest_id": backtest_id,
            "status": "started",
            "message": "Backtest initiated, check status with /analytics/backtest/{backtest_id}/status"
        }
    
    except Exception as e:
        logger.error(f"Error starting backtest: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/analytics/backtest/{backtest_id}/status")
async def get_backtest_status(backtest_id: str) -> Dict[str, Any]:
    """
    Get status and results of a backtest
    """
    try:
        status = await analytics_engine.get_backtest_status(backtest_id)
        
        if not status:
            raise HTTPException(
                status_code=404,
                detail=f"Backtest {backtest_id} not found"
            )
        
        return status
    
    except Exception as e:
        logger.error(f"Error getting backtest status: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/analytics/compare")
async def compare_leaders(
    addresses: List[str] = Query(..., min_items=2, max_items=10),
    days: int = Query(default=30, ge=7, le=365),
    metrics: List[str] = Query(default=["return", "sharpe", "max_drawdown", "win_rate"])
) -> Dict[str, Any]:
    """
    Compare multiple leaders across specified metrics
    """
    try:
        comparison = await analytics_engine.compare_leaders(
            addresses=addresses,
            days=days,
            metrics=metrics
        )
        
        return {
            "comparison": comparison,
            "leaders": addresses,
            "period_days": days,
            "metrics": metrics,
            "generated_at": datetime.utcnow()
        }
    
    except Exception as e:
        logger.error(f"Error comparing leaders: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/analytics/trending-leaders")
async def get_trending_leaders(
    timeframe: str = Query(default="7d", regex="^(1d|3d|7d|30d)$"),
    min_followers: int = Query(default=5, ge=1),
    limit: int = Query(default=20, ge=1, le=100)
) -> Dict[str, Any]:
    """
    Get trending/top performing leaders
    """
    try:
        trending = await analytics_engine.get_trending_leaders(
            timeframe=timeframe,
            min_followers=min_followers,
            limit=limit
        )
        
        return {
            "trending_leaders": trending,
            "timeframe": timeframe,
            "criteria": {
                "min_followers": min_followers,
                "limit": limit
            },
            "generated_at": datetime.utcnow()
        }
    
    except Exception as e:
        logger.error(f"Error getting trending leaders: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/analytics/health")
async def analytics_health() -> Dict[str, Any]:
    """
    Health check for analytics service components
    """
    try:
        health = await analytics_engine.health_check()
        
        return {
            "status": "healthy" if all(health.values()) else "degraded",
            "components": health,
            "timestamp": datetime.utcnow()
        }
    
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        raise HTTPException(status_code=500, detail="Health check failed")
