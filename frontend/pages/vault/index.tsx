import Layout from '../../components/Layout'
import { LockClosedIcon } from '@heroicons/react/24/solid'
import type { VaultSummary } from '../../types/models'

const vaults: VaultSummary[] = [
  {
    name: "Growth Vault",
    apr: 12.7,
    aum: "$82,500"
  },
  {
    name: "Stable Vault",
    apr: 7.3,
    aum: "$41,200"
  },
  {
    name: "Alpha Vault",
    apr: 18.1,
    aum: "$23,900"
  }
]

export default function Vault() {
  return (
    <Layout>
      <section className="container mx-auto py-8">
        <h2 className="text-3xl font-bold mb-8">Vaults</h2>
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
      </section>
    </Layout>
  )
}