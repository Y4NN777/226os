import { useEffect, useRef, useState } from 'react';
import { useI18n } from '../../i18n/I18nProvider';

const DURATION_MS = 1200;

function canAnimate() {
  return (
    typeof IntersectionObserver !== 'undefined' &&
    !(typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches)
  );
}

/**
 * A real total that counts up from 0 the first time it enters the screen, once. The final value
 * is in the DOM from the start for screen readers (`.lp-sr`); the moving number is `aria-hidden`.
 * Without IntersectionObserver or under reduced motion the final value is simply shown.
 */
export function CountUp({ value }: { value: number }) {
  const { locale } = useI18n();
  const format = new Intl.NumberFormat(locale);
  const [animated] = useState(canAnimate);
  const [shown, setShown] = useState(animated ? 0 : value);
  const node = useRef<HTMLSpanElement>(null);
  const played = useRef(false);

  useEffect(() => {
    if (!animated || !node.current) return;
    if (played.current) {
      setShown(value);
      return;
    }
    let frame = 0;
    const observer = new IntersectionObserver((entries) => {
      if (!entries.some((entry) => entry.isIntersecting)) return;
      observer.disconnect();
      played.current = true;
      let start: number | undefined;
      const tick = (time: number) => {
        start ??= time;
        const progress = Math.min(1, (time - start) / DURATION_MS);
        // Exponential ease-out: quick start, soft landing.
        const eased = progress === 1 ? 1 : 1 - Math.pow(2, -10 * progress);
        setShown(Math.round(value * eased));
        if (progress < 1) frame = requestAnimationFrame(tick);
      };
      frame = requestAnimationFrame(tick);
    });
    observer.observe(node.current);
    return () => {
      observer.disconnect();
      cancelAnimationFrame(frame);
    };
  }, [animated, value]);

  return (
    <>
      <span className="lp-sr">{format.format(value)}</span>
      <span ref={node} aria-hidden="true">
        {format.format(shown)}
      </span>
    </>
  );
}
