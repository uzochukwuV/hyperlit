// Hyperliquid Copy Trading Frontend Application
class CopyTradingApp {
    constructor() {
        this.apiBase = "http://localhost:8000/api/v1";
        this.currentSection = 'dashboard';
        this.charts = {};
        this.refreshInterval = null;
        this.isLoading = false;
        
        this.init();
    }

    async init() {
        this.setupNavigation();
        this.setupEventListeners();
        this.setupWebSocket();
        
        // Initial load
        await this.loadDashboard();
        this.startAutoRefresh();
    }

    // Navigation Management
    setupNavigation() {
        document.querySelectorAll('.nav-link[data-section]').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = link.getAttribute('data-section');
                this.showSection(section);
            });
        });
    }

    showSection(sectionName) {
        // Hide all sections
        document.querySelectorAll('.content-section').forEach(section => {
            section.classList.add('d-none');
        });
        
        // Show target section
        const targetSection = document.getElementById(sectionName);
        if (targetSection) {
            targetSection.classList.remove('d-none');
            this.currentSection = sectionName;
            
            // Update active nav
            document.querySelectorAll('.nav-link').forEach(link => {
                link.classList.remove('active');
            });
            document.querySelector(`[data-section="${sectionName}"]`).classList.add('active');
            
            // Load section data
            this.loadSectionData(sectionName);
        }
    }

    async loadSectionData(section) {
        switch (section) {
            case 'dashboard':
                await this.loadDashboard();
                break;
            case 'leaders':
                await this.loadLeaders();
                break;
            case 'followers':
                await this.loadFollowers();
                break;
            case 'trades':
                await this.loadTrades();
                break;
            case 'analytics':
                await this.loadAnalytics();
                break;
            case 'hyperliquid':
                await this.loadHyperliquid();
                break;
        }
    }

    // Dashboard
    async loadDashboard() {
        try {
            this.setLoading(true);
            
            // Load system status
            await this.loadSystemStatus();
            
            // Load portfolio performance
            await this.loadPortfolioPerformance();
            
            // Load recent activity
            await this.loadRecentActivity();
            
            // Load summary stats
            await this.loadSummaryStats();
            
        } catch (error) {
            console.error('Error loading dashboard:', error);
            this.showError('Failed to load dashboard data');
        } finally {
            this.setLoading(false);
        }
    }

    async loadSystemStatus() {
        try {
            const response = await fetch(`${this.apiBase}/health`);
            const data = await response.json();
            
            const statusElement = document.getElementById('system-status');
            if (data.success && data.data.services.database && data.data.services.websocket) {
                statusElement.innerHTML = '<i class="fas fa-circle"></i> Online';
                statusElement.className = 'text-success';
            } else {
                statusElement.innerHTML = '<i class="fas fa-circle"></i> Issues Detected';
                statusElement.className = 'text-warning';
            }
        } catch (error) {
            const statusElement = document.getElementById('system-status');
            statusElement.innerHTML = '<i class="fas fa-circle"></i> Offline';
            statusElement.className = 'text-danger';
        }
    }

    async loadSummaryStats() {
        try {
            const [followersRes, tradesRes] = await Promise.all([
                fetch(`${this.apiBase}/followers`),
                fetch(`${this.apiBase}/trades?limit=100`)
            ]);

            const followersData = await followersRes.json();
            const tradesData = await tradesRes.json();

            if (followersData.success) {
                const activeFollows = followersData.data.filter(f => f.is_active).length;
                document.getElementById('active-follows').textContent = activeFollows;
            }

            if (tradesData.success) {
                const trades = tradesData.data.trades || [];
                const filledTrades = trades.filter(t => t.status === 'filled');
                
                // Calculate P&L
                let totalPnL = 0;
                let profitableTrades = 0;
                
                filledTrades.forEach(trade => {
                    const pnl = trade.side === 'sell' ? 
                        trade.size * trade.price : 
                        -trade.size * trade.price;
                    totalPnL += pnl;
                    if (pnl > 0) profitableTrades++;
                });

                document.getElementById('total-pnl').textContent = `$${totalPnL.toFixed(2)}`;
                document.getElementById('total-pnl').className = totalPnL >= 0 ? 'text-success' : 'text-danger';
                
                const successRate = filledTrades.length > 0 ? 
                    (profitableTrades / filledTrades.length * 100).toFixed(1) : 0;
                document.getElementById('success-rate').textContent = `${successRate}%`;
            }
        } catch (error) {
            console.error('Error loading summary stats:', error);
        }
    }

    async loadPortfolioPerformance() {
        try {
            const response = await fetch(`${this.apiBase}/trades?limit=30`);
            const data = await response.json();
            
            if (data.success && data.data.trades) {
                this.createPortfolioChart(data.data.trades);
            }
        } catch (error) {
            console.error('Error loading portfolio performance:', error);
        }
    }

    createPortfolioChart(trades) {
        const ctx = document.getElementById('portfolioChart').getContext('2d');
        
        // Process trades for chart data
        const dailyPnL = this.processDailyPnL(trades);
        
        if (this.charts.portfolio) {
            this.charts.portfolio.destroy();
        }

        this.charts.portfolio = new Chart(ctx, {
            type: 'line',
            data: {
                labels: dailyPnL.labels,
                datasets: [{
                    label: 'Cumulative P&L ($)',
                    data: dailyPnL.cumulative,
                    borderColor: 'rgb(13, 110, 253)',
                    backgroundColor: 'rgba(13, 110, 253, 0.1)',
                    fill: true,
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: false,
                        ticks: {
                            callback: function(value) {
                                return '$' + value.toFixed(2);
                            }
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }

    processDailyPnL(trades) {
        const dailyData = {};
        
        trades.forEach(trade => {
            if (trade.status !== 'filled') return;
            
            const date = new Date(trade.executed_at).toISOString().split('T')[0];
            const pnl = trade.side === 'sell' ? 
                trade.size * trade.price : 
                -trade.size * trade.price;
            
            if (!dailyData[date]) {
                dailyData[date] = 0;
            }
            dailyData[date] += pnl;
        });

        const sortedDates = Object.keys(dailyData).sort();
        const labels = [];
        const cumulative = [];
        let runningTotal = 0;

        sortedDates.forEach(date => {
            runningTotal += dailyData[date];
            labels.push(new Date(date).toLocaleDateString());
            cumulative.push(runningTotal);
        });

        return { labels, cumulative };
    }

    async loadRecentActivity() {
        try {
            const response = await fetch(`${this.apiBase}/trades?limit=10`);
            const data = await response.json();
            
            const tbody = document.getElementById('recent-activity');
            
            if (data.success && data.data.trades) {
                const trades = data.data.trades.slice(0, 10);
                
                if (trades.length === 0) {
                    tbody.innerHTML = `
                        <tr>
                            <td colspan="8" class="text-center text-muted">
                                <i class="fas fa-inbox"></i> No recent activity
                            </td>
                        </tr>
                    `;
                    return;
                }

                tbody.innerHTML = trades.map(trade => `
                    <tr>
                        <td>${this.formatDateTime(trade.executed_at)}</td>
                        <td>${trade.is_leader_trade ? 'Leader' : 'Follow'}</td>
                        <td>${this.formatAddress(trade.leader_address)}</td>
                        <td><span class="badge bg-primary">${trade.asset}</span></td>
                        <td>
                            <span class="badge bg-${trade.side === 'buy' ? 'success' : 'danger'}">
                                ${trade.side.toUpperCase()}
                            </span>
                        </td>
                        <td>${this.formatNumber(trade.size)}</td>
                        <td>$${this.formatNumber(trade.price)}</td>
                        <td>
                            <span class="badge bg-${this.getStatusColor(trade.status)}">
                                ${trade.status.toUpperCase()}
                            </span>
                        </td>
                    </tr>
                `).join('');
            } else {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="8" class="text-center text-muted">
                            <i class="fas fa-exclamation-triangle"></i> Failed to load activity
                        </td>
                    </tr>
                `;
            }
        } catch (error) {
            console.error('Error loading recent activity:', error);
            const tbody = document.getElementById('recent-activity');
            tbody.innerHTML = `
                <tr>
                    <td colspan="8" class="text-center text-danger">
                        <i class="fas fa-exclamation-triangle"></i> Error loading data
                    </td>
                </tr>
            `;
        }
    }

    // Leaders
    async loadLeaders() {
        try {
            this.setLoading(true);
            
            const period = document.getElementById('period-filter')?.value || 30;
            const minFollowers = document.getElementById('min-followers')?.value || 5;
            
            const response = await fetch(`${this.apiBase}/leaders`);
            const data = await response.json();
            
            const grid = document.getElementById('leaders-grid');
            
            if (data.success && data.data) {
                const leaders = data.data;
                
                if (leaders.length === 0) {
                    grid.innerHTML = `
                        <div class="col-12 text-center text-muted">
                            <i class="fas fa-users fa-3x mb-3"></i>
                            <h5>No Leaders Found</h5>
                            <p>No leaders match your criteria. Try adjusting the filters.</p>
                        </div>
                    `;
                    return;
                }

                grid.innerHTML = leaders.map(leader => `
                    <div class="col-lg-4 col-md-6 mb-4">
                        <div class="card leader-card h-100">
                            <div class="card-body">
                                <div class="d-flex justify-content-between align-items-start mb-3">
                                    <h5 class="card-title">
                                        ${leader.name || this.formatAddress(leader.address)}
                                    </h5>
                                    <span class="badge bg-${leader.is_active ? 'success' : 'secondary'}">
                                        ${leader.is_active ? 'Active' : 'Inactive'}
                                    </span>
                                </div>
                                
                                <div class="row text-center mb-3">
                                    <div class="col-4">
                                        <small class="text-muted">Followers</small>
                                        <div class="fw-bold">${leader.total_followers}</div>
                                    </div>
                                    <div class="col-4">
                                        <small class="text-muted">Win Rate</small>
                                        <div class="fw-bold">${(leader.win_rate * 100).toFixed(1)}%</div>
                                    </div>
                                    <div class="col-4">
                                        <small class="text-muted">P&L (30d)</small>
                                        <div class="fw-bold text-${leader.pnl_30d >= 0 ? 'success' : 'danger'}">
                                            $${this.formatNumber(leader.pnl_30d)}
                                        </div>
                                    </div>
                                </div>
                                
                                <div class="mb-3">
                                    <small class="text-muted">Max Drawdown</small>
                                    <div class="progress mt-1">
                                        <div class="progress-bar bg-warning" role="progressbar" 
                                             style="width: ${Math.min(Math.abs(leader.max_drawdown), 50)}%">
                                        </div>
                                    </div>
                                    <small class="text-muted">${Math.abs(leader.max_drawdown).toFixed(1)}%</small>
                                </div>
                                
                                <div class="d-grid gap-2">
                                    <button class="btn btn-primary btn-sm" onclick="copyTrader('${leader.address}')">
                                        <i class="fas fa-copy"></i> Copy Trader
                                    </button>
                                    <button class="btn btn-outline-secondary btn-sm" onclick="viewLeaderDetails('${leader.address}')">
                                        <i class="fas fa-chart-line"></i> View Details
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                `).join('');
            } else {
                grid.innerHTML = `
                    <div class="col-12 text-center text-danger">
                        <i class="fas fa-exclamation-triangle fa-3x mb-3"></i>
                        <h5>Error Loading Leaders</h5>
                        <p>Failed to load leader data. Please try again.</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading leaders:', error);
            this.showError('Failed to load leaders');
        } finally {
            this.setLoading(false);
        }
    }

    // Followers
    async loadFollowers() {
        try {
            this.setLoading(true);
            
            const response = await fetch(`${this.apiBase}/followers`);
            const data = await response.json();
            
            const tbody = document.getElementById('followers-table');
            
            if (data.success && data.data) {
                const followers = data.data;
                
                if (followers.length === 0) {
                    tbody.innerHTML = `
                        <tr>
                            <td colspan="8" class="text-center text-muted">
                                <i class="fas fa-plus-circle fa-2x mb-2"></i>
                                <br>No copy trading positions yet
                                <br><small>Click "Add New Follow" to get started</small>
                            </td>
                        </tr>
                    `;
                    return;
                }

                tbody.innerHTML = followers.map(follower => `
                    <tr>
                        <td>
                            <div class="fw-bold">${this.formatAddress(follower.leader_address)}</div>
                            <small class="text-muted">Since ${this.formatDate(follower.created_at)}</small>
                        </td>
                        <td>${follower.copy_percentage.toFixed(1)}%</td>
                        <td>$${this.formatNumber(follower.max_position_size)}</td>
                        <td class="pnl-cell">
                            <span class="text-muted">
                                <i class="fas fa-spinner fa-spin"></i> Loading...
                            </span>
                        </td>
                        <td class="win-rate-cell">
                            <span class="text-muted">
                                <i class="fas fa-spinner fa-spin"></i> Loading...
                            </span>
                        </td>
                        <td>
                            <span class="badge bg-info">
                                ${this.calculateRiskLevel(follower)}
                            </span>
                        </td>
                        <td>
                            <span class="badge bg-${follower.is_active ? 'success' : 'secondary'}">
                                ${follower.is_active ? 'Active' : 'Paused'}
                            </span>
                        </td>
                        <td>
                            <div class="btn-group btn-group-sm">
                                <button class="btn btn-outline-primary" onclick="editFollower(${follower.id})" 
                                        title="Edit Settings">
                                    <i class="fas fa-edit"></i>
                                </button>
                                <button class="btn btn-outline-secondary" onclick="viewFollowerAnalytics(${follower.id})" 
                                        title="View Analytics">
                                    <i class="fas fa-chart-bar"></i>
                                </button>
                                <button class="btn btn-outline-danger" onclick="deleteFollower(${follower.id})" 
                                        title="Remove">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </td>
                    </tr>
                `).join('');

                // Load PnL data for each follower
                this.loadFollowerPnLData(followers);
            } else {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="8" class="text-center text-danger">
                            <i class="fas fa-exclamation-triangle"></i> Failed to load follows
                        </td>
                    </tr>
                `;
            }
        } catch (error) {
            console.error('Error loading followers:', error);
            this.showError('Failed to load followers');
        } finally {
            this.setLoading(false);
        }
    }

    async loadFollowerPnLData(followers) {
        for (const follower of followers) {
            try {
                const response = await fetch(`${this.apiBase}/analytics/${follower.id}/pnl?days=30`);
                const data = await response.json();
                
                if (data.success && data.data) {
                    const analytics = data.data;
                    
                    // Update P&L cell
                    const pnlCell = document.querySelector(`tr:nth-child(${followers.indexOf(follower) + 1}) .pnl-cell`);
                    if (pnlCell) {
                        pnlCell.innerHTML = `
                            <span class="text-${analytics.total_pnl >= 0 ? 'success' : 'danger'}">
                                $${this.formatNumber(analytics.total_pnl)}
                            </span>
                        `;
                    }
                    
                    // Update win rate cell
                    const winRateCell = document.querySelector(`tr:nth-child(${followers.indexOf(follower) + 1}) .win-rate-cell`);
                    if (winRateCell) {
                        winRateCell.innerHTML = `${(analytics.win_rate * 100).toFixed(1)}%`;
                    }
                }
            } catch (error) {
                console.error(`Error loading PnL for follower ${follower.id}:`, error);
            }
        }
    }

    // Trades
    async loadTrades() {
        try {
            this.setLoading(true);
            
            const period = document.getElementById('trade-period')?.value || 30;
            const status = document.getElementById('trade-status')?.value || '';
            
            let url = `${this.apiBase}/trades?limit=100`;
            if (status) url += `&status=${status}`;
            
            const response = await fetch(url);
            const data = await response.json();
            
            const tbody = document.getElementById('trades-table');
            
            if (data.success && data.data.trades) {
                const trades = data.data.trades;
                
                if (trades.length === 0) {
                    tbody.innerHTML = `
                        <tr>
                            <td colspan="10" class="text-center text-muted">
                                <i class="fas fa-inbox fa-2x mb-2"></i>
                                <br>No trades found
                            </td>
                        </tr>
                    `;
                    return;
                }

                tbody.innerHTML = trades.map(trade => {
                    const value = trade.size * trade.price;
                    const pnl = trade.side === 'sell' ? value : -value;
                    
                    return `
                        <tr>
                            <td>
                                <div>${this.formatDateTime(trade.executed_at)}</div>
                            </td>
                            <td>
                                <div class="fw-bold">${this.formatAddress(trade.leader_address)}</div>
                                <small class="text-muted">${trade.is_leader_trade ? 'Leader' : 'Follower'}</small>
                            </td>
                            <td>
                                <span class="badge bg-primary">${trade.asset}</span>
                            </td>
                            <td>
                                <span class="badge bg-${trade.side === 'buy' ? 'success' : 'danger'}">
                                    ${trade.side.toUpperCase()}
                                </span>
                            </td>
                            <td>${this.formatNumber(trade.size)}</td>
                            <td>$${this.formatNumber(trade.price)}</td>
                            <td>$${this.formatNumber(value)}</td>
                            <td class="text-${pnl >= 0 ? 'success' : 'danger'}">
                                $${this.formatNumber(pnl)}
                            </td>
                            <td>
                                <span class="badge bg-${this.getStatusColor(trade.status)}">
                                    ${trade.status.toUpperCase()}
                                </span>
                            </td>
                            <td>
                                ${trade.hyperliquid_tx_id ? 
                                    `<a href="https://app.hyperliquid.xyz/tx/${trade.hyperliquid_tx_id}" 
                                        target="_blank" class="btn btn-sm btn-outline-primary" title="View on Hyperliquid">
                                        <i class="fas fa-external-link-alt"></i>
                                    </a>` : 
                                    '<span class="text-muted">-</span>'
                                }
                            </td>
                        </tr>
                    `;
                }).join('');
            } else {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="10" class="text-center text-danger">
                            <i class="fas fa-exclamation-triangle"></i> Failed to load trades
                        </td>
                    </tr>
                `;
            }
        } catch (error) {
            console.error('Error loading trades:', error);
            this.showError('Failed to load trades');
        } finally {
            this.setLoading(false);
        }
    }

    // Analytics
    async loadAnalytics() {
        try {
            this.setLoading(true);
            
            // Load AI recommendations
            await this.loadAIRecommendations();
            
            // Load performance charts
            await this.loadPerformanceMetrics();
            
            // Load risk analysis
            await this.loadRiskAnalysis();
            
            // Load optimization
            await this.loadOptimization();
            
        } catch (error) {
            console.error('Error loading analytics:', error);
            this.showError('Failed to load analytics');
        } finally {
            this.setLoading(false);
        }
    }

    async loadAIRecommendations() {
        const container = document.getElementById('ai-recommendations');
        
        try {
            // Get a follower to generate recommendations for
            const followersRes = await fetch(`${this.apiBase}/followers`);
            const followersData = await followersRes.json();
            
            if (followersData.success && followersData.data.length > 0) {
                const follower = followersData.data[0];
                
                // Note: This endpoint may not be implemented yet
                try {
                    const response = await fetch(`${this.apiBase}/analytics/recommendations/${follower.id}`);
                    const data = await response.json();
                    
                    if (data.success && data.data) {
                        const recommendations = data.data;
                        container.innerHTML = recommendations.map(rec => `
                            <div class="recommendation-item mb-3">
                                <div class="d-flex justify-content-between">
                                    <strong>${rec.asset} ${rec.action.toUpperCase()}</strong>
                                    <span class="badge bg-${rec.confidence > 0.7 ? 'success' : 'warning'}">
                                        ${(rec.confidence * 100).toFixed(0)}%
                                    </span>
                                </div>
                                <small class="text-muted">${rec.reasoning}</small>
                                <div class="mt-1">
                                    <small>Expected: ${rec.expected_return_pct.toFixed(1)}% | Risk: ${rec.risk_pct.toFixed(1)}%</small>
                                </div>
                            </div>
                        `).join('');
                    } else {
                        throw new Error('No recommendations available');
                    }
                } catch (error) {
                    // Fallback to mock recommendations
                    container.innerHTML = `
                        <div class="alert alert-info">
                            <i class="fas fa-info-circle"></i>
                            <strong>AI Analytics Coming Soon</strong>
                            <p class="mb-0 mt-2">Advanced ML-powered recommendations will be available once sufficient trading data is collected.</p>
                        </div>
                    `;
                }
            } else {
                container.innerHTML = `
                    <div class="alert alert-warning">
                        <i class="fas fa-exclamation-triangle"></i>
                        <strong>No Active Follows</strong>
                        <p class="mb-0 mt-2">Add some copy trading positions to receive AI recommendations.</p>
                    </div>
                `;
            }
        } catch (error) {
            container.innerHTML = `
                <div class="alert alert-danger">
                    <i class="fas fa-exclamation-triangle"></i>
                    <strong>Error</strong>
                    <p class="mb-0 mt-2">Unable to load AI recommendations at this time.</p>
                </div>
            `;
        }
    }

    async loadPerformanceMetrics() {
        const ctx = document.getElementById('performanceChart').getContext('2d');
        
        try {
            const response = await fetch(`${this.apiBase}/trades?limit=50`);
            const data = await response.json();
            
            if (data.success && data.data.trades) {
                const trades = data.data.trades;
                const performanceData = this.calculatePerformanceMetrics(trades);
                
                if (this.charts.performance) {
                    this.charts.performance.destroy();
                }

                this.charts.performance = new Chart(ctx, {
                    type: 'line',
                    data: {
                        labels: performanceData.labels,
                        datasets: [{
                            label: 'Cumulative Return %',
                            data: performanceData.returns,
                            borderColor: 'rgb(40, 167, 69)',
                            backgroundColor: 'rgba(40, 167, 69, 0.1)',
                            fill: true
                        }, {
                            label: 'Drawdown %',
                            data: performanceData.drawdown,
                            borderColor: 'rgb(220, 53, 69)',
                            backgroundColor: 'rgba(220, 53, 69, 0.1)',
                            fill: true
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        scales: {
                            y: {
                                ticks: {
                                    callback: function(value) {
                                        return value.toFixed(1) + '%';
                                    }
                                }
                            }
                        }
                    }
                });
            }
        } catch (error) {
            console.error('Error loading performance metrics:', error);
        }
    }

    async loadRiskAnalysis() {
        const ctx = document.getElementById('riskChart').getContext('2d');
        
        try {
            // Mock risk data for now
            const riskData = {
                labels: ['Very Low', 'Low', 'Medium', 'High', 'Very High'],
                datasets: [{
                    label: 'Risk Distribution',
                    data: [15, 25, 35, 20, 5],
                    backgroundColor: [
                        'rgba(40, 167, 69, 0.8)',
                        'rgba(255, 193, 7, 0.8)',
                        'rgba(13, 110, 253, 0.8)',
                        'rgba(255, 133, 27, 0.8)',
                        'rgba(220, 53, 69, 0.8)'
                    ]
                }]
            };

            if (this.charts.risk) {
                this.charts.risk.destroy();
            }

            this.charts.risk = new Chart(ctx, {
                type: 'doughnut',
                data: riskData,
                options: {
                    responsive: true,
                    plugins: {
                        legend: {
                            position: 'bottom'
                        }
                    }
                }
            });
        } catch (error) {
            console.error('Error loading risk analysis:', error);
        }
    }

    async loadOptimization() {
        const container = document.getElementById('optimization-results');
        
        container.innerHTML = `
            <div class="alert alert-info">
                <i class="fas fa-cog"></i>
                <strong>Portfolio Optimization</strong>
                <p class="mb-2 mt-2">Advanced portfolio optimization features are being developed.</p>
                <ul class="mb-0">
                    <li>Risk-adjusted position sizing</li>
                    <li>Correlation-based allocation</li>
                    <li>Dynamic rebalancing</li>
                    <li>Drawdown protection</li>
                </ul>
            </div>
        `;
    }

    // Utility Functions
    setupEventListeners() {
        // Add follower form
        document.getElementById('addFollowerForm')?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.addFollower();
        });

        // Refresh buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('[data-refresh]')) {
                e.preventDefault();
                this.loadSectionData(this.currentSection);
            }
        });
    }

    setupWebSocket() {
        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws`;
            
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                console.log('WebSocket connected');
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleWebSocketMessage(data);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };
            
            this.ws.onclose = () => {
                console.log('WebSocket disconnected');
                // Attempt to reconnect after 5 seconds
                setTimeout(() => this.setupWebSocket(), 5000);
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        } catch (error) {
            console.error('WebSocket setup failed:', error);
        }
    }

    handleWebSocketMessage(data) {
        switch (data.type) {
            case 'trade_update':
                this.handleTradeUpdate(data.data);
                break;
            case 'system_status':
                this.handleSystemStatusUpdate(data.data);
                break;
            default:
                console.log('Unknown WebSocket message type:', data.type);
        }
    }

    handleTradeUpdate(tradeData) {
        // Refresh current section if it's trades or dashboard
        if (this.currentSection === 'trades' || this.currentSection === 'dashboard') {
            this.loadSectionData(this.currentSection);
        }
        
        // Show notification
        this.showNotification('New Trade', `${tradeData.side.toUpperCase()} ${tradeData.asset} executed`, 'success');
    }

    handleSystemStatusUpdate(statusData) {
        // Update system status indicator
        this.loadSystemStatus();
    }

    async addFollower() {
        try {
            const form = document.getElementById('addFollowerForm');
            const formData = new FormData(form);
            
            const followerData = {
                user_id: 'user_' + Date.now(), // In real app, get from auth
                leader_address: document.getElementById('leaderAddress').value,
                api_wallet_address: document.getElementById('apiWalletAddress').value,
                copy_percentage: parseFloat(document.getElementById('copyPercentage').value),
                max_position_size: parseFloat(document.getElementById('maxPositionSize').value),
                stop_loss_percentage: document.getElementById('stopLoss').value ? 
                    parseFloat(document.getElementById('stopLoss').value) : null,
                take_profit_percentage: document.getElementById('takeProfit').value ? 
                    parseFloat(document.getElementById('takeProfit').value) : null,
                is_active: document.getElementById('isActive').checked,
                risk_settings: {}
            };

            const response = await fetch(`${this.apiBase}/followers`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(followerData)
            });

            const data = await response.json();

            if (data.success) {
                this.showSuccess('Follow added successfully!');
                
                // Close modal
                const modal = bootstrap.Modal.getInstance(document.getElementById('addFollowerModal'));
                modal.hide();
                
                // Reset form
                form.reset();
                
                // Refresh followers if we're on that section
                if (this.currentSection === 'followers') {
                    await this.loadFollowers();
                }
            } else {
                this.showError(data.error || 'Failed to add follow');
            }
        } catch (error) {
            console.error('Error adding follower:', error);
            this.showError('Failed to add follow');
        }
    }

    calculatePerformanceMetrics(trades) {
        const dailyData = {};
        let cumulativeReturn = 0;
        let peak = 0;
        
        trades.forEach(trade => {
            if (trade.status !== 'filled') return;
            
            const date = new Date(trade.executed_at).toISOString().split('T')[0];
            const pnl = trade.side === 'sell' ? 
                trade.size * trade.price : 
                -trade.size * trade.price;
            
            if (!dailyData[date]) {
                dailyData[date] = { pnl: 0, return: 0, drawdown: 0 };
            }
            dailyData[date].pnl += pnl;
        });

        const sortedDates = Object.keys(dailyData).sort();
        const labels = [];
        const returns = [];
        const drawdown = [];

        sortedDates.forEach(date => {
            const dailyPnl = dailyData[date].pnl;
            cumulativeReturn += (dailyPnl / 10000) * 100; // Assuming $10k base
            
            if (cumulativeReturn > peak) {
                peak = cumulativeReturn;
            }
            
            const currentDrawdown = peak - cumulativeReturn;
            
            labels.push(new Date(date).toLocaleDateString());
            returns.push(cumulativeReturn);
            drawdown.push(-currentDrawdown);
        });

        return { labels, returns, drawdown };
    }

    calculateRiskLevel(follower) {
        const copyPct = follower.copy_percentage;
        const maxPos = follower.max_position_size;
        
        if (copyPct <= 5 && maxPos <= 500) return 'Low';
        if (copyPct <= 15 && maxPos <= 2000) return 'Medium';
        if (copyPct <= 25 && maxPos <= 5000) return 'High';
        return 'Very High';
    }

    processDailyPnL(trades) {
        const dailyData = {};
        
        trades.forEach(trade => {
            if (trade.status !== 'filled') return;
            
            const date = new Date(trade.executed_at).toISOString().split('T')[0];
            const pnl = trade.side === 'sell' ? 
                trade.size * trade.price : 
                -trade.size * trade.price;
            
            if (!dailyData[date]) {
                dailyData[date] = 0;
            }
            dailyData[date] += pnl;
        });

        const sortedDates = Object.keys(dailyData).sort();
        const labels = [];
        const cumulative = [];
        let runningTotal = 0;

        sortedDates.forEach(date => {
            runningTotal += dailyData[date];
            labels.push(new Date(date).toLocaleDateString());
            cumulative.push(runningTotal);
        });

        return { labels, cumulative };
    }

    startAutoRefresh() {
        // Refresh every 30 seconds
        this.refreshInterval = setInterval(() => {
            if (this.currentSection === 'dashboard') {
                this.loadSummaryStats();
                this.loadRecentActivity();
            }
        }, 30000);
    }

    setLoading(isLoading) {
        this.isLoading = isLoading;
        // You could add a global loading indicator here
    }

    formatNumber(num) {
        if (num === null || num === undefined) return '0';
        return new Intl.NumberFormat('en-US', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 6
        }).format(num);
    }

    formatAddress(address) {
        if (!address) return '';
        return address.slice(0, 6) + '...' + address.slice(-4);
    }

    formatDateTime(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
    }

    formatDate(dateString) {
        return new Date(dateString).toLocaleDateString();
    }

    getStatusColor(status) {
        const colors = {
            'filled': 'success',
            'pending': 'warning',
            'cancelled': 'secondary',
            'rejected': 'danger',
            'failed': 'danger'
        };
        return colors[status] || 'secondary';
    }

    showSuccess(message) {
        this.showNotification('Success', message, 'success');
    }

    showError(message) {
        this.showNotification('Error', message, 'danger');
    }

    showNotification(title, message, type = 'info') {
        const toast = document.getElementById('alertToast');
        const toastBody = document.getElementById('toastMessage');
        
        toast.className = `toast border-${type}`;
        toastBody.innerHTML = `
            <div class="d-flex">
                <i class="fas fa-${type === 'success' ? 'check-circle' : type === 'danger' ? 'exclamation-triangle' : 'info-circle'} me-2 text-${type}"></i>
                <div>
                    <strong>${title}</strong><br>
                    ${message}
                </div>
            </div>
        `;
        
        const bsToast = new bootstrap.Toast(toast);
        bsToast.show();
    }

    // Hyperliquid Section
    async loadHyperliquid() {
        try {
            this.setLoading(true);
            
            // Load protocol status
            await this.loadProtocolStatus();
            
            // Load market data
            await this.loadMarketData();
            
            // Load spot markets
            await this.loadSpotMarkets();
            
            // Load trending assets
            await this.loadTrendingAssets();
            
            // Update API stats
            this.updateAPIStats();
            
        } catch (error) {
            console.error('Error loading Hyperliquid data:', error);
            this.showError('Failed to load Hyperliquid data');
        } finally {
            this.setLoading(false);
        }
    }

    async loadProtocolStatus() {
        try {
            // Load perpetuals metadata
            const perpResponse = await fetch('https://api.hyperliquid.xyz/info', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: 'meta' })
            });
            
            if (perpResponse.ok) {
                const perpData = await perpResponse.json();
                document.getElementById('perp-count').textContent = perpData.universe?.length || 0;
                document.getElementById('exchange-status').className = 'badge bg-success';
                document.getElementById('exchange-status').textContent = 'Online';
            } else {
                throw new Error('Failed to fetch perp data');
            }

            // Load spot metadata
            const spotResponse = await fetch('https://api.hyperliquid.xyz/info', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: 'spotMeta' })
            });
            
            if (spotResponse.ok) {
                const spotData = await spotResponse.json();
                document.getElementById('spot-count').textContent = spotData.tokens?.length || 0;
            }

        } catch (error) {
            console.error('Error loading protocol status:', error);
            document.getElementById('exchange-status').className = 'badge bg-danger';
            document.getElementById('exchange-status').textContent = 'Offline';
            document.getElementById('perp-count').textContent = 'Error';
            document.getElementById('spot-count').textContent = 'Error';
        }
    }

    async loadMarketData() {
        try {
            const response = await fetch('https://api.hyperliquid.xyz/info', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: 'metaAndAssetCtxs' })
            });

            if (!response.ok) throw new Error('Failed to fetch market data');
            
            const [meta, assetCtxs] = await response.json();
            const tbody = document.getElementById('perp-markets');
            
            if (meta.universe && assetCtxs) {
                const topMarkets = meta.universe.slice(0, 10).map((asset, index) => {
                    const ctx = assetCtxs[index];
                    if (!ctx) return null;
                    
                    const change24h = ctx.prevDayPx ? 
                        ((parseFloat(ctx.markPx) - parseFloat(ctx.prevDayPx)) / parseFloat(ctx.prevDayPx) * 100) : 0;
                    
                    return {
                        name: asset.name,
                        markPx: ctx.markPx,
                        change24h: change24h,
                        volume24h: ctx.dayNtlVlm,
                        openInterest: ctx.openInterest,
                        funding: ctx.funding
                    };
                }).filter(Boolean);

                tbody.innerHTML = topMarkets.map(market => `
                    <tr>
                        <td>
                            <strong>${market.name}</strong>
                        </td>
                        <td>${this.formatNumber(parseFloat(market.markPx))}</td>
                        <td class="text-${market.change24h >= 0 ? 'success' : 'danger'}">
                            ${market.change24h >= 0 ? '+' : ''}${market.change24h.toFixed(2)}%
                        </td>
                        <td>${this.formatLargeNumber(parseFloat(market.volume24h))}</td>
                        <td>${this.formatLargeNumber(parseFloat(market.openInterest))}</td>
                        <td class="funding-${parseFloat(market.funding) >= 0 ? 'positive' : 'negative'}">
                            ${(parseFloat(market.funding) * 100).toFixed(4)}%
                        </td>
                    </tr>
                `).join('');
            } else {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="6" class="text-center text-muted">
                            No market data available
                        </td>
                    </tr>
                `;
            }
        } catch (error) {
            console.error('Error loading market data:', error);
            const tbody = document.getElementById('perp-markets');
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="text-center text-danger">
                        <i class="fas fa-exclamation-triangle"></i> Failed to load market data
                    </td>
                </tr>
            `;
        }
    }

    async loadSpotMarkets() {
        try {
            const response = await fetch('https://api.hyperliquid.xyz/info', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: 'spotMetaAndAssetCtxs' })
            });

            if (!response.ok) throw new Error('Failed to fetch spot data');
            
            const [meta, assetCtxs] = await response.json();
            const tbody = document.getElementById('spot-markets');
            
            if (meta.universe && assetCtxs) {
                const spotMarkets = meta.universe.slice(0, 8).map((pair, index) => {
                    const ctx = assetCtxs[index];
                    if (!ctx) return null;
                    
                    const change24h = ctx.prevDayPx ? 
                        ((parseFloat(ctx.markPx) - parseFloat(ctx.prevDayPx)) / parseFloat(ctx.prevDayPx) * 100) : 0;
                    
                    return {
                        name: pair.name,
                        markPx: ctx.markPx,
                        change24h: change24h,
                        volume24h: ctx.dayNtlVlm,
                        marketCap: parseFloat(ctx.markPx) * 1000000 // Mock market cap
                    };
                }).filter(Boolean);

                tbody.innerHTML = spotMarkets.map(market => `
                    <tr>
                        <td>
                            <strong>${market.name}</strong>
                        </td>
                        <td>${this.formatNumber(parseFloat(market.markPx))}</td>
                        <td class="text-${market.change24h >= 0 ? 'success' : 'danger'}">
                            ${market.change24h >= 0 ? '+' : ''}${market.change24h.toFixed(2)}%
                        </td>
                        <td>${this.formatLargeNumber(parseFloat(market.volume24h))}</td>
                        <td>${this.formatLargeNumber(market.marketCap)}</td>
                        <td>
                            <button class="btn btn-sm btn-outline-primary" onclick="viewSpotDetails('${market.name}')">
                                <i class="fas fa-eye"></i>
                            </button>
                        </td>
                    </tr>
                `).join('');
            } else {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="6" class="text-center text-muted">
                            No spot markets available
                        </td>
                    </tr>
                `;
            }
        } catch (error) {
            console.error('Error loading spot markets:', error);
            const tbody = document.getElementById('spot-markets');
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="text-center text-danger">
                        <i class="fas fa-exclamation-triangle"></i> Failed to load spot markets
                    </td>
                </tr>
            `;
        }
    }

    async loadTrendingAssets() {
        const container = document.getElementById('trending-assets');
        
        try {
            // Mock trending data based on volume
            const trendingAssets = [
                { name: 'BTC', change: '+5.2%', volume: '1.2B' },
                { name: 'ETH', change: '+3.8%', volume: '890M' },
                { name: 'SOL', change: '+12.4%', volume: '456M' },
                { name: 'AVAX', change: '-2.1%', volume: '234M' },
                { name: 'MATIC', change: '+8.7%', volume: '189M' }
            ];

            container.innerHTML = trendingAssets.map(asset => `
                <div class="trending-asset-item">
                    <div>
                        <strong>${asset.name}</strong>
                        <div class="small text-muted">Vol: ${asset.volume}</div>
                    </div>
                    <div class="text-${asset.change.startsWith('+') ? 'success' : 'danger'}">
                        ${asset.change}
                    </div>
                </div>
            `).join('');
        } catch (error) {
            console.error('Error loading trending assets:', error);
            container.innerHTML = `
                <div class="text-center text-danger">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p class="mt-2">Failed to load trending assets</p>
                </div>
            `;
        }
    }

    updateAPIStats() {
        // Mock API statistics
        document.getElementById('api-calls-today').textContent = Math.floor(Math.random() * 10000 + 5000);
        document.getElementById('success-rate-api').textContent = (99.5 + Math.random() * 0.4).toFixed(1) + '%';
        document.getElementById('avg-latency').textContent = Math.floor(Math.random() * 20 + 35) + 'ms';
    }

    formatLargeNumber(num) {
        if (num >= 1e9) return (num / 1e9).toFixed(1) + 'B';
        if (num >= 1e6) return (num / 1e6).toFixed(1) + 'M';
        if (num >= 1e3) return (num / 1e3).toFixed(1) + 'K';
        return num.toFixed(2);
    }
}

// Global functions for button handlers
window.copyTrader = function(address) {
    // Pre-fill the add follower modal with leader address
    document.getElementById('leaderAddress').value = address;
    const modal = new bootstrap.Modal(document.getElementById('addFollowerModal'));
    modal.show();
};

window.viewLeaderDetails = function(address) {
    // This would open a detailed view - for now just show placeholder
    app.showNotification('Coming Soon', 'Detailed leader analysis is being developed', 'info');
};

window.editFollower = function(id) {
    app.showNotification('Coming Soon', 'Edit follower functionality is being developed', 'info');
};

window.viewFollowerAnalytics = function(id) {
    app.showNotification('Coming Soon', 'Follower analytics view is being developed', 'info');
};

window.deleteFollower = async function(id) {
    if (confirm('Are you sure you want to remove this follow? This will stop copying this leader.')) {
        try {
            const response = await fetch(`${app.apiBase}/followers/${id}`, {
                method: 'DELETE'
            });
            
            const data = await response.json();
            
            if (data.success) {
                app.showSuccess('Follow removed successfully');
                await app.loadFollowers();
            } else {
                app.showError(data.error || 'Failed to remove follow');
            }
        } catch (error) {
            console.error('Error removing follower:', error);
            app.showError('Failed to remove follow');
        }
    }
};

window.loadLeaders = function() {
    app.loadLeaders();
};

window.loadTrades = function() {
    app.loadTrades();
};

window.addFollower = function() {
    app.addFollower();
};

window.refreshMarketData = function() {
    const button = document.querySelector('[onclick="refreshMarketData()"]');
    const icon = button.querySelector('i');
    icon.classList.add('refresh-spinning');
    
    app.loadMarketData().finally(() => {
        icon.classList.remove('refresh-spinning');
    });
};

window.viewSpotDetails = function(pair) {
    app.showNotification('Coming Soon', `Detailed view for ${pair} is being developed`, 'info');
};

// Initialize the application
let app;
document.addEventListener('DOMContentLoaded', () => {
    app = new CopyTradingApp();
});
