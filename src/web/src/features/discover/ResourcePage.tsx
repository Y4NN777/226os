import { Breadcrumbs, BreadcrumbItem } from '@astryxdesign/core/Breadcrumbs';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { VStack } from '@astryxdesign/core/VStack';
import { useParams } from '@tanstack/react-router';
import { useState, type ReactNode } from 'react';
import { ApiError } from '../../api/http';
import { DemoBadge } from '../../components/DemoBadge';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { relatedPath } from '../related';
import { categoryLabel } from './categories';
import { CountryCodes } from './CountryCodes';
import { ErrorReport, ErrorReportToggle } from './ErrorReport';
import {
  formatDate,
  hostOf,
  kindKey,
  languageName,
  showsVerifiedDate,
  sourceStatus,
  typeKey,
} from './present';
import { useResource, useResourceSources } from './queries';
import type { Resource } from './region';
import { ResourceThumb } from './ResourceThumb';
import { VerificationBadge } from './VerificationBadge';
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

/** A titled panel (prototype `.card.res-card`) labelled by its heading. */
function Panel({ id, title, children }: { id: string; title: string; children: ReactNode }) {
  return (
    <section className="discover-panel" aria-labelledby={id}>
      {/* Prototype `h2.h3`: a section heading at a smaller visual size. */}
      <Heading level={4} accessibilityLevel={2} id={id}>
        {title}
      </Heading>
      {children}
    </section>
  );
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
    <header className="discover-head">
      <Breadcrumbs label={t('discover.breadcrumb')} variant="supporting">
        <BreadcrumbItem href="/discover">{t('discover.title')}</BreadcrumbItem>
        <BreadcrumbItem isCurrent>{resource.name}</BreadcrumbItem>
      </Breadcrumbs>
      <div className="discover-head-row">
        <div className="discover-head-id">
          <ResourceThumb name={resource.name} size="lg" />
          <div className="discover-head-text">
            <p className="discover-eyebrow">
              {`${t(typeKey[resource.type])} · ${categoryLabel(resource.category, t)}`}
            </p>
            <Heading level={1}>{resource.name}</Heading>
            <Text type="supporting">
              {resource.provider.url ? (
                <ExternalLink href={resource.provider.url}>{resource.provider.name}</ExternalLink>
              ) : (
                resource.provider.name
              )}
            </Text>
          </div>
        </div>
        <div className="discover-head-actions">
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
        </div>
      </div>
      <div className="discover-facts">
        <VerificationBadge status={resource.verification_status} />
        {resource.provenance === 'demo' ? <DemoBadge /> : null}
        {resource.last_verified_at && showsVerifiedDate(resource.verification_status) ? (
          <DateText iso={resource.last_verified_at} labelKey="discover.verifiedOn" />
        ) : null}
        <DateText iso={resource.updated_at} labelKey="discover.updatedOn" />
      </div>
      <p className="discover-lede">{resource.summary}</p>
    </header>
  );
}

function Coverage({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const languages = resource.doc_languages.map((code) => languageName(code, locale)).join(', ');
  return (
    <Panel id="discover-coverage" title={t('discover.detail.coverage')}>
      <dl className="discover-dl">
        <dt>{t('discover.detail.countries')}</dt>
        <dd>
          {resource.countries.length > 0 ? (
            <CountryCodes codes={resource.countries} withNames />
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
    </Panel>
  );
}

function Limitations({ items }: { items: readonly string[] }) {
  const { t } = useI18n();
  return (
    <Panel id="discover-limitations" title={t('discover.detail.limitations')}>
      <ul className="discover-limits" aria-labelledby="discover-limitations">
        {items.map((limit) => (
          <li key={limit}>
            <Icon icon="warning" color="warning" size="sm" />
            <span>{limit}</span>
          </li>
        ))}
      </ul>
    </Panel>
  );
}

function Sources({ id }: { id: string }) {
  const { t } = useI18n();
  const sources = useResourceSources(id);
  return (
    <Panel id="discover-sources" title={t('discover.sources.title')}>
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
                  <VerificationBadge status={status} />
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
    </Panel>
  );
}

function Related({ links }: { links: Resource['related'] }) {
  const { t } = useI18n();
  return (
    <section className="discover-results" aria-labelledby="discover-related">
      <Heading level={4} accessibilityLevel={2} id="discover-related">
        {t('discover.detail.related')}
      </Heading>
      <ul className="discover-linked">
        {links.map((link) => (
          <li key={`${link.type}-${link.id}`} className="discover-linked-item">
            <span className="discover-linked-kind">{t(relatedKey[link.type])}</span>
            <Link href={relatedPath(link)} weight="semibold">
              {link.title}
            </Link>
          </li>
        ))}
      </ul>
    </section>
  );
}

function Detail({ resource }: { resource: Resource }) {
  const [isReportOpen, setReportOpen] = useState(false);
  const [isReportSent, setReportSent] = useState(false);
  return (
    <VStack gap={5}>
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
    </VStack>
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
          description={t('discover.notFound.body')}
          actions={<Button label={t('discover.back')} href="/discover" />}
        />
      );
    }
    return <ErrorState error={resource.error} onRetry={() => void resource.refetch()} />;
  }
  return <Detail resource={resource.data} />;
}
