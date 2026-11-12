import { InternationalizationProvider } from '@astryxdesign/core/i18n';
import frFR from '@astryxdesign/core/locales/fr-FR.generated.js';
import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { readStorage, writeStorage } from '../app/storage';
import { detectLocale } from './detect';
import { locales, type Locale, type MessageKey } from './messages';
import { translate, type Params } from './translate';

const STORAGE_KEY = '226os.locale';

interface I18nValue {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: MessageKey, params?: Params) => string;
}

const I18nContext = createContext<I18nValue | null>(null);

function initialLocale(): Locale {
  const stored = readStorage(STORAGE_KEY);
  const explicit = locales.find((locale) => locale === stored);
  return explicit ?? detectLocale(navigator.languages);
}

// Astryx ships its own component strings (e.g. the Button "Loading" announcement).
const astryxMessages = { 'fr-FR': frFR };
const astryxLocale: Record<Locale, string> = { fr: 'fr-FR', en: 'en' };

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(initialLocale);

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  const value = useMemo<I18nValue>(
    () => ({
      locale,
      setLocale: (next) => {
        writeStorage(STORAGE_KEY, next);
        setLocaleState(next);
      },
      t: (key, params) => translate(locale, key, params),
    }),
    [locale],
  );

  return (
    <I18nContext value={value}>
      <InternationalizationProvider locale={astryxLocale[locale]} messages={astryxMessages}>
        {children}
      </InternationalizationProvider>
    </I18nContext>
  );
}

export function useI18n(): I18nValue {
  const value = useContext(I18nContext);
  if (!value) throw new Error('useI18n must be used inside <I18nProvider>');
  return value;
}
