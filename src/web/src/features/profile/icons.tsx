import type { ReactNode, SVGProps } from 'react';

/*
 * The few glyphs the people pages need and Astryx does not ship. Paths come from the prototype's
 * own icon set (prototypes/html/assets/css/icons.css, original 24px stroke icons). Pass them to
 * Astryx `<Icon icon={…} />`, which sizes and colours them; they stay decorative.
 */
function glyph(paths: ReactNode) {
  function Glyph(props: SVGProps<SVGSVGElement>) {
    return (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        {...props}
      >
        {paths}
      </svg>
    );
  }
  return Glyph;
}

export const MapPinIcon = glyph(
  <>
    <path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0z" />
    <circle cx="12" cy="10" r="3" />
  </>,
);

export const LanguagesIcon = glyph(
  <path d="M4 5h8M8 3v2M5 5c0 4 3 7 6 8M11 5c0 3-3 7-7 8M13 21l4-10 4 10M14.5 18h5" />,
);

export const LockIcon = glyph(
  <>
    <rect x="4" y="11" width="16" height="10" rx="2" />
    <path d="M8 11V7a4 4 0 0 1 8 0v4" />
  </>,
);

export const EditIcon = glyph(<path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" />);

export const GitBranchIcon = glyph(
  <>
    <circle cx="6" cy="5" r="2.5" />
    <circle cx="6" cy="19" r="2.5" />
    <circle cx="18" cy="8" r="2.5" />
    <path d="M6 7.5v9M18 10.5c0 4-6 3-11 7" />
  </>,
);

export const UserIcon = glyph(
  <>
    <circle cx="12" cy="8" r="4" />
    <path d="M4 21a8 8 0 0 1 16 0" />
  </>,
);

export const UsersIcon = glyph(
  <>
    <circle cx="9" cy="8" r="3.5" />
    <path d="M2 21a7 7 0 0 1 14 0M17 4.5a3.5 3.5 0 0 1 0 7M22 21a7 7 0 0 0-4-6.3" />
  </>,
);

export const SlidersIcon = glyph(
  <path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6" />,
);

export const EyeIcon = glyph(
  <>
    <path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12z" />
    <circle cx="12" cy="12" r="3" />
  </>,
);

export const GlobeIcon = glyph(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="M2 12h20M12 2a15 15 0 0 1 0 20 15 15 0 0 1 0-20z" />
  </>,
);

export const CodeIcon = glyph(<path d="m16 18 6-6-6-6M8 6l-6 6 6 6" />);

export const CircleDotIcon = glyph(
  <>
    <circle cx="12" cy="12" r="10" />
    <circle cx="12" cy="12" r="1" />
  </>,
);
