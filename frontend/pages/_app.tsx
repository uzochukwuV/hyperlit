import type { AppProps } from 'next/app'
import '../styles/globals.css'

import '@rainbow-me/rainbowkit/styles.css';
import {
  getDefaultWallets,
  RainbowKitProvider,
  darkTheme,
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
import { publicProvider } from 'wagmi/providers/public';

const { chains, publicClient } = configureChains(
  [mainnet, polygon, optimism, arbitrum, goerli, sepolia, polygonMumbai],
  [publicProvider()]
);

const { connectors } = getDefaultWallets({
  appName: 'Hyperlit',
  projectId: 'hyperlit-app', // Use your real WalletConnect projectId in production
  chains,
});

const wagmiConfig = createConfig({
  autoConnect: true,
  connectors,
  publicClient,
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