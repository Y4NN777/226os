import { Icon } from '@astryxdesign/core/Icon';
import type { ReactNode } from 'react';
import type { SvgIcon } from './icons';

/** A line of small muted facts: author, date, counts (prototype `.meta`). Numbers line up. */
export function Meta({ children }: { children: ReactNode }) {
  return <div className="ui-meta">{children}</div>;
}

/** One fact; the optional icon is decorative, the text says it all. */
export function MetaItem({ icon, children }: { icon?: SvgIcon; children: ReactNode }) {
  return (
    <span className="ui-meta-item">
      {icon ? <Icon icon={icon} size="sm" color="inherit" /> : null}
      {children}
    </span>
  );
}
