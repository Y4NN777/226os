import { Link } from '@astryxdesign/core/Link';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';

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

/**
 * Tags as a list (prototype `.tag-list`). `hrefFor` makes each one a link; `more` ends a
 * shortened list with "+n", whose full sentence is what a screen reader hears.
 */
export function TagList({
  tags,
  label,
  hrefFor,
  more,
}: {
  tags: readonly string[];
  /** Accessible name of the list, when the context does not give one. */
  label?: string;
  hrefFor?: (tag: string) => string;
  more?: { count: number; label: string };
}) {
  return (
    <ul className="ui-tag-list" aria-label={label}>
      {tags.map((tag) => (
        <li key={tag}>
          <Tag label={tag} {...(hrefFor ? { href: hrefFor(tag) } : {})} />
        </li>
      ))}
      {more && more.count > 0 ? (
        <li className="ui-tag-more">
          <span aria-hidden="true">+{more.count}</span>
          <VisuallyHidden>{more.label}</VisuallyHidden>
        </li>
      ) : null}
    </ul>
  );
}
