import { Badge } from '@astryxdesign/core/Badge';
import { useI18n } from '../../i18n/I18nProvider';
import { statusKey, type VerificationStatus } from './present';

const variant = {
  verified: 'success',
  claimed: 'info',
  unverified: 'neutral',
  stale: 'warning',
  disputed: 'error',
} as const satisfies Record<
  VerificationStatus,
  'success' | 'info' | 'neutral' | 'warning' | 'error'
>;

/** Status as text first; the colour only reinforces it. */
export function VerificationBadge({ status }: { status: VerificationStatus }) {
  const { t } = useI18n();
  return <Badge variant={variant[status]} label={t(statusKey[status])} />;
}
