import axios from 'axios';
import type { Link, GroupWithCount, Group, Stats, Config } from '../types';

const client = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
  timeout: 15_000,
});

// ── Links ────────────────────────────────────────────────

export interface ListLinksResponse {
  links: Link[];
  total: number;
}

export async function listLinks(
  params?: {
    group_id?: string;
    status?: string;
    search?: string;
    sort?: string;
    order?: string;
    limit?: number;
    offset?: number;
  }
): Promise<ListLinksResponse> {
  const res = await client.get<ListLinksResponse>('/links', { params });
  return res.data;
}

export async function getLink(id: string): Promise<Link> {
  const res = await client.get<Link>(`/links/${id}`);
  return res.data;
}

export async function createLink(url: string): Promise<Link> {
  const res = await client.post<Link>('/links', { url });
  return res.data;
}

export interface BatchCreateResult {
  url: string;
  status: 'created' | 'duplicate' | 'error';
  error?: string;
  link?: Link;
}

export interface BatchCreateResponse {
  created: number;
  results: BatchCreateResult[];
}

export async function createLinksBatch(urls: string[]): Promise<BatchCreateResponse> {
  const res = await client.post<BatchCreateResponse>('/links/batch', { urls }, { timeout: 60_000 });
  return res.data;
}

export async function updateLink(
  id: string,
  data: { title?: string; source_type?: string; group_ids?: string[]; notes?: string; is_interesting?: boolean }
): Promise<Link> {
  const res = await client.put<Link>(`/links/${id}`, data);
  return res.data;
}

export async function deleteLink(id: string): Promise<void> {
  await client.delete(`/links/${id}`);
}

export async function reprocessLink(id: string): Promise<void> {
  await client.post(`/links/${id}/reprocess`);
}

export interface ReprocessAllResponse {
  queued: boolean;
  job_id: string;
}

export async function reprocessAllLinks(): Promise<ReprocessAllResponse> {
  const res = await client.post<ReprocessAllResponse>('/links/reprocess-all');
  return res.data;
}

// ── Groups ───────────────────────────────────────────────

export async function listGroups(): Promise<GroupWithCount[]> {
  const res = await client.get<GroupWithCount[]>('/groups');
  return res.data;
}

export async function createGroup(data: {
  name: string;
  color?: string;
  sort_order?: number;
}): Promise<Group> {
  const res = await client.post<Group>('/groups', data);
  return res.data;
}

export async function updateGroup(
  id: string,
  data: { name?: string; color?: string; sort_order?: number }
): Promise<Group> {
  const res = await client.put<Group>(`/groups/${id}`, data);
  return res.data;
}

export async function deleteGroup(id: string): Promise<void> {
  await client.delete(`/groups/${id}`);
}

// ── Stats ────────────────────────────────────────────────

export async function getStats(): Promise<Stats> {
  const res = await client.get<Stats>('/stats');
  return res.data;
}

// ── Config ───────────────────────────────────────────────

export async function getConfig(): Promise<Config> {
  const res = await client.get<Config>('/config');
  return res.data;
}

export async function updateConfig(data: Config): Promise<void> {
  await client.put('/config', data);
}
