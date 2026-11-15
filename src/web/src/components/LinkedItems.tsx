import { Link } from '@astryxdesign/core/Link';

export interface LinkedItem {
  key: string;
  /** The kind of content ("Question", "Projet"), written above the title, outside the link. */
  kind: string;
  title: string;
  href: string;
  /** Original language of the title. */
  lang?: string | undefined;
}

/**
 * Linked content as small tiles (prototype `.linked-items`): the kind above the title, the title
 * link covering the tile, so the link's name is the title alone.
 */
export function LinkedItems({ items, label }: { items: readonly LinkedItem[]; label?: string }) {
  return (
    <ul className="ui-linked" aria-label={label}>
      {items.map((item) => (
        <li key={item.key} className="ui-linked-item">
          <span className="ui-linked-kind">{item.kind}</span>
          <span className="ui-linked-title" lang={item.lang}>
            <Link href={item.href}>{item.title}</Link>
          </span>
        </li>
      ))}
    </ul>
  );
}
