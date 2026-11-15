import { Heading } from '@astryxdesign/core/Heading';
import { useId, type ReactNode } from 'react';

/**
 * A titled block of a page (prototype `.section`, `.card` with an `h2`), named by its heading so
 * it is a landmark region (or a complementary block with `as="aside"`). `isPanel` puts it on the
 * layered surface; panel titles are one step smaller, as in the prototype cards. One scale for
 * every page: ui.css owns the sizes, the page only picks plain or panel.
 */
export function Section({
  title,
  children,
  id,
  titleId,
  headingLevel = 2,
  isPanel = false,
  as: Element = 'section',
  titleAside,
  actions,
}: {
  title: string;
  children: ReactNode;
  /** Anchor of the section (`#besoins`). */
  id?: string;
  /** Id of the heading, when another element points at it. */
  titleId?: string;
  headingLevel?: 2 | 3;
  isPanel?: boolean;
  as?: 'section' | 'aside';
  /** Small note next to the title (a language code, a « Bientôt » pill). */
  titleAside?: ReactNode;
  /** Controls at the end of the title row. */
  actions?: ReactNode;
}) {
  const generated = useId();
  const headingId = titleId ?? `${generated}-title`;
  const className = isPanel ? 'ui-section ui-panel' : 'ui-section';
  return (
    <Element id={id} className={className} aria-labelledby={headingId}>
      <div className="ui-section-head">
        <div className="ui-section-title-row">
          <Heading level={headingLevel} id={headingId} className="ui-section-title">
            {title}
          </Heading>
          {titleAside}
        </div>
        {actions ? <div className="ui-section-actions">{actions}</div> : null}
      </div>
      {children}
    </Element>
  );
}
