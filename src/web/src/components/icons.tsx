import type { ReactNode, SVGProps } from 'react';

/**
 * Glyphs the Astryx registry does not name (home, book, compass, sun, moon, globe...), as SVG
 * components for the `icon` prop of Astryx Icon, IconButton and SideNavItem ("custom icons" in
 * `astryx docs icons`). Paths come from the 226os HTML prototype icon set
 * (prototypes/html/assets/css/icons.css): 24px, 2px stroke, round caps, colour = currentColor.
 */
export type SvgIcon = (props: SVGProps<SVGSVGElement>) => ReactNode;

function stroke(paths: ReactNode): SvgIcon {
  return function StrokeIcon(props: SVGProps<SVGSVGElement>) {
    return (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
        focusable="false"
        {...props}
      >
        {paths}
      </svg>
    );
  };
}

export const HomeIcon = stroke(<path d="m3 11 9-8 9 8M5 9.5V21h14V9.5" />);
export const BookIcon = stroke(
  <path d="M4 4.5A2.5 2.5 0 0 1 6.5 2H20v17H6.5A2.5 2.5 0 0 0 4 21.5zM4 21.5V4.5" />,
);
export const NetworkIcon = stroke(
  <>
    <circle cx="12" cy="5" r="2.5" />
    <circle cx="5" cy="19" r="2.5" />
    <circle cx="19" cy="19" r="2.5" />
    <path d="M12 7.5v4M12 11.5 6 17M12 11.5l6 5.5" />
  </>,
);
export const HelpIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="M9.1 9a3 3 0 0 1 5.8 1c0 2-3 2.5-3 4M12 17.5h.01" />
  </>,
);
export const CompassIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="m16 8-2 6-6 2 2-6z" />
  </>,
);
export const HammerIcon = stroke(
  <>
    <path d="m14 6 4 4M3 21l9-9M12 12l-2-2 5-5 5 5-5 5z" />
    <path d="m15 4 2-2 5 5-2 2" />
  </>,
);
export const SunIcon = stroke(
  <>
    <circle cx="12" cy="12" r="4" />
    <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
  </>,
);
export const MoonIcon = stroke(<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />);
export const GlobeIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="M2 12h20M12 2a15 15 0 0 1 0 20 15 15 0 0 1 0-20z" />
  </>,
);
export const SearchIcon = stroke(
  <>
    <circle cx="11" cy="11" r="7" />
    <path d="m21 21-4.3-4.3" />
  </>,
);
export const CheckCircleIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="m8.5 12.5 2.5 2.5 4.5-5" />
  </>,
);
export const ShieldCheckIcon = stroke(
  <>
    <path d="M12 3 4 6v6c0 5 3.5 8 8 9 4.5-1 8-4 8-9V6z" />
    <path d="m9 12 2 2 4-4" />
  </>,
);
export const ClockIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="M12 6v6l4 2" />
  </>,
);
export const MessageIcon = stroke(
  <path d="M21 15a2 2 0 0 1-2 2H8l-5 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />,
);
export const InfoIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="M12 16v-4M12 8h.01" />
  </>,
);
export const AlertCircleIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="M12 8v4M12 16h.01" />
  </>,
);
export const PlusIcon = stroke(<path d="M12 5v14M5 12h14" />);
export const MenuIcon = stroke(<path d="M4 6h16M4 12h16M4 18h16" />);
export const XIcon = stroke(<path d="M18 6 6 18M6 6l12 12" />);
export const ArrowRightIcon = stroke(<path d="M5 12h14M12 5l7 7-7 7" />);
export const UsersIcon = stroke(
  <>
    <circle cx="9" cy="8" r="3.5" />
    <path d="M2 21a7 7 0 0 1 14 0M17 4.5a3.5 3.5 0 0 1 0 7M22 21a7 7 0 0 0-4-6.3" />
  </>,
);
export const UserIcon = stroke(
  <>
    <circle cx="12" cy="8" r="4" />
    <path d="M4 21a8 8 0 0 1 16 0" />
  </>,
);
export const UserPlusIcon = stroke(
  <>
    <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
    <circle cx="9" cy="7" r="4" />
    <path d="M19 8v6M22 11h-6" />
  </>,
);
export const EditIcon = stroke(<path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" />);
export const SettingsIcon = stroke(
  <>
    <circle cx="12" cy="12" r="3" />
    <path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" />
  </>,
);
export const LogInIcon = stroke(
  <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4M10 17l5-5-5-5M15 12H3" />,
);
export const LogOutIcon = stroke(
  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" />,
);
export const SparklesIcon = stroke(
  <path d="m12 3 1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9zM19 16l.8 2.2L22 19l-2.2.8L19 22l-.8-2.2L16 19l2.2-.8z" />,
);
export const EyeIcon = stroke(
  <>
    <path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12z" />
    <circle cx="12" cy="12" r="3" />
  </>,
);
export const CodeIcon = stroke(<path d="m16 18 6-6-6-6M8 6l-6 6 6 6" />);
export const FileTextIcon = stroke(
  <>
    <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5z" />
    <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" />
  </>,
);
export const CircleDotIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <circle cx="12" cy="12" r="1" />
  </>,
);
export const InboxIcon = stroke(
  <>
    <path d="M22 12h-6l-2 3h-4l-2-3H2" />
    <path d="M5.5 5h13L22 12v6a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2v-6z" />
  </>,
);
/* Page glyphs, gathered from the solve and people pages' local sets (same prototype paths). */
export const LinkIcon = stroke(
  <>
    <path d="M10 13a5 5 0 0 0 7.5.5l3-3a5 5 0 0 0-7-7l-1.7 1.7" />
    <path d="M14 11a5 5 0 0 0-7.5-.5l-3 3a5 5 0 0 0 7 7l1.7-1.7" />
  </>,
);
export const LockIcon = stroke(
  <>
    <rect x="4" y="11" width="16" height="10" rx="2" />
    <path d="M8 11V7a4 4 0 0 1 8 0v4" />
  </>,
);
export const MapPinIcon = stroke(
  <>
    <path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0z" />
    <circle cx="12" cy="10" r="3" />
  </>,
);
export const LanguagesIcon = stroke(
  <path d="M4 5h8M8 3v2M5 5c0 4 3 7 6 8M11 5c0 3-3 7-7 8M13 21l4-10 4 10M14.5 18h5" />,
);
export const GitBranchIcon = stroke(
  <>
    <circle cx="6" cy="5" r="2.5" />
    <circle cx="6" cy="19" r="2.5" />
    <circle cx="18" cy="8" r="2.5" />
    <path d="M6 7.5v9M18 10.5c0 4-6 3-11 7" />
  </>,
);
export const SlidersIcon = stroke(
  <path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6" />,
);

/** Lucide `panel-left`: the sidebar collapse toggle. */
export const PanelLeftIcon = stroke(
  <>
    <rect width="18" height="18" x="3" y="3" rx="2" />
    <path d="M9 3v18" />
  </>,
);
