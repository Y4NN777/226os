import { Badge } from '@astryxdesign/core/Badge';
import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useParams } from '@tanstack/react-router';
import { useState } from 'react';
import { useSession } from '../../app/session';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { AnswerForm } from './AnswerForm';
import { AcceptButton, VoteControls } from './Interactions';
import { Byline, StateBadge, TagList } from './present';
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
  const { t } = useI18n();
  return (
    <li>
      <Card variant={answer.is_accepted ? 'green' : 'muted'}>
        <VStack gap={3}>
          {answer.is_accepted ? (
            <HStack>
              <Badge
                variant="success"
                icon={<Icon icon="check" size="sm" />}
                label={t('solve.answer.accepted')}
              />
            </HStack>
          ) : null}
          <RichText markdown={answer.body_md} snippets={answer.snippets} />
          <HStack gap={3} wrap="wrap" vAlign="center">
            <Byline name={answer.author.display_name} />
            <VoteControls
              target="answer"
              id={answer.id}
              score={answer.vote_score}
              viewerVote={answer.viewer_vote}
              questionId={answer.question_id}
              interactive={interactive}
            />
          </HStack>
          {canAccept ? (
            <AcceptButton
              questionId={questionId}
              answerId={answer.id}
              isAccepted={answer.is_accepted}
            />
          ) : null}
        </VStack>
      </Card>
    </li>
  );
}

function Answers({
  questionId,
  interactive,
  canAccept,
}: {
  questionId: string;
  interactive: boolean;
  canAccept: boolean;
}) {
  const { t } = useI18n();
  const answers = useAnswers(questionId);
  // The server already sends the accepted answer first; sorting keeps it so across pages.
  const items = (answers.data?.pages.flatMap((page) => page.items) ?? [])
    .map((answer, index) => ({ answer, index }))
    .sort(
      (a, b) => Number(b.answer.is_accepted) - Number(a.answer.is_accepted) || a.index - b.index,
    )
    .map(({ answer }) => answer);
  return (
    <VStack gap={3}>
      <Heading level={2}>{t('solve.answers.title')}</Heading>
      {answers.isPending ? <LoadingState /> : null}
      {answers.isError && !answers.data ? (
        <ErrorState error={answers.error} onRetry={() => void answers.refetch()} />
      ) : null}
      {answers.data && items.length === 0 ? <Text as="p">{t('solve.answers.none')}</Text> : null}
      {items.length > 0 ? (
        <VStack as="ul" gap={3} aria-label={t('solve.answers.title')}>
          {items.map((answer) => (
            <AnswerItem
              key={answer.id}
              answer={answer}
              questionId={questionId}
              interactive={interactive}
              canAccept={canAccept}
            />
          ))}
        </VStack>
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
    </VStack>
  );
}

function Related({ links }: { links: Question['related'] }) {
  const { t } = useI18n();
  if (links.length === 0) return null;
  return (
    <VStack gap={1}>
      <Text type="label" color="secondary">
        {t('solve.question.related')}
      </Text>
      <VStack as="ul" gap={1}>
        {links.map((link) => (
          <li key={`${link.type}-${link.id}`}>
            {link.type === 'resource' ? (
              <Link href={`/discover/${link.id}`}>{link.title}</Link>
            ) : (
              <Text>{link.title}</Text>
            )}
          </li>
        ))}
      </VStack>
    </VStack>
  );
}

export function QuestionPage() {
  const { t } = useI18n();
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
    <VStack gap={6}>
      <VStack gap={3} hAlign="start">
        <Link href="/solve">{t('solve.question.back')}</Link>
        {editing ? (
          <EditQuestion
            data={query.data}
            onDone={() => {
              setEditing(false);
            }}
            onReload={() => void query.refetch()}
          />
        ) : (
          <>
            <HStack gap={2} vAlign="center" wrap="wrap">
              <Heading level={1}>{question.title}</Heading>
              <StateBadge state={question.state} />
              {isAuthor && etag ? (
                <Button
                  variant="ghost"
                  label={t('solve.question.edit')}
                  onClick={() => {
                    setEditing(true);
                  }}
                />
              ) : null}
            </HStack>
            <HStack gap={3} wrap="wrap" vAlign="center">
              <Byline name={question.author.display_name} />
              <VoteControls
                target="question"
                id={question.id}
                score={question.vote_score}
                viewerVote={question.viewer_vote}
                questionId={question.id}
                interactive={interactive && !isAuthor}
              />
            </HStack>
            <TagList tags={question.tags} />
            <RichText markdown={question.body_md} snippets={question.snippets} />
            <Related links={question.related} />
          </>
        )}
      </VStack>
      <Answers
        questionId={question.id}
        interactive={interactive}
        canAccept={interactive && isAuthor}
      />
      {user ? (
        <VStack gap={3}>
          <AnswerForm questionId={question.id} />
        </VStack>
      ) : session.isPending ? null : (
        <HStack>
          <Button label={t('solve.answer.signIn')} href={signInHref(`/questions/${question.id}`)} />
        </HStack>
      )}
    </VStack>
  );
}
