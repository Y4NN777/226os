import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { ThemePreferenceProvider, useThemePreference } from './ThemeProvider';

function Probe() {
  const { mode, setMode } = useThemePreference();
  return (
    <div>
      <p>{`mode:${mode}`}</p>
      <button
        type="button"
        onClick={() => {
          setMode('dark');
        }}
      >
        dark
      </button>
    </div>
  );
}

function renderProbe() {
  return render(
    <ThemePreferenceProvider>
      <Probe />
    </ThemePreferenceProvider>,
  );
}

describe('ThemePreferenceProvider', () => {
  it('defaults to following the system', () => {
    renderProbe();
    expect(screen.getByText('mode:system')).toBeInTheDocument();
  });

  it('restores a stored preference and applies it to <html> through Astryx Theme', () => {
    localStorage.setItem('226os.theme', 'dark');
    renderProbe();
    expect(screen.getByText('mode:dark')).toBeInTheDocument();
    expect(document.documentElement).toHaveAttribute('data-theme', 'dark');
  });

  it('ignores an invalid stored preference', () => {
    localStorage.setItem('226os.theme', 'neon');
    renderProbe();
    expect(screen.getByText('mode:system')).toBeInTheDocument();
  });

  it('changes and persists the preference', async () => {
    renderProbe();
    await userEvent.click(screen.getByRole('button', { name: 'dark' }));
    expect(screen.getByText('mode:dark')).toBeInTheDocument();
    expect(document.documentElement).toHaveAttribute('data-theme', 'dark');
    expect(localStorage.getItem('226os.theme')).toBe('dark');
  });

  it('throws when used outside the provider', () => {
    expect(() => render(<Probe />)).toThrow(/ThemePreferenceProvider/);
  });
});
