import { useEffect, useState } from 'react';

/** Whole seconds left before a retry is allowed (from Retry-After); 0 means free. */
export function useCountdown() {
  const [seconds, setSeconds] = useState(0);
  useEffect(() => {
    if (seconds <= 0) return undefined;
    const id = setTimeout(() => {
      setSeconds((current) => current - 1);
    }, 1000);
    return () => {
      clearTimeout(id);
    };
  }, [seconds]);
  return { seconds, start: setSeconds };
}
