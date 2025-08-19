import Layout from '../components/Layout'
import Link from 'next/link'
import { MagnifyingGlassIcon, UserGroupIcon, LockClosedIcon } from '@heroicons/react/24/outline'

export default function Home() {
  return (
    <Layout>
      {/* Hero Section */}
      <section className="relative flex flex-col items-center justify-center min-h-[60vh] py-16 overflow-hidden">
        {/* Gradient or illustration */}
        <div className="absolute inset-0 bg-gradient-to-tr from-primary/10 via-secondary/10 to-amber-100 pointer-events-none z-0" />
        <div className="relative z-10 flex flex-col items-center">
          <h1 className="text-4xl md:text-5xl font-extrabold text-primary drop-shadow mb-4 text-center">
            Welcome to Hyperlit
          </h1>
          <p className="text-lg text-gray-700 max-w-xl text-center mb-8">
            The modern, fluid copy trading platform. Effortlessly discover, copy, and invest in the best trading strategies.
          </p>
          <Link href="/dashboard">
            <a className="inline-block px-8 py-3 bg-primary text-white font-semibold rounded-xl text-lg shadow-lg hover:bg-primary-dark transition-all duration-200">
              Get Started
            </a>
          </Link>
        </div>
        {/* Subtle illustration */}
        <svg className="absolute bottom-0 right-0 w-60 opacity-30 z-0" viewBox="0 0 300 200" fill="none">
          <ellipse cx="150" cy="100" rx="140" ry="70" fill="#6366f1" fillOpacity="0.07" />
        </svg>
      </section>

      {/* Features Grid */}
      <section className="container mx-auto py-12">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <FeatureCard
            icon={<MagnifyingGlassIcon className="h-10 w-10 text-primary" />}
            title="Discover"
            desc="Browse and analyze top-performing traders with deep insights and transparent track records."
          />
          <FeatureCard
            icon={<UserGroupIcon className="h-10 w-10 text-secondary" />}
            title="Copy"
            desc="Effortlessly copy trading strategies from the best. Mimic their trades automatically."
          />
          <FeatureCard
            icon={<LockClosedIcon className="h-10 w-10 text-amber-500" />}
            title="Vault"
            desc="Secure your assets and monitor your portfolio performance with robust vault management."
          />
        </div>
      </section>
    </Layout>
  )
}

function FeatureCard({
  icon,
  title,
  desc,
}: {
  icon: React.ReactNode
  title: string
  desc: string
}) {
  return (
    <div className="flex flex-col items-center bg-white rounded-2xl shadow-md hover:shadow-xl transition-shadow duration-200 border border-gray-100 p-8 group cursor-pointer hover:-translate-y-1">
      <div className="mb-4 group-hover:scale-110 transition-transform duration-200">{icon}</div>
      <h3 className="text-xl font-semibold mb-2 text-gray-900">{title}</h3>
      <p className="text-gray-600 text-center">{desc}</p>
    </div>
  )
}