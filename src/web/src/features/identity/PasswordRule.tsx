import { Text } from '@astryxdesign/core/Text';
import { useI18n } from '../../i18n/I18nProvider';
import { isPasswordLongEnough } from './validation';

/** Live length rule: met or not met is said in words, never by colour alone. */
export function PasswordRule({ value }: { value: string }) {
  const { t } = useI18n();
  const met = isPasswordLongEnough(value);
  return (
    <div role="status">
      <Text type="supporting">{t(met ? 'identity.rule.met' : 'identity.rule.unmet')}</Text>
    </div>
  );
}
