import { Button } from '@astryxdesign/core/Button';
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
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { SearchIcon } from '../../components/icons';
import { List } from '../../components/List';
import { LoadingState } from '../../components/LoadingState';
import { countryOptions } from '../../components/countries';
import { PageHeader } from '../../components/PageHeader';
import { ResultsMeta } from '../../components/ResultsMeta';
import { CardGrid } from '../../components/CardGrid';
import { SortSelect } from '../../components/SortSelect';
import { useListView, type ListView } from '../../components/useListView';
import { ViewToggle } from '../../components/ViewToggle';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { categoryCodes, categoryKey, isCategoryCode } from './categories';
import { useResourceList } from './queries';
import { ResourceRow } from './ResourceRow';
import { validateDiscoverSearch, type DiscoverSearch } from './routeSearch';
import './discover.css';

type SearchPatch = { [K in keyof DiscoverSearch]?: DiscoverSearch[K] | undefined };

/** Applies a patch to the URL filters; an undefined value removes the filter. */
function patched(base: DiscoverSearch, patch: SearchPatch): DiscoverSearch {
  const entries = Object.entries({ ...base, ...patch }).filter(([, value]) => value !== undefined);
  return Object.fromEntries(entries);
}

/** Results count from the API total, which covers every page. */
function countText(n: number, t: (key: MessageKey, p?: { n: number }) => string) {
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
        <TextInput
          type="search"
          size="lg"
          startIcon={<Icon icon="search" />}
          label={t('discover.search.label')}
          isLabelHidden
          placeholder={t('discover.search.placeholder')}
          value={draft}
          onChange={setDraft}
          enterKeyHint="search"
          width="100%"
        />
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
        <div className="discover-filter discover-filter-type">
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
  const { view: addressView, ...search } = validateDiscoverSearch(
    useSearch({ from: '/app/discover' }),
  );
  const { view, remember } = useListView('discover', addressView);
  const navigate = useNavigate();
  const list = useResourceList(search);
  const westView = search.region !== 'all' && search.country === undefined;
  const hasFilters = Object.keys(search).length > 0;
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const total = list.data?.pages.at(-1)?.total;
  const sortOptions = (['name', 'verified'] as const).map((sort) => ({
    value: sort,
    label: t(`discover.sort.${sort}`),
  }));

  // The view stays in the address when the address already holds it.
  const keepView = (next: DiscoverSearch): DiscoverSearch =>
    addressView ? { ...next, view: addressView } : next;
  const change = (patch: SearchPatch) =>
    void navigate({ to: '/discover', search: keepView(patched(search, patch)) });
  const switchView = (next: ListView) => {
    remember(next);
    void navigate({ to: '/discover', search: { ...search, view: next } });
  };
  const resetButton = (variant: 'ghost' | 'secondary') => (
    <Button
      variant={variant}
      size={variant === 'ghost' ? 'sm' : 'md'}
      label={t('discover.reset')}
      onClick={() => void navigate({ to: '/discover', search: keepView({}) })}
    />
  );

  return (
    <div className="ui-page">
      <PageHeader
        title={t('discover.title')}
        breadcrumbs={[{ label: t('nav.home'), href: '/' }, { label: t('discover.title') }]}
        lead={t('discover.lede')}
      />
      <Filters
        key={search.q ?? ''}
        search={search}
        seenCountries={items.flatMap((item) => item.countries)}
        onChange={change}
      />
      <section aria-labelledby="discover-results-h" className="ui-section">
        {/* Prototype `h2.visually-hidden`: keeps the outline h1 > h2 > h3 for the rows. */}
        <VisuallyHidden as="h2" id="discover-results-h">
          {t('discover.results.label')}
        </VisuallyHidden>
        {/* Always mounted, so a count that appears later is announced. */}
        <ResultsMeta summary={items.length > 0 && total !== undefined ? countText(total, t) : null}>
          <SortSelect
            label={t('discover.sort.label')}
            options={sortOptions}
            value={search.sort ?? 'name'}
            onChange={(sort) => {
              change({ sort: sort === 'name' ? undefined : sort });
            }}
          />
          {hasFilters && items.length > 0 ? resetButton('ghost') : null}
          <ViewToggle value={view} onChange={switchView} />
        </ResultsMeta>
        {list.isPending ? <LoadingState rows={4} /> : null}
        {list.isError && !list.data ? (
          <ErrorState error={list.error} onRetry={() => void list.refetch()} />
        ) : null}
        {list.data && items.length === 0 ? (
          <EmptyState
            headingLevel={3}
            icon={SearchIcon}
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
        ) : null}
        {items.length > 0 ? (
          <VStack gap={3}>
            {view === 'cards' ? (
              <CardGrid label={t('discover.results.label')}>
                {items.map((resource) => (
                  <ResourceRow key={resource.id} resource={resource} />
                ))}
              </CardGrid>
            ) : (
              <List label={t('discover.results.label')}>
                {items.map((resource) => (
                  <ResourceRow key={resource.id} resource={resource} />
                ))}
              </List>
            )}
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
    </div>
  );
}
