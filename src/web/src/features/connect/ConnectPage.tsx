import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { Divider } from '@astryxdesign/core/Divider';
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
import { ApiError } from '../../api/http';
import { countryOptions } from '../../components/countries';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { CodeIcon, GlobeIcon, UsersIcon, type SvgIcon } from '../../components/icons';
import { List } from '../../components/List';
import { LoadingState } from '../../components/LoadingState';
import { PageHeader } from '../../components/PageHeader';
import { Section } from '../../components/Section';
import { StatusBadge } from '../../components/StatusBadge';
import { useI18n } from '../../i18n/I18nProvider';
import { GithubDevRow } from './GithubDevRow';
import { PersonRow } from './PersonRow';
import { useProfileDirectory } from './queries';
import { languageTag, skillTag, validateConnectSearch, type ConnectSearch } from './routeSearch';

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
        const next: ConnectSearch = {
          ...(country ? { country } : {}),
          ...(search.source ? { source: search.source } : {}),
        };
        const lang = languageTag(language);
        const tag = skillTag(skill);
        if (lang) next.language = lang;
        if (tag) next.skill = tag;
        go(next);
      }}
    >
      {/* The prototype's filter bar: one borderless panel holding the three filters. */}
      <div className="ui-panel ui-filter-bar">
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
          <SegmentedControl
            label={t('connect.source.label')}
            value={search.source ?? 'all'}
            onChange={(value) => {
              const next: ConnectSearch = { ...search };
              if (value === 'members' || value === 'github') next.source = value;
              else delete next.source;
              go(next);
            }}
          >
            <SegmentedControlItem value="all" label={t('connect.source.all')} />
            <SegmentedControlItem value="members" label={t('connect.source.members')} />
            <SegmentedControlItem value="github" label={t('connect.source.github')} />
          </SegmentedControl>
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
      </div>
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
  icon: SvgIcon;
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
  return (
    <Section
      title={t('connect.spaces.title')}
      titleAside={<StatusBadge kind="later" label={t('connect.spaces.badge')} />}
    >
      <Text as="p" type="supporting">
        {t('connect.soon')}
      </Text>
      <div className="ui-panel">
        <VStack as="ul" gap={0}>
          {spaces.map((space, index) => (
            <li key={space.key}>
              <HStack gap={3} vAlign="start" wrap="nowrap" paddingBlock={2}>
                <Icon icon={space.icon} color="secondary" />
                <Text weight="semibold">{t(`connect.space.${space.key}`)}</Text>
              </HStack>
              {index < spaces.length - 1 ? <Divider variant="subtle" /> : null}
            </li>
          ))}
        </VStack>
      </div>
    </Section>
  );
}

/** Where the external entries come from, when they were generated, and how to be removed (ADR-0016). */
function ExternalFooter({ generatedAt }: { generatedAt: string }) {
  const { t, locale } = useI18n();
  const date = new Intl.DateTimeFormat(locale, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date(generatedAt));
  return (
    <Text as="p" type="supporting">
      <span>{t('connect.external.footer', { date })}</span> ·{' '}
      <Link href={OPT_OUT_URL}>{t('connect.external.optout')}</Link>
    </Text>
  );
}

// Opt-out requests (ADR-0016 point 6) go privately to support, never to a public issue (owner, 2026-11-15).
const OPT_OUT_URL = `mailto:support@226os.tech?subject=${encodeURIComponent('Retrait de l’annuaire Connect')}`;

/**
 * `connect`: the people directory, filtered by the URL. Members open to collaboration come first; public
 * GitHub developers ranked by committers.top follow in their own labelled list (ADR-0016).
 */
export function ConnectPage() {
  const { t } = useI18n();
  // Router merges the raw parent search over the validated one, so invalid values are dropped again here.
  const search = validateConnectSearch(useSearch({ from: '/app/connect' }));
  const list = useProfileDirectory(search);
  const navigate = useNavigate();
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const external = list.data?.pages.flatMap((page) => page.external) ?? [];
  const seenCountries = [
    ...items.flatMap((person) => (person.country ? [person.country] : [])),
    ...external.map((dev) => dev.country),
  ];
  const firstExternal = external[0];
  const filtered = Object.keys(search).length > 0;
  const showMembers = search.source !== 'github' && (items.length > 0 || external.length === 0);
  const empty = list.data !== undefined && items.length === 0 && external.length === 0;

  // Loading, failure and the empty state sit in the first section on the page.
  const status = (
    <>
      {list.isPending ? <LoadingState rows={4} /> : null}
      {list.isError && !list.data ? (
        <LoadFailure error={list.error} onRetry={() => void list.refetch()} />
      ) : null}
      {empty ? (
        <EmptyState
          headingLevel={3}
          title={t(filtered ? 'connect.empty.filtered.title' : 'connect.empty.title')}
          {...(filtered
            ? {
                description: t('connect.empty.filtered.body'),
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
    </>
  );
  const more = (
    <>
      {list.isError && list.data ? (
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
    </>
  );

  return (
    <div className="ui-page">
      <PageHeader
        title={t('connect.title')}
        breadcrumbs={[{ label: t('nav.home'), href: '/' }, { label: t('connect.title') }]}
        lead={t('connect.lede')}
      />
      <Filters
        key={`${search.country ?? ''}|${search.language ?? ''}|${search.skill ?? ''}|${search.source ?? ''}`}
        search={search}
        seenCountries={seenCountries}
      />
      {showMembers ? (
        <Section title={t('connect.results.label')}>
          {status}
          {items.length > 0 ? (
            <VStack gap={4}>
              <List label={t('connect.results.label')}>
                {items.map((person) => (
                  <PersonRow key={person.username} person={person} />
                ))}
              </List>
              {external.length === 0 ? more : null}
            </VStack>
          ) : null}
        </Section>
      ) : null}
      {!showMembers || external.length > 0 ? (
        <Section title={t('connect.external.title')}>
          {showMembers ? null : status}
          {external.length > 0 ? (
            <VStack gap={4}>
              <List label={t('connect.external.title')}>
                {external.map((dev) => (
                  <GithubDevRow key={`${dev.country}/${dev.login}`} dev={dev} />
                ))}
              </List>
              {more}
              {firstExternal ? <ExternalFooter generatedAt={firstExternal.generated_at} /> : null}
            </VStack>
          ) : null}
        </Section>
      ) : null}
      <SpacesSoon />
    </div>
  );
}
