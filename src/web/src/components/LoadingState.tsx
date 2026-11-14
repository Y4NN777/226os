import { Spinner } from '@astryxdesign/core/Spinner';
import { useI18n } from '../i18n/I18nProvider';
import type { MessageKey } from '../i18n/messages';

/** Polite, labelled loading indicator. */
export function LoadingState({ label = 'state.loading' }: { label?: MessageKey }) {
  const { t } = useI18n();
  return (
    <div role="status" className="state-loading">
      <Spinner size="lg" label={t(label)} />
    </div>
  );
}
