import { Button } from '@astryxdesign/core/Button';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useLocation, useNavigate, useSearch } from '@tanstack/react-router';
import { useEffect, useRef, useState } from 'react';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { PlusIcon } from '../../components/icons';
import { List } from '../../components/List';
import { LoadingState } from '../../components/LoadingState';
import { PageHeader } from '../../components/PageHeader';
import { useI18n } from '../../i18n/I18nProvider';
import { useSession } from '../../app/session';
import { AssistantPanel } from '../search/AssistantPanel';
import { useIdempotencyKey } from './idempotency';
import { QuestionForm, type LinkDraft, type QuestionDraft } from './QuestionForm';
import { QuestionRow } from './QuestionRow';
import { useCreateQuestion, useQuestionList } from './queries';
import { parseLink, questionStates, type SolveSearch } from './routeSearch';
import { signInHref } from './signIn';
import './solve.css';

const ALL = 'all';

/** Search field, state filter (the prototype's segmented control) and the tag filter panel. */
function Filters({ search }: { search: SolveSearch }) {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [q, setQ] = useState(search.q ?? '');
  const [tag, setTag] = useState(search.tag ?? '');
  const [panelOpen, setPanelOpen] = useState(Boolean(search.tag));
  const hasFilters = Boolean(search.q ?? search.tag ?? search.state);

  return (
    <form
      role="search"
      className="solve-search"
      onSubmit={(event) => {
        event.preventDefault();
        const next: SolveSearch = {
          ...(q.trim() ? { q: q.trim() } : {}),
          ...(tag.trim() ? { tag: tag.trim().toLowerCase() } : {}),
          ...(search.state ? { state: search.state } : {}),
        };
        void navigate({ to: '/solve', search: next });
      }}
    >
      <div className="solve-searchbar">
        <div className="solve-grow">
          <TextInput
            type="search"
            label={t('solve.search.label')}
            isLabelHidden
            size="lg"
            startIcon={<Icon icon="search" />}
            placeholder={t('solve.search.placeholder')}
            value={q}
            onChange={setQ}
            enterKeyHint="search"
            width="100%"
          />
        </div>
        {/* Enter submits, from this field and from the tag filter: the icon sits in the field. */}
        <button type="submit" hidden tabIndex={-1} aria-hidden="true" />
      </div>
      <div className="solve-toolbar">
        <div className="solve-seg">
          <SegmentedControl
            label={t('solve.filter.state')}
            value={search.state ?? ALL}
            onChange={(value) => {
              const state = questionStates.find((candidate) => candidate === value);
              const next: SolveSearch = { ...search };
              delete next.state;
              if (state) next.state = state;
              void navigate({ to: '/solve', search: next });
            }}
          >
            <SegmentedControlItem value={ALL} label={t('solve.filter.all')} />
            <SegmentedControlItem value="open" label={t('solve.filter.open')} />
            <SegmentedControlItem value="answered" label={t('solve.filter.answered')} />
            <SegmentedControlItem value="resolved" label={t('solve.filter.resolved')} />
          </SegmentedControl>
        </div>
        <div className="solve-toolbar-end">
          {hasFilters ? (
            <Button
              variant="ghost"
              size="sm"
              label={t('solve.reset')}
              onClick={() => void navigate({ to: '/solve', search: {} })}
            />
          ) : null}
          <IconButton
            variant="ghost"
            label={t('solve.filters')}
            icon={<Icon icon="funnel" />}
            aria-expanded={panelOpen}
            aria-controls="solve-filters"
            onClick={() => {
              setPanelOpen((open) => !open);
            }}
          />
        </div>
      </div>
      {/* Always in the DOM so aria-controls always points at it; hidden while collapsed. */}
      <div id="solve-filters" className="solve-filters" hidden={!panelOpen}>
        <div className="solve-grow">
          <TextInput
            label={t('solve.tag.label')}
            placeholder={t('solve.tag.placeholder')}
            value={tag}
            onChange={setTag}
            autoCapitalize="none"
            hasAutoCorrect={false}
            hasSpellCheck={false}
            width="100%"
          />
        </div>
      </div>
    </form>
  );
}

/** The prototype's #ask view: back to the list, a heading, the form and the assistant beside it. */
function AskView({
  onClose,
  link,
  focusField,
}: {
  onClose: () => void;
  link: LinkDraft | undefined;
  focusField: boolean;
}) {
  const { t, locale } = useI18n();
  const navigate = useNavigate();
  const create = useCreateQuestion();
  const idempotency = useIdempotencyKey();
  const heading = useRef<HTMLHeadingElement>(null);

  // The view replaces the list: move focus to its heading, as a route change would.
  // Opened by a link while already on the list, it goes straight to the first field.
  useEffect(() => {
    if (focusField) document.querySelector<HTMLElement>('#ask input, #ask textarea')?.focus();
    else heading.current?.focus();
  }, [focusField]);

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
    <div id="ask" className="ui-page">
      {/* The ask view is local state on /solve, so the way back is a button, not a crumb. */}
      <PageHeader
        title={t('solve.ask.h1')}
        headingRef={heading}
        breadcrumbs={[{ label: t('nav.home'), href: '/' }, { label: t('nav.solve') }]}
        lead={t('solve.ask.intro')}
        actions={
          <button type="button" className="solve-back" onClick={onClose}>
            <Icon icon="chevronLeft" size="sm" />
            {t('solve.back')}
          </button>
        }
      />
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
    </div>
  );
}

export function SolvePage() {
  const { t } = useI18n();
  const search = useSearch({ from: '/app/solve' });
  const hash = useLocation({ select: (location) => location.hash });
  const session = useSession();
  const prefilled = parseLink(search.link);
  const [asking, setAsking] = useState(prefilled !== undefined || hash === 'ask');
  const [fromLink, setFromLink] = useState(false);
  const firstHash = useRef<string | undefined>(hash);
  // « Poser une question » links to /solve#ask: it must also work from the list itself.
  useEffect(() => {
    if (hash === 'ask' && hash !== firstHash.current) {
      setAsking(true);
      setFromLink(true);
    }
    firstHash.current = undefined;
  }, [hash]);
  const navigate = useNavigate();
  const list = useQuestionList(search);
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const signedIn = session.data?.authenticated === true;

  if (asking && signedIn) {
    return (
      <AskView
        link={prefilled}
        focusField={fromLink}
        onClose={() => {
          setFromLink(false);
          setAsking(false);
          // Leaving #ask lets the shell link open the view again.
          if (hash === 'ask') void navigate({ to: '/solve', search, hash: '', replace: true });
        }}
      />
    );
  }

  const askAction = signedIn ? (
    <Button
      variant="primary"
      label={t('solve.ask')}
      icon={<Icon icon={PlusIcon} />}
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
  );

  return (
    <div className="ui-page">
      <PageHeader
        title={t('solve.h1')}
        breadcrumbs={[{ label: t('nav.home'), href: '/' }, { label: t('nav.solve') }]}
        actions={askAction}
        lead={t('solve.list.lede')}
      />
      <Filters key={`${search.q ?? ''}|${search.tag ?? ''}`} search={search} />
      <section aria-label={t('solve.results.label')} className="ui-section">
        {list.isPending ? <LoadingState rows={4} /> : null}
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
          <VStack gap={4}>
            <List label={t('solve.results.label')}>
              {items.map((question) => (
                <QuestionRow key={question.id} question={question} />
              ))}
            </List>
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
    </div>
  );
}
