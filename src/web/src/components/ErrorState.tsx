import { Button } from '@astryxdesign/core/Button';
import { Icon } from '@astryxdesign/core/Icon';
import { Text } from '@astryxdesign/core/Text';
import { ApiError } from '../api/http';
import { useI18n } from '../i18n/I18nProvider';
import type { MessageKey } from '../i18n/messages';
import { errorMessage } from './errorMessage';
import { AlertCircleIcon } from './icons';

/**
 * A failed load (prototype `.state.state-error`): what happened, a retry when it can help, and
 * the support reference. The title is strong text, not a heading, so it never breaks the
 * outline of the page it appears in.
 */
export function ErrorState({
  error,
  onRetry,
  title = 'error.title',
}: {
  error: unknown;
  onRetry?: () => void;
  title?: MessageKey;
}) {
  const { t } = useI18n();
  const { key, params } = errorMessage(error);
  const requestId = error instanceof ApiError ? error.requestId : undefined;
  return (
    <div role="alert" className="ui-state ui-state-error">
      <Icon icon={AlertCircleIcon} color="error" size="lg" />
      <p className="ui-state-title">{t(title)}</p>
      <Text as="p" color="secondary">
        {t(key, params)}
      </Text>
      {requestId ? (
        <Text as="p" type="supporting" color="secondary">
          {t('state.reference', { id: requestId })}
        </Text>
      ) : null}
      {onRetry ? (
        <div className="ui-state-actions">
          <Button label={t('state.retry')} onClick={onRetry} />
        </div>
      ) : null}
    </div>
  );
}
