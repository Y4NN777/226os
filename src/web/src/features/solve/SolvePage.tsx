import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { useSession } from '../../app/session';
import { AssistantPanel } from '../search/AssistantPanel';
import { useIdempotencyKey } from './idempotency';
import { Byline, StateBadge, TagList } from './present';
import { QuestionForm, type LinkDraft, type QuestionDraft } from './QuestionForm';
import { useCreateQuestion, useQuestionList, type Question } from './queries';
import { parseLink, type SolveSearch } from './routeSearch';
import { signInHref } from './signIn';

function Filters({ search }: { search: SolveSearch }) {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [q, setQ] = useState(search.q ?? '');
  const [tag, setTag] = useState(search.tag ?? '');
  return (
    <form
      role="search"
      onSubmit={(event) => {
        event.preventDefault();
        const next: SolveSearch = {
          ...(q.trim() ? { q: q.trim() } : {}),
          ...(tag.trim() ? { tag: tag.trim().toLowerCase() } : {}),
        };
        void navigate({ to: '/solve', search: next });
      }}
    >
      <HStack gap={2} vAlign="end" wrap="wrap">
        {/* One row on desktop: the question field takes the free width, the tag filter and the button follow. */}
        <div className="grow-field">
          <TextInput
            type="search"
            label={t('solve.search.label')}
            isLabelHidden
            placeholder={t('solve.search.placeholder')}
            value={q}
            onChange={setQ}
            width="100%"
          />
        </div>
        <TextInput
          label={t('solve.tag.label')}
          isLabelHidden
          placeholder={t('solve.tag.placeholder')}
          value={tag}
          onChange={setTag}
        />
        <IconButton type="submit" label={t('solve.search.submit')} icon={<Icon icon="search" />} />
      </HStack>
    </form>
  );
}

function QuestionCard({ question }: { question: Question }) {
  const { t } = useI18n();
  return (
    <Card variant="muted">
      <VStack gap={2}>
        <HStack gap={2} wrap="wrap" vAlign="center">
          <Link href={`/questions/${question.id}`} type="large" weight="semibold">
            {question.title}
          </Link>
          <StateBadge state={question.state} />
        </HStack>
        <TagList tags={question.tags} />
        <HStack gap={3} wrap="wrap">
          <Byline name={question.author.display_name} />
          {/* With no answer the state badge already says so. */}
          {question.answer_count > 0 ? (
            <Text type="supporting">{t('solve.answers.count', { n: question.answer_count })}</Text>
          ) : null}
          <Text type="supporting">{t('solve.score', { n: question.vote_score })}</Text>
        </HStack>
      </VStack>
    </Card>
  );
}

function AskForm({ onClose, link }: { onClose: () => void; link: LinkDraft | undefined }) {
  const { t, locale } = useI18n();
  const navigate = useNavigate();
  const create = useCreateQuestion();
  const idempotency = useIdempotencyKey();

  const submit = async (draft: QuestionDraft) => {
    const body = {
      title: draft.title,
      body_md: draft.body_md,
      lang: draft.lang,
      tags: draft.tags,
      ...(draft.snippets.length > 0 ? { snippets: draft.snippets } : {}),
      ...(draft.links.length > 0 ? { links: draft.links } : {}),
    };
    const created = await create.mutateAsync({ body, key: idempotency.keyFor(body) });
    idempotency.settle();
    await navigate({ to: '/questions/$questionId', params: { questionId: created.id } });
  };

  return (
    <QuestionForm
      mode="create"
      initial={{ title: '', body: '', tags: [], lang: locale }}
      submitLabel={t('solve.form.submit')}
      onSubmit={submit}
      onCancel={onClose}
      initialLink={link}
      assistant={(draft) => (
        <AssistantPanel
          q={[draft.title.trim(), draft.body.trim()].filter(Boolean).join('\n\n')}
          free={false}
          signInNext="/solve"
        />
      )}
    />
  );
}

export function SolvePage() {
  const { t } = useI18n();
  const search = useSearch({ from: '/app/solve' });
  const navigate = useNavigate();
  const session = useSession();
  const prefilled = parseLink(search.link);
  const [asking, setAsking] = useState(prefilled !== undefined);
  const list = useQuestionList(search);
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const signedIn = session.data?.authenticated === true;
  const hasFilters = Object.keys(search).length > 0;

  return (
    <VStack gap={5}>
      <VStack gap={1}>
        <Heading level={1}>{t('solve.title')}</Heading>
        <Text as="p">{t('solve.lede')}</Text>
      </VStack>
      {asking && signedIn ? (
        <AskForm
          link={prefilled}
          onClose={() => {
            setAsking(false);
          }}
        />
      ) : (
        <HStack>
          {signedIn ? (
            <Button
              variant="primary"
              label={t('solve.ask')}
              onClick={() => {
                setAsking(true);
              }}
            />
          ) : session.isPending ? null : (
            <Button
              variant="primary"
              label={t('solve.ask.signIn')}
              href={signInHref(search.link ? `/solve?link=${search.link}` : '/solve')}
            />
          )}
        </HStack>
      )}
      <Filters key={`${search.q ?? ''}|${search.tag ?? ''}`} search={search} />
      {hasFilters ? (
        <HStack>
          <Button
            variant="ghost"
            label={t('solve.reset')}
            onClick={() => void navigate({ to: '/solve', search: {} })}
          />
        </HStack>
      ) : null}
      <section aria-label={t('solve.results.label')}>
        {list.isPending ? <LoadingState /> : null}
        {list.isError && !list.data ? (
          <ErrorState error={list.error} onRetry={() => void list.refetch()} />
        ) : null}
        {list.data && items.length === 0 ? (
          <EmptyState
            headingLevel={2}
            title={t('solve.empty.title')}
            description={t('solve.empty.body')}
          />
        ) : null}
        {items.length > 0 ? (
          <VStack gap={3}>
            <VStack as="ul" gap={3}>
              {items.map((question) => (
                <li key={question.id}>
                  <QuestionCard question={question} />
                </li>
              ))}
            </VStack>
            {list.isError ? (
              <ErrorState error={list.error} onRetry={() => void list.fetchNextPage()} />
            ) : null}
            {list.hasNextPage ? (
              <HStack hAlign="center">
                <Button
                  label={t('solve.more')}
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
