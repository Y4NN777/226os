import { EmptyState as AstryxEmptyState } from '@astryxdesign/core/EmptyState';
import { Icon } from '@astryxdesign/core/Icon';
import type { ReactNode } from 'react';
import { InboxIcon, type SvgIcon } from './icons';

/**
 * Nothing to show yet, with the next step (prototype `.state`): icon, title, one sentence and
 * an optional action, on a soft fill. The page writes its own copy for each area.
 */
export function EmptyState({
  title,
  description,
  actions,
  icon = InboxIcon,
  headingLevel = 2,
}: {
  title: string;
  description?: string;
  actions?: ReactNode;
  icon?: SvgIcon;
  /** 1 only when the state is the whole page (404). */
  headingLevel?: 1 | 2 | 3 | 4;
}) {
  return (
    <AstryxEmptyState
      className="ui-state"
      title={title}
      headingLevel={headingLevel}
      icon={<Icon icon={icon} color="secondary" />}
      {...(description ? { description } : {})}
      {...(actions ? { actions } : {})}
    />
  );
}
