import { Card } from '@astryxdesign/core/Card';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import type { ComponentProps, ReactNode } from 'react';
import { useI18n } from '../../i18n/I18nProvider';
import { isLocale, locales } from '../../i18n/messages';
import { useThemePreference, type ThemeMode } from '../../theme/ThemeProvider';
import { EyeIcon, SlidersIcon, UserIcon } from '../profile/icons';
import { DisplayName } from './DisplayName';
import { ProfileVisibility } from './ProfileVisibility';
import './settings.css';

const themeModes = ['light', 'dark', 'system'] as const satisfies readonly ThemeMode[];

function isThemeMode(value: string): value is ThemeMode {
  return (themeModes as readonly string[]).includes(value);
}

/** Section anchors: stable French ids, so `/parametres#confidentialite` links work. */
const sections = [
  { id: 'profil', key: 'profile', icon: UserIcon },
  { id: 'preferences', key: 'preferences', icon: SlidersIcon },
  { id: 'confidentialite', key: 'privacy', icon: EyeIcon },
] as const satisfies readonly {
  id: string;
  key: string;
  icon: ComponentProps<typeof Icon>['icon'];
}[];

type SectionKey = (typeof sections)[number]['key'];

/** One prototype `set-section`: a borderless panel with its title and one-line lede. */
function SettingsSection({
  id,
  sectionKey,
  children,
}: {
  id: string;
  sectionKey: SectionKey;
  children: ReactNode;
}) {
  const { t } = useI18n();
  const headingId = `${id}-titre`;
  return (
    <section id={id} aria-labelledby={headingId}>
      <Card variant="muted" padding={6}>
        <VStack gap={5}>
          <VStack gap={1}>
            <Heading level={2} id={headingId}>
              {t(`profile.settings.${sectionKey}`)}
            </Heading>
            <Text as="p" type="supporting">
              {t(`profile.settings.${sectionKey}.lede`)}
            </Text>
          </VStack>
          {children}
        </VStack>
      </Card>
    </section>
  );
}

function Preferences() {
  const { t, locale, setLocale } = useI18n();
  const { mode, setMode } = useThemePreference();
  return (
    <VStack gap={5} hAlign="start">
      {/* The control's own label is only an aria-label: show it too. */}
      <VStack gap={2} hAlign="start">
        <Text weight="semibold" aria-hidden="true">
          {t('settings.theme.label')}
        </Text>
        <SegmentedControl
          label={t('settings.theme.label')}
          value={mode}
          onChange={(value) => {
            if (isThemeMode(value)) setMode(value);
          }}
        >
          {themeModes.map((option) => (
            <SegmentedControlItem
              key={option}
              value={option}
              label={t(`settings.theme.${option}`)}
            />
          ))}
        </SegmentedControl>
      </VStack>

      <VStack gap={2} hAlign="start">
        <Text weight="semibold" aria-hidden="true">
          {t('settings.language.label')}
        </Text>
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
      </VStack>
    </VStack>
  );
}

/**
 * `parametres`, laid out as the prototype: a section nav and one panel per section. Every
 * section is on the page at once (the nav jumps to it), so an anchor link lands on a loaded
 * form and nothing is hidden behind a tab.
 */
export function SettingsPage() {
  const { t } = useI18n();

  return (
    <VStack gap={6}>
      <Heading level={1}>{t('settings.title')}</Heading>
      <div className="settings-layout">
        <nav className="settings-nav" aria-label={t('profile.settings.nav')}>
          <ul>
            {sections.map((section) => (
              <li key={section.id}>
                <a href={`#${section.id}`}>
                  <Icon icon={section.icon} size="sm" />
                  {t(`profile.settings.${section.key}`)}
                </a>
              </li>
            ))}
          </ul>
        </nav>
        <VStack gap={5} maxWidth={760}>
          <SettingsSection id="profil" sectionKey="profile">
            <DisplayName />
          </SettingsSection>
          <SettingsSection id="preferences" sectionKey="preferences">
            <Preferences />
          </SettingsSection>
          <SettingsSection id="confidentialite" sectionKey="privacy">
            <ProfileVisibility />
          </SettingsSection>
        </VStack>
      </div>
    </VStack>
  );
}
