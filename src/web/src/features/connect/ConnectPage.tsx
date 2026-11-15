import { Badge } from '@astryxdesign/core/Badge';
import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { Divider } from '@astryxdesign/core/Divider';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useId, useState, type ComponentType, type SVGProps } from 'react';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale } from '../../i18n/messages';
import { countryName } from '../discover/present';
import { WEST_AFRICA } from '../discover/region';
import { CodeIcon, GlobeIcon, UsersIcon } from '../profile/icons';
import { PersonRow } from './PersonRow';
import { useProfileDirectory } from './queries';
import { languageTag, skillTag, validateConnectSearch, type ConnectSearch } from './routeSearch';

/** West Africa first (Discover's list), then any other country the page has seen. */
function countryOptions(extra: Iterable<string>, locale: Locale) {
  const byName = (codes: Iterable<string>) =>
    [...new Set(codes)]
      .map((code) => ({ value: code, label: countryName(code, locale) }))
      .sort((a, b) => a.label.localeCompare(b.label, locale));
  const west: readonly string[] = WEST_AFRICA;
  return [...byName(west), ...byName([...extra].filter((code) => !west.includes(code)))];
}

function Filters({
  search,
  seenCountries,
}: {
  search: ConnectSearch;
  seenCountries: readonly string[];
}) {
  const { t, locale } = useI18n();
  const navigate = useNavigate();
  const [language, setLanguage] = useState(search.language ?? '');
  const [skill, setSkill] = useState(search.skill ?? '');
  const active = Object.keys(search).length > 0;
  const go = (next: ConnectSearch) => {
    void navigate({ to: '/connect', search: next });
  };
  const country = search.country;
  return (
    <form
      aria-label={t('connect.filters.label')}
      onSubmit={(event) => {
        event.preventDefault();
        const next: ConnectSearch = { ...(country ? { country } : {}) };
        const lang = languageTag(language);
        const tag = skillTag(skill);
        if (lang) next.language = lang;
        if (tag) next.skill = tag;
        go(next);
      }}
    >
      {/* The prototype's filter bar: one borderless panel holding the three filters. */}
      <Card variant="muted" padding={3}>
        <HStack gap={3} vAlign="end" wrap="wrap">
          <Selector
            label={t('connect.country.label')}
            placeholder={t('connect.country.placeholder')}
            hasClear
            options={countryOptions([...seenCountries, ...(country ? [country] : [])], locale)}
            value={country ?? null}
            onChange={(value) => {
              const next: ConnectSearch = { ...search };
              if (value) next.country = value;
              else delete next.country;
              go(next);
            }}
          />
          <TextInput
            label={t('connect.language.label')}
            placeholder={t('connect.language.placeholder')}
            value={language}
            onChange={setLanguage}
            autoCapitalize="none"
            hasAutoCorrect={false}
            hasSpellCheck={false}
          />
          <TextInput
            label={t('connect.skill.label')}
            placeholder={t('connect.skill.placeholder')}
            value={skill}
            onChange={setSkill}
            autoCapitalize="none"
            hasAutoCorrect={false}
            hasSpellCheck={false}
          />
          <IconButton type="submit" label={t('connect.apply')} icon={<Icon icon="search" />} />
          {active ? (
            <Button
              variant="ghost"
              label={t('connect.reset')}
              onClick={() => {
                go({});
              }}
            />
          ) : null}
        </HStack>
      </Card>
    </form>
  );
}

/** A 429 names the wait the server asked for; every other failure uses the shared state. */
function LoadFailure({ error, onRetry }: { error: unknown; onRetry: () => void }) {
  const { t } = useI18n();
  if (!(error instanceof ApiError) || error.code !== 'RATE_LIMITED') {
    return <ErrorState error={error} onRetry={onRetry} />;
  }
  return (
    <div role="alert">
      <Banner
        status="warning"
        title={
          error.retryAfter
            ? t('connect.rate', { seconds: error.retryAfter })
            : t('connect.rate.now')
        }
        collapsible={false}
        endContent={<Button label={t('state.retry')} onClick={onRetry} />}
      />
    </div>
  );
}

const spaces: readonly {
  key: 'country' | 'tech' | 'topic';
  icon: ComponentType<SVGProps<SVGSVGElement>>;
}[] = [
  { key: 'country', icon: GlobeIcon },
  { key: 'tech', icon: CodeIcon },
  { key: 'topic', icon: UsersIcon },
];

