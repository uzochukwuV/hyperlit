/**
 * API Endpoint constants for Hyperlit frontend.
 */
const API_PREFIX = "/api";

export const ENDPOINTS = {
  TRADERS: `${API_PREFIX}/info?type=frontendOpenOrders`,
  VAULTS: `${API_PREFIX}/vaults`, // or `/info?type=vaultDetails` if required
  PORTFOLIO: `${API_PREFIX}/info?type=portfolio`,
  PROFILE: `${API_PREFIX}/user/profile`,
};