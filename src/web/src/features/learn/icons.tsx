import type { ReactNode, SVGProps } from 'react';

/**
 * Glyphs the Astryx registry does not name, as SVG components for the `icon` prop of Astryx Icon
 * ("custom icons" in the Icon docs). Paths come from the 226os HTML prototype icon set
 * (prototypes/html/assets/css/icons.css): 24px, 2px stroke, round caps, colour = currentColor.
 * Local to Learn until the shared set lands; the names match the shell's `components/icons.tsx`.
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

export const BookIcon = stroke(
  <path d="M4 4.5A2.5 2.5 0 0 1 6.5 2H20v17H6.5A2.5 2.5 0 0 0 4 21.5zM4 21.5V4.5" />,
);
export const FileTextIcon = stroke(
  <>
    <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5z" />
    <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" />
  </>,
);
export const CodeIcon = stroke(<path d="m16 18 6-6-6-6M8 6l-6 6 6 6" />);
export const EditIcon = stroke(<path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" />);
export const SparklesIcon = stroke(
  <path d="m12 3 1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9zM19 16l.8 2.2L22 19l-2.2.8L19 22l-.8-2.2L16 19l2.2-.8z" />,
);
export const LinkIcon = stroke(
  <>
    <path d="M10 13a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1 1" />
    <path d="M14 11a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7l1-1" />
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
export const ArrowRightIcon = stroke(<path d="M5 12h14M12 5l7 7-7 7" />);
export const CheckCircleIcon = stroke(
  <>
    <circle cx="12" cy="12" r="10" />
    <path d="m8.5 12.5 2.5 2.5 4.5-5" />
  </>,
);
export const SendIcon = stroke(<path d="m22 2-7 20-4-9-9-4zM22 2 11 13" />);
export const EyeIcon = stroke(
  <>
    <path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12z" />
    <circle cx="12" cy="12" r="3" />
  </>,
);
