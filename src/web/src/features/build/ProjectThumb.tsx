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

/**
 * A project's thumbnail: the repository's GitHub preview, loaded lazily with no referrer, or the
 * monogram tile when there is no preview or it fails to load (offline, private repository, slow
 * network). Always decorative: the project name is written next to it.
 */
export function ProjectThumb({
  name,
  githubUrl,
  size = 'md',
}: {
  name: string;
  githubUrl: string;
  size?: 'md' | 'lg';
}) {
  const [failed, setFailed] = useState(false);
  const src = repoPreviewUrl(githubUrl);
  if (!src || failed) return <Avatar name={name} variant="tile" size={size} />;
  return (
    <img
      className={`build-thumb build-thumb-${size}`}
      src={src}
      alt=""
      loading="lazy"
      decoding="async"
      referrerPolicy="no-referrer"
      width={size === 'lg' ? 320 : 96}
      height={size === 'lg' ? 160 : 48}
      onError={() => {
        setFailed(true);
      }}
    />
  );
}
