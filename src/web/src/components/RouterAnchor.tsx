import { useRouter } from '@tanstack/react-router';
import type { AnchorHTMLAttributes, MouseEvent, Ref } from 'react';
import { isSafeInternalPath } from './safePath';

type Props = AnchorHTMLAttributes<HTMLAnchorElement> & { ref?: Ref<HTMLAnchorElement> };

function isPlainLeftClick(event: MouseEvent<HTMLAnchorElement>) {
  return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
}

/**
 * The link element every Astryx component renders (LinkProvider): same-origin paths
 * navigate through TanStack Router without a reload; anything else stays a plain anchor.
 */
export function RouterAnchor({ href, onClick, target, ref, children, ...rest }: Props) {
  const router = useRouter();
  const internal = href !== undefined && isSafeInternalPath(href);
  return (
    <a
      {...rest}
      ref={ref}
      href={href}
      target={target}
      onClick={(event) => {
        onClick?.(event);
        if (event.defaultPrevented || !internal || target || !isPlainLeftClick(event)) return;
        event.preventDefault();
        router.history.push(href);
      }}
    >
      {children}
    </a>
  );
}
