import {
  catalogs as defaultCatalogs,
  defaultLocale,
  type Catalogs,
  type Locale,
  type MessageKey,
} from './messages';

export type Params = Record<string, string | number>;

/** Message for `locale`, else the French one, else the key itself. */
export function translate(
  locale: Locale,
  key: MessageKey,
  params?: Params,
  catalogs: Catalogs = defaultCatalogs,
): string {
  const template = catalogs[locale][key] ?? catalogs[defaultLocale][key] ?? key;
  if (!params) return template;
  return template.replace(/\{(\w+)\}/g, (placeholder, name: string) =>
    name in params ? String(params[name]) : placeholder,
  );
}
