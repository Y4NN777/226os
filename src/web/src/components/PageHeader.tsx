import { Heading } from '@astryxdesign/core/Heading';
import { Text } from '@astryxdesign/core/Text';
import type { ReactNode } from 'react';
import { Breadcrumbs, type Crumb } from './Breadcrumbs';

export interface PageHeaderProps {
  /** The page's one h1. */
  title: string;
  /** One or two sentences under the title, in the muted tone. */
  lead?: ReactNode;
  breadcrumbs?: readonly Crumb[];
  /** The page's main action(s), at the end of the title row. */
  actions?: ReactNode;
}

/**
 * Top of a page: breadcrumbs, the h1 with its actions, and the lead (prototype `.page-header`,
 * `.page-header-row`, `.lede`).
 *
 * Convention: a page renders its breadcrumbs only through this header (`breadcrumbs`), never a
 * `Breadcrumbs` of its own, so every trail sits in the same place with the same look. The last
 * crumb is the current page, as text.
 */
export function PageHeader({ title, lead, breadcrumbs, actions }: PageHeaderProps) {
  return (
    <div className="ui-page-header">
      {breadcrumbs && breadcrumbs.length > 0 ? <Breadcrumbs items={breadcrumbs} /> : null}
      <div className="ui-page-header-row">
        <Heading level={1}>{title}</Heading>
        {actions ? <div className="ui-page-header-actions">{actions}</div> : null}
      </div>
      {lead ? (
        <Text as="p" color="secondary" className="ui-page-header-lead">
          {lead}
        </Text>
      ) : null}
    </div>
  );
}
