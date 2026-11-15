import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { useParams } from '@tanstack/react-router';
import { useId, useState } from 'react';
import { useSession } from '../../app/session';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { RouterAnchor } from '../../components/RouterAnchor';
import { useI18n } from '../../i18n/I18nProvider';
import { relatedPath } from '../related';
import { AnswerForm } from './AnswerForm';
import { ChatGlyph, LinkGlyph, PencilGlyph } from './icons';
import { AcceptButton, useVote, VoteColumn, VoteError } from './Interactions';
import { countryName, DateText, LangMark, Person, StatePill, TagList } from './present';
import { QuestionForm, type QuestionDraft } from './QuestionForm';
import {
  useAnswers,
  useInteractions,
  useQuestion,
  useUpdateQuestion,
  type Answer,
  type Question,
  type QuestionUpdate,
  type QuestionWithTag,
} from './queries';
import { RichText } from './RichText';
import { signInHref } from './signIn';
import './solve.css';

/** The merge patch holds only what changed, so an untouched field never overwrites a newer one. */
function diffOf(
  question: Question,
  draft: Pick<QuestionDraft, 'title' | 'body_md' | 'tags'>,
): QuestionUpdate {
  const patch: QuestionUpdate = {};
  if (draft.title !== question.title) patch.title = draft.title;
  if (draft.body_md !== question.body_md) patch.body_md = draft.body_md;
  if (draft.tags.join(',') !== question.tags.join(',')) patch.tags = draft.tags;
  return patch;
}

/** "1 réponse", "2 réponses" (French counts 0 and 1 as singular; Intl knows the rule). */
function useAnswerCount() {
  const { t, locale } = useI18n();
  const rules = new Intl.PluralRules(locale);
  return (n: number) =>
    rules.select(n) === 'one' ? t('solve.answers.one', { n }) : t('solve.answers.many', { n });
}

function EditQuestion({
  data,
  onDone,
  onReload,
}: {
  data: QuestionWithTag;
  onDone: () => void;
  onReload: () => void;
}) {
  const { t } = useI18n();
  const { question, etag } = data;
  const update = useUpdateQuestion(question.id);

  const submit = async (draft: QuestionDraft) => {
    const body = diffOf(question, draft);
    // Without an ETag the edit cannot be sent safely: answer like a stale version (reload prompt).
    if (!etag) throw new ApiError({ status: 412, code: 'PRECONDITION_FAILED' });
    await update.mutateAsync({ body, etag });
    onDone();
  };

  return (
    <QuestionForm
      mode="edit"
      initial={{
        title: question.title,
        body: question.body_md,
        tags: question.tags,
        lang: question.lang,
      }}
      submitLabel={t('solve.question.save')}
      onSubmit={submit}
      onCancel={onDone}
      isUnchanged={(draft) => Object.keys(diffOf(question, draft)).length === 0}
      failureNotice={(error) =>
        error instanceof ApiError && error.code === 'PRECONDITION_FAILED' ? (
          <div role="alert">
            <Banner
              status="warning"
              title={t('solve.stale.title')}
              description={t('solve.stale.body')}
              collapsible={false}
              endContent={<Button label={t('solve.stale.reload')} onClick={onReload} />}
            />
          </div>
        ) : undefined
      }
    />
  );
}

/** `new URL` in a try: URL.canParse is missing from Safari before 17. */
function parseUrl(value: string): URL | null {
  try {
    return new URL(value);
  } catch {
    return null;
  }
}

/** Answer sources: http(s) addresses become external links, anything else stays plain text. */
function Sources({ sources }: { sources: readonly string[] }) {
  const { t } = useI18n();
  const id = useId();
  if (sources.length === 0) return null;
  return (
    <div className="solve-sources">
      <p id={id} className="solve-sources-h">
        <Icon icon={LinkGlyph} size="sm" />
        {t('solve.answer.sources')}
      </p>
      <ul aria-labelledby={id}>
        {sources.map((source) => {
          const url = parseUrl(source);
          const safe = url !== null && (url.protocol === 'https:' || url.protocol === 'http:');
          const label = safe ? `${url.host}${url.pathname === '/' ? '' : url.pathname}` : source;
          return (
            <li key={source}>
              {safe ? (
                <Link href={url.href} isExternalLink newTabLabel={t('solve.newTab')}>
                  {label}
                </Link>
              ) : (
                <Text type="supporting">{label}</Text>
              )}
            </li>
          );
        })}
      </ul>
    </div>
  );
}

