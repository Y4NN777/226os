import type { components } from '../../api/schema';

type ResourceType = components['schemas']['ResourceType'];

/** Discover filters kept in the URL (shareable, back-button friendly). No personal data. */
export interface DiscoverSearch {
  q?: string;
  /** Omitted means the West Africa default; 'all' is "Toute l'Afrique". */
  region?: 'all';
  country?: string;
  category?: string;
  type?: ResourceType;
}

export function validateDiscoverSearch(raw: Record<string, unknown>): DiscoverSearch {
  const out: DiscoverSearch = {};
  if (typeof raw.q === 'string' && raw.q.trim()) out.q = raw.q.slice(0, 100);
  if (raw.region === 'all') out.region = 'all';
  if (typeof raw.country === 'string' && /^[A-Z]{2}$/.test(raw.country)) out.country = raw.country;
  if (typeof raw.category === 'string' && /^[a-z0-9-]{1,64}$/.test(raw.category)) {
    out.category = raw.category;
  }
  if (raw.type === 'api' || raw.type === 'sdk' || raw.type === 'tool') out.type = raw.type;
  return out;
}
