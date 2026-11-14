import type { MessageKey } from '../../i18n/messages';

/** The contract's github_url pattern (api/openapi.yaml, ProjectCreate). */
const GITHUB_RE = /^https:\/\/github\.com\/[A-Za-z0-9._-]+\/[A-Za-z0-9._-]+$/;
const TAG_RE = /^[a-z0-9][a-z0-9.+#-]{0,31}$/;

export const DESCRIPTION_MAX = 1000;
export const TECHNOLOGIES_MAX = 15;
export const NEED_SKILLS_MAX = 10;
export const NEED_TITLE_MIN = 3;
export const NEED_TITLE_MAX = 150;
export const NEED_DESCRIPTION_MAX = 2000;
export const MESSAGE_MAX = 2000;

const length = (value: string) => Array.from(value).length;

/** "Go, mobile-money" becomes ["go", "mobile-money"]: lowercase, trimmed, no duplicates. */
export function parseTags(value: string): string[] {
  const tags = value
    .split(',')
    .map((tag) => tag.trim().toLowerCase())
    .filter(Boolean);
  return [...new Set(tags)];
}

export function checkGithubUrl(value: string): MessageKey | undefined {
  const url = value.trim();
  return url.length <= 200 && GITHUB_RE.test(url) ? undefined : 'build.err.github';
}

export function checkDescription(value: string): MessageKey | undefined {
  return length(value.trim()) > DESCRIPTION_MAX ? 'build.err.description' : undefined;
}

/** Optional: an empty list is fine for a project. */
export function checkTechnologies(value: string): MessageKey | undefined {
  const tags = parseTags(value);
  return tags.length <= TECHNOLOGIES_MAX && tags.every((tag) => TAG_RE.test(tag))
    ? undefined
    : 'build.err.technologies';
}

export function checkNeedTitle(value: string): MessageKey | undefined {
  const size = length(value.trim());
  return size < NEED_TITLE_MIN || size > NEED_TITLE_MAX ? 'project.err.needTitle' : undefined;
}

export function checkNeedDescription(value: string): MessageKey | undefined {
  const size = length(value.trim());
  return size < 1 || size > NEED_DESCRIPTION_MAX ? 'project.err.needDescription' : undefined;
}

export function checkNeedSkills(value: string): MessageKey | undefined {
  const tags = parseTags(value);
  const ok =
    tags.length >= 1 && tags.length <= NEED_SKILLS_MAX && tags.every((t) => TAG_RE.test(t));
  return ok ? undefined : 'project.err.needSkills';
}

export function checkApplicationMessage(value: string): MessageKey | undefined {
  const size = length(value.trim());
  return size < 1 || size > MESSAGE_MAX ? 'project.err.message' : undefined;
}
