import { Avatar } from '../../components/Avatar';
import { CountryCode } from '../../components/CountryCode';
import { StatusBadge } from '../../components/StatusBadge';
import { useI18n } from '../../i18n/I18nProvider';
import type { Question } from './queries';

type Author = Question['author'];

/**
 * The state pill (shared StatusBadge). Only the two states the prototype labels get one:
 * "answered" is already told by the answer count next to it.
 */
export function QuestionState({ state }: { state: Question['state'] }) {
  const { t } = useI18n();
  if (state === 'answered') return null;
  return state === 'resolved' ? (
    <StatusBadge kind="accepted" label={t('solve.pill.resolved')} />
  ) : (
    <StatusBadge kind="open" label={t('solve.pill.open')} />
  );
}

/** Initials avatar (no remote image), display name and country code. */
export function Person({ author }: { author: Author }) {
  return (
    <span className="solve-person">
      <Avatar name={author.display_name} size="xs" />
      <span>{author.display_name}</span>
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
