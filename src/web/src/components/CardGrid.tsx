import type { ReactNode } from 'react';

/**
 * The card view of a results list: the same rows as `List`, laid out as cards in a grid (1 column
 * on a phone, 2 from 768px, 3 from 1024px). Children are `ListRow`s (or a feature's own row): the
 * title link covers the whole card, and the lead (thumbnail or tile) sits at the top.
 */
export function CardGrid({ label, children }: { label: string; children: ReactNode }) {
  return (
    <ul className="ui-card-grid" aria-label={label}>
      {children}
    </ul>
  );
}
