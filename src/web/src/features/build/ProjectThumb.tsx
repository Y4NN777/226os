import { useState } from 'react';
import { Avatar } from '../../components/Avatar';

/** Same shape as the contract's `github_url` pattern: exactly https://github.com/<owner>/<repo>. */
const repoPattern = /^https:\/\/github\.com\/([A-Za-z0-9._-]+)\/([A-Za-z0-9._-]+)$/u;

/**
 * The social preview GitHub already renders for every public repository. Nothing is stored or
 * uploaded on our side; anything that is not a plain repository URL gets no preview.
 */
export function repoPreviewUrl(githubUrl: string): string | null {
  const match = repoPattern.exec(githubUrl);
  if (!match) return null;
  const [, owner, repo] = match;
  return `https://opengraph.githubassets.com/1/${owner ?? ''}/${repo ?? ''}`;
}

const dimensions = {
  md: { width: 96, height: 48 },
  lg: { width: 320, height: 160 },
  card: { width: 400, height: 200 },
} as const;

/**
 * A project's thumbnail: the repository's GitHub preview, loaded lazily with no referrer, or the
 * monogram tile when there is no preview or it fails to load (offline, private repository, slow
 * network). Always decorative: the project name is written next to it.
 */
type ThumbProps = {
  name: string;
  githubUrl: string;
  /** `card` fills the width of a card, 2:1. */
  size?: 'md' | 'lg' | 'card';
  /** false: render nothing instead of the monogram tile (where the page already shows that tile). */
  fallback?: boolean;
};

export function ProjectThumb(props: ThumbProps) {
  // A new repository URL is a new image: the failed state of the previous one must not stick.
  return <Thumb key={props.githubUrl} {...props} />;
}

function Thumb({ name, githubUrl, size = 'md', fallback = true }: ThumbProps) {
  const [failed, setFailed] = useState(false);
  const src = repoPreviewUrl(githubUrl);
  if (!src || failed) {
    if (!fallback) return null;
    return size === 'card' ? (
      <span className="build-thumb-fallback">
        <Avatar name={name} variant="tile" size="lg" />
      </span>
    ) : (
      <Avatar name={name} variant="tile" size={size} />
    );
  }
  return (
    <img
      className={`build-thumb build-thumb-${size}`}
      src={src}
      alt=""
      loading="lazy"
      decoding="async"
      referrerPolicy="no-referrer"
      width={dimensions[size].width}
      height={dimensions[size].height}
      onError={() => {
        setFailed(true);
      }}
    />
  );
}
