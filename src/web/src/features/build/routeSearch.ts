import type { components } from '../../api/schema';

type ProjectStage = components['schemas']['ProjectStage'];
type ProjectSort = components['schemas']['ProjectSort'];

export const projectStages = [
  'idea',
  'prototype',
  'active',
  'maintained',
] as const satisfies readonly ProjectStage[];

/** The first sort is the API default: it is never written in the address. */
export const projectSorts = [
  'updated',
  'created',
  'needs',
] as const satisfies readonly ProjectSort[];

/** Build filters kept in the URL (shareable, back-button friendly). No personal data. */
export interface BuildSearch {
  q?: string;
  /** "Has open needs" (listProjects `open_needs`). */
  open?: true;
  stage?: ProjectStage;
  country?: string;
  /** Absent means « updated », the default. */
  sort?: Exclude<ProjectSort, 'updated'>;
}

export function validateBuildSearch(raw: Record<string, unknown>): BuildSearch {
  const out: BuildSearch = {};
  if (typeof raw.q === 'string' && raw.q.trim()) out.q = raw.q.trim().slice(0, 200);
  if (raw.open === true || raw.open === 'true') out.open = true;
  const stage = projectStages.find((candidate) => candidate === raw.stage);
  if (stage) out.stage = stage;
  if (typeof raw.country === 'string' && /^[A-Z]{2}$/.test(raw.country)) out.country = raw.country;
  if (raw.sort === 'created' || raw.sort === 'needs') out.sort = raw.sort;
  return out;
}
