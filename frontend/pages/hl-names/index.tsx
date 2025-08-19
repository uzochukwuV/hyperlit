import React, { useState } from "react";
import { resolveName, reverseResolve } from "../../utils/hlnames";
import type { HlnResolveResponse, HlnReverseResponse } from "../../types/HlnRecord";
import { useRouter } from "next/router";

const isAddress = (input: string) => /^0x[a-fA-F0-9]{40}$/.test(input.trim());

export default function HlNamesResolverPage() {
  const [input, setInput] = useState("");
  const [result, setResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  const handleResolveName = async () => {
    setError(null);
    setResult(null);
    setLoading(true);
    try {
      const { address }: HlnResolveResponse = await resolveName(input.trim());
      setResult(address ? \`Address: \${address}\` : "No address found.");
    } catch (e: any) {
      setError(e?.response?.data?.error || "Failed to resolve name.");
    }
    setLoading(false);
  };

  const handleReverseResolve = async () => {
    setError(null);
    setResult(null);
    setLoading(true);
    try {
      const { name }: HlnReverseResponse = await reverseResolve(input.trim());
      setResult(name ? \`Name: \${name}\` : "No .hl name found.");
    } catch (e: any) {
      setError(e?.response?.data?.error || "Failed to resolve address.");
    }
    setLoading(false);
  };

  const handleProfile = () => {
    if (input.trim().endsWith(".hl")) {
      router.push(\`/hl-names/\${encodeURIComponent(input.trim())}\`);
    }
  };

  return (
    <div className="max-w-lg mx-auto mt-16 p-8 bg-white rounded-xl shadow-lg">
      <h1 className="text-2xl font-bold mb-6 text-center">Hyperliquid Name Resolver (.hl)</h1>
      <input
        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-400 mb-4"
        placeholder="Enter .hl name or address"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        spellCheck={false}
      />
      <div className="flex gap-2 mb-4">
        <button
          className="flex-1 bg-indigo-600 text-white py-2 rounded-lg font-semibold hover:bg-indigo-700 transition"
          onClick={handleResolveName}
          disabled={loading || !input.trim().endsWith(".hl")}
        >
          Resolve Name→Address
        </button>
        <button
          className="flex-1 bg-blue-500 text-white py-2 rounded-lg font-semibold hover:bg-blue-600 transition"
          onClick={handleReverseResolve}
          disabled={loading || !isAddress(input.trim())}
        >
          Resolve Address→Name
        </button>
      </div>
      <div className="mb-2">
        <button
          className="w-full bg-gray-100 text-gray-700 py-2 rounded-lg hover:bg-gray-200 transition font-medium"
          onClick={handleProfile}
          disabled={!input.trim().endsWith(".hl")}
        >
          View .hl Profile
        </button>
      </div>
      {loading && (
        <div className="text-center text-gray-500 mt-4">Loading...</div>
      )}
      {error && (
        <div className="text-red-600 text-center mt-4">{error}</div>
      )}
      {result && (
        <div className="bg-green-50 border border-green-200 rounded-lg text-green-700 px-4 py-3 mt-4 text-center">
          {result}
        </div>
      )}
      <div className="mt-8 text-center text-gray-400 text-xs">
        Powered by <a href="https://hlnames.xyz" className="underline hover:text-indigo-500" target="_blank" rel="noopener noreferrer">Hyperliquid Names API</a>
      </div>
    </div>
  );
}