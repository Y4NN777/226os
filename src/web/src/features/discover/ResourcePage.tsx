import { Badge } from '@astryxdesign/core/Badge';
import { Breadcrumbs, BreadcrumbItem } from '@astryxdesign/core/Breadcrumbs';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useParams } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { ApiError } from '../../api/http';
import { DemoBadge } from '../../components/DemoBadge';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import {
  countryName,
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
import { VerificationBadge } from './VerificationBadge';

function ExternalLink({ href, children }: { href: string; children: ReactNode }) {
  const { t } = useI18n();
  return (
    <Link href={href} isExternalLink newTabLabel={t('discover.newTab')}>
      {children}
    </Link>
  );
}

function Fact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <VStack gap={0.5}>
      <Text type="label" color="secondary">
        {label}
      </Text>
      {children}
    </VStack>
  );
}

const relatedKey = {
  question: 'discover.related.question',
  resource: 'discover.related.resource',
  project: 'discover.related.project',
  article: 'discover.related.article',
  series: 'discover.related.series',
} as const satisfies Record<Resource['related'][number]['type'], MessageKey>;

function Sources({ id }: { id: string }) {
  const { t, locale } = useI18n();
  const sources = useResourceSources(id);
  return (
    <VStack gap={3}>
      <Heading level={2} id="discover-sources">
        {t('discover.sources.title')}
      </Heading>
      {sources.isPending ? <LoadingState /> : null}
      {sources.isError ? (
        <ErrorState error={sources.error} onRetry={() => void sources.refetch()} />
      ) : null}
      {sources.data && sources.data.items.length === 0 ? (
        <Text as="p">{t('discover.sources.none')}</Text>
      ) : null}
      {sources.data && sources.data.items.length > 0 ? (
        <VStack as="ul" gap={3} aria-labelledby="discover-sources">
          {sources.data.items.map((source) => {
            const status = sourceStatus(source);
            return (
              <li key={source.id}>
                <VStack gap={1}>
                  <ExternalLink href={source.url}>
                    {source.title ?? hostOf(source.url)}
                  </ExternalLink>
                  <HStack gap={2} wrap="wrap" vAlign="center">
                    <Text type="supporting">{t(kindKey[source.kind])}</Text>
                    <VerificationBadge status={status} />
                    {source.verified_at && showsVerifiedDate(status) ? (
                      <Text type="supporting">
                        {t('discover.verifiedOn', { date: formatDate(source.verified_at, locale) })}
                      </Text>
                    ) : null}
                  </HStack>
                </VStack>
              </li>
            );
          })}
        </VStack>
      ) : null}
    </VStack>
  );
}

function Detail({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const countries = resource.countries.map((code) => countryName(code, locale)).join(', ');
  const languages = resource.doc_languages.map((code) => languageName(code, locale)).join(', ');
  return (
    <VStack gap={5}>
      <VStack gap={2}>
        <Breadcrumbs label={t('discover.breadcrumb')} variant="supporting">
          <BreadcrumbItem href="/discover">{t('discover.title')}</BreadcrumbItem>
          <BreadcrumbItem isCurrent>{resource.name}</BreadcrumbItem>
        </Breadcrumbs>
        <Heading level={1}>{resource.name}</Heading>
        <HStack gap={2} wrap="wrap" vAlign="center">
          <Badge label={t(typeKey[resource.type])} />
          <Badge label={resource.category} />
          <VerificationBadge status={resource.verification_status} />
          {resource.provenance === 'demo' ? <DemoBadge /> : null}
          {resource.last_verified_at && showsVerifiedDate(resource.verification_status) ? (
            <Text type="supporting">
              {t('discover.verifiedOn', { date: formatDate(resource.last_verified_at, locale) })}
            </Text>
          ) : null}
        </HStack>
        <Text as="p">{resource.summary}</Text>
      </VStack>

      <VStack gap={3}>
        <Fact label={t('discover.detail.provider')}>
          {resource.provider.url ? (
            <ExternalLink href={resource.provider.url}>{resource.provider.name}</ExternalLink>
          ) : (
            <Text>{resource.provider.name}</Text>
          )}
        </Fact>
        <Fact label={t('discover.detail.countries')}>
          <Text>{countries || t('discover.detail.countries.none')}</Text>
        </Fact>
        {resource.doc_languages.length > 0 ? (
          <Fact label={t('discover.detail.docLanguages')}>
            <Text>{languages}</Text>
          </Fact>
        ) : null}
        {resource.docs_url ? (
          <ExternalLink href={resource.docs_url}>{t('discover.detail.docs')}</ExternalLink>
        ) : null}
        {resource.sdk_languages.length > 0 ? (
          <Fact label={t('discover.detail.sdks')}>
            <HStack gap={3} wrap="wrap">
              {resource.sdk_languages.map((sdk) => (
                <ExternalLink key={sdk.language} href={sdk.url}>
                  {sdk.name ?? sdk.language}
                </ExternalLink>
              ))}
            </HStack>
          </Fact>
        ) : null}
      </VStack>

      {resource.limitations && resource.limitations.length > 0 ? (
        <VStack gap={2}>
          <Heading level={2} id="discover-limitations">
            {t('discover.detail.limitations')}
          </Heading>
          <ul aria-labelledby="discover-limitations">
            {resource.limitations.map((limit) => (
              <li key={limit}>{limit}</li>
            ))}
          </ul>
        </VStack>
      ) : null}

      <Sources id={resource.id} />

      {resource.related.length > 0 ? (
        <VStack gap={2}>
          <Heading level={2} id="discover-related">
            {t('discover.detail.related')}
          </Heading>
          <VStack as="ul" gap={2} aria-labelledby="discover-related">
            {resource.related.map((link) => (
              <li key={`${link.type}-${link.id}`}>
                <HStack gap={2} wrap="wrap">
                  <Text type="supporting">{t(relatedKey[link.type])}</Text>
                  {link.type === 'resource' ? (
                    <Link href={`/discover/${link.id}`}>{link.title}</Link>
                  ) : (
                    <Text>{link.title}</Text>
                  )}
                </HStack>
              </li>
            ))}
          </VStack>
        </VStack>
      ) : null}
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
