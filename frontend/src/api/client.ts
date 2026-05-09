import type { Instance, AuditRecord, AgentInfo } from '../types';

const BASE = '/api/v1';

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, options);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export const api = {
  health: () => fetchJSON<{ status: string }>(`${BASE}/health`),

  createTask: (data: {
    scene: string;
    title: string;
    description: string;
    input: string;
    alerts?: Array<{ id: string; rule: string; source_ip?: string; detail: string; count?: number }>;
  }) =>
    fetchJSON<Instance>(`${BASE}/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    }),

  getTask: (id: string) => fetchJSON<Instance>(`${BASE}/tasks/${id}`),

  cancelTask: (id: string) =>
    fetchJSON<{ status: string }>(`${BASE}/tasks/${id}/cancel`, { method: 'POST' }),

  getAuditTrail: (id: string) => fetchJSON<AuditRecord[]>(`${BASE}/tasks/${id}/audit`),

  getAgents: () => fetchJSON<{ agents: AgentInfo[] }>(`${BASE}/agents`),

  getInstances: () => fetchJSON<Instance[]>(`${BASE}/instances`),

  getApprovals: () => fetchJSON<AuditRecord[]>(`${BASE}/approvals`),

  approveAction: (id: string) =>
    fetchJSON<{ status: string }>(`${BASE}/approvals/${id}/approve`, { method: 'POST' }),
};
