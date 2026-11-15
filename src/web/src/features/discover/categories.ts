import type { MessageKey } from '../../i18n/messages';

/**
 * The 18 resource categories (resource_categories.code), in their sort order, as seeded by
 * migrations 00002 and 00011 (E2). The API has no category-list endpoint yet, so the codes live
 * here; a code outside this list still shows, as the raw code.
 */
export const categoryCodes = [
  'mobile-money',
  'online-checkout',
  'cross-border',
  'open-finance',
  'e-invoicing',
  'identity-kyc',
  'sms-ussd',
  'chat-messaging',
  'maps-addressing',
  'logistics',
  'open-gov',
  'energy-climate',
  'ai-language',
  'cloud-hosting',
  'dev-tools',
  'ai-models',
  'ai-tooling',
  'ai-datasets',
] as const;

export type CategoryCode = (typeof categoryCodes)[number];

export function isCategoryCode(code: string): code is CategoryCode {
  return categoryCodes.some((known) => known === code);
}

export function categoryKey(code: CategoryCode): MessageKey {
  return `discover.category.${code}`;
}

/** Localized label of a category, or the code itself when it is not one of the 18. */
export function categoryLabel(code: string, t: (key: MessageKey) => string): string {
  return isCategoryCode(code) ? t(categoryKey(code)) : code;
}
