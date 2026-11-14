import { Button } from '@astryxdesign/core/Button';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Text } from '@astryxdesign/core/Text';
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
  type VoteState,
  type VoteTarget,
} from './queries';

/**
 * Score with up and down buttons (when votes are live and the visitor is a member). The score
 * after a vote is the one the server returned; the buttons only send intent.
 */
export function VoteControls({
  target,
  id,
  score,
  interactive,
}: {
  target: VoteTarget;
  id: string;
  score: number;
  interactive: boolean;
}) {
  const { t } = useI18n();
  const cast = useCastVote();
  const retract = useRetractVote();
  const { absorbNotLive } = useInteractions();
  const [state, setState] = useState<VoteState | null>(null);
  const [error, setError] = useState<unknown>(null);
  const value = state?.value ?? 0;

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

  const pending = cast.isPending || retract.isPending;
  return (
    <VStack gap={1}>
      <HStack gap={1} vAlign="center">
        {interactive ? (
          <IconButton
            size="sm"
            variant="ghost"
            label={t(`solve.vote.up.${target}`)}
            icon={<Icon icon="arrowUp" />}
            isDisabled={pending}
            aria-pressed={value === 1}
            onClick={() => void press(1)}
          />
        ) : null}
        <Text type="supporting">{t('solve.score', { n: state?.score ?? score })}</Text>
        {interactive ? (
          <IconButton
            size="sm"
            variant="ghost"
            label={t(`solve.vote.down.${target}`)}
            icon={<Icon icon="arrowDown" />}
            isDisabled={pending}
            aria-pressed={value === -1}
            onClick={() => void press(-1)}
          />
        ) : null}
      </HStack>
      {error ? (
        error instanceof ApiError && error.code === 'SELF_VOTE_FORBIDDEN' ? (
          <Text role="alert" type="supporting">
            {t('solve.err.selfVote')}
          </Text>
        ) : (
          <ErrorState error={error} />
        )
      ) : null}
    </VStack>
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