function AnswerItem({
  answer,
  questionId,
  interactive,
  canAccept,
}: {
  answer: Answer;
  questionId: string;
  interactive: boolean;
  canAccept: boolean;
}) {
  const vote = useVote({
    target: 'answer',
    id: answer.id,
    score: answer.vote_score,
    viewerVote: answer.viewer_vote,
    questionId: answer.question_id,
  });
  return (
    <li className="solve-surface">
      {answer.is_accepted ? (
        <div className="solve-answer-head">
          <StatePill state="resolved" />
        </div>
      ) : null}
      <div className="solve-post">
        <VoteColumn vote={vote} interactive={interactive} />
        <div className="solve-post-main">
          <div lang={answer.lang}>
            <RichText markdown={answer.body_md} snippets={answer.snippets} />
          </div>
          <VoteError vote={vote} />
          <Sources sources={answer.sources} />
          <div className="solve-answer-foot">
            <div className="solve-meta">
              <Person author={answer.author} />
              <DateText iso={answer.created_at} />
              <LangMark lang={answer.lang} />
            </div>
            {canAccept ? (
              <AcceptButton
                questionId={questionId}
                answerId={answer.id}
                isAccepted={answer.is_accepted}
              />
            ) : null}
          </div>
        </div>
      </div>
    </li>
  );
}

function Answers({
  questionId,
  count,
  interactive,
  canAccept,
}: {
  questionId: string;
  count: number;
  interactive: boolean;
  canAccept: boolean;
}) {
  const { t } = useI18n();
  const answerCount = useAnswerCount();
  const answers = useAnswers(questionId);
  // The server already sends the accepted answer first; sorting keeps it so across pages.
  const items = (answers.data?.pages.flatMap((page) => page.items) ?? [])
    .map((answer, index) => ({ answer, index }))
    .sort(
      (a, b) => Number(b.answer.is_accepted) - Number(a.answer.is_accepted) || a.index - b.index,
    )
    .map(({ answer }) => answer);
  return (
    <section className="solve-detail-main">
      <Heading level={2}>{count > 0 ? answerCount(count) : t('solve.answers.title')}</Heading>
      {answers.isPending ? <LoadingState /> : null}
      {answers.isError && !answers.data ? (
        <ErrorState error={answers.error} onRetry={() => void answers.refetch()} />
      ) : null}
      {answers.data && items.length === 0 ? <Text as="p">{t('solve.answers.none')}</Text> : null}
      {items.length > 0 ? (
        <ul className="solve-answer-list" aria-label={t('solve.answers.title')}>
          {items.map((answer) => (
            <AnswerItem
              key={answer.id}
              answer={answer}
              questionId={questionId}
              interactive={interactive}
              canAccept={canAccept}
            />
          ))}
        </ul>
      ) : null}
      {answers.hasNextPage ? (
        <HStack hAlign="center">
          <Button
            label={t('solve.more')}
            isLoading={answers.isFetchingNextPage}
            onClick={() => void answers.fetchNextPage()}
          />
        </HStack>
      ) : null}
    </section>
  );
}

const kindKey = {
  question: 'solve.kind.question',
  resource: 'solve.kind.resource',
  project: 'solve.kind.project',
  article: 'solve.kind.article',
  series: 'solve.kind.series',
} as const;

/** Context and linked items, the prototype's side column (only what the contract carries). */
function Aside({ question }: { question: Question }) {
  const { t, locale } = useI18n();
  const contextId = useId();
  const linkedId = useId();
  if (!question.country && question.related.length === 0) return null;
  return (
    <div className="solve-aside">
      {question.country ? (
        <aside className="solve-aside-block" aria-labelledby={contextId}>
          <h2 id={contextId} className="solve-aside-h">
            {t('solve.question.context')}
          </h2>
          <dl className="solve-context">
            <dt>{t('solve.question.country')}</dt>
            <dd>
              <span className="solve-inline">
                <span className="solve-country" aria-hidden="true">
                  {question.country}
                </span>
                {countryName(question.country, locale)}
              </span>
            </dd>
          </dl>
        </aside>
      ) : null}
      {question.related.length > 0 ? (
        <aside className="solve-aside-block" aria-labelledby={linkedId}>
          <h2 id={linkedId} className="solve-aside-h">
            {t('solve.question.linked')}
          </h2>
          <ul className="solve-linked">
            {question.related.map((link) => (
              <li key={`${link.type}-${link.id}`}>
                <span className="solve-linked-kind">{t(kindKey[link.type])}</span>
                <RouterAnchor href={relatedPath(link)} lang={link.lang}>
                  {link.title}
                </RouterAnchor>
              </li>
            ))}
          </ul>
        </aside>
      ) : null}
    </div>
  );
}

