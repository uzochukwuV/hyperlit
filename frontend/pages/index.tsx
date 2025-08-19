import Layout from '../components/Layout'

export default function Home() {
  return (
    <Layout>
      <section className="flex flex-col items-center justify-center min-h-[60vh] py-12">
        <h1 className="text-4xl md:text-5xl font-extrabold text-primary mb-4">
          Welcome to Hyperlit
        </h1>
        <p className="text-lg text-gray-700 max-w-xl text-center mb-8">
          The modern, fluid copy trading platform. Effortlessly discover, copy, and invest in the best trading strategies.
        </p>
      </section>
    </Layout>
  )
}