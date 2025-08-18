# Hyperliquid Copy Trading Setup Guide

## Prerequisites

1. **Go 1.21+** installed on your system
2. **PostgreSQL** database with TimescaleDB extension
3. **Ethereum/Arbitrum wallet** with private key for trading
4. **Hyperliquid account** (optional for testing)

## Quick Start

### 1. Clone and Setup Environment

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your configuration
nano .env
```

### 2. Configure Database

```bash
# Create PostgreSQL database
createdb copytrading

# Initialize the database schema
psql -d copytrading -f sql/init.sql
```

### 3. Set Your Private Key

**IMPORTANT**: You need to set your wallet private key in the `.env` file:

```env
API_WALLET_PRIVATE_KEY=your_private_key_without_0x_prefix
```

**Security Notes:**
- Never commit your private key to version control
- Use a dedicated trading wallet for copy trading
- Consider using environment variables in production
- For testing, you can generate a new wallet

### 4. Build and Run

```bash
# Install dependencies
go mod tidy

# Build the application
go build -o hyperlit.exe ./main.go

# Run the application
./hyperlit.exe
```

## Configuration Options

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `API_WALLET_PRIVATE_KEY` | Your wallet private key (without 0x) | `abc123...` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost/db` |

### Optional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HYPERLIQUID_API_URL` | `https://api.hyperliquid.xyz` | Hyperliquid API endpoint |
| `ENVIRONMENT` | `development` | Environment mode |
| `LOG_LEVEL` | `info` | Logging level |
| `MAX_FOLLOWERS_PER_LEADER` | `100` | Max followers per trader |
| `SIGNATURE_CHAIN_ID` | `42161` | Arbitrum chain ID |

## Features

### Traditional Copy Trading
- Follow registered traders
- Configurable copy percentages
- Risk management settings
- Real-time trade copying

### Permissionless Copy Trading (NEW)
- Copy ANY trader's address without registration
- Auto-discovery of high-performing traders
- Advanced filtering and risk management
- Smart order execution with slippage protection

## API Endpoints

### Traditional Copy Trading
- `POST /api/v1/followers` - Create follower
- `GET /api/v1/followers` - List followers
- `GET /api/v1/leaders` - List leaders
- `GET /api/v1/trades` - Get trade history

### Permissionless Copy Trading
- `POST /api/v1/permissionless/followers` - Follow any trader
- `GET /api/v1/permissionless/traders` - Discover traders
- `GET /api/v1/permissionless/leaderboard` - Top performers
- `GET /api/v1/permissionless/insights` - AI analytics

## Security

- Private keys are encrypted in the database
- Row-level security policies
- Rate limiting aligned with Hyperliquid
- Input validation and sanitization

## Support

For issues and feature requests, please check the project documentation or create an issue in the repository.