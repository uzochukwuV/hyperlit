# Hyperlit Project Memory

## Current Project Context

### Hyperliquid Community Hackathon
**Building on the blockchain to house all of finance**

- **Event Type**: 4-week global online hackathon
- **Prize Pool**: $250,000 total
- **Participants**: 417 registered

#### Timeline
- **Week 3 - Building**: Aug 11, 2025 - 15:00
- **Week 4 - Submissions and Closing Ceremony**: Aug 18, 2025 - 15:00
- **Demo Day**: Aug 20, 2025
- **Project Submission Deadline**: Aug 24, 2025
- **Winners Announced**: Sep 5, 2025

#### Core Tracks
1. **🛠️ Public Goods Track**
   - Build open and reusable tools, infrastructure, resources
   - Examples: SDKs, indexers, educational materials
   - Must be open source
   - Judged on: Quality, Usefulness to Ecosystem, Documentation & Reusability

2. **🚀 Hyperliquid Frontier Track** 
   - Grow the Hyperliquid ecosystem technically, socially, culturally
   - Build dApps, integrations, communities, content
   - Judged on: Quality and Ecosystem Impact

3. **🧪 Precompiles + CoreWriter Exploration Track**
   - Push boundaries across HyperCore and HyperEVM
   - Unlock new use cases, performance gains
   - Judged on: Quality, Novelty, and Technicality

#### Prize Structure (per track)
- 🥇 1st Place: $30,000
- 🥈 2nd Place: $15,000  
- 🥉 3rd Place: $5,000

#### Notable Bounties
- HyperEVM Transaction Simulator: $30,000
- HyperEVM RPC Improvements: $20,000
- Best use of GlueX Router API: $7,000
- Best use of Liquid Labs: $7,000
- Various smaller bounties from $1,000-$5,000

#### Key Requirements
- Projects must align with at least one core track
- All submissions need demo + judge access to materials
- Public Goods Track must be open source
- Winners must complete KYC/KYB
- Prizes paid in USDC/USDT on Hyperliquid

## Project Development Status

### Hyperliquid Copy Trading Platform
**Architecture**: Multi-service with Go and Python components

#### Critical Bug Fixes Implemented
1. **Asset ID Mapping Fix**
   - Fixed perpetuals: return index directly for perps
   - Fixed spot markets: use formula `10000 + pair.Index`
   - Added delisted asset validation

2. **EIP-712 Signature Implementation**
   - Corrected domain structure for Hyperliquid
   - Added proper Arbitrum Chain ID (42161)
   - Fixed signature format with r, s, v values

3. **URL Configuration Corrections**
   - Mainnet API: `https://api.hyperliquid.xyz`
   - Mainnet WebSocket: `wss://api.hyperliquid.xyz/ws`
   - Testnet API: `https://api.hyperliquid-testnet.xyz`
   - Testnet WebSocket: `wss://api.hyperliquid-testnet.xyz/ws`

#### Advanced Features

##### Permissionless Copy Trading System
**Innovation**: Copy any trader's address without requiring registration

**Core Components**:
- `PermissionlessCopyEngine`: Main orchestration engine
- Trader discovery and performance analysis
- Real-time WebSocket monitoring
- Intelligent copy filters and risk management

**Key Features**:
- **Auto-Discovery**: Automatically finds high-performing traders
- **Smart Filtering**: Asset whitelist/blacklist, position size limits, time restrictions
- **Risk Management**: Copy percentage, max position size, slippage protection
- **Performance Tracking**: Win rate, Sharpe ratio, drawdown analysis

**Copy Filters Available**:
- Position value limits (min/max)
- Time-based execution delays
- Trading hour restrictions
- Asset filtering (whitelist/blacklist)
- Leverage limits
- Profitability filters

**Technical Implementation**:
- WebSocket subscriptions to `userFills`, `orderUpdates`, `userEvents`
- Real-time trade replication with configurable delays
- Database tracking of all copy trades
- Performance analytics for trader recommendations

#### Enhanced Configuration
- Rate limiting aligned with Hyperliquid limits (1200 req/min, 2000 WS msg/min)
- Proper connection pooling and retry logic
- Security features for API key management
- Multi-environment support (mainnet/testnet)

#### Smart Order Execution
- Market impact analysis
- TWAP execution for large orders
- Slippage protection
- Price improvement tracking
- Partial execution handling

#### Analytics & Insights
- Deep trader performance analysis
- Risk metrics calculation
- Trading pattern recognition
- Market condition adaptation
- Seasonal performance tracking