import { Avatar } from '@astryxdesign/core/Avatar';
import { Icon } from '@astryxdesign/core/Icon';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { CheckCircleGlyph, CircleDotGlyph } from './icons';
import type { Question } from './queries';

type Author = Question['author'];

const pillKey = {
  open: 'solve.pill.open',
  resolved: 'solve.pill.resolved',
} as const satisfies Partial<Record<Question['state'], MessageKey>>;

/**
 * Icon and word on a neutral fill (the prototype's state pill). Only the two states the
 * prototype labels get one: "answered" is already told by the answer count next to it.
 */
export function StatePill({ state }: { state: Question['state'] }) {
  const { t } = useI18n();
  if (state === 'answered') return null;
  return (
    <span className={`solve-pill solve-pill-${state}`}>
      <Icon icon={state === 'resolved' ? CheckCircleGlyph : CircleDotGlyph} size="sm" />
      {t(pillKey[state])}
    </span>
  );
}

/** Technical tags as quiet monospace words; the "#" comes from CSS, so it is not read out. */
export function TagList({ tags }: { tags: readonly string[] }) {
  if (tags.length === 0) return null;
  return (
    <ul className="solve-tags">
      {tags.map((tag) => (
        <li key={tag} className="solve-tag">
          {tag}
        </li>
      ))}
    </ul>
  );
}

export function countryName(code: string, locale: string): string {
  try {
    return new Intl.DisplayNames([locale], { type: 'region' }).of(code) ?? code;
  } catch {
    return code;
  }
}

/** The country as its text code (never a flag), with the full name for screen readers. */
export function CountryCode({ code }: { code: string }) {
  const { locale } = useI18n();
  return (
    <span className="solve-country">
      <span aria-hidden="true">{code}</span>
      <VisuallyHidden>{countryName(code, locale)}</VisuallyHidden>
    </span>
  );
}

/** Initials avatar (no remote image), display name and country code. */
export function Person({ author }: { author: Author }) {
  return (
    <span className="solve-person">
      {/* The name is written next to it: the avatar is decorative, with no name tooltip. */}
      <span aria-hidden="true" className="solve-person-avatar">
        <Avatar name={author.display_name} size="xsm" tooltip={false} />
      </span>
      <span className="solve-person-name">{author.display_name}</span>
      {author.country ? <CountryCode code={author.country} /> : null}
    </span>
  );
}

/** "10 nov. 2026" / "10 Nov 2026", in UTC so the day never shifts with the visitor's zone. */
export function DateText({ iso }: { iso: string }) {
  const { locale } = useI18n();
  const label = new Intl.DateTimeFormat(locale === 'en' ? 'en-GB' : locale, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    timeZone: 'UTC',
  }).format(new Date(iso));
  return <time dateTime={iso}>{label}</time>;
}

/** The original language of a piece of user content ("FR", "EN"). */
export function LangMark({ lang }: { lang: string }) {
  return <span className="solve-lang">{lang.toUpperCase()}</span>;
}

/**
 * A one-paragraph plain-text preview of a Markdown body: code blocks dropped, links reduced to
 * their text, emphasis and heading marks removed. Rendered as text, never as HTML.
 */
export function excerptOf(markdown: string, max = 240): string {
  const text = markdown
    .replace(/```[\s\S]*?(```|$)/g, ' ')
    .replace(/!\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/^\s{0,3}(#{1,6}|>|[-*+]|\d+[.)])\s+/gm, '')
    .replace(/[`*_~]/g, '')
    .replace(/\s+/g, ' ')
    .trim();
  return text.length > max ? `${text.slice(0, max - 1).trimEnd()}…` : text;
}
