import { Heading } from '@astryxdesign/core/Heading';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { VStack } from '@astryxdesign/core/VStack';
import { useI18n } from '../../i18n/I18nProvider';
import { isLocale, locales } from '../../i18n/messages';
import { useThemePreference, type ThemeMode } from '../../theme/ThemeProvider';
import { DisplayName } from './DisplayName';
import { ProfileVisibility } from './ProfileVisibility';

const themeModes = ['light', 'dark', 'system'] as const satisfies readonly ThemeMode[];

function isThemeMode(value: string): value is ThemeMode {
  return (themeModes as readonly string[]).includes(value);
}

export function SettingsPage() {
  const { t, locale, setLocale } = useI18n();
  const { mode, setMode } = useThemePreference();

  return (
    <VStack gap={6} hAlign="start" maxWidth={560}>
      <Heading level={1}>{t('settings.title')}</Heading>

      <SegmentedControl
        label={t('settings.theme.label')}
        value={mode}
        onChange={(value) => {
          if (isThemeMode(value)) setMode(value);
        }}
      >
        {themeModes.map((option) => (
          <SegmentedControlItem key={option} value={option} label={t(`settings.theme.${option}`)} />
        ))}
      </SegmentedControl>

      <SegmentedControl
        label={t('settings.language.label')}
        value={locale}
        onChange={(value) => {
          if (isLocale(value)) setLocale(value);
        }}
      >
        {locales.map((option) => (
          <SegmentedControlItem
            key={option}
            value={option}
            label={t(`settings.language.${option}`)}
          />
        ))}
      </SegmentedControl>

      <DisplayName />

      <ProfileVisibility />
    </VStack>
  );
}
