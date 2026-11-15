import { useCallback } from 'react';

/**
 * Fade-and-rise-in when an element first scrolls into view. Progressive enhancement: content is
 * visible by default, and `data-reveal` is only set when IntersectionObserver exists and the
 * visitor has not asked for reduced motion. Without JS, without the observer, or with reduced
 * motion, nothing is ever hidden.
 */
export function useReveal() {
  return useCallback((node: HTMLElement | null) => {
    if (!node || typeof IntersectionObserver === 'undefined') return;
    if (typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches)
      return;

    node.setAttribute('data-reveal', 'hidden');
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          node.setAttribute('data-reveal', 'shown');
          observer.disconnect();
        }
      },
      { rootMargin: '0px 0px -10% 0px' },
    );
    observer.observe(node);
    return () => {
      observer.disconnect();
      node.removeAttribute('data-reveal');
    };
  }, []);
}
