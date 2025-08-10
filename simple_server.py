
import asyncio
import json
import os
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional

from fastapi import FastAPI, HTTPException, Request
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse, JSONResponse
from pydantic import BaseModel
import uvicorn

app = FastAPI(title="Hyperliquid Copy Trading")

# Serve static frontend files
if Path("frontend").exists():
    app.mount("/static", StaticFiles(directory="frontend"), name="static")

# Mock data for demonstration
MOCK_LEADERS = [
    {
        "address": "0x1234...5678",
        "name": "AlphaTrder",
        "total_followers": 150,
        "win_rate": 0.72,
        "pnl_30d": 15420.50,
        "max_drawdown": -8.5,
        "is_active": True,
        "sharpe_ratio": 1.85,
        "total_volume": 2450000
    },
    {
        "address": "0xabcd...efgh", 
        "name": "CryptoKing",
        "total_followers": 89,
        "win_rate": 0.68,
        "pnl_30d": 8975.25,
        "max_drawdown": -12.3,
        "is_active": True,
        "sharpe_ratio": 1.42,
        "total_volume": 1850000
    },
    {
        "address": "0x9876...5432",
        "name": "DegenMaster",
        "total_followers": 234,
        "win_rate": 0.81,
        "pnl_30d": 32150.75,
        "max_drawdown": -15.2,
        "is_active": True,
        "sharpe_ratio": 2.15,
        "total_volume": 3200000
    }
]

MOCK_TRADES = [
    {
        "id": 1,
        "leader_address": "0x1234...5678",
        "asset": "ETH",
        "side": "buy",
        "size": 10.5,
        "price": 3450.25,
        "status": "filled",
        "executed_at": "2024-01-10T14:30:00Z",
        "is_leader_trade": True
    },
    {
        "id": 2,
        "leader_address": "0xabcd...efgh",
        "asset": "BTC", 
        "side": "sell",
        "size": 0.25,
        "price": 45250.00,
        "status": "filled",
        "executed_at": "2024-01-10T13:45:00Z",
        "is_leader_trade": False
    }
]

@app.get("/")
async def root():
    """Serve the main frontend page"""
    if Path("frontend/index.html").exists():
        return FileResponse("frontend/index.html")
    return {"message": "Hyperliquid Copy Trading API", "status": "running"}

@app.get("/api/v1/health")
async def health_check():
    """Health check endpoint"""
    return {
        "success": True,
        "timestamp": datetime.utcnow().isoformat(),
        "data": {
            "status": "healthy",
            "services": {
                "database": True,
                "websocket": True,
                "analytics": True
            }
        }
    }

@app.get("/api/v1/leaders")
async def get_leaders():
    """Get top trading leaders"""
    return {
        "success": True,
        "data": MOCK_LEADERS
    }

@app.get("/api/v1/trades")
async def get_trades(limit: int = 50):
    """Get recent trades"""
    return {
        "success": True,
        "data": {
            "trades": MOCK_TRADES[:limit],
            "total": len(MOCK_TRADES)
        }
    }

@app.get("/api/v1/followers")
async def get_followers():
    """Get user's active follows"""
    mock_followers = [
        {
            "leader_address": "0x1234...5678",
            "leader_name": "AlphaTrder", 
            "copy_percentage": 5.0,
            "max_position": 1000.0,
            "is_active": True,
            "created_at": "2024-01-01T00:00:00Z"
        }
    ]
    return {
        "success": True,
        "data": mock_followers
    }

@app.get("/api/v1/analytics/performance")
async def get_performance():
    """Get performance analytics"""
    return {
        "success": True,
        "data": {
            "total_pnl": 12450.75,
            "win_rate": 0.74,
            "sharpe_ratio": 1.65,
            "max_drawdown": -8.2,
            "daily_returns": [
                {"date": "2024-01-01", "return": 150.25},
                {"date": "2024-01-02", "return": 225.50},
                {"date": "2024-01-03", "return": -75.30},
                {"date": "2024-01-04", "return": 320.15},
                {"date": "2024-01-05", "return": 180.75}
            ]
        }
    }

if __name__ == "__main__":
    print("🚀 Starting Hyperliquid Copy Trading Platform")
    print("📊 Frontend: http://localhost:5000")  
    print("🔌 API: http://localhost:5000/api/v1/health")
    
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=5000,
        log_level="info",
        reload=False
    )
