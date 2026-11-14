import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import { DemoBadge } from '../../components/DemoBadge';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale } from '../../i18n/messages';
import { countryName, formatDate, showsVerifiedDate, typeKey } from './present';
import { useResourceList } from './queries';
import { WEST_AFRICA, type Resource } from './region';
import type { DiscoverSearch } from './routeSearch';
import { VerificationBadge } from './VerificationBadge';

type SearchPatch = { [K in keyof DiscoverSearch]?: DiscoverSearch[K] | undefined };

/** Applies a patch to the URL filters; an undefined value removes the filter. */
function patched(base: DiscoverSearch, patch: SearchPatch): DiscoverSearch {
  const entries = Object.entries({ ...base, ...patch }).filter(([, value]) => value !== undefined);
  return Object.fromEntries(entries);
}

function sortedByName(codes: Iterable<string>, locale: Locale) {
  return [...new Set(codes)]
    .map((code) => ({ value: code, label: countryName(code, locale) }))
    .sort((a, b) => a.label.localeCompare(b.label, locale));
}

/** West Africa first, then any other code the loaded items or the URL mention. */
function countryOptions(extra: Iterable<string>, locale: Locale) {
  const west: readonly string[] = WEST_AFRICA;
  const others = [...new Set(extra)].filter((code) => !west.includes(code));
  return [...sortedByName(west, locale), ...sortedByName(others, locale)];
}

function Filters({
  search,
  seenCountries,
  onChange,
}: {
  search: DiscoverSearch;
  seenCountries: readonly string[];
  onChange: (patch: SearchPatch) => void;
}) {
  const { t, locale } = useI18n();
  const [draft, setDraft] = useState(search.q ?? '');
  const wholeContinent = search.region === 'all' || search.country !== undefined;
  const options = countryOptions(
    [...seenCountries, ...(search.country ? [search.country] : [])],
    locale,
  );
  return (
    <VStack gap={3}>
      <form
        role="search"
        onSubmit={(event) => {
          event.preventDefault();
          onChange({ q: draft.trim() || undefined });
        }}
      >
        <HStack gap={2} vAlign="end" wrap="nowrap">
          <TextInput
            type="search"
            label={t('discover.search.label')}
            isLabelHidden
            placeholder={t('discover.search.placeholder')}
            value={draft}
            onChange={setDraft}
            width="100%"
          />
          <IconButton
            type="submit"
            label={t('discover.search.submit')}
            icon={<Icon icon="search" />}
          />
        </HStack>
      </form>
      <HStack gap={3} wrap="wrap" vAlign="end">
        <SegmentedControl
          label={t('discover.region.label')}
          value={wholeContinent ? 'all' : 'west'}
          onChange={(value) => {
            onChange({ region: value === 'all' ? 'all' : undefined, country: undefined });
          }}
        >
          <SegmentedControlItem value="west" label={t('discover.region.west')} />
          <SegmentedControlItem value="all" label={t('discover.region.all')} />
        </SegmentedControl>
        <Selector
          label={t('discover.country.label')}
          isLabelHidden
          placeholder={t('discover.country.placeholder')}
          options={options}
          hasClear
          hasSearch
          value={search.country ?? null}
          onChange={(value) => {
            onChange({ country: value || undefined });
          }}
        />
        <SegmentedControl
          label={t('discover.type.label')}
          value={search.type ?? 'all'}
          onChange={(value) => {
            onChange({ type: value === 'all' ? undefined : (value as DiscoverSearch['type']) });
          }}
        >
          <SegmentedControlItem value="all" label={t('discover.type.all')} />
          <SegmentedControlItem value="api" label={t('discover.type.api')} />
          <SegmentedControlItem value="sdk" label={t('discover.type.sdk')} />
          <SegmentedControlItem value="tool" label={t('discover.type.tool')} />
        </SegmentedControl>
      </HStack>
      {search.category ? (
        <HStack gap={1} vAlign="center">
          <Text type="supporting">{t('discover.category.active', { name: search.category })}</Text>
          <IconButton
            size="sm"
            variant="ghost"
            label={t('discover.category.remove')}
            icon={<Icon icon="close" />}
            onClick={() => {
              onChange({ category: undefined });
            }}
          />
        </HStack>
      ) : null}
    </VStack>
  );
}

