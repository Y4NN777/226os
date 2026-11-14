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

const westAfrica: readonly string[] = WEST_AFRICA;

/** True when the resource covers at least one West African country (none listed: false). */
export function inWestAfrica(countries: readonly string[]): boolean {
  return countries.some((code) => westAfrica.includes(code));
}

type ListQuery = NonNullable<paths['/resources']['get']['parameters']['query']>;

/** Server-side filters. A picked country goes to the API as a one-item list; the region is still filtered client-side. */
export function listQuery(search: DiscoverSearch): ListQuery {
  const query: ListQuery = { limit: 50 };
  if (search.q) query.q = search.q;
  if (search.type) query.type = search.type;
  if (search.category) query.category = search.category;
  if (search.country) query.country = [search.country];
  return query;
}

export type Resource = components['schemas']['Resource'];
