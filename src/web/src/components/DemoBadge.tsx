import { Badge } from '@astryxdesign/core/Badge';
import { useI18n } from '../i18n/I18nProvider';

/**
 * Marks sample data so it is never mistaken for real activity (ROADMAP C15). The API
 * exposes no `provenance` field yet, so it labels typed client fixtures only.
 */
export function DemoBadge() {
  const { t } = useI18n();
  return (
    <span title={t('state.demo.hint')}>
      <Badge variant="orange" label={t('state.demo')} />
    </span>
  );
}
