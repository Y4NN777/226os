import { useCallback, useEffect, useRef, useState } from 'react';

/** After a failed submit, moves focus to the first field the form marks as invalid. */
export function useInvalidFocus() {
  const formRef = useRef<HTMLFormElement>(null);
  const [requests, setRequests] = useState(0);
  useEffect(() => {
    if (requests === 0) return;
    formRef.current?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
  }, [requests]);
  const focusFirstInvalid = useCallback(() => {
    setRequests((count) => count + 1);
  }, []);
  return { formRef, focusFirstInvalid };
}
