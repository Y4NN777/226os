import type { components } from '../../api/schema';
import { topicCodes, type TopicCode } from './topics';

/**
 * Link kinds an article can carry. The API also names `article` (and once `series`), but it answers
 * 422 for those pairs for now, so they are never offered (issue #124).
 */
export const linkTypes = ['question', 'resource', 'project'] as const;
export type LinkType = (typeof linkTypes)[number];

export interface LinkDraft {
  type: LinkType;
  id: string;
}

export interface WriteSearch {
  /** "question:<uuid>": the content the article is about, set by the page that links here. */
  link?: string;
}

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/** "question:<uuid>" becomes a typed link; anything else (kind, id, shape) is dropped. */
export function parseLink(value: string | undefined): LinkDraft | undefined {
  const [type, id, ...rest] = (value ?? '').split(':');
  if (rest.length > 0 || id === undefined) return undefined;
  const known = linkTypes.find((candidate) => candidate === type);
  return known && UUID_RE.test(id) ? { type: known, id: id.toLowerCase() } : undefined;
}

export function validateWriteSearch(raw: Record<string, unknown>): WriteSearch {
  return typeof raw.link === 'string' && parseLink(raw.link)
    ? { link: raw.link.toLowerCase() }
    : {};
}

/** Type chips: the two article kinds (listLearnArticles `type`) and the training series. */
export const learnTypes = ['article', 'tutorial', 'series'] as const;
export type LearnType = (typeof learnTypes)[number];
export const learnLevels = [
  'beginner',
  'intermediate',
  'advanced',
] as const satisfies readonly components['schemas']['Level'][];
export type LearnLevel = (typeof learnLevels)[number];
export const learnLangs = ['fr', 'en'] as const satisfies readonly components['schemas']['Lang'][];
export type LearnLang = (typeof learnLangs)[number];

/** Learn hub filters kept in the URL (shareable, back-button friendly). No personal data. */
export interface LearnSearch {
  type?: LearnType;
  /** Text search: listLearnArticles only (1..200 characters); listTrainingSeries has no `q`. */
  q?: string;
  topic?: TopicCode;
  level?: LearnLevel;
  lang?: LearnLang;
}

function pick<T extends string>(values: readonly T[], raw: unknown): T | undefined {
  return values.find((value) => value === raw);
}

/** Unknown values are dropped; `q` is dropped under the series type, which cannot be searched. */
export function validateLearnSearch(raw: Record<string, unknown>): LearnSearch {
  const out: LearnSearch = {};
  const type = pick(learnTypes, raw.type);
  if (type) out.type = type;
  const q = typeof raw.q === 'string' ? raw.q.trim().slice(0, 200) : '';
  if (q && type !== 'series') out.q = q;
  const topic = pick(topicCodes, raw.topic);
  if (topic) out.topic = topic;
  const level = pick(learnLevels, raw.level);
  if (level) out.level = level;
  const lang = pick(learnLangs, raw.lang);
  if (lang) out.lang = lang;
  return out;
}
