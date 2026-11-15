import { Avatar } from '@astryxdesign/core/Avatar';
import { Link } from '@astryxdesign/core/Link';
import type { ReactNode } from 'react';
import type { components } from '../../api/schema';
import { DemoBadge } from '../../components/DemoBadge';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale } from '../../i18n/messages';
import { FeedIcon, type FeedIconName } from './FeedIcon';
import './home.css';

type Question = components['schemas']['Question'];
type Resource = components['schemas']['Resource'];
type Project = components['schemas']['Project'];
type Article = components['schemas']['ArticleSummary'];
type VerificationStatus = components['schemas']['VerificationStatus'];

/**
 * Dense list rows after the prototype (`ul.list > li.row`: lead, title, excerpt, meta, aside).
 * Local to the home-auth pages; the lead swaps them for the shared ui/ rows later.
 */

export function formatDate(iso: string, locale: Locale): string {
  return new Intl.DateTimeFormat(locale === 'en' ? 'en-GB' : locale, {
    dateStyle: 'medium',
    timeZone: 'UTC',
  }).format(new Date(iso));
}

function countryName(code: string, locale: Locale): string {
  try {
    return new Intl.DisplayNames([locale], { type: 'region' }).of(code) ?? code;
  } catch {
    return code;
  }
}

export function RowList({ children }: { children: ReactNode }) {
  return <ul className="feed-list">{children}</ul>;
}

/** Icon + word on a neutral fill: state never rests on colour alone. */
export type PillKind = 'accepted' | 'verified' | 'stale' | 'unverified' | 'open';

const pillIcon: Record<PillKind, FeedIconName> = {
  accepted: 'checkCircle',
  verified: 'shieldCheck',
  stale: 'clock',
  unverified: 'help',
  open: 'circleDot',
};

export function StatePill({ kind, label }: { kind: PillKind; label: string }) {
  return (
    <span className={`feed-pill feed-pill-${kind}`}>
      <FeedIcon name={pillIcon[kind]} />
      {label}
    </span>
  );
}

/** Square tinted lead with a glyph, for things that have no face (APIs, projects, articles). */
export function Thumb({ icon, strong = false }: { icon: FeedIconName; strong?: boolean }) {
  return (
    <span className={strong ? 'feed-thumb feed-thumb-strong' : 'feed-thumb'} aria-hidden="true">
      <FeedIcon name={icon} />
    </span>
  );
}

export function Tags({ tags, max = 3 }: { tags: readonly string[]; max?: number }) {
  if (tags.length === 0) return null;
  return (
    <ul className="feed-tags">
      {tags.slice(0, max).map((tag) => (
        <li key={tag} className="feed-tag">
          {tag}
        </li>
      ))}
    </ul>
  );
}

/** Country as a text code, the full name for assistive technology (no flags). */
export function Country({ code }: { code: string }) {
  const { locale } = useI18n();
  return (
    <span className="feed-country">
      <abbr title={countryName(code, locale)}>{code}</abbr>
    </span>
  );
}

export function Row({
  lead,
  above,
  title,
  href,
  lang,
  excerpt,
  meta,
  aside,
}: {
  lead?: ReactNode;
  above?: ReactNode;
  title: string;
  href: string;
  lang?: string | undefined;
  excerpt?: string | undefined;
  meta?: ReactNode;
  aside?: ReactNode;
}) {
  const langProps = lang ? { lang } : {};
  return (
    <li className={lead ? 'feed-row' : 'feed-row feed-row-nolead'}>
      {lead ? <div className="feed-row-lead">{lead}</div> : null}
      <div className="feed-row-main">
        {above ? <div className="feed-row-above">{above}</div> : null}
        <h3 className="feed-row-title" {...langProps}>
          <Link href={href} color="primary" weight="semibold">
            {title}
          </Link>
        </h3>
        {excerpt ? (
          <p className="feed-row-excerpt" {...langProps}>
            {excerpt}
          </p>
        ) : null}
        {meta ? <div className="feed-meta">{meta}</div> : null}
      </div>
      {aside ? <div className="feed-row-aside">{aside}</div> : null}
    </li>
  );
}

