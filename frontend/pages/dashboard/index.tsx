import Layout from '../../components/Layout'
import { BanknotesIcon, ChartBarIcon, UserGroupIcon, ArrowTrendingUpIcon } from '@heroicons/react/24/solid'
import ChartPlaceholder from '../../components/ChartPlaceholder'

const stats = [
  {
    title: "Total Equity",
    value: "$18,250",
    icon: <BanknotesIcon className="h-8 w-8 text-primary" />,
    change: "+5.1%"
  },
  {
    title: "P&L",
    value: "+$1,340",
    icon: <ArrowTrendingUpIcon className="h-8 w-8 text-green-500" />,
    change: "+8.2%"
  },
  {
    title: "Active Traders",
    value: "7",
    icon: <UserGroupIcon className="h-8 w-8 text-secondary" />,
    change: ""
  },
  {
    title: "Open Positions",
    value: "14",
    icon: <ChartBarIcon className="h-8 w-8 text-amber-500" />,
    change: ""
  }
]

export default function Dashboard() {
  return (
    <Layout>
      <section className="container mx-auto py-8">
        <h2 className="text-3xl font-bold mb-8">Dashboard</h2>
        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          {stats.map((s, i) => (
            <div
              key={i}
              className="bg-white rounded-2xl shadow-sm p-6 flex flex-col items-start border border-gray-100 hover:shadow-lg transition duration-200"
            >
              <div className="mb-3">{s.icon}</div>
              <div className="text-xl font-bold text-gray-900 mb-1">{s.value}</div>
              <div className="text-gray-500 mb-2">{s.title}</div>
              {s.change && (
                <div className="text-sm font-medium text-green-600">{s.change}</div>
              )}
              {/* Show chart next to P&L */}
              {s.title === "P&L" && (
                <div className="w-full mt-4">
                  <ChartPlaceholder />
                </div>
              )}
            </div>
          ))}
        </div>
      </section>
    </Layout>
  )
}