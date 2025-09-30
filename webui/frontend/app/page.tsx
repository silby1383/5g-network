export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="z-10 max-w-5xl w-full items-center justify-center font-mono text-sm">
        <h1 className="text-4xl font-bold text-center mb-8">
          5G Core Network Management
        </h1>
        <p className="text-center text-xl mb-12">
          Cloud-Native 5G Network Functions
        </p>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
          <div className="border border-gray-300 rounded-lg p-6 hover:border-primary-500 transition-colors">
            <h2 className="text-xl font-semibold mb-2">Network Functions</h2>
            <p className="text-gray-600">
              Monitor and manage AMF, SMF, UPF, and all core NFs
            </p>
          </div>
          
          <div className="border border-gray-300 rounded-lg p-6 hover:border-primary-500 transition-colors">
            <h2 className="text-xl font-semibold mb-2">Subscribers</h2>
            <p className="text-gray-600">
              Manage subscriber data and PDU sessions
            </p>
          </div>
          
          <div className="border border-gray-300 rounded-lg p-6 hover:border-primary-500 transition-colors">
            <h2 className="text-xl font-semibold mb-2">Analytics</h2>
            <p className="text-gray-600">
              Real-time metrics and distributed tracing
            </p>
          </div>
        </div>
        
        <div className="text-center">
          <p className="text-sm text-gray-500">
            Built with Next.js, TypeScript, and Tailwind CSS
          </p>
          <p className="text-sm text-gray-500 mt-2">
            eBPF-based observability • 3GPP compliant • Cloud-native
          </p>
        </div>
      </div>
    </main>
  )
}
