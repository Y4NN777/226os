import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { Text } from '@astryxdesign/core/Text';
import { ApiError } from '../api/http';
import { useI18n } from '../i18n/I18nProvider';
import type { MessageKey } from '../i18n/messages';
import { errorMessage } from './errorMessage';

/** A failed load: what happened, a retry when it can help, and the support reference. */
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
    <div role="alert">
      <Banner
        status="error"
        title={t(title)}
        collapsible={false}
        description={
          <>
            <Text as="p">{t(key, params)}</Text>
            {requestId ? (
              <Text as="p" type="supporting">
                {t('state.reference', { id: requestId })}
              </Text>
            ) : null}
          </>
        }
        {...(onRetry ? { endContent: <Button label={t('state.retry')} onClick={onRetry} /> } : {})}
      />
    </div>
  );
}
