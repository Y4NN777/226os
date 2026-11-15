import { Button } from '@astryxdesign/core/Button';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { useParams } from '@tanstack/react-router';
import { useState, type ReactNode } from 'react';
import { ApiError } from '../../api/http';
import { Avatar } from '../../components/Avatar';
import { CountryCode } from '../../components/CountryCode';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { LinkedItems } from '../../components/LinkedItems';
import { LoadingState } from '../../components/LoadingState';
import { PageHeader } from '../../components/PageHeader';
import { Section } from '../../components/Section';
import { StatusBadge } from '../../components/StatusBadge';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { relatedPath } from '../related';
import { categoryLabel } from './categories';
import { ErrorReport, ErrorReportToggle } from './ErrorReport';
import {
  formatDate,
  hostOf,
  kindKey,
  languageName,
  showsVerifiedDate,
  sourceStatus,
  statusKey,
  statusKind,
  typeKey,
  type VerificationStatus,
} from './present';
import { useResource, useResourceSources } from './queries';
import type { Resource } from './region';
import './discover.css';

const REPORT_PANEL = 'discover-report';

function ExternalLink({ href, children }: { href: string; children: ReactNode }) {
  const { t } = useI18n();
  return (
    <Link href={href} isExternalLink newTabLabel={t('discover.newTab')}>
      {children}
    </Link>
  );
}

/** A "Vérifié le …" or "Mis à jour le …" phrase kept in one element, with a machine date. */
function DateText({ iso, labelKey }: { iso: string; labelKey: MessageKey }) {
  const { t, locale } = useI18n();
  return <time dateTime={iso}>{t(labelKey, { date: formatDate(iso, locale) })}</time>;
}

function Verification({ status }: { status: VerificationStatus }) {
  const { t } = useI18n();
  return <StatusBadge kind={statusKind[status]} label={t(statusKey[status])} />;
}

const relatedKey = {
  question: 'discover.related.question',
  resource: 'discover.related.resource',
  project: 'discover.related.project',
  article: 'discover.related.article',
  series: 'discover.related.series',
} as const satisfies Record<Resource['related'][number]['type'], MessageKey>;

function Header({
  resource,
  isReportOpen,
  onToggleReport,
}: {
  resource: Resource;
  isReportOpen: boolean;
  onToggleReport: () => void;
}) {
  const { t } = useI18n();
  const docsLabel = `${t('discover.detail.docs')} ${t('discover.newTab')}`;
  return (
    <PageHeader
      title={resource.name}
      breadcrumbs={[
        { label: t('nav.home'), href: '/' },
        { label: t('discover.title'), href: '/discover' },
        { label: resource.name },
      ]}
      media={<Avatar name={resource.name} variant="tile" size="lg" />}
      eyebrow={`${t(typeKey[resource.type])} · ${categoryLabel(resource.category, t)}`}
      subtitle={
        resource.provider.url ? (
          <ExternalLink href={resource.provider.url}>{resource.provider.name}</ExternalLink>
        ) : (
          resource.provider.name
        )
      }
      meta={
        <>
          <Verification status={resource.verification_status} />
          {resource.provenance === 'demo' ? <StatusBadge kind="demo" /> : null}
          {resource.last_verified_at && showsVerifiedDate(resource.verification_status) ? (
            <DateText iso={resource.last_verified_at} labelKey="discover.verifiedOn" />
          ) : null}
          <DateText iso={resource.updated_at} labelKey="discover.updatedOn" />
        </>
      }
      actions={
        <>
          {resource.docs_url ? (
            <Button
              variant="primary"
              href={resource.docs_url}
              target="_blank"
              rel="noopener noreferrer"
              label={docsLabel}
              endContent={<Icon icon="externalLink" />}
            >
              {t('discover.detail.docs')}
              <VisuallyHidden>{` ${t('discover.newTab')}`}</VisuallyHidden>
            </Button>
          ) : null}
          <ErrorReportToggle
            isOpen={isReportOpen}
            panelId={REPORT_PANEL}
            onToggle={onToggleReport}
          />
        </>
      }
      lead={resource.summary}
    />
  );
}

