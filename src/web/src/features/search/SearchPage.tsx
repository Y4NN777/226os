import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import type { components } from '../../api/schema';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { AssistantPanel } from './AssistantPanel';
import { isSearchable, normalizeQuery, searchQuery } from './searchQuery';

type SearchResults = components['schemas']['SearchResults'];
type SearchHit = components['schemas']['SearchHit'];
type GroupName = keyof SearchResults['groups'];

const groupOrder = [
  ['resources', 'search.group.resources'],
  ['questions', 'search.group.questions'],
  ['learn', 'search.group.learn'],
  ['projects', 'search.group.projects'],
] as const satisfies readonly (readonly [GroupName, MessageKey])[];

function SearchForm({ initial }: { initial: string }) {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [value, setValue] = useState(initial);
  const submit = () => {
    const q = normalizeQuery(value);
    void navigate({ to: '/recherche', search: q ? { q } : {} });
  };
  return (
    <form
      role="search"
      onSubmit={(event) => {
        event.preventDefault();
        submit();
      }}
    >
      <HStack gap={2} vAlign="end">
        <TextInput
          type="search"
          label={t('search.label')}
          value={value}
          onChange={setValue}
          htmlName="q"
          enterKeyHint="search"
          autoComplete="off"
          width="100%"
        />
        {/* The field takes the free width; the button keeps its full label. */}
        <span className="no-shrink">
          <Button type="submit" variant="primary" label={t('search.submit')} />
        </span>
      </HStack>
    </form>
  );
}

function Hit({ hit, linked }: { hit: SearchHit; linked: boolean }) {
  const { locale } = useI18n();
  const lang = hit.lang && hit.lang !== locale ? hit.lang : undefined;
  return (
    <li lang={lang}>
      <VStack gap={0.5} hAlign="start">
        {linked ? (
          <Link href={`/discover/${hit.id}`}>{hit.title}</Link>
        ) : (
          <Text weight="semibold">{hit.title}</Text>
        )}
        {hit.snippet ? (
          <Text as="p" type="supporting">
            {hit.snippet}
          </Text>
        ) : null}
      </VStack>
    </li>
  );
}

function Results({ q, data }: { q: string; data: SearchResults }) {
  const { t } = useI18n();
  const groups = groupOrder.flatMap(([name, label]) => {
    const group = data.groups[name];
    return group && group.items.length > 0 ? [{ name, label, group }] : [];
  });
  if (groups.length === 0) {
    return <EmptyState title={t('search.none.title', { q })} description={t('search.none.body')} />;
  }
  return (
    <VStack gap={6} hAlign="stretch">
      {groups.map(({ name, label, group }) => (
        <section key={name} aria-labelledby={`search-group-${name}`}>
          <VStack gap={3} hAlign="stretch">
            <Heading level={2} id={`search-group-${name}`}>
              {t(label)}
            </Heading>
            <VStack as="ul" gap={3} hAlign="stretch">
              {group.items.map((hit) => (
                <Hit key={hit.id} hit={hit} linked={name === 'resources'} />
              ))}
            </VStack>
            {group.has_more ? (
              <Text as="p" type="supporting">
                {t('search.more')}
              </Text>
            ) : null}
          </VStack>
        </section>
      ))}
    </VStack>
  );
}

function Outcome({ q }: { q: string }) {
  const query = useQuery(searchQuery(q));
  if (query.isPending) return <LoadingState />;
  if (query.isError) {
    return (
      <ErrorState
        error={query.error}
        title="search.error.title"
        onRetry={() => void query.refetch()}
      />
    );
  }
  return <Results q={q} data={query.data} />;
}

export function SearchPage() {
  const { t } = useI18n();
  const { q: rawQ } = useSearch({ strict: false });
  const q = normalizeQuery(rawQ ?? '');
  return (
    <VStack gap={6} hAlign="stretch" maxWidth={720}>
      <Heading level={1}>{t('search.title')}</Heading>
      <SearchForm key={q} initial={q} />
      {q === '' ? (
        <EmptyState title={t('search.start.title')} description={t('search.start.body')} />
      ) : !isSearchable(q) ? (
        <EmptyState title={t('search.short.title')} description={t('search.short.body')} />
      ) : (
        <>
          <Outcome q={q} />
          <AssistantPanel key={q} q={q} />
        </>
      )}
    </VStack>
  );
}
