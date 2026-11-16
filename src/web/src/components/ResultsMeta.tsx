import type { ReactNode } from 'react';

/**
 * The line above a result list: the count (announced politely when it changes) and a slot for
 * the sort and view controls (prototype `.results-meta`). The page writes the count sentence.
 */
export function ResultsMeta({ summary, children }: { summary: ReactNode; children?: ReactNode }) {
  return (
    <div className="ui-results-meta">
      <p className="ui-results-meta-summary" aria-live="polite">
        {summary}
      </p>
      {children ? <div className="ui-results-meta-controls">{children}</div> : null}
    </div>
  );
}
