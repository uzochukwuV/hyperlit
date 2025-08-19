import Layout from '../../components/Layout'
import { LockClosedIcon } from '@heroicons/react/24/solid'
import type { VaultSummary } from '../../types/models'
import { useApi } from '../../hooks/useApi'
import { ENDPOINTS } from '../../utils/endpoints'

export default function Vault() {
  const { data: vaults, loading, error } = useApi<VaultSummary[]>(ENDPOINTS.VAULTS)

  return (
    <Layout>
      <section className="container mx-auto py-8">
        <h2 className="text-3xl font-bold mb-8">Vaults</h2>
        {loading && (
          <div className="flex justify-center items-center py-16"><Spinner /></div>
        )}
        {error && (
          <div className="flex justify-center items-center py-8 text-red-600 font-semibold">
            Failed to load vaults. Please try again.
          </div>
        )}
        {!loading && !error && vaults && vaults.length === 0 && (
          <div className="text-center text-gray-500 py-12">No vaults found.</div>
        )}
        {!loading && !error && vaults && vaults.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
            {vaults.map((vault) => (
              <div
                key={vault.name}
                className="bg-white rounded-2xl border border-gray-100 shadow-sm hover:shadow-xl transition-shadow duration-200 p-6 flex flex-col"
              >
                <div className="flex items-center gap-3 mb-2">
                  <LockClosedIcon className="w-7 h-7 text-amber-500" />
                  <span className="text-lg font-semibold">{vault.name}</span>
                </div>
                <div className="flex-1 flex flex-col justify-center">
                  <div className="text-sm text-gray-500 mb-2">APR</div>
                  <div className="text-2xl font-bold text-green-600 mb-2">{vault.apr}%</div>
                  <div className="text-sm text-gray-500 mb-1">Assets Under Management</div>
                  <div className="text-xl font-semibold text-gray-800">{vault.aum}</div>
                </div>
                <button className="mt-6 w-full bg-primary text-white font-semibold rounded-lg py-2 hover:bg-primary-dark transition-all duration-150 shadow-md">
                  Manage
                </button>
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