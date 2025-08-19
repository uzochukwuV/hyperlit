import type { AppProps } from 'next/app'
import '../styles/globals.css'

import '@rainbow-me/rainbowkit/styles.css';
import {
  getDefaultWallets,
  RainbowKitProvider,
  lightTheme,
} from '@rainbow-me/rainbowkit';
import { configureChains, createConfig, WagmiConfig } from 'wagmi';
import {
  mainnet,
  goerli,
  sepolia,
  polygon,
  polygonMumbai,
  optimism,
  arbitrum,
} from 'wagmi/chains';

// --- Alchemy integration imports ---
import { alchemyProvider } from '@alchemy/wagmi';
import { SmartWalletConnector } from '@alchemy/wagmi';
import { Alchemy, Network } from '@alchemy/sdk';

// Get API key from env
const alchemyApiKey = process.env.NEXT_PUBLIC_ALCHEMY_API_KEY as string;

// --- Configure chains with Alchemy's provider pointing to HyperEVM ---
const hyperEvmChain = {
  id: 13370,
  name: 'HyperEVM',
  network: 'hyperevm',
  nativeCurrency: {
    name: 'HyperEVM',
    symbol: 'HYEV',
    decimals: 18,
  },
  rpcUrls: {
    default: {
      http: [`https://hyperevm.g.alchemy.com/v2/${alchemyApiKey}`],
    },
    public: {
      http: [`https://hyperevm.g.alchemy.com/v2/${alchemyApiKey}`],
    },
  },
  blockExplorers: {
    default: { name: "Hyperevm Explorer", url: "https://hyperevm.alchemy.com/explorer" },
  },
  testnet: false,
};

// Configure chains (add HyperEVM first, then other chains)
const { chains, publicClient } = configureChains(
  [
    hyperEvmChain,
    mainnet,
    polygon,
    optimism,
    arbitrum,
    goerli,
    sepolia,
    polygonMumbai
  ],
  [
    alchemyProvider({
      apiKey: alchemyApiKey,
      // Optionally override RPC URL for HyperEVM
      priority: 0,
      rpcUrls: {
        [hyperEvmChain.id]: hyperEvmChain.rpcUrls.default.http[0],
      }
    }),
    // Optionally add fallback publicProvider if needed:
    // publicProvider()
  ]
);

// --- Set up connectors, including Alchemy SmartWalletConnector ---
import { connectorsForWallets } from '@rainbow-me/rainbowkit';

const smartWalletConnector = new SmartWalletConnector({
  chains,
  options: {
    apiKey: alchemyApiKey,
    // Enable email/social login
    enableSocialLogin: true,
    enableEmailLogin: true,
    // Optionally, you can add more custom config here.
  },
});

const { wallets } = getDefaultWallets({
  appName: 'Hyperlit',
  projectId: 'hyperlit-app', // Use your real WalletConnect projectId in production
  chains,
});

const connectors = connectorsForWallets([
  {
    groupName: 'Recommended',
    wallets: [
      smartWalletConnector,
      ...wallets[0].wallets, // Injected (MetaMask, etc.)
      ...wallets[1].wallets, // WalletConnect, etc.
    ],
  },
]);

const wagmiConfig = createConfig({
  autoConnect: true,
  connectors,
  publicClient,
});

// --- Initialize Alchemy SDK for other on-chain/read ops ---
export const alchemy = new Alchemy({
  apiKey: alchemyApiKey,
  network: Network.ETH_MAINNET, // Or set to custom if HyperEVM is a supported network in the SDK.
});

export default function App({ Component, pageProps }: AppProps) {
  return (
    <WagmiConfig config={wagmiConfig}>
      <RainbowKitProvider
        chains={chains}
        theme={{
          ...lightTheme(),
          colors: {
            ...lightTheme().colors,
            accentColor: '#6366f1',
          },
        }}
        modalSize="compact"
        showRecentTransactions={true}
      >
        <Component {...pageProps} />
      </RainbowKitProvider>
    </WagmiConfig>
  );
}