export const DISPLAY_NAME_KEY = '226os.displayName';
export const DISPLAY_NAME_MIN_LENGTH = 2;

export type DisplayNameError = 'form.name.required' | 'form.name.tooShort';

export function validateDisplayName(value: string): DisplayNameError | undefined {
  const trimmed = value.trim();
  if (trimmed.length === 0) return 'form.name.required';
  if (trimmed.length < DISPLAY_NAME_MIN_LENGTH) return 'form.name.tooShort';
  return undefined;
}