/**
 * The spaces block the DISCLAIMER keeps on Connect, labelled « Bientôt » once: plain rows with no
 * link or action, so nothing looks usable before it is.
 */
function SpacesSoon() {
  const { t } = useI18n();
  const headingId = useId();
  return (
    <section aria-labelledby={headingId}>
      <VStack gap={3}>
        <VStack gap={1}>
          <HStack gap={2} vAlign="center" wrap="wrap">
            <Heading level={2} id={headingId}>
              {t('connect.spaces.title')}
            </Heading>
            <Badge variant="neutral" label={t('connect.spaces.badge')} />
          </HStack>
          <Text as="p" type="supporting">
            {t('connect.soon')}
          </Text>
        </VStack>
        <Card variant="muted" padding={3}>
          <VStack as="ul" gap={0}>
            {spaces.map((space, index) => (
              <li key={space.key}>
                <HStack gap={3} vAlign="start" wrap="nowrap" paddingBlock={2}>
                  <Icon icon={space.icon} color="secondary" />
                  <VStack gap={0.5}>
                    <Text weight="semibold">{t(`connect.space.${space.key}`)}</Text>
                    <Text type="supporting">{t(`connect.space.${space.key}.body`)}</Text>
                  </VStack>
                </HStack>
                {index < spaces.length - 1 ? <Divider variant="subtle" /> : null}
              </li>
            ))}
          </VStack>
        </Card>
      </VStack>
    </section>
  );
}

/** `connect`: the people directory (real members open to collaboration), filtered by the URL. */
export function ConnectPage() {
  const { t } = useI18n();
  const resultsId = useId();
  // Router merges the raw parent search over the validated one, so invalid values are dropped again here.
  const search = validateConnectSearch(useSearch({ from: '/app/connect' }));
  const list = useProfileDirectory(search);
  const navigate = useNavigate();
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const seenCountries = items.flatMap((person) => (person.country ? [person.country] : []));
  const filtered = Object.keys(search).length > 0;

  return (
    <VStack gap={8}>
      <VStack gap={5}>
        <VStack gap={1} hAlign="start">
          <Heading level={1}>{t('connect.title')}</Heading>
          <Text as="p" type="supporting">
            {t('connect.lede')}
          </Text>
        </VStack>
        <Filters
          key={`${search.country ?? ''}|${search.language ?? ''}|${search.skill ?? ''}`}
          search={search}
          seenCountries={seenCountries}
        />
      </VStack>
      <section aria-labelledby={resultsId}>
        <VStack gap={2}>
          <Heading level={2} id={resultsId}>
            {t('connect.results.label')}
          </Heading>
          {list.isPending ? <LoadingState /> : null}
          {list.isError && !list.data ? (
            <LoadFailure error={list.error} onRetry={() => void list.refetch()} />
          ) : null}
          {list.data && items.length === 0 ? (
            <EmptyState
              headingLevel={3}
              title={t(filtered ? 'connect.empty.filtered.title' : 'connect.empty.title')}
              description={t(filtered ? 'connect.empty.filtered.body' : 'connect.empty.body')}
              {...(filtered
                ? {
                    actions: (
                      <Button
                        label={t('connect.reset')}
                        onClick={() => void navigate({ to: '/connect', search: {} })}
                      />
                    ),
                  }
                : {})}
            />
          ) : null}
          {items.length > 0 ? (
            <VStack gap={4}>
              <VStack as="ul" gap={0}>
                {items.map((person, index) => (
                  <PersonRow
                    key={person.username}
                    person={person}
                    isLast={index === items.length - 1}
                  />
                ))}
              </VStack>
              {list.isError ? (
                <LoadFailure error={list.error} onRetry={() => void list.fetchNextPage()} />
              ) : null}
              {list.hasNextPage ? (
                <HStack hAlign="center">
                  <Button
                    label={t('connect.more')}
                    isLoading={list.isFetchingNextPage}
                    onClick={() => void list.fetchNextPage()}
                  />
                </HStack>
              ) : null}
            </VStack>
          ) : null}
        </VStack>
      </section>
      <SpacesSoon />
    </VStack>
  );
}
