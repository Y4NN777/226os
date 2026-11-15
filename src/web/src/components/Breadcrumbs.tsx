import { BreadcrumbItem, Breadcrumbs as AstryxBreadcrumbs } from '@astryxdesign/core/Breadcrumbs';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { useSyncExternalStore } from 'react';
import { useI18n } from '../i18n/I18nProvider';

export interface Crumb {
  label: string;
  /** Omitted on the last crumb: the current page is text, not a link. */
  href?: string;
}

const phoneQuery = '(max-width: 767px)';

function subscribePhone(notify: () => void) {
  const query = window.matchMedia(phoneQuery);
  query.addEventListener('change', notify);
  return () => {
    query.removeEventListener('change', notify);
  };
}

/** Whether the viewport is a phone, where a deep trail would wrap onto several lines. */
function useIsPhone() {
  return useSyncExternalStore(
    subscribePhone,
    () => window.matchMedia(phoneQuery).matches,
    () => false,
  );
}

/**
 * Where the page sits: links up the trail, the current page last (prototype `.breadcrumbs`).
 * On a phone a trail of three or more shrinks to its parent as a « ‹ Parent » back link: the
 * full trail wraps, and the last crumb repeats the h1 right below.
 */
export function Breadcrumbs({ items }: { items: readonly Crumb[] }) {
  const { t } = useI18n();
  const isPhone = useIsPhone();
  const parent = items.length >= 3 ? items[items.length - 2] : undefined;
  if (isPhone && parent?.href) {
    return (
      <nav aria-label={t('shell.breadcrumbs')} className="ui-breadcrumbs">
        <Link href={parent.href} className="ui-breadcrumbs-back">
          <Icon icon="chevronLeft" size="sm" />
          {parent.label}
        </Link>
      </nav>
    );
  }
  return (
    <AstryxBreadcrumbs
      label={t('shell.breadcrumbs')}
      variant="supporting"
      className="ui-breadcrumbs"
    >
      {items.map((item, index) => {
        const isLast = index === items.length - 1;
        return (
          <BreadcrumbItem
            key={`${item.label}-${String(index)}`}
            {...(item.href && !isLast ? { href: item.href } : { isCurrent: true })}
          >
            {item.label}
          </BreadcrumbItem>
        );
      })}
    </AstryxBreadcrumbs>
  );
}
