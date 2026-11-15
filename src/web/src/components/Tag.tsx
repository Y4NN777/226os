import { Link } from '@astryxdesign/core/Link';

/** A technology tag in mono on a soft fill (prototype `.tag`); a link when it filters a list. */
export function Tag({ label, href }: { label: string; href?: string }) {
  return href ? (
    <Link href={href} className="ui-tag">
      {label}
    </Link>
  ) : (
    <span className="ui-tag">{label}</span>
  );
}

/** Tags as a list (prototype `.tag-list`). `hrefFor` makes each one a link. */
export function TagList({
  tags,
  label,
  hrefFor,
}: {
  tags: readonly string[];
  /** Accessible name of the list, when the context does not give one. */
  label?: string;
  hrefFor?: (tag: string) => string;
}) {
  return (
    <ul className="ui-tag-list" aria-label={label}>
      {tags.map((tag) => (
        <li key={tag}>
          <Tag label={tag} {...(hrefFor ? { href: hrefFor(tag) } : {})} />
        </li>
      ))}
    </ul>
  );
}
