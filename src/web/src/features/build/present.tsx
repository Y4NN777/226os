import { Link } from '@astryxdesign/core/Link';
import type { StatusKind } from '../../components/StatusBadge';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale } from '../../i18n/messages';

export function needsLabelKey(count: number) {
  if (count === 0) return 'build.needs.none' as const;
  return count === 1 ? ('build.needs.one' as const) : ('build.needs.many' as const);
}

/** The shared state pill for a project's open needs: open while any is open. */
export function needsKind(count: number): StatusKind {
  return count > 0 ? 'open' : 'closed';
}

/** Person link to the public profile; the profile itself decides what the caller may see. */
export function PersonLink({ username, name }: { username: string; name: string }) {
  return <Link href={`/profils/${username}`}>{name}</Link>;
}

export function formatDate(iso: string, locale: Locale): string {
  return new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeZone: 'UTC' }).format(
    new Date(iso),
  );
}

/** A date as a `<time>` element, written in the UI language. */
export function DateText({ iso }: { iso: string }) {
  const { locale } = useI18n();
  return <time dateTime={iso}>{formatDate(iso, locale)}</time>;
}
