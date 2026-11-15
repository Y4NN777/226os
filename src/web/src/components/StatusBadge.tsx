import { Badge } from '@astryxdesign/core/Badge';
import { Icon } from '@astryxdesign/core/Icon';
import { useI18n } from '../i18n/I18nProvider';
import type { MessageKey } from '../i18n/messages';
import {
  CheckCircleIcon,
  CircleDotIcon,
  ClockIcon,
  HelpIcon,
  InfoIcon,
  MessageIcon,
  ShieldCheckIcon,
  type SvgIcon,
} from './icons';

export type StatusKind =
  'verified' | 'stale' | 'unverified' | 'accepted' | 'answered' | 'later' | 'demo' | 'open';

type Tint = 'info' | 'warning' | 'secondary' | 'success';

const kinds: Record<StatusKind, { icon: SvgIcon; tint: Tint; label: MessageKey }> = {
  verified: { icon: ShieldCheckIcon, tint: 'info', label: 'status.verified' },
  stale: { icon: ClockIcon, tint: 'warning', label: 'status.stale' },
  unverified: { icon: HelpIcon, tint: 'secondary', label: 'status.unverified' },
  accepted: { icon: CheckCircleIcon, tint: 'success', label: 'status.accepted' },
  answered: { icon: MessageIcon, tint: 'success', label: 'status.answered' },
  later: { icon: ClockIcon, tint: 'secondary', label: 'status.later' },
  demo: { icon: InfoIcon, tint: 'secondary', label: 'state.demo' },
  open: { icon: CircleDotIcon, tint: 'success', label: 'status.open' },
};

/**
 * A state as icon + word on a neutral pill (prototype `.state-pill`): the word carries the
 * meaning, the tinted icon only reinforces it, never colour alone. `label` replaces the default
 * word when the page needs its own (grammar, a more precise status).
 */
export function StatusBadge({ kind, label }: { kind: StatusKind; label?: string }) {
  const { t } = useI18n();
  const { icon, tint, label: key } = kinds[kind];
  const badge = (
    <Badge
      variant="neutral"
      className={`ui-status ui-status-${kind}`}
      icon={<Icon icon={icon} color={tint} size="sm" />}
      label={label ?? t(key)}
    />
  );
  // Demo data says why it is marked, like DemoBadge (ROADMAP C15).
  return kind === 'demo' ? <span title={t('state.demo.hint')}>{badge}</span> : badge;
}
