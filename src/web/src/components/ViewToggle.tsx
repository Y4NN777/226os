import { ToggleButton, ToggleButtonGroup } from '@astryxdesign/core/ToggleButton';
import { useI18n } from '../i18n/I18nProvider';
import { parseView, type ListView } from './useListView';

/**
 * The cards / list switch of a results line. Two toggle buttons with `aria-pressed`, one always
 * pressed: pressing the current one does nothing. Targets are 44px (see `.ui-view-toggle`).
 */
export function ViewToggle({
  value,
  onChange,
}: {
  value: ListView;
  onChange: (next: ListView) => void;
}) {
  const { t } = useI18n();
  return (
    <ToggleButtonGroup
      className="ui-view-toggle"
      label={t('view.label')}
      value={value}
      onChange={(next) => {
        const view = parseView(next);
        if (view && view !== value) onChange(view);
      }}
    >
      <ToggleButton value="cards" label={t('view.cards')} />
      <ToggleButton value="list" label={t('view.list')} />
    </ToggleButtonGroup>
  );
}
