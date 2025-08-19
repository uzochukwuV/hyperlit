import React, { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { getTextRecords, resolveName } from "../../utils/hlnames";
import type { HlnTextRecords, HlnResolveResponse } from "../../types/HlnRecord";

export default function HlNameProfile() {
  const router = useRouter();
  const { name } = router.query as { name?: string };

  const [records, setRecords] = useState<HlnTextRecords | null>(null);
  const [address, setAddress] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!name) return;
    setLoading(true);
    setError(null);
    Promise.all([
      getTextRecords(name).catch((e) => {
        setError(e?.response?.data?.error || "Failed to fetch text records.");
        return {};
      }),
      resolveName(name).catch(() => ({ address: null })),
    ]).then(([recordsRes, resolveRes]) => {
      setRecords(recordsRes);
      setAddress(resolveRes?.address || null);
      setLoading(false);
    });
  }, [name]);

  return (
    <div className="max-w-xl mx-auto mt-16 p-8 bg-white rounded-xl shadow-lg">
      <h1 className="text-2xl font-bold mb-6 text-center">
        <span className="text-indigo-600">{name}</span> Profile
      </h1>
      {loading && (
        <div className="text-center text-gray-500">Loading...</div>
      )}
      {error && (
        <div className="text-red-600 text-center mb-4">{error}</div>
      )}
      {!loading && !error && (
        <div>
          <div className="mb-6">
            <span className="block text-sm text-gray-500 mb-1">Resolved Address</span>
            <div className="font-mono bg-gray-100 px-3 py-2 rounded-lg break-all text-gray-800">
              {address || <span className="text-gray-400">Not found</span>}
            </div>
          </div>
          <div>
            <span className="block text-sm text-gray-500 mb-2">Text Records</span>
            <div className="bg-gray-50 border border-gray-200 rounded-xl p-4">
              {records && Object.keys(records).length > 0 ? (
                <ul className="space-y-2">
                  {Object.entries(records).map(([key, value]) => (
                    <li key={key} className="flex flex-col sm:flex-row sm:items-center">
                      <span className="w-32 font-semibold text-gray-700">{key}</span>
                      <span className="flex-1 font-mono text-gray-900 break-all">{value}</span>
                    </li>
                  ))}
                </ul>
              ) : (
                <div className="text-gray-400 italic">No records found.</div>
              )}
            </div>
          </div>
        </div>
      )}
      <div className="mt-8 text-center">
        <button
          className="bg-gray-100 text-gray-700 py-2 px-6 rounded-lg hover:bg-gray-200 transition font-medium"
          onClick={() => router.push("/hl-names")}
        >
          ← Back to Resolver
        </button>
      </div>
    </div>
  );
}