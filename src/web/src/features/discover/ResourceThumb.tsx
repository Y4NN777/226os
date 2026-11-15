/** Removes a trailing "(…)" such as "(démo)", so demo entries do not all read "D". */
function baseName(name: string): string {
  return name.replace(/\s*\([^)]*\)\s*$/u, '').trim() || name;
}

/** First letter or digit of the first two words, as in the prototype thumbs ("Kalimba Pay" -> KP). */
export function initials(name: string): string {
  const words = baseName(name)
    .split(/\s+/u)
    .map((word) => word.replace(/[^\p{L}\p{N}]/gu, ''))
    .filter(Boolean);
  const letters = words.slice(0, 2).map((word) => (word[0] ?? '').toUpperCase());
  return letters.join('') || '?';
}

/** One of five token pairs, picked from the name so a listing keeps its colour everywhere. */
export function tone(name: string): number {
  let sum = 0;
  for (const char of baseName(name)) sum = (sum + (char.codePointAt(0) ?? 0)) % 997;
  return sum % 5;
}

/**
 * Decorative monogram tile (prototype `App.thumb`): the real name is always next to it, so it is
 * hidden from assistive technology. No remote logo is loaded (DESIGN.md: privacy, low bandwidth).
 */
export function ResourceThumb({ name, size = 'md' }: { name: string; size?: 'md' | 'lg' }) {
  return (
    <span
      className={`discover-thumb discover-thumb-${size} discover-tone-${String(tone(name))}`}
      aria-hidden="true"
    >
      {initials(name)}
    </span>
  );
}
