import { Heading } from '@astryxdesign/core/Heading';
import { Text } from '@astryxdesign/core/Text';
import type { ReactNode, Ref } from 'react';
import { Breadcrumbs, type Crumb } from './Breadcrumbs';

export interface PageHeaderProps {
  /** The page's one h1. */
  title: string;
  /** One or two sentences under the title, in the muted tone. */
  lead?: ReactNode;
  breadcrumbs?: readonly Crumb[];
  /** The page's main action(s), at the end of the title row. */
  actions?: ReactNode;
  /** Original language of the title when it is user content (a question, a project). */
  titleLang?: string;
  /** Id of the h1, when a landmark is labelled by it. */
  headingId?: string;
  /** For a view that moves focus to its title; the h1 then takes `tabIndex={-1}`. */
  headingRef?: Ref<HTMLHeadingElement>;
  /** The item's tile or the person's avatar, before the title (detail pages). */
  media?: ReactNode;
  /** A short line above the title (kind · category). */
  eyebrow?: ReactNode;
  /** A short line under the title (provider, @handle). */
  subtitle?: ReactNode;
  /** Pills and facts under the title (state, author, dates). */
  meta?: ReactNode;
  /** More header content after the lead (skills, links). */
  children?: ReactNode;
}

/**
 * Top of a page (prototype `header.page-header`): breadcrumbs, the h1 with its actions on the
 * right, and the lead. Detail pages add the item's tile, an eyebrow, a subtitle and a meta line.
 * Every page uses it, so the title, lead and spacing share one scale.
 */
export function PageHeader({
  title,
  lead,
  breadcrumbs,
  actions,
  titleLang,
  headingId,
  headingRef,
  media,
  eyebrow,
  subtitle,
  meta,
  children,
}: PageHeaderProps) {
  const heading = (
    <Heading
      level={1}
      className="ui-page-title"
      {...(headingId ? { id: headingId } : {})}
      {...(headingRef ? { ref: headingRef, tabIndex: -1 } : {})}
    >
      {title}
    </Heading>
  );
  const hasIdentity = Boolean(media ?? eyebrow ?? subtitle ?? meta);
  return (
    <header className="ui-page-header">
      {breadcrumbs && breadcrumbs.length > 0 ? <Breadcrumbs items={breadcrumbs} /> : null}
      <div className="ui-page-header-row">
        {hasIdentity ? (
          <div className="ui-page-header-id">
            {media ? <div className="ui-page-header-media">{media}</div> : null}
            <div className="ui-page-header-text">
              {eyebrow ? <p className="ui-page-header-eyebrow">{eyebrow}</p> : null}
              {titleLang ? <div lang={titleLang}>{heading}</div> : heading}
              {subtitle ? <p className="ui-page-header-subtitle">{subtitle}</p> : null}
              {meta ? <div className="ui-page-header-meta">{meta}</div> : null}
            </div>
          </div>
        ) : titleLang ? (
          <div lang={titleLang}>{heading}</div>
        ) : (
          heading
        )}
        {actions ? <div className="ui-page-header-actions">{actions}</div> : null}
      </div>
      {lead ? (
        <Text as="p" color="secondary" className="ui-page-header-lead">
          {lead}
        </Text>
      ) : null}
      {children}
    </header>
  );
}
