import { BreadcrumbItem, Breadcrumbs as AstryxBreadcrumbs } from '@astryxdesign/core/Breadcrumbs';
import { useI18n } from '../i18n/I18nProvider';

export interface Crumb {
  label: string;
  /** Omitted on the last crumb: the current page is text, not a link. */
  href?: string;
}

/** Where the page sits: links up the trail, the current page last (prototype `.breadcrumbs`). */
export function Breadcrumbs({ items }: { items: readonly Crumb[] }) {
  const { t } = useI18n();
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
