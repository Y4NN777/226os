import { Selector } from '@astryxdesign/core/Selector';
import { TextInput } from '@astryxdesign/core/TextInput';
import { countryOptions } from '../../components/countries';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { errorStatus } from '../identity/validation';
import { stageKey } from './present';
import type { Project } from './queries';
import { projectStages } from './routeSearch';

/** The three optional project fields, shared by the create form and the owner edit form. */
type FieldProps<T> = {
  value: T;
  onChange: (value: T) => void;
  error?: MessageKey | undefined;
};

export function StageField({ value, onChange, error }: FieldProps<Project['stage']>) {
  const { t } = useI18n();
  return (
    <Selector
      label={t('build.form.stage')}
      options={projectStages.map((stage) => ({ value: stage, label: t(stageKey(stage)) }))}
      value={value}
      onChange={(picked) => {
        const stage = projectStages.find((candidate) => candidate === picked);
        if (stage) onChange(stage);
      }}
      {...errorStatus(error ? t(error) : undefined)}
    />
  );
}

export function LicenseField({ value, onChange, error }: FieldProps<string>) {
  const { t } = useI18n();
  return (
    <TextInput
      label={t('build.form.license')}
      description={t('build.form.license.hint')}
      isOptional
      value={value}
      htmlName="license"
      autoCapitalize="none"
      hasAutoCorrect={false}
      hasSpellCheck={false}
      onChange={onChange}
      {...errorStatus(error ? t(error) : undefined)}
    />
  );
}

export function CountryField({ value, onChange, error }: FieldProps<string | null>) {
  const { t, locale } = useI18n();
  return (
    <Selector
      label={t('build.form.country')}
      placeholder={t('build.filter.country.any')}
      hasClear
      hasSearch
      options={countryOptions(value ? [value] : [], locale)}
      value={value}
      onChange={onChange}
      {...errorStatus(error ? t(error) : undefined)}
    />
  );
}
