import { act, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders, setLanguages } from '../test/render';
import { ViewToggle } from './ViewToggle';

describe('ViewToggle', () => {
  it('names the two views in French and reflects the current one with aria-pressed', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<ViewToggle value="cards" onChange={vi.fn()} />);
    expect(screen.getByRole('group', { name: 'Affichage' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cartes' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'Liste' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('names the two views in English', () => {
    setLanguages(['en-US']);
    renderWithProviders(<ViewToggle value="list" onChange={vi.fn()} />);
    expect(screen.getByRole('group', { name: 'View' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cards' })).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByRole('button', { name: 'List' })).toHaveAttribute('aria-pressed', 'true');
  });

  it('asks for the other view on click and on the keyboard, never for the current one', async () => {
    setLanguages(['fr-FR']);
    const onChange = vi.fn();
    const user = userEvent.setup({ delay: null });
    renderWithProviders(<ViewToggle value="cards" onChange={onChange} />);
    await user.click(screen.getByRole('button', { name: 'Cartes' }));
    expect(onChange).not.toHaveBeenCalled();
    act(() => {
      screen.getByRole('button', { name: 'Liste' }).focus();
    });
    await user.keyboard('{Enter}');
    expect(onChange).toHaveBeenCalledExactlyOnceWith('list');
    await user.keyboard(' ');
    expect(onChange).toHaveBeenCalledTimes(2);
  });
});
