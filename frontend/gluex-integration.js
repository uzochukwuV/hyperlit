// GlueX Router API Integration for Hyperliquid Copy Trading
class GlueXIntegration {
    constructor(apiKey) {
        this.apiKey = apiKey;
        this.baseURL = 'https://api.gluex.io/v1';
        this.supportedChains = ['ethereum', 'arbitrum', 'polygon', 'bsc', 'optimism'];
    }

    // 1. CROSS-CHAIN COPY TRADING
    async executeCrosschainCopyTrade(leaderTrade, followerWallet) {
        const { asset, amount, targetChain } = leaderTrade;
        const { chain: followerChain, availableAssets } = followerWallet;

        // Find best route to replicate trade across chains
        const route = await this.findOptimalRoute({
            fromChain: followerChain,
            toChain: targetChain,
            fromAsset: availableAssets[0], // Use available asset
            toAsset: asset,
            amount: amount
        });

        return {
            route,
            estimatedGas: route.gasEstimate,
            priceImpact: route.priceImpact,
            executionTime: route.estimatedTime
        };
    }

    // 2. YIELD-OPTIMIZED IDLE FUNDS
    async optimizeIdleFunds(walletBalance) {
        const strategies = await Promise.all([
            this.getBestLendingRates(walletBalance.usdc),
            this.getBestStakingOptions(walletBalance.eth),
            this.getBestLPOpportunities(walletBalance)
        ]);

        return {
            recommended: strategies.sort((a, b) => b.apy - a.apy)[0],
            alternatives: strategies.slice(1),
            autoCompoundOptions: strategies.filter(s => s.autoCompound)
        };
    }

    // 3. INTELLIGENT ASSET CONVERSION
    async convertForCopyTrade(fromAsset, toAsset, amount, maxSlippage = 0.5) {
        const routes = await this.getRoutes({
            fromToken: fromAsset,
            toToken: toAsset,
            amount: amount,
            maxSlippage: maxSlippage,
            includeProtocols: ['uniswap', '1inch', 'paraswap', 'cowswap']
        });

        // Select best route based on output amount and gas cost
        const bestRoute = routes.reduce((best, current) => {
            const bestValue = best.outputAmount - (best.gasEstimate * best.gasPrice);
            const currentValue = current.outputAmount - (current.gasEstimate * current.gasPrice);
            return currentValue > bestValue ? current : best;
        });

        return bestRoute;
    }

    // 4. PROFIT REINVESTMENT AUTOMATION
    async createReinvestmentStrategy(profits, riskProfile) {
        const strategies = {
            conservative: {
                lending: 70,    // 70% to lending protocols
                staking: 20,    // 20% to ETH staking
                trading: 10     // 10% keep for copy trading
            },
            moderate: {
                lending: 40,
                staking: 30,
                lp: 20,
                trading: 10
            },
            aggressive: {
                lending: 20,
                staking: 20,
                lp: 40,
                trading: 20
            }
        };

        const allocation = strategies[riskProfile];
        const reinvestmentPlan = [];

        for (const [strategy, percentage] of Object.entries(allocation)) {
            const amount = (profits * percentage) / 100;
            const route = await this.findBestProtocol(strategy, amount);
            reinvestmentPlan.push({ strategy, amount, route });
        }

        return reinvestmentPlan;
    }

    // 5. MULTI-PROTOCOL ANALYTICS
    async getPortfolioAnalytics(walletAddress) {
        const positions = await this.getAllPositions(walletAddress);
        
        return {
            totalValue: positions.reduce((sum, p) => sum + p.value, 0),
            yieldEarning: positions.filter(p => p.type === 'yield').reduce((sum, p) => sum + p.apy * p.value, 0),
            protocolDistribution: this.groupBy(positions, 'protocol'),
            riskMetrics: await this.calculateRiskMetrics(positions),
            rebalanceRecommendations: await this.getRebalanceRecommendations(positions)
        };
    }

    // CORE API METHODS
    async findOptimalRoute(params) {
        const response = await fetch(`${this.baseURL}/route`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.apiKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(params)
        });
        return response.json();
    }

    async getBestLendingRates(amount) {
        const response = await fetch(`${this.baseURL}/lending/rates?amount=${amount}`, {
            headers: { 'Authorization': `Bearer ${this.apiKey}` }
        });
        return response.json();
    }

    async getBestStakingOptions(amount) {
        const response = await fetch(`${this.baseURL}/staking/options?amount=${amount}`, {
            headers: { 'Authorization': `Bearer ${this.apiKey}` }
        });
        return response.json();
    }

    async getBestLPOpportunities(balance) {
        const response = await fetch(`${this.baseURL}/liquidity/opportunities`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.apiKey}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ balance })
        });
        return response.json();
    }

    // UTILITY METHODS
    groupBy(array, key) {
        return array.reduce((result, item) => {
            const group = item[key];
            result[group] = result[group] || [];
            result[group].push(item);
            return result;
        }, {});
    }
}

// Integration with existing copy trading app
class EnhancedCopyTradingApp extends CopyTradingApp {
    constructor() {
        super();
        this.gluex = new GlueXIntegration(process.env.GLUEX_API_KEY);
        this.yieldOptimization = true;
        this.crossChainEnabled = true;
    }

    // Override the original trade execution with GlueX routing
    async executeCopyTrade(leaderTrade, followerConfig) {
        const follower = await this.getFollowerWallet(followerConfig.walletAddress);
        
        // Check if we need cross-chain execution
        if (leaderTrade.chain !== follower.chain) {
            return await this.executeCrosschainTrade(leaderTrade, follower);
        }

        // Check if we need asset conversion
        if (!follower.hasAsset(leaderTrade.asset)) {
            const conversion = await this.gluex.convertForCopyTrade(
                follower.primaryAsset,
                leaderTrade.asset,
                leaderTrade.amount * followerConfig.copyPercentage
            );
            
            // Execute conversion first, then the trade
            await this.executeConversion(conversion);
        }

        // Execute the copy trade with optimal routing
        return await this.executeOptimizedTrade(leaderTrade, followerConfig);
    }

    // New: Yield optimization for idle funds
    async optimizeIdleFunds() {
        const followers = await this.getActiveFollowers();
        
        for (const follower of followers) {
            const idleFunds = await this.calculateIdleFunds(follower);
            if (idleFunds.total > 100) { // Only optimize if > $100
                const strategy = await this.gluex.optimizeIdleFunds(idleFunds);
                await this.executeYieldStrategy(follower, strategy.recommended);
            }
        }
    }

    // New: Profit reinvestment automation
    async autoReinvestProfits() {
        const followers = await this.getActiveFollowers();
        
        for (const follower of followers) {
            const profits = await this.calculateProfits(follower, '24h');
            if (profits > follower.reinvestmentThreshold) {
                const plan = await this.gluex.createReinvestmentStrategy(
                    profits, 
                    follower.riskProfile
                );
                await this.executeReinvestmentPlan(follower, plan);
            }
        }
    }
}

// Export for use in main app
window.GlueXIntegration = GlueXIntegration;
window.EnhancedCopyTradingApp = EnhancedCopyTradingApp;