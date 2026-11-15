import { Link } from '@astryxdesign/core/Link';
import type { ReactNode } from 'react';

/**
 * A dense list of items (prototype `ul.list > li.row`): one surface, one hairline between rows,
 * no stack of cards. Rows are `ListRow`s.
 *
 * Not Astryx `List`/`ListItem`: their label is not a heading and an `href` turns the whole item
 * into one link, while a row here needs a heading title (page outline), its own excerpt, tags,
 * meta and badges, and other links that stay reachable above the row link.
 */
export function List({
  label,
  isGrid = false,
  children,
}: {
  /** Accessible name of the list, e.g. "Ressources". */
  label: string;
  /** Rows as cards in a grid (the prototype's grid view). */
  isGrid?: boolean;
  children: ReactNode;
}) {
  return (
    <ul className={isGrid ? 'ui-list is-grid' : 'ui-list'} aria-label={label}>
      {children}
    </ul>
  );
}

export interface ListRowProps {
  title: string;
  /** The item page. The title link covers the whole row. */
  href: string;
  /** Heading level of the title in the page outline (h3 under a section h2 by default). */
  headingLevel?: 2 | 3 | 4;
  /** Leading avatar, initials or thumbnail; decorative next to the title. */
  lead?: ReactNode;
  /** One or two lines of summary, clamped. */
  excerpt?: ReactNode;
  /** A TagList. */
  tags?: ReactNode;
  /** A Meta line (author, date, counts that exist in the contract). */
  meta?: ReactNode;
  /** Badges and dates at the end of the row. */
  aside?: ReactNode;
  /** Other links or controls; they stay clickable above the row link. */
  above?: ReactNode;
}

export function ListRow({
  title,
  href,
  headingLevel = 3,
  lead,
  excerpt,
  tags,
  meta,
  aside,
  above,
}: ListRowProps) {
  const Title = `h${String(headingLevel)}` as 'h2' | 'h3' | 'h4';
  return (
    <li className={lead ? 'ui-row has-lead' : 'ui-row'}>
      {lead ? <div className="ui-row-lead">{lead}</div> : null}
      <div className="ui-row-main">
        <Title className="ui-row-title">
          <Link href={href} className="ui-row-link">
            {title}
          </Link>
        </Title>
        {excerpt ? <p className="ui-row-excerpt">{excerpt}</p> : null}
        {tags ? <div className="ui-row-above">{tags}</div> : null}
        {meta}
        {above ? <div className="ui-row-above">{above}</div> : null}
      </div>
      {aside ? <div className="ui-row-aside">{aside}</div> : null}
    </li>
  );
}
