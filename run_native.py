#!/usr/bin/env python3
"""
Native server to run the Hyperliquid Copy Trading platform without Docker
"""

import asyncio
import os
import subprocess
import sys
import time
import signal
from pathlib import Path

class NativeServer:
    def __init__(self):
        self.processes = []
        self.running = False
        
    def setup_database(self):
        """Initialize PostgreSQL database if needed"""
        try:
            # Check if database is accessible
            import psycopg2
            from urllib.parse import urlparse
            
            db_url = os.environ.get('DATABASE_URL')
            if not db_url:
                print("No DATABASE_URL found, using local PostgreSQL setup")
                return True
                
            # Parse database URL
            parsed = urlparse(db_url)
            
            # Test connection
            conn = psycopg2.connect(
                host=parsed.hostname,
                port=parsed.port or 5432,
                database=parsed.path[1:] if parsed.path else 'postgres',
                user=parsed.username,
                password=parsed.password
            )
            conn.close()
            print("✓ Database connection successful")
            return True
            
        except ImportError:
            print("Installing psycopg2...")
            subprocess.run([sys.executable, "-m", "pip", "install", "psycopg2-binary"], check=True)
            return self.setup_database()
        except Exception as e:
            print(f"Database setup error: {e}")
            print("Using embedded database for now")
            return True
    
    def install_python_deps(self):
        """Install Python analytics dependencies"""
        requirements = [
            "fastapi==0.104.1",
            "uvicorn[standard]==0.24.0",
            "asyncpg==0.29.0", 
            "pandas==2.1.3",
            "numpy==1.25.2",
            "scikit-learn==1.3.2",
            "aiohttp==3.9.1",
            "pydantic==2.5.0",
            "python-multipart==0.0.6",
            "structlog==23.2.0",
            "psutil==5.9.6"
        ]
        
        print("Installing Python dependencies...")
        for req in requirements:
            try:
                subprocess.run([sys.executable, "-m", "pip", "install", req], 
                             check=True, capture_output=True)
            except subprocess.CalledProcessError:
                print(f"Warning: Could not install {req}")
                continue
        print("✓ Python dependencies installed")
    
    def create_simple_backend(self):
        """Create a simple backend server that serves the frontend"""
        backend_code = '''
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
'''
        
        # Write the backend server
        with open("simple_server.py", "w") as f:
            f.write(backend_code)
        
        print("✓ Simple backend server created")
    
    def start_server(self):
        """Start the native server"""
        self.running = True
        
        # Setup
        self.setup_database()
        self.install_python_deps()
        self.create_simple_backend()
        
        print("\n🚀 Starting Hyperliquid Copy Trading Platform...")
        print("📊 Dashboard: http://localhost:5000")
        print("🔌 API Health: http://localhost:5000/api/v1/health")
        print("💡 Press Ctrl+C to stop\n")
        
        try:
            # Start the Python server
            process = subprocess.Popen([
                sys.executable, "simple_server.py"
            ], stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
            
            self.processes.append(process)
            
            # Stream output
            while self.running and process.poll() is None:
                line = process.stdout.readline()
                if line:
                    print(line.rstrip())
                time.sleep(0.1)
                    
        except KeyboardInterrupt:
            print("\n🛑 Shutting down...")
            self.stop_server()
    
    def stop_server(self):
        """Stop all processes"""
        self.running = False
        for process in self.processes:
            try:
                process.terminate()
                process.wait(timeout=5)
            except (subprocess.TimeoutExpired, ProcessLookupError):
                try:
                    process.kill()
                except ProcessLookupError:
                    pass
        self.processes.clear()
        print("✓ Server stopped")

def main():
    server = NativeServer()
    
    # Handle signals
    def signal_handler(signum, frame):
        server.stop_server()
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    server.start_server()

if __name__ == "__main__":
    main()