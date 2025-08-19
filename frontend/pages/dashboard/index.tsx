import Layout from '../../components/Layout'
import { BanknotesIcon, ChartBarIcon, UserGroupIcon, ArrowTrendingUpIcon } from '@heroicons/react/24/solid'
import ChartPlaceholder from '../../components/ChartPlaceholder'
import { useApi } from '../../hooks/useApi'
import { ENDPOINTS } from '../../utils/endpoints'
import type { TraderSummary } from '../../types/models'

interface PortfolioSummary {
  total_equity: number;
  pnl: number;
  positions: { id: string }[]; // minimal for counting open positions
}

export default function Dashboard() {
  // Get portfolio summary and traders
  const { data: portfolio, loading: loadingPortfolio, error: errorPortfolio } = useApi<PortfolioSummary>(ENDPOINTS.PORTFOLIO)
  const { data: traders, loading: loadingTraders, error: errorTraders } = useApi<TraderSummary[]>(ENDPOINTS.TRADERS)

  const loading = loadingPortfolio || loadingTraders;
  const error = errorPortfolio || errorTraders;

  // Map API data to stat cards
  const stats = [
    {
      title: "Total Equity",
      value: portfolio ? `${portfolio.total_equity.toLocaleString()}` : "--",
      icon: <BanknotesIcon className="h-8 w-8 text-primary" />,
      change: undefined
    },
    {
      title: "P&L",
      value: portfolio ? `${portfolio.pnl >= 0 ? "+" : ""}${portfolio.pnl.toLocaleString()}` : "--",
      icon: <ArrowTrendingUpIcon className="h-8 w-8 text-green-500" />,
      change: undefined
    },
    {
      title: "Active Traders",
      value: traders ? traders.length.toString() : "--",
      icon: <UserGroupIcon className="h-8 w-8 text-secondary" />,
      change: undefined
    },
    {
      title: "Open Positions",
      value: portfolio ? portfolio.positions.length.toString() : "--",
      icon: <ChartBarIcon className="h-8 w-8 text-amber-500" />,
      change: undefined
    }
  ]

  return (
    <Layout>
      <section className="container mx-auto py-8">
        <h2 className="text-3xl font-bold mb-8">Dashboard</h2>
        {loading && (
          <div className="flex justify-center items-center py-16"><Spinner /></div>
        )}
        {error && (
          <div className="flex justify-center items-center py-8 text-red-600 font-semibold">
            Failed to load dashboard data. Please try again.
          </div>
        )}
        {!loading && !error && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            {stats.map((s, i) => (
              <div
                key={i}
                className="bg-white rounded-2xl shadow-sm p-6 flex flex-col items-start border border-gray-100 hover:shadow-lg transition duration-200"
              >
                <div className="mb-3">{s.icon}</div>
                <div className="text-xl font-bold text-gray-900 mb-1">{s.value}</div>
                <div className="text-gray-500 mb-2">{s.title}</div>
                {/* For demo, show ChartPlaceholder for P&L only if not loading/error */}
                {s.title === "P&L" && (
                  <div className="w-full mt-4">
                    {loadingPortfolio ? <div className="flex justify-center"><Spinner /></div> : <ChartPlaceholder />}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </section>
    </Layout>
  )
}

function Spinner() {
  return (
    <svg className="animate-spin h-8 w-8 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
      <circle className="opacity-20" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
      <path className="opacity-70" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"/>
    </svg>
  )
}