import Layout from '../../components/Layout'
import { useState, useEffect } from 'react'
import type { ProfileUser } from '../../types/models'
import { useApi } from '../../hooks/useApi'
import { ENDPOINTS } from '../../utils/endpoints'
import api from '../../utils/api'

export default function Profile() {
  const { data, loading, error } = useApi<ProfileUser>(ENDPOINTS.PROFILE)
  const [user, setUser] = useState<ProfileUser | null>(null)
  const [saving, setSaving] = useState(false)
  const [success, setSuccess] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  // Populate form fields when data loads
  useEffect(() => {
    if (data) setUser(data)
  }, [data])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaveError(null)
    setSaving(true)
    try {
      // Only send allowed fields
      await api.post(ENDPOINTS.PROFILE, {
        user_id: user?.user_id,
        address: user?.address,
        name: user?.name,
        email: user?.email,
      })
      setSuccess(true)
      setTimeout(() => setSuccess(false), 2000)
    } catch (err: any) {
      setSaveError('Failed to update profile. Please try again.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Layout>
      <section className="container mx-auto py-8 max-w-xl">
        <h2 className="text-3xl font-bold mb-8">Profile</h2>
        {loading && (
          <div className="flex justify-center items-center py-16"><Spinner /></div>
        )}
        {error && (
          <div className="flex justify-center items-center py-8 text-red-600 font-semibold">
            Failed to load profile. Please try again.
          </div>
        )}
        {!loading && !error && user && (
          <form
            onSubmit={handleSubmit}
            className="bg-white rounded-2xl shadow-md p-8 flex flex-col gap-6 border border-gray-100"
          >
            <div>
              <label className="block text-gray-700 font-semibold mb-1">
                Wallet Address
              </label>
              <input
                className="w-full bg-gray-50 border border-gray-200 rounded-lg px-4 py-2 text-gray-700 cursor-not-allowed"
                value={user.address}
                readOnly
              />
            </div>
            <div>
              <label className="block text-gray-700 font-semibold mb-1">
                Display Name
              </label>
              <input
                className="w-full bg-gray-50 border border-gray-200 rounded-lg px-4 py-2 text-gray-900 focus:outline-none focus:border-primary"
                value={user.name}
                onChange={e => setUser(u => u ? { ...u, name: e.target.value } : u)}
                required
                maxLength={32}
              />
            </div>
            <div>
              <label className="block text-gray-700 font-semibold mb-1">
                Email
              </label>
              <input
                type="email"
                className="w-full bg-gray-50 border border-gray-200 rounded-lg px-4 py-2 text-gray-900 focus:outline-none focus:border-primary"
                value={user.email || ''}
                onChange={e => setUser(u => u ? { ...u, email: e.target.value } : u)}
                required
              />
            </div>
            <button
              type="submit"
              className={`w-full py-3 rounded-xl font-semibold text-lg transition-all duration-150 shadow-md ${
                saving || success
                  ? 'bg-green-500 text-white'
                  : 'bg-primary text-white hover:bg-primary-dark'
              }`}
              disabled={saving}
            >
              {saving ? 'Saving...' : success ? 'Saved!' : 'Save Changes'}
            </button>
            {saveError && (
              <div className="text-red-600 text-center font-semibold mt-2">{saveError}</div>
            )}
          </form>
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