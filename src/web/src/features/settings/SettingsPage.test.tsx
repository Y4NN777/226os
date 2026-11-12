import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { renderWithProviders, setLanguages } from '../../test/render';
import { SettingsPage } from './SettingsPage';

describe('SettingsPage display-name form', () => {
  it('rejects an empty name on submit and stores nothing', async () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<SettingsPage />);
    await userEvent.click(screen.getByRole('button', { name: 'Enregistrer' }));
    // Astryx shows the message and mirrors it in an assertive live region.
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Saisissez un nom d’affichage.');
    });
    expect(screen.getByLabelText(/Nom d’affichage/)).toBeInvalid();
    expect(localStorage.getItem('226os.displayName')).toBeNull();
  });

  it('rejects a one-character name on blur with the minimum length', async () => {
    setLanguages(['en-US']);
    renderWithProviders(<SettingsPage />);
    await userEvent.type(screen.getByLabelText(/Display name/), 'A');
    await userEvent.tab();
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'The name must be at least 2 characters long.',
      );
    });
  });

  it('saves a valid trimmed name on this device and confirms it', async () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<SettingsPage />);
    await userEvent.type(screen.getByLabelText(/Nom d’affichage/), '  Awa  ');
    await userEvent.click(screen.getByRole('button', { name: 'Enregistrer' }));
    expect(await screen.findByText('Nom enregistré sur cet appareil.')).toBeInTheDocument();
    expect(localStorage.getItem('226os.displayName')).toBe('Awa');
  });

  it('restores a previously saved name', () => {
    setLanguages(['fr-FR']);
    localStorage.setItem('226os.displayName', 'Awa');
    renderWithProviders(<SettingsPage />);
    expect(screen.getByLabelText(/Nom d’affichage/)).toHaveValue('Awa');
  });

  it('keeps unsaved edits when the language changes, and only relabels the UI', async () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<SettingsPage />);
    await userEvent.type(screen.getByLabelText(/Nom d’affichage/), 'Moussa');
    await userEvent.click(screen.getByRole('radio', { name: 'English' }));
    const input = await screen.findByLabelText(/Display name/);
    expect(input).toHaveValue('Moussa');
    expect(localStorage.getItem('226os.locale')).toBe('en');
  });

  it('lets the user choose a theme preference', async () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<SettingsPage />);
    expect(screen.getByRole('radio', { name: 'Système' })).toBeChecked();
    await userEvent.click(screen.getByRole('radio', { name: 'Sombre' }));
    expect(screen.getByRole('radio', { name: 'Sombre' })).toBeChecked();
    expect(document.documentElement).toHaveAttribute('data-theme', 'dark');
    expect(localStorage.getItem('226os.theme')).toBe('dark');
  });
});
