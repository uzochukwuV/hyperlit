import Layout from '../../components/Layout'
import { useState } from 'react'

export default function Profile() {
  // Placeholder wallet address for demonstration
  const walletAddress = '0x8f3a...d9B2'
  const [displayName, setDisplayName] = useState('TraderJoe')
  const [email, setEmail] = useState('traderjoe@email.com')
  const [submitted, setSubmitted] = useState(false)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSubmitted(true)
    setTimeout(() => setSubmitted(false), 1800)
  }

  return (
    <Layout>
      <section className="container mx-auto py-8 max-w-xl">
        <h2 className="text-3xl font-bold mb-8">Profile</h2>
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
              value={walletAddress}
              readOnly
            />
          </div>
          <div>
            <label className="block text-gray-700 font-semibold mb-1">
              Display Name
            </label>
            <input
              className="w-full bg-gray-50 border border-gray-200 rounded-lg px-4 py-2 text-gray-900 focus:outline-none focus:border-primary"
              value={displayName}
              onChange={e => setDisplayName(e.target.value)}
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
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
            />
          </div>
          <button
            type="submit"
            className={`w-full py-3 rounded-xl font-semibold text-lg transition-all duration-150 shadow-md ${
              submitted
                ? 'bg-green-500 text-white'
                : 'bg-primary text-white hover:bg-primary-dark'
            }`}
            disabled={submitted}
          >
            {submitted ? 'Saved!' : 'Save Changes'}
          </button>
        </form>
      </section>
    </Layout>
  )
}