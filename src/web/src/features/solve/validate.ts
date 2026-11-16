import type { components } from '../../api/schema';
import type { MessageKey } from '../../i18n/messages';

type Snippet = components['schemas']['Snippet'];

export const TITLE_MIN = 5;
export const TITLE_MAX = 200;
export const BODY_MAX = 20_000;
export const MAX_TAGS = 5;
export const SNIPPET_MAX = 10_000;

const TAG_RE = /^[a-z0-9][a-z0-9.+#-]{0,31}$/;
const LANGUAGE_RE = /^[a-z0-9+#.-]{1,32}$/;
const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/** Only the kinds the API accepts as question links (QuestionLinkRef). */
export const linkTypes = ['resource', 'project', 'article'] as const;
export type LinkType = (typeof linkTypes)[number];

const length = (value: string) => Array.from(value).length;

export function checkTitle(value: string): MessageKey | undefined {
  const size = length(value.trim());
  return size < TITLE_MIN || size > TITLE_MAX ? 'solve.err.title' : undefined;
}

export function checkBody(value: string): MessageKey | undefined {
  const size = length(value.trim());
  return size < 1 || size > BODY_MAX ? 'solve.err.body' : undefined;
}

export function checkAnswer(value: string): MessageKey | undefined {
  return checkBody(value) ? 'solve.err.answer' : undefined;
}

/** "Go, webhook" becomes ["go", "webhook"]: lowercase, trimmed, no duplicates. */
export function parseTags(value: string): string[] {
  const tags = value
    .split(',')
    .map((tag) => tag.trim().toLowerCase())
    .filter(Boolean);
  return [...new Set(tags)];
}

export function checkTags(value: string): MessageKey | undefined {
  const tags = parseTags(value);
  const ok = tags.length >= 1 && tags.length <= MAX_TAGS && tags.every((tag) => TAG_RE.test(tag));
  return ok ? undefined : 'solve.err.tags';
}

export function checkSnippetLanguage(language: string, code: string): MessageKey | undefined {
  if (!code.trim()) return undefined;
  return LANGUAGE_RE.test(language.trim().toLowerCase()) ? undefined : 'solve.err.snippet.language';
}

export function checkSnippetCode(code: string): MessageKey | undefined {
  return length(code) > SNIPPET_MAX ? 'solve.err.snippet.code' : undefined;
}

export function isValidLinkId(value: string): boolean {
  return UUID_RE.test(value.trim());
}

/** At most one snippet per question (S3 decision): none when the code is empty. */
export function snippetsOf(language: string, code: string): Snippet[] {
  return code.trim() ? [{ language: language.trim().toLowerCase(), code }] : [];
}
