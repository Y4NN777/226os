import { useCallback, useState } from 'react';

export interface Failure {
  /** Changes on every failure, so a notice remounts and takes focus again. */
  n: number;
  error: unknown;
}

export function useFailure() {
  const [failure, setFailure] = useState<Failure | null>(null);
  const report = useCallback((error: unknown) => {
    setFailure((previous) => ({ n: (previous?.n ?? 0) + 1, error }));
  }, []);
  const clear = useCallback(() => {
    setFailure(null);
  }, []);
  return { failure, report, clear };
}
