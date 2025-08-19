import { useRouter } from "next/router";
import { useAccount } from "wagmi";
import { ConnectButton } from "@rainbow-me/rainbowkit";
import { useEffect } from "react";

type Props = {
  children: React.ReactNode;
};

/**
 * Protects pages by requiring a connected wallet.
 * If not connected, shows a wallet connect prompt.
 */
export default function ProtectedRoute({ children }: Props) {
  const { isConnected } = useAccount();
  const router = useRouter();

  // Optionally, you can redirect home if not connected and not on login
  // useEffect(() => {
  //   if (!isConnected) router.replace("/");
  // }, [isConnected, router]);

  if (!isConnected) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center bg-gray-50">
        <div className="bg-white shadow rounded-2xl px-8 py-12 flex flex-col items-center border border-gray-100">
          <h2 className="text-2xl font-bold mb-4 text-primary">Connect your wallet</h2>
          <p className="text-gray-600 mb-6 text-center">
            To access this page, please connect your wallet.
          </p>
          <ConnectButton />
        </div>
      </div>
    );
  }

  return <>{children}</>;
}