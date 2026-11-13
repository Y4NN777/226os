import type { components } from '../../api/schema';
import type { Locale } from '../../i18n/messages';
import type { MessageKey } from '../../i18n/messages';

export type VerificationStatus = components['schemas']['VerificationStatus'];

export const statusKey = {
  verified: 'discover.status.verified',
  claimed: 'discover.status.claimed',
  unverified: 'discover.status.unverified',
  stale: 'discover.status.stale',
  disputed: 'discover.status.disputed',
} as const satisfies Record<VerificationStatus, MessageKey>;

export const typeKey = {
  api: 'discover.type.api',
  sdk: 'discover.type.sdk',
  tool: 'discover.type.tool',
} as const satisfies Record<components['schemas']['ResourceType'], MessageKey>;

/** A date is only meaningful as "verified on" when the status says someone checked it. */
export function showsVerifiedDate(status: VerificationStatus): boolean {
  return status !== 'claimed' && status !== 'unverified';
}

export function formatDate(iso: string, locale: Locale): string {
  return new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeZone: 'UTC' }).format(
    new Date(iso),
  );
}

/** Localized region name from the platform, never a hand-kept list; the code itself as fallback. */
export function countryName(code: string, locale: Locale): string {
  try {
    return new Intl.DisplayNames([locale], { type: 'region' }).of(code) ?? code;
  } catch {
    return code;
  }
}

export function languageName(code: string, locale: Locale): string {
  try {
    return new Intl.DisplayNames([locale], { type: 'language' }).of(code) ?? code;
  } catch {
    return code;
  }
}

export type SourceKind = components['schemas']['ResourceSource']['kind'];

export const kindKey = {
  official_docs: 'discover.kind.official_docs',
  changelog: 'discover.kind.changelog',
  pricing: 'discover.kind.pricing',
  community: 'discover.kind.community',
  other: 'discover.kind.other',
  claim: 'discover.kind.claim',
} as const satisfies Record<SourceKind, MessageKey>;

/** A provider claim is evidence of what the provider says, never proof: it cannot show as verified. */
export function sourceStatus(source: components['schemas']['ResourceSource']): VerificationStatus {
  return source.kind === 'claim' && source.status === 'verified' ? 'claimed' : source.status;
}

export function hostOf(url: string): string {
  try {
    return new URL(url).host;
  } catch {
    return url;
  }
}