function ResourceCard({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const names = resource.countries.map((code) => countryName(code, locale)).join(', ');
  return (
    <Card variant="muted">
      <VStack gap={2}>
        <HStack gap={2} wrap="wrap" vAlign="center">
          <Link href={`/discover/${resource.id}`} type="large" weight="semibold">
            {resource.name}
          </Link>
          <Badge label={t(typeKey[resource.type])} />
          <VerificationBadge status={resource.verification_status} />
          {resource.provenance === 'demo' ? <DemoBadge /> : null}
        </HStack>
        <Text type="supporting">{t('discover.by', { name: resource.provider.name })}</Text>
        <Text as="p">{resource.summary}</Text>
        {names ? <Text type="supporting">{t('discover.countries', { list: names })}</Text> : null}
        <HStack gap={3} wrap="wrap">
          {resource.last_verified_at && showsVerifiedDate(resource.verification_status) ? (
            <Text type="supporting">
              {t('discover.verifiedOn', { date: formatDate(resource.last_verified_at, locale) })}
            </Text>
          ) : null}
          <Text type="supporting">{t('discover.sources.count', { n: resource.source_count })}</Text>
        </HStack>
      </VStack>
    </Card>
  );
}

export function DiscoverPage() {
  const { t } = useI18n();
  const search = useSearch({ from: '/app/discover' });
  const navigate = useNavigate();
  const list = useResourceList(search);
  const westView = search.region !== 'all' && search.country === undefined;
  const hasFilters = Object.keys(search).length > 0;
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];

  const change = (patch: SearchPatch) =>
    void navigate({ to: '/discover', search: patched(search, patch) });

  return (
    <VStack gap={5}>
      <VStack gap={1}>
        <Heading level={1}>{t('discover.title')}</Heading>
        <Text as="p">{t('discover.lede')}</Text>
        <Text as="p" type="supporting">
          {t('discover.positioning')}
        </Text>
      </VStack>
      <Filters
        key={search.q ?? ''}
        search={search}
        seenCountries={items.flatMap((item) => item.countries)}
        onChange={change}
      />
      {hasFilters ? (
        <HStack>
          <Button
            variant="ghost"
            label={t('discover.reset')}
            onClick={() => void navigate({ to: '/discover', search: {} })}
          />
        </HStack>
      ) : null}
      <section aria-label={t('discover.results.label')}>
        {list.isPending ? <LoadingState /> : null}
        {list.isError && !list.data ? (
          <ErrorState error={list.error} onRetry={() => void list.refetch()} />
        ) : null}
        {list.data && items.length === 0 ? (
          <EmptyState
            headingLevel={2}
            title={t('discover.empty.title')}
            description={t(westView ? 'discover.empty.west' : 'discover.empty.other')}
            actions={
              westView ? (
                <Button
                  label={t('discover.empty.west.action')}
                  onClick={() => {
                    change({ region: 'all' });
                  }}
                />
              ) : undefined
            }
          />
        ) : null}
        {items.length > 0 ? (
          <VStack gap={3}>
            <VStack as="ul" gap={3}>
              {items.map((resource) => (
                <li key={resource.id}>
                  <ResourceCard resource={resource} />
                </li>
              ))}
            </VStack>
            {list.isError ? (
              <ErrorState error={list.error} onRetry={() => void list.fetchNextPage()} />
            ) : null}
            {list.hasNextPage ? (
              <HStack hAlign="center">
                <Button
                  label={t('discover.more')}
                  isLoading={list.isFetchingNextPage}
                  onClick={() => void list.fetchNextPage()}
                />
              </HStack>
            ) : null}
          </VStack>
        ) : null}
      </section>
    </VStack>
  );
}