function QuestionPost({ question, interactive }: { question: Question; interactive: boolean }) {
  const vote = useVote({
    target: 'question',
    id: question.id,
    score: question.vote_score,
    viewerVote: question.viewer_vote,
    questionId: question.id,
  });
  return (
    <article className="solve-surface solve-post">
      <VoteColumn vote={vote} interactive={interactive} />
      <div className="solve-post-main">
        <div lang={question.lang}>
          <RichText markdown={question.body_md} snippets={question.snippets} />
        </div>
        <VoteError vote={vote} />
        <TagList tags={question.tags} />
      </div>
    </article>
  );
}

export function QuestionPage() {
  const { t } = useI18n();
  const answerCount = useAnswerCount();
  const { questionId } = useParams({ from: '/app/questions/$questionId' });
  const session = useSession();
  const query = useQuestion(questionId);
  const [editing, setEditing] = useState(false);
  const user = session.data?.authenticated ? session.data.user : null;
  const { available } = useInteractions();
  const interactive = user !== null && available;

  if (query.isPending) return <LoadingState />;
  if (query.isError) {
    if (query.error instanceof ApiError && query.error.status === 404) {
      return (
        <EmptyState
          headingLevel={1}
          title={t('solve.question.notFound.title')}
          description={t('solve.question.notFound.body')}
          actions={<Button label={t('solve.question.back')} href="/solve" />}
        />
      );
    }
    return <ErrorState error={query.error} onRetry={() => void query.refetch()} />;
  }

  const { question, etag } = query.data;
  const isAuthor = user !== null && user.id === question.author.id;

  return (
    <div className="solve-page solve-page-wide">
      <RouterAnchor className="solve-back" href="/solve">
        <Icon icon="chevronLeft" size="sm" />
        {t('solve.back')}
      </RouterAnchor>
      {editing ? (
        <EditQuestion
          data={query.data}
          onDone={() => {
            setEditing(false);
          }}
          onReload={() => void query.refetch()}
        />
      ) : (
        <header className="solve-qhead">
          <div className="solve-qhead-title">
            <div lang={question.lang}>
              <Heading level={1}>{question.title}</Heading>
            </div>
            {isAuthor && etag ? (
              <IconButton
                variant="ghost"
                label={t('solve.question.edit')}
                icon={<Icon icon={PencilGlyph} />}
                onClick={() => {
                  setEditing(true);
                }}
              />
            ) : null}
          </div>
          <div className="solve-meta">
            <StatePill state={question.state} />
            <Person author={question.author} />
            <DateText iso={question.created_at} />
            <span className="solve-inline">
              <Icon icon={ChatGlyph} size="sm" />
              {answerCount(question.answer_count)}
            </span>
            <LangMark lang={question.lang} />
          </div>
        </header>
      )}
      <div className="solve-detail">
        <div className="solve-detail-main">
          {editing ? null : (
            <QuestionPost question={question} interactive={interactive && !isAuthor} />
          )}
          <Answers
            questionId={question.id}
            count={question.answer_count}
            interactive={interactive}
            canAccept={interactive && isAuthor}
          />
          {user ? (
            <section className="solve-surface" aria-labelledby="solve-answer-h">
              <Heading level={2} id="solve-answer-h">
                {t('solve.answer.yours')}
              </Heading>
              <AnswerForm questionId={question.id} />
            </section>
          ) : session.isPending ? null : (
            <div className="solve-compose-cta">
              <ChatGlyph width={20} height={20} />
              <p>{t('solve.answer.signInNote')}</p>
              <Button
                variant="primary"
                label={t('solve.answer.signIn')}
                href={signInHref(`/questions/${question.id}`)}
              />
            </div>
          )}
        </div>
        <Aside question={question} />
      </div>
    </div>
  );
}
