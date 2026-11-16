import { Badge } from '@astryxdesign/core/Badge';
import { Icon } from '@astryxdesign/core/Icon';
import { useI18n } from '../i18n/I18nProvider';
import type { MessageKey } from '../i18n/messages';
import {
  AlertCircleIcon,
  CheckCircleIcon,
  CircleDotIcon,
  ClockIcon,
  HelpIcon,
  InfoIcon,
  LockIcon,
  MessageIcon,
  ShieldCheckIcon,
  type SvgIcon,
} from './icons';

export type StatusKind =
  | 'verified'
  | 'stale'
  | 'unverified'
  | 'accepted'
  | 'answered'
  | 'later'
  | 'demo'
  | 'open'
  | 'closed'
  | 'claimed'
  | 'disputed'
  | 'pending'
  | 'private';

type Tint = 'info' | 'warning' | 'secondary' | 'success' | 'error';

// Icons and tints follow the prototype's App.statePill (assets/js/app.js, components.css).
const kinds: Record<StatusKind, { icon: SvgIcon; tint: Tint; label: MessageKey }> = {
  verified: { icon: ShieldCheckIcon, tint: 'info', label: 'status.verified' },
  stale: { icon: ClockIcon, tint: 'warning', label: 'status.stale' },
  unverified: { icon: HelpIcon, tint: 'secondary', label: 'status.unverified' },
  accepted: { icon: CheckCircleIcon, tint: 'success', label: 'status.accepted' },
  answered: { icon: MessageIcon, tint: 'success', label: 'status.answered' },
  later: { icon: ClockIcon, tint: 'secondary', label: 'status.later' },
  demo: { icon: InfoIcon, tint: 'secondary', label: 'state.demo' },
  open: { icon: CircleDotIcon, tint: 'success', label: 'status.open' },
  closed: { icon: LockIcon, tint: 'secondary', label: 'status.closed' },
  claimed: { icon: InfoIcon, tint: 'info', label: 'status.claimed' },
  disputed: { icon: AlertCircleIcon, tint: 'error', label: 'status.disputed' },
  pending: { icon: ClockIcon, tint: 'warning', label: 'status.pending' },
  private: { icon: LockIcon, tint: 'secondary', label: 'status.private' },
};

/**
 * A state as icon + word on a neutral pill (prototype `.state-pill`): the word carries the
 * meaning, the tinted icon only reinforces it, never colour alone. `label` replaces the default
 * word when the page needs its own (grammar, a more precise status); `title` adds a hint.
 */
export function StatusBadge({
  kind,
  label,
  title,
}: {
  kind: StatusKind;
  label?: string;
  title?: string;
}) {
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
  const hint = title ?? (kind === 'demo' ? t('state.demo.hint') : undefined);
  return hint ? (
    <span className="ui-status-wrap" title={hint}>
      {badge}
    </span>
  ) : (
    badge
  );
}
