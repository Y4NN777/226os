import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';

/**
 * A country as its ISO code in a small mono chip, never a flag (prototype `.country`). Screen
 * readers hear the full name instead of the letters.
 */
export function CountryCode({ code, name }: { code: string; name: string }) {
  return (
    <span className="ui-country">
      <span className="ui-country-code" aria-hidden="true">
        {code}
      </span>
      <VisuallyHidden>{name}</VisuallyHidden>
    </span>
  );
}
