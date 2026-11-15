import type { SVGProps } from 'react';

/**
 * Line glyphs the Astryx fallback registry does not ship (it has no speech bubble, pencil, plus,
 * open-state dot or link). Same drawing rules as the registry: 24x24 view box, currentColor
 * stroke of 1.5, decorative only. Passed to Astryx `Icon` as icon components; the lead can swap
 * them for the shared icon set once the shell lands.
 */
function Glyph({ children, ...props }: SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.5}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
      {...props}
    >
      {children}
    </svg>
  );
}

export function ChatGlyph(props: SVGProps<SVGSVGElement>) {
  return (
    <Glyph {...props}>
      <path d="M20 15a2 2 0 0 1-2 2H8l-4 4V5a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2z" />
    </Glyph>
  );
}

export function CheckCircleGlyph(props: SVGProps<SVGSVGElement>) {
  return (
    <Glyph {...props}>
      <circle cx="12" cy="12" r="9" />
      <path d="m8.5 12.5 2.5 2.5 4.5-5" />
    </Glyph>
  );
}

export function CircleDotGlyph(props: SVGProps<SVGSVGElement>) {
  return (
    <Glyph {...props}>
      <circle cx="12" cy="12" r="9" />
      <circle cx="12" cy="12" r="1.5" fill="currentColor" />
    </Glyph>
  );
}

export function PencilGlyph(props: SVGProps<SVGSVGElement>) {
  return (
    <Glyph {...props}>
      <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" />
    </Glyph>
  );
}

export function PlusGlyph(props: SVGProps<SVGSVGElement>) {
  return (
    <Glyph {...props}>
      <path d="M12 5v14M5 12h14" />
    </Glyph>
  );
}

export function LinkGlyph(props: SVGProps<SVGSVGElement>) {
  return (
    <Glyph {...props}>
      <path d="M10 13a5 5 0 0 0 7.5.5l3-3a5 5 0 0 0-7-7l-1.7 1.7" />
      <path d="M14 11a5 5 0 0 0-7.5-.5l-3 3a5 5 0 0 0 7 7l1.7-1.7" />
    </Glyph>
  );
}
