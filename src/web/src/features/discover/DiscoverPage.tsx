import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale, MessageKey } from '../../i18n/messages';
import { categoryCodes, categoryKey, isCategoryCode } from './categories';
import { countryName } from './present';
import { useResourceList } from './queries';
import { WEST_AFRICA } from './region';
import { ResourceRow } from './ResourceRow';
import type { DiscoverSearch } from './routeSearch';
import './discover.css';

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

/** Results count: a total only once the last page is loaded, since pages are cursor-based. */
function countText(n: number, hasMore: boolean, t: (key: MessageKey, p?: { n: number }) => string) {
  if (hasMore) return t('discover.count.first', { n });
  return n === 1 ? t('discover.count.one') : t('discover.count.other', { n });
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
  const knownCategory =
    search.category !== undefined && isCategoryCode(search.category) ? search.category : null;
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
            size="lg"
            startIcon="search"
            label={t('discover.search.label')}
            isLabelHidden
            placeholder={t('discover.search.placeholder')}
            value={draft}
            onChange={setDraft}
            width="100%"
          />
          <IconButton
            type="submit"
            size="lg"
            variant="ghost"
            label={t('discover.search.submit')}
            icon={<Icon icon="search" />}
          />
        </HStack>
      </form>
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
      <div className="discover-filters">
        <Selector
          label={t('discover.country.label')}
          placeholder={t(
            // The default view is West Africa only, so « Tous les pays » would overstate it.
            wholeContinent ? 'discover.country.placeholder.all' : 'discover.country.placeholder',
          )}
          options={options}
          hasClear
          hasSearch
          width="100%"
          value={search.country ?? null}
          onChange={(value) => {
            onChange({ country: value || undefined });
          }}
        />
        <Selector
          label={t('discover.category.label')}
          placeholder={t('discover.category.placeholder')}
          options={categoryCodes.map((code) => ({ value: code, label: t(categoryKey(code)) }))}
          hasClear
          hasSearch
          width="100%"
          value={knownCategory}
          onChange={(value) => {
            onChange({ category: value || undefined });
          }}
        />
        <div className="discover-filter">
          <span className="discover-filter-label" aria-hidden="true">
            {t('discover.type.label')}
          </span>
          <SegmentedControl
            label={t('discover.type.label')}
            layout="fill"
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
        </div>
      </div>
      {search.category && knownCategory === null ? (
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
  const resetButton = (variant: 'ghost' | 'secondary') => (
    <Button
      variant={variant}
      size={variant === 'ghost' ? 'sm' : 'md'}
      label={t('discover.reset')}
      onClick={() => void navigate({ to: '/discover', search: {} })}
    />
  );

  return (
    <VStack gap={5}>
      <header className="discover-page-header">
        <Heading level={1}>{t('discover.title')}</Heading>
        <Text as="p" color="secondary">
          {t('discover.lede')}
        </Text>
        <Text as="p" type="supporting">
          {t('discover.positioning')}
        </Text>
      </header>
      <Filters
        key={search.q ?? ''}
        search={search}
        seenCountries={items.flatMap((item) => item.countries)}
        onChange={change}
      />
      <section aria-labelledby="discover-results-h" className="discover-results">
        {/* Prototype `h2.visually-hidden`: keeps the outline h1 > h2 > h3 for the rows. */}
        <VisuallyHidden as="h2" id="discover-results-h">
          {t('discover.results.label')}
        </VisuallyHidden>
        <div className="discover-results-bar">
          {/* Always mounted, so a count that appears later is announced. */}
          <p className="discover-count" role="status">
            {items.length > 0 ? countText(items.length, list.hasNextPage, t) : null}
          </p>
          {hasFilters && items.length > 0 ? resetButton('ghost') : null}
        </div>
        {list.isPending ? <LoadingState /> : null}
        {list.isError && !list.data ? (
          <ErrorState error={list.error} onRetry={() => void list.refetch()} />
        ) : null}
        {list.data && items.length === 0 ? (
          <div className="discover-panel">
            <EmptyState
              headingLevel={3}
              icon={<Icon icon="search" />}
              title={t('discover.empty.title')}
              description={t(westView ? 'discover.empty.west' : 'discover.empty.other')}
              actions={
                westView || hasFilters ? (
                  <HStack gap={2} hAlign="center" wrap="wrap">
                    {westView ? (
                      <Button
                        label={t('discover.empty.west.action')}
                        onClick={() => {
                          change({ region: 'all' });
                        }}
                      />
                    ) : null}
                    {hasFilters ? resetButton('secondary') : null}
                  </HStack>
                ) : undefined
              }
            />
          </div>
        ) : null}
        {items.length > 0 ? (
          <VStack gap={3}>
            <ul className="discover-rows">
              {items.map((resource) => (
                <ResourceRow key={resource.id} resource={resource} />
              ))}
            </ul>
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
