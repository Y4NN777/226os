import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { readStorage, writeStorage } from '../../app/storage';
import { useI18n } from '../../i18n/I18nProvider';
import { isLocale, locales } from '../../i18n/messages';
import { useThemePreference, type ThemeMode } from '../../theme/ThemeProvider';
import { DISPLAY_NAME_KEY, DISPLAY_NAME_MIN_LENGTH, validateDisplayName } from './useDisplayName';

const themeModes = ['light', 'dark', 'system'] as const satisfies readonly ThemeMode[];

function isThemeMode(value: string): value is ThemeMode {
  return (themeModes as readonly string[]).includes(value);
}

export function SettingsPage() {
  const { t, locale, setLocale } = useI18n();
  const { mode, setMode } = useThemePreference();
  const [saved, setSaved] = useState(false);

  const form = useForm({
    defaultValues: { displayName: readStorage(DISPLAY_NAME_KEY) ?? '' },
    onSubmit: ({ value }) => {
      writeStorage(DISPLAY_NAME_KEY, value.displayName.trim());
      setSaved(true);
    },
  });

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

      <form
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          setSaved(false);
          void form.handleSubmit();
        }}
      >
        <VStack gap={3} hAlign="start">
          <form.Field
            name="displayName"
            validators={{
              onBlur: ({ value }) => validateDisplayName(value),
              onSubmit: ({ value }) => validateDisplayName(value),
            }}
          >
            {(field) => {
              const errorKey = field.state.meta.errors[0];
              return (
                <TextInput
                  label={t('form.name.label')}
                  description={t('form.name.description')}
                  value={field.state.value}
                  htmlName={field.name}
                  autoComplete="nickname"
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    setSaved(false);
                    field.handleChange(value);
                  }}
                  {...(errorKey
                    ? {
                        status: {
                          type: 'error' as const,
                          message: t(errorKey, { min: DISPLAY_NAME_MIN_LENGTH }),
                        },
                      }
                    : {})}
                />
              );
            }}
          </form.Field>
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <Button
                type="submit"
                variant="primary"
                label={t('form.save')}
                isLoading={isSubmitting}
              />
            )}
          </form.Subscribe>
          <div role="status">{saved ? <Text>{t('form.saved')}</Text> : null}</div>
        </VStack>
      </form>
    </VStack>
  );
}
