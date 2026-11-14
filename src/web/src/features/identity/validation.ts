import type { MessageKey } from '../../i18n/messages';

export const PASSWORD_MIN = 12;
export const PASSWORD_MAX = 128;
export const TOKEN_MIN = 16;
export const TOKEN_MAX = 256;
export const USERNAME_RE = /^[a-z0-9][a-z0-9_-]{2,29}$/;
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

export type Check = MessageKey | undefined;

/** Length in code points of the NFKC form, as the server counts it. */
const size = (value: string) => Array.from(value.normalize('NFKC')).length;

export function checkEmail(value: string): Check {
  const email = value.trim();
  if (!email) return 'identity.err.email.required';
  return EMAIL_RE.test(email) ? undefined : 'identity.err.email.format';
}

export function checkPasswordPresent(value: string): Check {
  return value ? undefined : 'identity.err.password.required';
}

export function checkNewPassword(value: string): Check {
  if (!value) return 'identity.err.password.required';
  if (size(value) < PASSWORD_MIN) return 'identity.err.password.tooShort';
  return size(value) > PASSWORD_MAX ? 'identity.err.password.tooLong' : undefined;
}

export function checkToken(value: string): Check {
  const token = value.trim();
  if (!token) return 'identity.err.code.required';
  return size(token) < TOKEN_MIN || size(token) > TOKEN_MAX
    ? 'identity.err.code.length'
    : undefined;
}

export function isPasswordLongEnough(value: string): boolean {
  return size(value) >= PASSWORD_MIN;
}

/** TextInput `status` prop for a message, or nothing. */
export function errorStatus(message: string | undefined) {
  return message ? { status: { type: 'error' as const, message } } : {};
}
