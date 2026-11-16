import type { components } from '../../api/schema';
import type { MessageKey } from '../../i18n/messages';

type Intent = components['schemas']['Intent'];
type ProfileUpdate = components['schemas']['ProfileUpdate'];

export const intents = [
  'ask',
  'help',
  'integrate',
  'contribute',
  'learn',
] as const satisfies readonly Intent[];

export function isIntent(value: string): value is Intent {
  return (intents as readonly string[]).includes(value);
}

/** The checklist fields a member can complete in place. */
export const checklistItems = [
  'intent',
  'country',
  'languages',
  'skills',
  'github',
  'bio',
  'open_to_collab',
] as const;
export type ChecklistItem = (typeof checklistItems)[number];

const TAG_RE = /^[a-z0-9][a-z0-9.+#-]{0,31}$/;
const COUNTRY_RE = /^[A-Z]{2}$/;
const LANGUAGE_RE = /^[a-z]{2,3}$/;
const GITHUB_RE = /^[A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?$/;
const BIO_MAX = 500;

function list(value: string): string[] {
  return [
    ...new Set(
      value
        .split(',')
        .map((part) => part.trim().toLowerCase())
        .filter(Boolean),
    ),
  ];
}

interface Rule {
  check: (value: string) => MessageKey | undefined;
  body: (value: string) => ProfileUpdate;
}

const rules: Record<Exclude<ChecklistItem, 'intent' | 'open_to_collab'>, Rule> = {
  country: {
    check: (value) =>
      COUNTRY_RE.test(value.trim().toUpperCase()) ? undefined : 'profile.err.country',
    body: (value) => ({ country: value.trim().toUpperCase() }),
  },
  languages: {
    check: (value) => {
      const codes = list(value);
      const ok = codes.length >= 1 && codes.length <= 10 && codes.every((c) => LANGUAGE_RE.test(c));
      return ok ? undefined : 'profile.err.languages';
    },
    body: (value) => ({ languages: list(value) }),
  },
  skills: {
    check: (value) => {
      const tags = list(value);
      const ok = tags.length >= 1 && tags.length <= 30 && tags.every((tag) => TAG_RE.test(tag));
      return ok ? undefined : 'profile.err.skills';
    },
    body: (value) => ({ skills: list(value) }),
  },
  github: {
    check: (value) => (GITHUB_RE.test(value.trim()) ? undefined : 'profile.err.github'),
    body: (value) => ({ github: value.trim() }),
  },
  bio: {
    check: (value) => {
      const size = Array.from(value.trim()).length;
      return size >= 1 && size <= BIO_MAX ? undefined : 'profile.err.bio';
    },
    body: (value) => ({ bio: value.trim() }),
  },
};

export function checkItem(item: ChecklistItem, value: string): MessageKey | undefined {
  if (item === 'intent') return isIntent(value) ? undefined : 'profile.err.intent';
  if (item === 'open_to_collab') return undefined;
  return rules[item].check(value);
}

export function bodyOf(item: ChecklistItem, value: string): ProfileUpdate {
  if (item === 'intent') return { intent: isIntent(value) ? value : null };
  if (item === 'open_to_collab') return { open_to_collab: value === 'true' };
  return rules[item].body(value);
}
