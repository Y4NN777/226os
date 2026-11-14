import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextArea } from '@astryxdesign/core/TextArea';
import { VStack } from '@astryxdesign/core/VStack';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { isSafeInternalPath } from '../../components/safePath';
import { useI18n } from '../../i18n/I18nProvider';
import { signInHref } from '../solve/signIn';

type AssistantAnswer = components['schemas']['AssistantAnswer'];
type AssistantSource = components['schemas']['AssistantSource'];

const MAX_QUESTION = 1000;
const unavailableKey = ['assistant', 'unavailable'] as const;

/** True once the assistant answered 503 in this session: the panel then stays hidden. */
function useUnavailable() {
  const queryClient = useQueryClient();
  const flag = useQuery({
    queryKey: unavailableKey,
    queryFn: () => false,
    staleTime: Infinity,
    gcTime: Infinity,
    enabled: false,
    initialData: false,
  });
  return [
    flag.data,
    () => {
      queryClient.setQueryData(unavailableKey, true);
    },
  ] as const;
}

function useAsk(onUnavailable: () => void) {
  return useMutation({
    mutationFn: (body: components['schemas']['AssistantQuestion']) =>
      api.call(() => api.client.POST('/assistant/answers', { body })),
    retry: false,
    onError: (error) => {
      if (error instanceof ApiError && error.status === 503) onUnavailable();
    },
  });
}

function SourceLink({ source }: { source: AssistantSource }) {
  if (isSafeInternalPath(source.url)) return <Link href={source.url}>{source.title}</Link>;
  if (/^https?:\/\//i.test(source.url)) {
    return (
      <Link href={source.url} isExternalLink>
        {source.title}
      </Link>
    );
  }
  return <Text>{source.title}</Text>;
}

function Answer({ answer }: { answer: AssistantAnswer }) {
  const { t, locale } = useI18n();
  const lang = answer.lang === locale ? undefined : answer.lang;
  return (
    <div lang={lang}>
      <VStack gap={3} hAlign="start">
        <Text as="p" type="supporting">
          {t('search.assistant.generated')}
        </Text>
        <Text as="p" style={{ whiteSpace: 'pre-line' }}>
          {answer.answer}
        </Text>
        <Heading level={3}>{t('search.assistant.verify')}</Heading>
        <Text as="p" style={{ whiteSpace: 'pre-line' }}>
          {answer.how_to_verify}
        </Text>
        {answer.sources.length > 0 ? (
          <>
            <Heading level={3}>{t('search.assistant.sources')}</Heading>
            <VStack as="ul" gap={1} hAlign="start">
              {answer.sources.map((source) => (
                <li key={source.id}>
                  <SourceLink source={source} />
                </li>
              ))}
            </VStack>
          </>
        ) : null}
      </VStack>
    </div>
  );
}

function Refusal({ answer, q }: { answer: AssistantAnswer; q: string }) {
  const { t } = useI18n();
  const href =
    answer.search_url && isSafeInternalPath(answer.search_url)
      ? answer.search_url
      : `/recherche?q=${encodeURIComponent(q)}`;
  return (
    <VStack gap={2} hAlign="start">
      <Heading level={3}>{t('search.assistant.nosource.title')}</Heading>
      <Text as="p">{t('search.assistant.nosource.body')}</Text>
      <Link href={href}>{t('search.assistant.nosource.link')}</Link>
    </VStack>
  );
}

function AskForm({ q }: { q: string }) {
  const { t, locale } = useI18n();
  const [free, setFree] = useState('');
  const [, markUnavailable] = useUnavailable();
  const ask = useAsk(markUnavailable);
  const unavailable = ask.error instanceof ApiError && ask.error.status === 503;
  const submit = () => {
    const question = free.trim() || q;
    ask.mutate({ question, lang: locale });
  };

  if (unavailable) return <Text as="p">{t('search.assistant.unavailable')}</Text>;
  return (
    <VStack gap={3} hAlign="start">
      <form
        onSubmit={(event) => {
          event.preventDefault();
          if (!ask.isPending) submit();
        }}
      >
        <VStack gap={3} hAlign="start">
          <TextArea
            label={t('search.assistant.field')}
            description={t('search.assistant.hint', { q })}
            rows={2}
            maxLength={MAX_QUESTION}
            value={free}
            onChange={setFree}
          />
          <Button
            type="submit"
            variant="primary"
            label={t('search.assistant.ask')}
            isLoading={ask.isPending}
          />
        </VStack>
      </form>
      {ask.isPending ? (
        <Text as="p" type="supporting" role="status">
          {t('search.assistant.loading')}
        </Text>
      ) : null}
      {ask.isError ? <ErrorState error={ask.error} /> : null}
      {ask.data?.outcome === 'answered' ? <Answer answer={ask.data} /> : null}
      {ask.data?.outcome === 'no_source' ? <Refusal answer={ask.data} q={q} /> : null}
    </VStack>
  );
}

/**
 * S9 assistant panel: members only, one POST /assistant/answers per click, answer shown
 * as plain text with its sources. No vote, accept or solved control (AI-9).
 */
export function AssistantPanel({ q }: { q: string }) {
  const { t } = useI18n();
  const session = useSession();
  const [unavailable] = useUnavailable();
  const signedIn = session.data?.authenticated === true;
  const [seen] = useState(unavailable);
  if (seen || session.isPending) return null;

  return (
    <section aria-labelledby="assistant-title">
      <VStack gap={3} hAlign="start">
        <HStack gap={2} vAlign="center" wrap="wrap">
          <Heading level={2} id="assistant-title">
            {t('search.assistant.title')}
          </Heading>
          <Badge variant="neutral" label={t('search.assistant.ai')} />
        </HStack>
        {signedIn ? (
          <>
            <Text as="p" type="supporting">
              {t('search.assistant.notice')}
            </Text>
            <AskForm q={q} />
          </>
        ) : (
          <>
            <Text as="p">{t('search.assistant.signIn')}</Text>
            <Button
              label={t('search.assistant.signIn.action')}
              href={signInHref(`/recherche?q=${encodeURIComponent(q)}`)}
            />
          </>
        )}
      </VStack>
    </section>
  );
}