export function QuestionRow({ question }: { question: Question }) {
  const { t, locale } = useI18n();
  const answers =
    question.answer_count === 0
      ? t('feed.row.noAnswer')
      : question.answer_count === 1
        ? t('feed.row.answers.one')
        : t('feed.row.answers.many', { n: question.answer_count });
  return (
    <Row
      lead={<Avatar name={question.author.display_name} size={40} />}
      title={question.title}
      href={`/questions/${question.id}`}
      lang={question.lang}
      meta={
        <>
          <Tags tags={question.tags} />
          {question.country ? <Country code={question.country} /> : null}
          <time dateTime={question.created_at}>{formatDate(question.created_at, locale)}</time>
        </>
      }
      aside={
        <>
          {question.accepted_answer_id ? (
            <StatePill kind="accepted" label={t('feed.pill.accepted')} />
          ) : null}
          <span className="feed-aside-text">
            <FeedIcon name="message" />
            {answers}
          </span>
        </>
      }
    />
  );
}

const verificationPill: Record<VerificationStatus, PillKind> = {
  verified: 'verified',
  claimed: 'unverified',
  unverified: 'unverified',
  stale: 'stale',
  disputed: 'stale',
};

export function ResourceRow({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const status = resource.verification_status;
  const checked = status === 'verified' && resource.last_verified_at;
  return (
    <Row
      lead={<Thumb icon="compass" strong={status === 'verified'} />}
      title={resource.name}
      href={`/discover/${resource.id}`}
      excerpt={resource.summary}
      meta={
        <>
          <span>{t('feed.catalog.provider', { name: resource.provider.name })}</span>
          {resource.countries.slice(0, 4).map((code) => (
            <Country key={code} code={code} />
          ))}
        </>
      }
      aside={
        <>
          <StatePill kind={verificationPill[status]} label={t(`feed.verification.${status}`)} />
          {resource.provenance === 'demo' ? <DemoBadge /> : null}
          {checked ? (
            <time dateTime={checked}>
              {t('feed.row.checkedOn', { date: formatDate(checked, locale) })}
            </time>
          ) : null}
        </>
      }
    />
  );
}

export function ProjectRow({ project }: { project: Project }) {
  const { t } = useI18n();
  const open = project.open_needs_count;
  return (
    <Row
      lead={<Thumb icon="hammer" />}
      title={project.name}
      href={`/build/${project.id}`}
      lang={project.lang}
      excerpt={project.description}
      meta={
        <>
          <Tags tags={project.technologies} />
          <span>{t('feed.by', { name: project.owner.display_name })}</span>
          {project.owner.country ? <Country code={project.owner.country} /> : null}
        </>
      }
      aside={
        open > 0 ? (
          <>
            <StatePill kind="open" label={t('feed.pill.open')} />
            <span className="feed-aside-text">
              {open === 1 ? t('feed.row.needs.one') : t('feed.row.needs.many', { n: open })}
            </span>
          </>
        ) : undefined
      }
    />
  );
}

export function ArticleRow({ article }: { article: Article }) {
  const { t, locale } = useI18n();
  const tutorial = article.type === 'tutorial';
  return (
    <Row
      lead={<Thumb icon={tutorial ? 'code' : 'fileText'} strong={!tutorial} />}
      above={
        <span className="feed-kind">
          <FeedIcon name={tutorial ? 'code' : 'fileText'} />
          {t(tutorial ? 'feed.type.tutorial' : 'feed.type.article')}
        </span>
      }
      title={article.title}
      href={`/apprendre/articles/${article.id}`}
      lang={article.lang}
      excerpt={article.summary}
      meta={
        <>
          <span>{article.author.display_name}</span>
          <span>{t(`feed.level.${article.level}`)}</span>
          <span className="feed-lang">{article.lang.toUpperCase()}</span>
        </>
      }
      aside={
        <time dateTime={article.published_at}>{formatDate(article.published_at, locale)}</time>
      }
    />
  );
}
