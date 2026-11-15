import { Button } from '@astryxdesign/core/Button';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Text } from '@astryxdesign/core/Text';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { VStack } from '@astryxdesign/core/VStack';
import { useState } from 'react';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { useI18n } from '../../i18n/I18nProvider';
import {
  useAccept,
  useCastVote,
  useInteractions,
  useRetractVote,
  type ViewerVote,
  type VoteState,
  type VoteTarget,
} from './queries';

export interface VoteSubject {
  target: VoteTarget;
  id: string;
  score: number;
  /** The caller's vote as the server last returned it (0 when none or anonymous). */
  viewerVote: ViewerVote;
  /** The question whose cached detail and answers are refreshed after a vote. */
  questionId: string;
}

/**
 * The vote on one question or answer. The pressed button starts from the viewer's vote in the
 * contract; the score after a vote is the one the server returned, and the cached question and
 * answers are refreshed. The buttons only send intent.
 */
export function useVote({ target, id, score, viewerVote, questionId }: VoteSubject) {
  const cast = useCastVote(questionId);
  const retract = useRetractVote(questionId);
  const { absorbNotLive } = useInteractions();
  const [state, setState] = useState<VoteState | null>(null);
  const [error, setError] = useState<unknown>(null);
  const value = state?.value ?? viewerVote;

  const press = async (next: 1 | -1) => {
    setError(null);
    try {
      setState(
        next === value
          ? await retract.mutateAsync({ target_type: target, target_id: id })
          : await cast.mutateAsync({ target_type: target, target_id: id, value: next }),
      );
    } catch (failure) {
      if (!absorbNotLive(failure)) setError(failure);
    }
  };

  return {
    target,
    value,
    score: state?.score ?? score,
    pending: cast.isPending || retract.isPending,
    error,
    press,
  };
}

export type Vote = ReturnType<typeof useVote>;

/**
 * The prototype's vote column: up, score, down (buttons only when votes are live and the visitor
 * is a member). The score is shown as a number and read out as "Score : n".
 */
export function VoteColumn({ vote, interactive }: { vote: Vote; interactive: boolean }) {
  const { t } = useI18n();
  return (
    <div className="solve-vote">
      {interactive ? (
        <IconButton
          size="sm"
          variant="ghost"
          label={t(`solve.vote.up.${vote.target}`)}
          icon={<Icon icon="arrowUp" />}
          isDisabled={vote.pending}
          aria-pressed={vote.value === 1}
          onClick={() => void vote.press(1)}
        />
      ) : (
        <Icon icon="arrowUp" size="sm" color="secondary" />
      )}
      <span className="solve-vote-score">
        <span aria-hidden="true">{vote.score}</span>
        <VisuallyHidden>{t('solve.score', { n: vote.score })}</VisuallyHidden>
      </span>
      {interactive ? (
        <IconButton
          size="sm"
          variant="ghost"
          label={t(`solve.vote.down.${vote.target}`)}
          icon={<Icon icon="arrowDown" />}
          isDisabled={vote.pending}
          aria-pressed={vote.value === -1}
          onClick={() => void vote.press(-1)}
        />
      ) : null}
    </div>
  );
}

/** A refused vote, said next to the content (the column is too narrow for a sentence). */
export function VoteError({ vote }: { vote: Vote }) {
  const { t } = useI18n();
  if (!vote.error) return null;
  return vote.error instanceof ApiError && vote.error.code === 'SELF_VOTE_FORBIDDEN' ? (
    <Text role="alert" type="supporting">
      {t('solve.err.selfVote')}
    </Text>
  ) : (
    <ErrorState error={vote.error} />
  );
}

/** Accept or withdraw the acceptance of one answer; only shown to the question author. */
export function AcceptButton({
  questionId,
  answerId,
  isAccepted,
}: {
  questionId: string;
  answerId: string;
  isAccepted: boolean;
}) {
  const { t } = useI18n();
  const accept = useAccept(questionId);
  const { absorbNotLive } = useInteractions();
  return (
    <VStack gap={1} hAlign="start">
      <Button
        size="sm"
        variant={isAccepted ? 'ghost' : 'secondary'}
        label={t(isAccepted ? 'solve.unaccept' : 'solve.accept')}
        isLoading={accept.isPending}
        onClick={() => {
          accept.mutate(isAccepted ? null : answerId, { onError: absorbNotLive });
        }}
      />
      {accept.isError && !(accept.error instanceof ApiError && accept.error.status === 501) ? (
        <ErrorState error={accept.error} />
      ) : null}
    </VStack>
  );
}
