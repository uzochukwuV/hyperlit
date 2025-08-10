import numpy as np
import pandas as pd
from typing import Dict, List, Optional, Any
from datetime import datetime, timedelta
import logging
import pickle
import asyncio
from sklearn.ensemble import RandomForestRegressor, GradientBoostingClassifier
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_squared_error, accuracy_score

from database import get_db

logger = logging.getLogger(__name__)

class MLPredictor:
    def __init__(self):
        self.db = get_db()
        self.models = {}
        self.scalers = {}
        self.feature_importance = {}
        self.model_versions = {}
        self.is_trained = False

    async def predict_leader_performance(
        self,
        leader_address: str,
        horizon_days: int = 7,
        confidence_threshold: float = 0.7
    ) -> Optional[Dict[str, Any]]:
        """
        Predict leader performance using ML models
        """
        try:
            # Get leader's historical data
            trades = await self.db.get_leader_trades(leader_address, days=90)
            if len(trades) < 20:  # Minimum trades required
                return None

            # Extract features
            features = await self._extract_features(trades, leader_address)
            if not features:
                return None

            # Ensure models are trained
            if not self.is_trained:
                await self._train_models()

            # Make predictions
            predictions = await self._make_predictions(features, horizon_days)
            
            # Calculate confidence
            confidence = await self._calculate_prediction_confidence(features, predictions)
            
            if confidence < confidence_threshold:
                return None

            return {
                "leader_address": leader_address,
                "horizon_days": horizon_days,
                "predicted_return_pct": predictions["return_pct"],
                "probability_of_profit": predictions["profit_probability"],
                "expected_max_drawdown_pct": predictions["max_drawdown_pct"],
                "confidence": confidence,
                "feature_importance": self.feature_importance.get("performance_model", {}),
                "model_version": self.model_versions.get("performance_model", "1.0"),
                "prediction_timestamp": datetime.utcnow()
            }

        except Exception as e:
            logger.error(f"Error predicting leader performance: {e}")
            return None

    async def _extract_features(
        self,
        trades: List[Dict],
        leader_address: str
    ) -> Optional[Dict[str, float]]:
        """
        Extract ML features from trade history
        """
        try:
            df = pd.DataFrame(trades)
            if df.empty:
                return None

            df['executed_at'] = pd.to_datetime(df['executed_at'])
            df['pnl'] = df.apply(lambda row: 
                row['size'] * row['price'] if row['side'] == 'sell'
                else -row['size'] * row['price'], axis=1
            )

            # Time-based features
            df['hour'] = df['executed_at'].dt.hour
            df['day_of_week'] = df['executed_at'].dt.dayofweek
            df['days_since_start'] = (df['executed_at'] - df['executed_at'].min()).dt.days

            # Performance features
            total_pnl = df['pnl'].sum()
            win_rate = (df['pnl'] > 0).mean()
            avg_win = df[df['pnl'] > 0]['pnl'].mean() if len(df[df['pnl'] > 0]) > 0 else 0
            avg_loss = df[df['pnl'] < 0]['pnl'].mean() if len(df[df['pnl'] < 0]) > 0 else 0
            profit_factor = abs(df[df['pnl'] > 0]['pnl'].sum() / df[df['pnl'] < 0]['pnl'].sum()) if df[df['pnl'] < 0]['pnl'].sum() != 0 else 0

            # Volatility features
            daily_pnl = df.groupby(df['executed_at'].dt.date)['pnl'].sum()
            volatility = daily_pnl.std()
            
            # Trend features
            returns = daily_pnl.pct_change().dropna()
            momentum = returns.tail(7).mean()  # 7-day momentum
            
            # Risk features
            max_consecutive_losses = self._calculate_max_consecutive_losses(df['pnl'])
            drawdown = self._calculate_current_drawdown(daily_pnl)

            # Trading frequency features
            avg_trades_per_day = len(df) / df['executed_at'].dt.date.nunique()
            trade_size_variance = df['size'].var()

            # Asset diversity features
            unique_assets = df['asset'].nunique()
            asset_concentration = df.groupby('asset')['pnl'].sum().abs().max() / abs(total_pnl) if total_pnl != 0 else 0

            # Time pattern features
            most_active_hour = df['hour'].mode().iloc[0] if not df['hour'].mode().empty else 12
            weekend_trading_ratio = (df['day_of_week'] >= 5).mean()

            # Recent performance features
            recent_trades = df.tail(10)
            recent_win_rate = (recent_trades['pnl'] > 0).mean()
            recent_avg_pnl = recent_trades['pnl'].mean()

            features = {
                "total_pnl": total_pnl,
                "win_rate": win_rate,
                "avg_win": avg_win,
                "avg_loss": abs(avg_loss),
                "profit_factor": profit_factor,
                "volatility": volatility,
                "momentum": momentum,
                "max_consecutive_losses": max_consecutive_losses,
                "drawdown": drawdown,
                "avg_trades_per_day": avg_trades_per_day,
                "trade_size_variance": trade_size_variance,
                "unique_assets": unique_assets,
                "asset_concentration": asset_concentration,
                "most_active_hour": most_active_hour,
                "weekend_trading_ratio": weekend_trading_ratio,
                "recent_win_rate": recent_win_rate,
                "recent_avg_pnl": recent_avg_pnl,
                "total_trades": len(df),
                "trading_days": df['executed_at'].dt.date.nunique()
            }

            return features

        except Exception as e:
            logger.error(f"Error extracting features: {e}")
            return None

    def _calculate_max_consecutive_losses(self, pnl_series: pd.Series) -> int:
        """Calculate maximum consecutive losses"""
        consecutive_losses = 0
        max_consecutive = 0
        
        for pnl in pnl_series:
            if pnl < 0:
                consecutive_losses += 1
                max_consecutive = max(max_consecutive, consecutive_losses)
            else:
                consecutive_losses = 0
        
        return max_consecutive

    def _calculate_current_drawdown(self, daily_pnl: pd.Series) -> float:
        """Calculate current drawdown"""
        cumulative = daily_pnl.cumsum()
        running_max = cumulative.expanding().max()
        drawdown = cumulative - running_max
        return abs(drawdown.iloc[-1]) if not drawdown.empty else 0

    async def _train_models(self):
        """
        Train ML models using historical data
        """
        try:
            logger.info("Training ML models...")
            
            # Get training data from multiple leaders
            training_data = await self._prepare_training_data()
            
            if not training_data or len(training_data) < 100:
                logger.warning("Insufficient training data, using pre-trained models")
                await self._load_pretrained_models()
                return

            # Prepare features and targets
            X = pd.DataFrame([data["features"] for data in training_data])
            y_return = [data["target_return"] for data in training_data]
            y_profit = [1 if ret > 0 else 0 for ret in y_return]

            # Split data
            X_train, X_test, y_return_train, y_return_test, y_profit_train, y_profit_test = train_test_split(
                X, y_return, y_profit, test_size=0.2, random_state=42
            )

            # Scale features
            scaler = StandardScaler()
            X_train_scaled = scaler.fit_transform(X_train)
            X_test_scaled = scaler.transform(X_test)

            # Train return prediction model
            return_model = RandomForestRegressor(
                n_estimators=100,
                max_depth=10,
                random_state=42,
                n_jobs=-1
            )
            return_model.fit(X_train_scaled, y_return_train)

            # Train profit probability model
            profit_model = GradientBoostingClassifier(
                n_estimators=100,
                max_depth=6,
                random_state=42
            )
            profit_model.fit(X_train_scaled, y_profit_train)

            # Evaluate models
            return_pred = return_model.predict(X_test_scaled)
            profit_pred = profit_model.predict(X_test_scaled)

            return_mse = mean_squared_error(y_return_test, return_pred)
            profit_accuracy = accuracy_score(y_profit_test, profit_pred)

            logger.info(f"Model performance - Return MSE: {return_mse:.4f}, Profit Accuracy: {profit_accuracy:.4f}")

            # Store models
            self.models["performance_model"] = return_model
            self.models["profit_model"] = profit_model
            self.scalers["feature_scaler"] = scaler

            # Store feature importance
            feature_names = X.columns.tolist()
            self.feature_importance["performance_model"] = dict(
                zip(feature_names, return_model.feature_importances_)
            )

            self.model_versions["performance_model"] = "1.0"
            self.is_trained = True

            logger.info("ML models trained successfully")

        except Exception as e:
            logger.error(f"Error training models: {e}")
            await self._load_pretrained_models()

    async def _prepare_training_data(self) -> List[Dict[str, Any]]:
        """
        Prepare training data from historical leader performance
        """
        try:
            # This would query multiple leaders' historical data
            # For now, return empty to use pretrained models
            return []

        except Exception as e:
            logger.error(f"Error preparing training data: {e}")
            return []

    async def _load_pretrained_models(self):
        """
        Load pretrained models (placeholder implementation)
        """
        logger.info("Loading pretrained models...")
        
        # In a real implementation, you would load saved model files
        # For now, create simple placeholder models
        
        self.models["performance_model"] = RandomForestRegressor(
            n_estimators=50, random_state=42
        )
        self.models["profit_model"] = GradientBoostingClassifier(
            n_estimators=50, random_state=42
        )
        self.scalers["feature_scaler"] = StandardScaler()
        
        # Create dummy training data to fit the models
        dummy_X = np.random.randn(100, 18)  # 18 features
        dummy_y_return = np.random.randn(100) * 5  # Random returns
        dummy_y_profit = np.random.choice([0, 1], 100)  # Random profit/loss
        
        # Fit models with dummy data
        self.scalers["feature_scaler"].fit(dummy_X)
        dummy_X_scaled = self.scalers["feature_scaler"].transform(dummy_X)
        
        self.models["performance_model"].fit(dummy_X_scaled, dummy_y_return)
        self.models["profit_model"].fit(dummy_X_scaled, dummy_y_profit)
        
        # Set feature importance
        feature_names = [f"feature_{i}" for i in range(18)]
        self.feature_importance["performance_model"] = dict(
            zip(feature_names, self.models["performance_model"].feature_importances_)
        )
        
        self.model_versions["performance_model"] = "pretrained_1.0"
        self.is_trained = True
        
        logger.info("Pretrained models loaded")

    async def _make_predictions(
        self,
        features: Dict[str, float],
        horizon_days: int
    ) -> Dict[str, float]:
        """
        Make predictions using trained models
        """
        try:
            # Convert features to array
            feature_names = [
                "total_pnl", "win_rate", "avg_win", "avg_loss", "profit_factor",
                "volatility", "momentum", "max_consecutive_losses", "drawdown",
                "avg_trades_per_day", "trade_size_variance", "unique_assets",
                "asset_concentration", "most_active_hour", "weekend_trading_ratio",
                "recent_win_rate", "recent_avg_pnl", "total_trades"
            ]
            
            feature_array = np.array([[features.get(name, 0) for name in feature_names]])
            
            # Scale features
            feature_array_scaled = self.scalers["feature_scaler"].transform(feature_array)
            
            # Make predictions
            predicted_return = self.models["performance_model"].predict(feature_array_scaled)[0]
            profit_probability = self.models["profit_model"].predict_proba(feature_array_scaled)[0][1]
            
            # Adjust for time horizon
            adjusted_return = predicted_return * (horizon_days / 7)  # Scale to horizon
            
            # Estimate max drawdown (simple heuristic)
            volatility = features.get("volatility", 0)
            estimated_drawdown = abs(adjusted_return) * 0.3 + volatility * 0.1
            
            return {
                "return_pct": adjusted_return,
                "profit_probability": profit_probability,
                "max_drawdown_pct": estimated_drawdown
            }

        except Exception as e:
            logger.error(f"Error making predictions: {e}")
            return {
                "return_pct": 0.0,
                "profit_probability": 0.5,
                "max_drawdown_pct": 5.0
            }

    async def _calculate_prediction_confidence(
        self,
        features: Dict[str, float],
        predictions: Dict[str, float]
    ) -> float:
        """
        Calculate confidence in predictions based on feature quality and model certainty
        """
        try:
            confidence = 0.5  # Base confidence
            
            # Increase confidence based on data quality
            total_trades = features.get("total_trades", 0)
            trading_days = features.get("trading_days", 0)
            
            if total_trades > 50:
                confidence += 0.1
            if total_trades > 100:
                confidence += 0.1
            
            if trading_days > 30:
                confidence += 0.1
            if trading_days > 60:
                confidence += 0.1
                
            # Adjust based on prediction consistency
            profit_probability = predictions.get("profit_probability", 0.5)
            if profit_probability > 0.7 or profit_probability < 0.3:
                confidence += 0.1  # More confident in extreme probabilities
            
            # Penalize for high volatility
            volatility = features.get("volatility", 0)
            if volatility > 10:
                confidence -= 0.1
                
            return min(max(confidence, 0.0), 1.0)  # Clamp between 0 and 1

        except Exception as e:
            logger.error(f"Error calculating confidence: {e}")
            return 0.5

    async def retrain_models(self):
        """
        Retrain models with new data
        """
        logger.info("Starting model retraining...")
        self.is_trained = False
        await self._train_models()

    async def predict_market_regime(
        self,
        market_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Predict market regime (bull/bear/sideways)
        """
        try:
            # Simplified market regime prediction
            # In reality, would use more sophisticated models
            
            volatility = market_data.get("volatility", 20)
            momentum = market_data.get("momentum", 0)
            
            if momentum > 5 and volatility < 25:
                regime = "bull"
                confidence = 0.8
            elif momentum < -5 and volatility > 30:
                regime = "bear"
                confidence = 0.8
            else:
                regime = "sideways"
                confidence = 0.6
            
            return {
                "regime": regime,
                "confidence": confidence,
                "volatility_forecast": volatility * 1.1,
                "momentum_forecast": momentum * 0.8,
                "prediction_horizon_days": 14
            }

        except Exception as e:
            logger.error(f"Error predicting market regime: {e}")
            return {
                "regime": "sideways",
                "confidence": 0.5,
                "volatility_forecast": 20.0,
                "momentum_forecast": 0.0,
                "prediction_horizon_days": 14
            }

    async def optimize_portfolio_allocation(
        self,
        leaders: List[str],
        risk_tolerance: float,
        target_return: float
    ) -> Dict[str, float]:
        """
        Optimize portfolio allocation using ML
        """
        try:
            # Simplified portfolio optimization
            # In reality, would use modern portfolio theory with ML enhancements
            
            num_leaders = len(leaders)
            if num_leaders == 0:
                return {}
            
            # Base equal allocation
            base_weight = 1.0 / num_leaders
            
            # Adjust based on risk tolerance
            allocation = {}
            for leader in leaders:
                allocation[leader] = base_weight
            
            return allocation

        except Exception as e:
            logger.error(f"Error optimizing portfolio: {e}")
            return {}

    async def health_check(self) -> bool:
        """
        Check ML model health
        """
        try:
            if not self.is_trained:
                return False
            
            # Test prediction with dummy data
            dummy_features = {
                "total_pnl": 1000,
                "win_rate": 0.6,
                "avg_win": 50,
                "avg_loss": 30,
                "profit_factor": 1.5,
                "volatility": 15,
                "momentum": 2,
                "max_consecutive_losses": 3,
                "drawdown": 5,
                "avg_trades_per_day": 2,
                "trade_size_variance": 100,
                "unique_assets": 3,
                "asset_concentration": 0.4,
                "most_active_hour": 14,
                "weekend_trading_ratio": 0.2,
                "recent_win_rate": 0.7,
                "recent_avg_pnl": 25,
                "total_trades": 100
            }
            
            predictions = await self._make_predictions(dummy_features, 7)
            
            # Check if predictions are reasonable
            if (predictions["return_pct"] is not None and 
                0 <= predictions["profit_probability"] <= 1):
                return True
            
            return False

        except Exception as e:
            logger.error(f"ML health check failed: {e}")
            return False

    async def get_model_info(self) -> Dict[str, Any]:
        """
        Get information about loaded models
        """
        return {
            "is_trained": self.is_trained,
            "models": list(self.models.keys()),
            "model_versions": self.model_versions,
            "feature_importance": self.feature_importance,
            "last_training": datetime.utcnow().isoformat()  # Would track actual training time
        }

    async def save_models(self, filepath: str):
        """
        Save trained models to disk
        """
        try:
            model_data = {
                "models": self.models,
                "scalers": self.scalers,
                "feature_importance": self.feature_importance,
                "model_versions": self.model_versions,
                "trained_at": datetime.utcnow()
            }
            
            with open(filepath, 'wb') as f:
                pickle.dump(model_data, f)
                
            logger.info(f"Models saved to {filepath}")

        except Exception as e:
            logger.error(f"Error saving models: {e}")

    async def load_models(self, filepath: str):
        """
        Load trained models from disk
        """
        try:
            with open(filepath, 'rb') as f:
                model_data = pickle.load(f)
            
            self.models = model_data.get("models", {})
            self.scalers = model_data.get("scalers", {})
            self.feature_importance = model_data.get("feature_importance", {})
            self.model_versions = model_data.get("model_versions", {})
            self.is_trained = len(self.models) > 0
            
            logger.info(f"Models loaded from {filepath}")

        except Exception as e:
            logger.error(f"Error loading models: {e}")
