import { useCallback, type CSSProperties } from 'react';

/** Where an item comes from when it enters, and goes to when it leaves. */
export type RevealKind = 'up' | 'down' | 'left' | 'right' | 'zoom-in' | 'zoom-out';

/**
 * Attributes for an element that animates with the scroll: `reveal('left', 2)` is the third item
 * of a staggered group. The look is in landing.css (`[data-reveal]`), transform and opacity only.
 */
export function reveal(kind: RevealKind, index = 0) {
  return {
    'data-reveal': kind,
    style: { ['--rd' as string]: index } as CSSProperties,
  };
}

function prefersReducedMotion() {
  return typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Callback ref for the page root: every `[data-reveal]` below it gets `data-reveal-state`:
 * `before` (still below the viewport), `in` (on screen), `after` (scrolled past). It is a
 * progressive enhancement: without IntersectionObserver, or under reduced motion, no state is ever
 * set and the stylesheet leaves everything visible. Items added later (data arriving) are picked
 * up by a MutationObserver.
 */
export function useReveal() {
  return useCallback((root: HTMLElement | null) => {
    if (!root || typeof IntersectionObserver === 'undefined' || prefersReducedMotion()) return;

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          const above = entry.boundingClientRect.top < 0;
          entry.target.setAttribute(
            'data-reveal-state',
            entry.isIntersecting ? 'in' : above ? 'after' : 'before',
          );
        }
      },
      { rootMargin: '-6% 0px -8% 0px', threshold: 0 },
    );
    const watched = new WeakSet<Element>();
    const scan = () => {
      for (const node of root.querySelectorAll('[data-reveal]')) {
        if (watched.has(node)) continue;
        watched.add(node);
        node.setAttribute('data-reveal-state', 'before');
        observer.observe(node);
      }
    };
    scan();
    const mutations = typeof MutationObserver === 'undefined' ? null : new MutationObserver(scan);
    mutations?.observe(root, { childList: true, subtree: true });

    return () => {
      observer.disconnect();
      mutations?.disconnect();
      for (const node of root.querySelectorAll('[data-reveal-state]')) {
        node.removeAttribute('data-reveal-state');
      }
    };
  }, []);
}
