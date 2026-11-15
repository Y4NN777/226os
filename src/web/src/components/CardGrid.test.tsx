import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../test/render';
import { CardGrid } from './CardGrid';
import { ListRow } from './List';

describe('CardGrid', () => {
  it('is a labelled list of cards, each one a link to its item', () => {
    renderWithProviders(
      <CardGrid label="Projets">
        <ListRow title="kora-sdk" href="/build/1" excerpt="Un SDK." />
        <ListRow title="sahel-docs" href="/build/2" />
      </CardGrid>,
    );
    const list = screen.getByRole('list', { name: 'Projets' });
    expect(list).toHaveClass('ui-card-grid');
    expect(within(list).getAllByRole('listitem')).toHaveLength(2);
    expect(screen.getByRole('link', { name: 'kora-sdk' })).toHaveAttribute('href', '/build/1');
  });
});
