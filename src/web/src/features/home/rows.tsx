import type { components } from '../../api/schema';
import { Avatar } from '../../components/Avatar';
import { CountryCode } from '../../components/CountryCode';
import { CodeIcon, FileTextIcon, MessageIcon } from '../../components/icons';
import { ListRow } from '../../components/List';
import { Meta, MetaItem } from '../../components/Meta';
import { StatusBadge, type StatusKind } from '../../components/StatusBadge';
import { Tag, TagList } from '../../components/Tag';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale } from '../../i18n/messages';

type Question = components['schemas']['Question'];
type Resource = components['schemas']['Resource'];
type Project = components['schemas']['Project'];
type Article = components['schemas']['ArticleSummary'];
type VerificationStatus = components['schemas']['VerificationStatus'];

/**
 * The member home's rows (prototype accueil.html `ul.list > li.row`), built on the shared
 * List/ListRow, StatusBadge, TagList, CountryCode, Meta and Avatar. Only contract fields.
 */

export function formatDate(iso: string, locale: Locale): string {
  return new Intl.DateTimeFormat(locale === 'en' ? 'en-GB' : locale, {
    dateStyle: 'medium',
    timeZone: 'UTC',
  }).format(new Date(iso));
}

export function countryName(code: string, locale: Locale): string {
  try {
    return new Intl.DisplayNames([locale], { type: 'region' }).of(code) ?? code;
  } catch {
    return code;
  }
}

/** A country code chip with its full name for assistive technology. */
export function Country({ code }: { code: string }) {
  const { locale } = useI18n();
  return <CountryCode code={code} name={countryName(code, locale)} />;
}

/** Verification as a status pill; claimed and disputed keep their own word. */
const verificationKind: Record<VerificationStatus, StatusKind> = {
  verified: 'verified',
  claimed: 'unverified',
  unverified: 'unverified',
  stale: 'stale',
  disputed: 'stale',
};

export function VerificationStatusBadge({ status }: { status: VerificationStatus }) {
  const { t } = useI18n();
  return <StatusBadge kind={verificationKind[status]} label={t(`feed.verification.${status}`)} />;
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
    <ListRow
      title={question.title}
      href={`/questions/${question.id}`}
      lang={question.lang}
      lead={<Avatar name={question.author.display_name} />}
      tags={question.tags.length > 0 ? <TagList tags={question.tags.slice(0, 3)} /> : undefined}
      meta={
        <Meta>
          <MetaItem>{question.author.display_name}</MetaItem>
          {question.country ? <Country code={question.country} /> : null}
          <MetaItem>
            <time dateTime={question.created_at}>{formatDate(question.created_at, locale)}</time>
          </MetaItem>
        </Meta>
      }
      aside={
        <>
          {question.accepted_answer_id ? <StatusBadge kind="accepted" /> : null}
          <MetaItem icon={MessageIcon}>{answers}</MetaItem>
        </>
      }
    />
  );
}

export function ResourceRow({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const checked = resource.verification_status === 'verified' ? resource.last_verified_at : null;
  return (
    <ListRow
      title={resource.name}
      href={`/discover/${resource.id}`}
      lead={<Avatar name={resource.name} />}
      excerpt={resource.summary}
      meta={
        <Meta>
          <MetaItem>{t('feed.catalog.provider', { name: resource.provider.name })}</MetaItem>
          {resource.countries.slice(0, 4).map((code) => (
            <Country key={code} code={code} />
          ))}
        </Meta>
      }
      aside={
        <>
          <VerificationStatusBadge status={resource.verification_status} />
          {resource.provenance === 'demo' ? <StatusBadge kind="demo" /> : null}
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
    <ListRow
      title={project.name}
      href={`/build/${project.id}`}
      lead={<Avatar name={project.name} />}
      lang={project.lang}
      excerpt={project.description}
      tags={
        project.technologies.length > 0 ? (
          <TagList tags={project.technologies.slice(0, 3)} />
        ) : undefined
      }
      meta={
        <Meta>
          <MetaItem>{t('feed.by', { name: project.owner.display_name })}</MetaItem>
          {project.owner.country ? <Country code={project.owner.country} /> : null}
        </Meta>
      }
      aside={
        open > 0 ? (
          <>
            <StatusBadge kind="open" />
            <span>
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
    <ListRow
      title={article.title}
      href={`/apprendre/articles/${article.id}`}
      lead={<Avatar name={article.author.display_name} />}
      lang={article.lang}
      excerpt={article.summary}
      meta={
        <Meta>
          <MetaItem icon={tutorial ? CodeIcon : FileTextIcon}>
            {t(tutorial ? 'feed.type.tutorial' : 'feed.type.article')}
          </MetaItem>
          <MetaItem>{article.author.display_name}</MetaItem>
          <MetaItem>{t(`feed.level.${article.level}`)}</MetaItem>
          <Tag label={article.lang.toUpperCase()} />
        </Meta>
      }
      aside={
        <time dateTime={article.published_at}>{formatDate(article.published_at, locale)}</time>
      }
    />
  );
}
