import { Link } from '@astryxdesign/core/Link';
import { countryName } from '../../components/CountryCode';
import type { StatusKind } from '../../components/StatusBadge';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale, MessageKey } from '../../i18n/messages';
import { WEST_AFRICA } from '../discover/region';
import type { Project } from './queries';

export function needsLabelKey(count: number) {
  if (count === 0) return 'build.needs.none' as const;
  return count === 1 ? ('build.needs.one' as const) : ('build.needs.many' as const);
}

/** The shared state pill for a project's open needs: open while any is open. */
export function needsKind(count: number): StatusKind {
  return count > 0 ? 'open' : 'closed';
}

const stageKeys = {
  idea: 'build.stage.idea',
  prototype: 'build.stage.prototype',
  active: 'build.stage.active',
  maintained: 'build.stage.maintained',
} as const satisfies Record<Project['stage'], MessageKey>;

export function stageKey(stage: Project['stage']) {
  return stageKeys[stage];
}

/** West Africa first, then the other codes seen (the page's rows, the URL, the project). */
export function countryOptions(extra: Iterable<string>, locale: Locale) {
  const byName = (codes: Iterable<string>) =>
    [...new Set(codes)]
      .map((code) => ({ value: code, label: countryName(code, locale) }))
      .sort((a, b) => a.label.localeCompare(b.label, locale));
  const west: readonly string[] = WEST_AFRICA;
  return [...byName(west), ...byName([...extra].filter((code) => !west.includes(code)))];
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
