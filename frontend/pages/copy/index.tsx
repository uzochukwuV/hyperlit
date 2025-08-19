import Layout from '../../components/Layout'
import { UserCircleIcon } from '@heroicons/react/24/outline'
import type { TraderSummary } from '../../types/models'
import { useApi } from '../../hooks/useApi'
import { ENDPOINTS } from '../../utils/endpoints'
import ProtectedRoute from '../../components/ProtectedRoute'

function CopyInner() {
  const { data: traders, loading, error } = useApi<TraderSummary[]>(ENDPOINTS.TRADERS)

  return (
    <section className="container mx-auto py-8">
      <h2 className="text-3xl font-bold mb-8">Copy Traders</h2>
      {loading && (
        <div className="flex justify-center items-center py-16">
          <Spinner />
        </div>
      )}
      {error && (
        <div className="flex justify-center items-center py-8 text-red-600 font-semibold">
          Failed to load traders. Please try again later.
        </div>
      )}
      {!loading && !error && traders && traders.length === 0 && (
        <div className="text-center text-gray-500 py-12">No traders found.</div>
      )}
      {!loading && !error && traders && traders.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
          {traders.map((trader) => (
            <div
              key={trader.id}
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
      )}
    </section>
  )
}

export default function Copy() {
  return (
    <Layout>
      <ProtectedRoute>
        <CopyInner />
      </ProtectedRoute>
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