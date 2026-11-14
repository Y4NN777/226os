// Same bounds as the contract and the server (1 to 80 characters after trimming; owner decision on #146).
export const DISPLAY_NAME_MAX_LENGTH = 80;

export type DisplayNameError = 'form.name.required' | 'form.name.tooLong';

export function validateDisplayName(value: string): DisplayNameError | undefined {
  const length = Array.from(value.trim()).length;
  if (length === 0) return 'form.name.required';
  if (length > DISPLAY_NAME_MAX_LENGTH) return 'form.name.tooLong';
  return undefined;
}

/** Server violation codes on /display_name, as form messages. Others are shown as a failure. */
export function displayNameViolation(code: string): DisplayNameError | undefined {
  switch (code) {
    case 'required':
    case 'too_short':
      return 'form.name.required';
    case 'too_long':
      return 'form.name.tooLong';
    default:
      return undefined;
  }
}
