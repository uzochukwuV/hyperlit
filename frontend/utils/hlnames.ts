import axios from 'axios';
import type { HlnResolveResponse, HlnReverseResponse, HlnTextRecords } from '../types/HlnRecord';

const HLN_API_BASE = 'https://api.hlnames.xyz/api';

export async function resolveName(name: string): Promise<HlnResolveResponse> {
  const { data } = await axios.get<HlnResolveResponse>(\`\${HLN_API_BASE}/resolve/\${encodeURIComponent(name)}\`);
  return data;
}

export async function reverseResolve(address: string): Promise<HlnReverseResponse> {
  const { data } = await axios.get<HlnReverseResponse>(\`\${HLN_API_BASE}/reverse/\${encodeURIComponent(address)}\`);
  return data;
}

export async function getTextRecords(name: string): Promise<HlnTextRecords> {
  const { data } = await axios.get<HlnTextRecords>(\`\${HLN_API_BASE}/records/\${encodeURIComponent(name)}\`);
  return data;
}