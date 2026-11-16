import { Banner } from '@astryxdesign/core/Banner';
import { useEffect, useRef, type ReactNode } from 'react';
import { ErrorState } from '../../components/ErrorState';
import { useI18n } from '../../i18n/I18nProvider';

/** Takes focus when it appears, so a keyboard or screen-reader user lands on the outcome. */
export function FocusRegion({
  role,
  children,
}: {
  role?: 'alert' | 'status';
  children: ReactNode;
}) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    ref.current?.focus();
  }, []);
  return (
    <div ref={ref} tabIndex={-1} {...(role ? { role } : {})}>
      {children}
    </div>
  );
}

/** 429: calm, without the exact wait (the countdown lives on the submit button). */
export function RateLimitNotice() {
  const { t } = useI18n();
  return (
    <FocusRegion role="alert">
      <Banner
        status="warning"
        title={t('identity.limit.title')}
        description={t('identity.limit.body')}
        collapsible={false}
      />
    </FocusRegion>
  );
}

/** Network, 5xx and anything unexpected: localized, with the support reference. */
export function FailureNotice({ error }: { error: unknown }) {
  return (
    <FocusRegion>
      <ErrorState error={error} title="identity.error.title" />
    </FocusRegion>
  );
}
