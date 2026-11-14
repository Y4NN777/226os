import { Avatar } from '@astryxdesign/core/Avatar';
import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { Locale } from '../../i18n/messages';
import { countryName, languageName } from '../discover/present';
import { WEST_AFRICA } from '../discover/region';
import { TagList } from '../learn/present';
import { useProfileDirectory, type ProfileSummary } from './queries';
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
      <HStack gap={2} vAlign="end" wrap="wrap">
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
    </form>
  );
}

function PersonCard({ person }: { person: ProfileSummary }) {
  const { t, locale } = useI18n();
  const facts = [
    person.country ? countryName(person.country, locale) : undefined,
    person.languages && person.languages.length > 0
      ? person.languages.map((code) => languageName(code, locale)).join(', ')
      : undefined,
  ].filter((fact): fact is string => fact !== undefined);
  return (
    <HStack gap={3} vAlign="start" wrap="nowrap">
      <Avatar name={person.display_name} size="md" />
      <VStack gap={1}>
        <HStack gap={2} wrap="wrap" vAlign="center">
          <Link href={`/profils/${person.username}`} type="large" weight="semibold">
            {person.display_name}
          </Link>
          <Text type="supporting">@{person.username}</Text>
        </HStack>
        {facts.map((fact) => (
          <Text key={fact} type="supporting">
            {fact}
          </Text>
        ))}
        {person.skills && person.skills.length > 0 ? <TagList tags={person.skills} /> : null}
        {person.github ? (
          <Link
            href={`https://github.com/${person.github}`}
            aria-label={t('connect.github', { name: person.display_name })}
          >
            GitHub
          </Link>
        ) : null}
      </VStack>
    </HStack>
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

/** `connect`: the people directory (real members open to collaboration), filtered by the URL. */
export function ConnectPage() {
  const { t } = useI18n();
  // Router merges the raw parent search over the validated one, so invalid values are dropped again here.
  const search = validateConnectSearch(useSearch({ from: '/app/connect' }));
  const list = useProfileDirectory(search);
  const navigate = useNavigate();
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const seenCountries = items.flatMap((person) => (person.country ? [person.country] : []));
  const filtered = Object.keys(search).length > 0;

  return (
    <VStack gap={5}>
      <VStack gap={1} hAlign="start">
        <Heading level={1}>{t('connect.title')}</Heading>
        <Text as="p">{t('connect.lede')}</Text>
      </VStack>
      <Filters
        key={`${search.country ?? ''}|${search.language ?? ''}|${search.skill ?? ''}`}
        search={search}
        seenCountries={seenCountries}
      />
      <section aria-label={t('connect.results.label')}>
        {list.isPending ? <LoadingState /> : null}
        {list.isError && !list.data ? (
          <LoadFailure error={list.error} onRetry={() => void list.refetch()} />
        ) : null}
        {list.data && items.length === 0 ? (
          <EmptyState
            headingLevel={2}
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
            <VStack as="ul" gap={4}>
              {items.map((person) => (
                <li key={person.username}>
                  <PersonCard person={person} />
                </li>
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
      </section>
      <Text as="p" type="supporting">
        {t('connect.soon')}
      </Text>
    </VStack>
  );
}
