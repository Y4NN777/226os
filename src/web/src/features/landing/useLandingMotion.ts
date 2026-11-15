import { useCallback } from 'react';

/**
 * Drives the hero loops. Sets `data-motion` (`full` or `reduced`, from prefers-reduced-motion)
 * and `data-live` (`false` while the hero is off screen or the tab is hidden), which the
 * stylesheet turns into `animation: none` and `animation-play-state: paused`. Without
 * IntersectionObserver the hero simply stays live.
 */
export function useLandingMotion() {
  return useCallback((node: HTMLElement | null) => {
    if (!node) return;
    const reduced =
      typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;
    node.setAttribute('data-motion', reduced ? 'reduced' : 'full');
    node.setAttribute('data-live', 'true');
    if (reduced) return;

    let onScreen = true;
    const apply = () => {
      node.setAttribute('data-live', onScreen && !document.hidden ? 'true' : 'false');
    };
    document.addEventListener('visibilitychange', apply);
    const observer =
      typeof IntersectionObserver === 'undefined'
        ? null
        : new IntersectionObserver((entries) => {
            onScreen = entries.some((entry) => entry.isIntersecting);
            apply();
          });
    observer?.observe(node);
    return () => {
      document.removeEventListener('visibilitychange', apply);
      observer?.disconnect();
    };
  }, []);
}
