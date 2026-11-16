import { Avatar as AstryxAvatar, type AvatarSize } from '@astryxdesign/core/Avatar';

export type AvatarScale = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

/** Person sizes on the Astryx scale; `xl` is the profile hero (prototype `.avatar-lg` 64, 72 here). */
const personSize: Record<AvatarScale, AvatarSize> = {
  xs: 'xsm',
  sm: 'sm',
  md: 'md',
  lg: 'lg',
  xl: 72,
};

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

/** One of five token pairs, picked from the name so an item keeps its colour everywhere. */
export function tone(name: string): number {
  let sum = 0;
  for (const char of baseName(name)) sum = (sum + (char.codePointAt(0) ?? 0)) % 997;
  return sum % 5;
}

/**
 * A person's initials in a circle (prototype `.avatar`), or an item's monogram tile (prototype
 * `App.thumb`: resources, projects) with `variant="tile"`. No remote picture is loaded (privacy,
 * low bandwidth). Next to the name (the usual case) it is decorative and hidden from assistive
 * technology; a person alone takes `isDecorative={false}` and is an image named by the person.
 * A tile is always decorative: the item's name is written next to it.
 */
export function Avatar({
  name,
  size = 'md',
  variant = 'person',
  isDecorative = true,
}: {
  name: string;
  size?: AvatarScale;
  variant?: 'person' | 'tile';
  isDecorative?: boolean;
}) {
  if (variant === 'tile') {
    return (
      <span className={`ui-tile ui-tile-${size} ui-tone-${String(tone(name))}`} aria-hidden="true">
        {initials(name)}
      </span>
    );
  }
  const avatar = (
    <AstryxAvatar name={name} size={personSize[size]} tooltip={false} className="ui-avatar" />
  );
  return isDecorative ? (
    <span aria-hidden="true" className="ui-avatar-wrap">
      {avatar}
    </span>
  ) : (
    avatar
  );
}
