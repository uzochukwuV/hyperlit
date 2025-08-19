/**
 * TypeScript interfaces mapping Go backend structs.
 */

export interface Leader {
  id: string;
  name: string;
  returns: number;
  followers: number;
}

export interface TraderSummary {
  id: string;
  name: string;
  returns: number;
  followers: number;
}

export interface FollowerSettings {
  id: string;
  user_id: string;
  leader_id: string;
  allocation: number; // e.g. USD or %
  max_drawdown?: number;
  enabled: boolean;
}

export interface CopyTrade {
  id: string;
  trader_id: string;
  follower_id: string;
  symbol: string;
  size: number;
  pnl: number;
  opened_at: string;
  closed_at?: string;
  status: "OPEN" | "CLOSED";
}

export interface VaultSummary {
  name: string;
  apr: number;
  aum: string; // e.g. "$82,500"
}

export interface ProfileUser {
  user_id: string;
  address: string;
  name: string;
  email?: string;
}