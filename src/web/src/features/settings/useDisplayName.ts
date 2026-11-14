export const DISPLAY_NAME_MIN_LENGTH = 2;
export const DISPLAY_NAME_MAX_LENGTH = 80;

export type DisplayNameError = 'form.name.required' | 'form.name.tooShort' | 'form.name.tooLong';

export function validateDisplayName(value: string): DisplayNameError | undefined {
  const length = Array.from(value.trim()).length;
  if (length === 0) return 'form.name.required';
  if (length < DISPLAY_NAME_MIN_LENGTH) return 'form.name.tooShort';
  if (length > DISPLAY_NAME_MAX_LENGTH) return 'form.name.tooLong';
  return undefined;
}

/** Server violation codes on /display_name, as form messages. Others are shown as a failure. */
export function displayNameViolation(code: string): DisplayNameError | undefined {
  switch (code) {
    case 'required':
      return 'form.name.required';
    case 'too_short':
      return 'form.name.tooShort';
    case 'too_long':
      return 'form.name.tooLong';
    default:
      return undefined;
  }
}
