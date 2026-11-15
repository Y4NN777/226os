import { Badge } from '@astryxdesign/core/Badge';
import { BreadcrumbItem, Breadcrumbs } from '@astryxdesign/core/Breadcrumbs';
import { Icon, type IconColor, type IconName } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale } from '../../i18n/messages';

export function needsLabelKey(count: number) {
  if (count === 0) return 'build.needs.none' as const;
  return count === 1 ? ('build.needs.one' as const) : ('build.needs.many' as const);
}

/** Technology or skill slugs as quiet neutral tags (the prototype's `.tag-list`). */
export function Technologies({ items }: { items: readonly string[] }) {
  if (items.length === 0) return null;
  return (
    <ul className="build-tags">
      {items.map((item) => (
        <li key={item} className="build-tag">
          {item}
        </li>
      ))}
    </ul>
  );
}

/** Person link to the public profile; the profile itself decides what the caller may see. */
export function PersonLink({ username, name }: { username: string; name: string }) {
  return <Link href={`/profils/${username}`}>{name}</Link>;
}

export type PillTone = 'open' | 'done' | 'muted' | 'pending';

const pillIcon: Record<PillTone, { icon: IconName; color: IconColor }> = {
  open: { icon: 'success', color: 'success' },
  done: { icon: 'checkDouble', color: 'success' },
  muted: { icon: 'stop', color: 'secondary' },
  pending: { icon: 'clock', color: 'warning' },
};

/**
 * State as icon plus word on a neutral fill (the prototype's `.state-pill`): never colour alone.
 * The icon is decorative; the label carries the state.
 */
export function StatePill({ tone, label }: { tone: PillTone; label: string }) {
  const { icon, color } = pillIcon[tone];
  return (
    <span className="build-pill">
      <Badge variant="neutral" icon={<Icon icon={icon} color={color} size="xsm" />} label={label} />
    </span>
  );
}

/** A plain neutral badge for metadata such as the private visibility or a content language. */
export function MetaBadge({ label, title }: { label: string; title?: string }) {
  return (
    <span title={title}>
      <Badge variant="neutral" label={label} />
    </span>
  );
}

const tones = ['green', 'blue', 'orange', 'teal', 'gray'] as const;

/** Initials of a project name, for its monogram tile. */
export function initials(name: string): string {
  const words = name
    .split(/[\s\-_.]+/u)
    .map((word) => word.replace(/[^\p{L}\p{N}]/gu, ''))
    .filter((word) => word.length > 0);
  const letters =
    words.length > 1
      ? `${words[0]?.[0] ?? ''}${words[1]?.[0] ?? ''}`
      : (words[0] ?? '').slice(0, 2);
  return letters.toUpperCase();
}

/**
 * Decorative monogram tile, its tone derived from the name so it never changes between visits.
 * Hidden from assistive technology: the name is always written next to it.
 */
export function Monogram({ name, size = 'md' }: { name: string; size?: 'md' | 'lg' }) {
  let hash = 0;
  for (const char of name) hash = (hash * 31 + (char.codePointAt(0) ?? 0)) >>> 0;
  const tone = tones[hash % tones.length] ?? 'gray';
  return (
    <span aria-hidden="true" className={`build-monogram build-monogram-${size} build-tone-${tone}`}>
      {initials(name)}
    </span>
  );
}

export function formatDate(iso: string, locale: Locale): string {
  return new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeZone: 'UTC' }).format(
    new Date(iso),
  );
}

/** Country name in the UI language; the code itself when the runtime has no name for it. */
export function countryName(code: string, locale: Locale): string {
  try {
    return new Intl.DisplayNames([locale], { type: 'region' }).of(code) ?? code;
  } catch {
    return code;
  }
}

export interface Crumb {
  label: string;
  /** Omitted on the last crumb: the current page is text, not a link. */
  href?: string;
}

/**
 * Where the page sits (prototype `.breadcrumbs`). Same shape as the shared shell breadcrumbs, so
 * the lead can swap this local copy for the shared one.
 */
export function BuildBreadcrumbs({ items }: { items: readonly Crumb[] }) {
  const { t } = useI18n();
  return (
    <Breadcrumbs label={t('build.breadcrumbs')} variant="supporting">
      {items.map((item, index) => {
        const isLast = index === items.length - 1;
        return (
          <BreadcrumbItem
            key={`${item.label}-${String(index)}`}
            {...(item.href && !isLast ? { href: item.href } : { isCurrent: true })}
          >
            {item.label}
          </BreadcrumbItem>
        );
      })}
    </Breadcrumbs>
  );
}

/** A date as a `<time>` element, written in the UI language. */
export function DateText({ iso }: { iso: string }) {
  const { locale } = useI18n();
  return <time dateTime={iso}>{formatDate(iso, locale)}</time>;
}