function Coverage({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const languages = resource.doc_languages.map((code) => languageName(code, locale)).join(', ');
  return (
    <Section title={t('discover.detail.coverage')} isPanel>
      <dl className="ui-facts">
        <dt>{t('discover.detail.countries')}</dt>
        <dd>
          {resource.countries.length > 0 ? (
            <span className="ui-countries">
              {resource.countries.map((code) => (
                <CountryCode key={code} code={code} withName />
              ))}
            </span>
          ) : (
            t('discover.detail.countries.none')
          )}
        </dd>
        <dt>{t('discover.detail.category')}</dt>
        <dd>{categoryLabel(resource.category, t)}</dd>
        <dt>{t('discover.detail.kind')}</dt>
        <dd>{t(typeKey[resource.type])}</dd>
        {resource.doc_languages.length > 0 ? (
          <>
            <dt>{t('discover.detail.docLanguages')}</dt>
            <dd>{languages}</dd>
          </>
        ) : null}
        {resource.sdk_languages.length > 0 ? (
          <>
            <dt>{t('discover.detail.sdks')}</dt>
            <dd>
              <ul className="discover-inline-list">
                {resource.sdk_languages.map((sdk) => (
                  <li key={sdk.language}>
                    <ExternalLink href={sdk.url}>{sdk.name ?? sdk.language}</ExternalLink>
                  </li>
                ))}
              </ul>
            </dd>
          </>
        ) : null}
      </dl>
    </Section>
  );
}

function Limitations({ items }: { items: readonly string[] }) {
  const { t } = useI18n();
  return (
    <Section title={t('discover.detail.limitations')} titleId="discover-limitations" isPanel>
      <ul className="discover-limits" aria-labelledby="discover-limitations">
        {items.map((limit) => (
          <li key={limit}>
            <Icon icon="warning" color="warning" size="sm" />
            <span>{limit}</span>
          </li>
        ))}
      </ul>
    </Section>
  );
}

function Sources({ id }: { id: string }) {
  const { t } = useI18n();
  const sources = useResourceSources(id);
  return (
    <Section title={t('discover.sources.title')} titleId="discover-sources" isPanel>
      {sources.isPending ? <LoadingState /> : null}
      {sources.isError ? (
        <ErrorState error={sources.error} onRetry={() => void sources.refetch()} />
      ) : null}
      {sources.data && sources.data.items.length === 0 ? (
        <Text as="p" color="secondary">
          {t('discover.sources.none')}
        </Text>
      ) : null}
      {sources.data && sources.data.items.length > 0 ? (
        <ul className="discover-sources" aria-labelledby="discover-sources">
          {sources.data.items.map((source) => {
            const status = sourceStatus(source);
            return (
              <li key={source.id} className="discover-source">
                <span className="discover-source-main">
                  <ExternalLink href={source.url}>
                    {source.title ?? hostOf(source.url)}
                  </ExternalLink>
                  <Text type="supporting">{t(kindKey[source.kind])}</Text>
                </span>
                <span className="discover-source-aside">
                  <Verification status={status} />
                  {source.verified_at && showsVerifiedDate(status) ? (
                    <Text type="supporting">
                      <DateText iso={source.verified_at} labelKey="discover.verifiedOn" />
                    </Text>
                  ) : null}
                </span>
              </li>
            );
          })}
        </ul>
      ) : null}
    </Section>
  );
}

function Related({ links }: { links: Resource['related'] }) {
  const { t } = useI18n();
  return (
    <Section title={t('discover.detail.related')}>
      <LinkedItems
        items={links.map((link) => ({
          key: `${link.type}-${link.id}`,
          kind: t(relatedKey[link.type]),
          title: link.title,
          href: relatedPath(link),
        }))}
      />
    </Section>
  );
}

function Detail({ resource }: { resource: Resource }) {
  const [isReportOpen, setReportOpen] = useState(false);
  const [isReportSent, setReportSent] = useState(false);
  return (
    <div className="ui-page">
      <Header
        resource={resource}
        isReportOpen={isReportOpen}
        onToggleReport={() => {
          setReportOpen((open) => !open);
        }}
      />
      {isReportOpen ? (
        <ErrorReport
          id={REPORT_PANEL}
          resourceId={resource.id}
          sent={isReportSent}
          onSent={() => {
            setReportSent(true);
          }}
        />
      ) : null}
      <Coverage resource={resource} />
      {resource.limitations && resource.limitations.length > 0 ? (
        <Limitations items={resource.limitations} />
      ) : null}
      <Sources id={resource.id} />
      {resource.related.length > 0 ? <Related links={resource.related} /> : null}
    </div>
  );
}

export function ResourcePage() {
  const { t } = useI18n();
  const { resourceId } = useParams({ from: '/app/discover/$resourceId' });
  const resource = useResource(resourceId);

  if (resource.isPending) return <LoadingState />;
  if (resource.isError) {
    if (resource.error instanceof ApiError && resource.error.status === 404) {
      return (
        <EmptyState
          headingLevel={1}
          title={t('discover.notFound.title')}
          actions={<Button label={t('discover.back')} href="/discover" />}
        />
      );
    }
    return <ErrorState error={resource.error} onRetry={() => void resource.refetch()} />;
  }
  return <Detail resource={resource.data} />;
}
