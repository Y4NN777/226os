import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { setLanguages } from '../test/render';
import { I18nProvider, useI18n } from './I18nProvider';

function Probe() {
  const { locale, setLocale, t } = useI18n();
  return (
    <div>
      <p>{`${locale}|${t('nav.settings')}`}</p>
      <button
        type="button"
        onClick={() => {
          setLocale('en');
        }}
      >
        en
      </button>
    </div>
  );
}

function renderProbe() {
  return render(
    <I18nProvider>
      <Probe />
    </I18nProvider>,
  );
}

describe('I18nProvider', () => {
  it('detects English from the browser languages and sets <html lang>', () => {
    setLanguages(['en-US']);
    renderProbe();
    expect(screen.getByText('en|Settings')).toBeInTheDocument();
    expect(document.documentElement.lang).toBe('en');
  });

  it('falls back to French for unsupported languages', () => {
    setLanguages(['de-DE']);
    renderProbe();
    expect(screen.getByText('fr|Réglages')).toBeInTheDocument();
    expect(document.documentElement.lang).toBe('fr');
  });

  it('lets an explicit stored choice win over the browser language', () => {
    setLanguages(['fr-FR']);
    localStorage.setItem('226os.locale', 'en');
    renderProbe();
    expect(screen.getByText('en|Settings')).toBeInTheDocument();
  });

  it('ignores an invalid stored value', () => {
    setLanguages(['en-US']);
    localStorage.setItem('226os.locale', 'klingon');
    renderProbe();
    expect(screen.getByText('en|Settings')).toBeInTheDocument();
  });

  it('switches locale at runtime and persists the choice', async () => {
    setLanguages(['fr-FR']);
    renderProbe();
    await userEvent.click(screen.getByRole('button', { name: 'en' }));
    expect(screen.getByText('en|Settings')).toBeInTheDocument();
    expect(document.documentElement.lang).toBe('en');
    expect(localStorage.getItem('226os.locale')).toBe('en');
  });

  it('throws when useI18n is used outside the provider', () => {
    expect(() => render(<Probe />)).toThrow(/I18nProvider/);
  });
});
