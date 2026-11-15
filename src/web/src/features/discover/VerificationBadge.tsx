import { Badge } from '@astryxdesign/core/Badge';
import { Icon } from '@astryxdesign/core/Icon';
import { useI18n } from '../../i18n/I18nProvider';
import { statusKey, type VerificationStatus } from './present';

/** Prototype state pill: a neutral fill, a status icon in its colour, and the word. */
const glyph = {
  verified: { icon: 'success', color: 'success' },
  claimed: { icon: 'info', color: 'info' },
  unverified: { icon: 'info', color: 'secondary' },
  stale: { icon: 'clock', color: 'warning' },
  disputed: { icon: 'warning', color: 'error' },
} as const satisfies Record<
  VerificationStatus,
  {
    icon: 'success' | 'info' | 'clock' | 'warning';
    color: 'success' | 'info' | 'secondary' | 'warning' | 'error';
  }
>;

/** Status as text first; the icon and its colour only reinforce it. */
export function VerificationBadge({ status }: { status: VerificationStatus }) {
  const { t } = useI18n();
  const { icon, color } = glyph[status];
  return (
    <Badge
      variant="neutral"
      icon={<Icon icon={icon} color={color} size="sm" />}
      label={t(statusKey[status])}
    />
  );
}
