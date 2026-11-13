import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { TextInput } from '@astryxdesign/core/TextInput';
import { useState, type SVGProps } from 'react';
import { useI18n } from '../../i18n/I18nProvider';
import { errorStatus } from './validation';

/** Astryx has no `eye` icon (only `eyeSlash`), so the open eye is a small inline SVG. */
function EyeIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      {...props}
    >
      <path d="M1.5 10S4.5 4.5 10 4.5 18.5 10 18.5 10 15.5 15.5 10 15.5 1.5 10 1.5 10Z" />
      <circle cx="10" cy="10" r="2.5" />
    </svg>
  );
}

interface PasswordFieldProps {
  label: string;
  value: string;
  name: string;
  autoComplete: 'current-password' | 'new-password';
  onChange: (value: string) => void;
  onBlur: () => void;
  description?: string;
  error?: string | undefined;
}

/**
 * Password input with the show/hide toggle beside it. TextInput has no end slot and InputGroup
 * merges the group label into the input name (tried: it breaks the accessible name), so the
 * toggle sits next to the field. Paste stays allowed.
 */
export function PasswordField({
  label,
  value,
  name,
  autoComplete,
  onChange,
  onBlur,
  description,
  error,
}: PasswordFieldProps) {
  const { t } = useI18n();
  const [visible, setVisible] = useState(false);
  return (
    <HStack gap={2} vAlign="end">
      <TextInput
        label={label}
        type={visible ? 'text' : 'password'}
        value={value}
        htmlName={name}
        autoComplete={autoComplete}
        width="100%"
        onChange={onChange}
        onBlur={onBlur}
        {...(description ? { description } : {})}
        {...errorStatus(error)}
      />
      <IconButton
        label={t('identity.password.show')}
        aria-pressed={visible}
        variant="ghost"
        icon={<Icon icon={visible ? 'eyeSlash' : EyeIcon} />}
        onClick={() => {
          setVisible((current) => !current);
        }}
      />
    </HStack>
  );
}
