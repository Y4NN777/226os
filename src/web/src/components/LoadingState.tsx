import { Skeleton } from '@astryxdesign/core/Skeleton';
import { Spinner } from '@astryxdesign/core/Spinner';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { useI18n } from '../i18n/I18nProvider';
import type { MessageKey } from '../i18n/messages';

/**
 * Polite, labelled loading indicator. With `rows`, placeholder rows in the shape of a list
 * (prototype `.skeleton-row`) instead of the spinner, so the page does not jump when it loads.
 */
export function LoadingState({
  label = 'state.loading',
  rows,
}: {
  label?: MessageKey;
  rows?: number;
}) {
  const { t } = useI18n();
  if (rows) {
    return (
      <div role="status" className="ui-skeleton-list">
        <VisuallyHidden>{t(label)}</VisuallyHidden>
        {Array.from({ length: rows }, (_, index) => (
          <div key={index} className="ui-skeleton-row" aria-hidden="true">
            <Skeleton width="70%" height={18} index={index} />
            <Skeleton width="45%" height={13} index={index} />
          </div>
        ))}
      </div>
    );
  }
  return (
    <div role="status" className="state-loading">
      <Spinner size="lg" label={t(label)} />
    </div>
  );
}
