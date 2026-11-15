import { useState } from 'react';
import { Avatar, type AvatarScale } from './Avatar';

/** Pixel sizes of the `Avatar` scale (Astryx `xsm` 20, `sm` 24, `md` 36, `lg` 48; hero 72). */
const pixels: Record<AvatarScale, number> = { xs: 20, sm: 24, md: 36, lg: 48, xl: 72 };

// Mirrors the `github` pattern of the API contract.
const GITHUB_LOGIN = /^[A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?$/;

/**
 * A person's GitHub avatar when the API gave a valid username (the server already hides it when
 * the member made the field private), the initials `Avatar` otherwise or when the picture fails.
 * The login is validated before it enters the URL. No referrer is sent. Decorative next to the
 * name; alone (`isDecorative={false}`) it is an image named by the person.
 */
export function ProfilePhoto({
  name,
  github,
  size = 'md',
  isDecorative = true,
}: {
  name: string;
  github?: string | null | undefined;
  size?: AvatarScale;
  isDecorative?: boolean;
}) {
  const [failed, setFailed] = useState<string | null>(null);
  if (!github || !GITHUB_LOGIN.test(github) || failed === github) {
    return <Avatar name={name} size={size} isDecorative={isDecorative} />;
  }
  const px = pixels[size];
  return (
    <img
      src={`https://avatars.githubusercontent.com/${github}?s=${String(px * 2)}`}
      alt={isDecorative ? '' : name}
      width={px}
      height={px}
      loading="lazy"
      decoding="async"
      referrerPolicy="no-referrer"
      style={{ borderRadius: '50%', objectFit: 'cover', flex: 'none' }}
      onError={() => {
        setFailed(github);
      }}
    />
  );
}
