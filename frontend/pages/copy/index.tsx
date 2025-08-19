import Layout from '../../components/Layout'
import { UserCircleIcon } from '@heroicons/react/24/outline'

const traders = [
  {
    name: "Alice Chen",
    returns: 27.3,
    followers: 1120
  },
  {
    name: "Markus Stein",
    returns: 19.8,
    followers: 980
  },
  {
    name: "Sofia Li",
    returns: 32.5,
    followers: 1510
  },
  {
    name: "David Kim",
    returns: 14.7,
    followers: 670
  },
  {
    name: "Emma Rossi",
    returns: 23.4,
    followers: 865
  },
  {
    name: "Nina Patel",
    returns: 21.2,
    followers: 990
  }
]

export default function Copy() {
  return (
    <Layout>
      <section className="container mx-auto py-8">
        <h2 className="text-3xl font-bold mb-8">Copy Traders</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
          {traders.map((trader, idx) => (
            <div
              key={idx}
              className="bg-white rounded-2xl border border-gray-100 shadow-sm hover:shadow-xl transition-shadow duration-200 p-6 flex flex-col items-center"
            >
              <UserCircleIcon className="w-14 h-14 text-primary mb-3" />
              <div className="text-lg font-semibold mb-1">{trader.name}</div>
              <div className="text-sm text-gray-500 mb-2">
                <span className="font-medium text-green-600">{trader.returns > 0 ? "+" : ""}{trader.returns}%</span> returns
              </div>
              <div className="text-xs text-gray-400 mb-4">{trader.followers} followers</div>
              <button className="w-full bg-primary text-white font-semibold rounded-lg py-2 mt-auto hover:bg-primary-dark transition-all duration-150 shadow-md">
                Copy
              </button>
            </div>
          ))}
        </div>
      </section>
    </Layout>
  )
}