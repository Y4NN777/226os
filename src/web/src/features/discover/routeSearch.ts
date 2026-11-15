import type { components } from '../../api/schema';
import { parseView, type ListView } from '../../components/useListView';

type ResourceType = components['schemas']['ResourceType'];

/** Discover filters kept in the URL (shareable, back-button friendly). No personal data. */
export interface DiscoverSearch {
  q?: string;
  /** Omitted means the West Africa default; 'all' is "Toute l'Afrique". */
  region?: 'all';
  country?: string;
  category?: string;
  type?: ResourceType;
  /** Absent means « name », the default order. */
  sort?: 'verified';
  /** Cards or list; absent means the browser's remembered choice, then cards. Not a filter. */
  view?: ListView;
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
  if (raw.sort === 'verified') out.sort = 'verified';
  const view = parseView(raw.view);
  if (view) out.view = view;
  return out;
}
