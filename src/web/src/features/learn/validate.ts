import type { MessageKey } from '../../i18n/messages';

export const TITLE_MIN = 5;
export const TITLE_MAX = 200;
export const SUMMARY_MAX = 500;
export const BODY_MAX = 50_000;
export const MAX_TAGS = 5;

const TAG_RE = /^[a-z0-9][a-z0-9.+#-]{0,31}$/;

const length = (value: string) => Array.from(value).length;

export function checkTitle(value: string): MessageKey | undefined {
  const size = length(value.trim());
  return size < TITLE_MIN || size > TITLE_MAX ? 'learn.err.title' : undefined;
}

export function checkSummary(value: string): MessageKey | undefined {
  return length(value.trim()) > SUMMARY_MAX ? 'learn.err.summary' : undefined;
}

export function checkBody(value: string): MessageKey | undefined {
  const size = length(value.trim());
  return size < 1 || size > BODY_MAX ? 'learn.err.body' : undefined;
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
  return ok ? undefined : 'learn.err.tags';
}
