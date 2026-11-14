import type { components, paths } from '../../api/schema';
import type { DiscoverSearch } from './routeSearch';

/**
 * UN M49 "Western Africa", restricted to the countries in our table.
 * The owner may adjust this list; the Discover default view reads nothing else.
 */
export const WEST_AFRICA = [
  'BJ',
  'BF',
  'CV',
  'CI',
  'GM',
  'GH',
  'GN',
  'GW',
  'LR',
  'ML',
  'MR',
  'NE',
  'NG',
  'SN',
  'SL',
  'TG',
] as const;

type ListQuery = NonNullable<paths['/resources']['get']['parameters']['query']>;

/**
 * Server-side filters. With no country picked and no `region=all`, the view is West
 * Africa: the 16 codes go to the API as one comma-separated `country` (a union).
 */
export function listQuery(search: DiscoverSearch): ListQuery {
  const query: ListQuery = { limit: 50 };
  if (search.q) query.q = search.q;
  if (search.type) query.type = search.type;
  if (search.category) query.category = search.category;
  if (search.country) query.country = [search.country];
  else if (search.region !== 'all') query.country = [...WEST_AFRICA];
  return query;
}

export type Resource = components['schemas']['Resource'];
