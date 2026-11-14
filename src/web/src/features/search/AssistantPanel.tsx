import { Badge } from '@astryxdesign/core/Badge';
import { Collapsible } from '@astryxdesign/core/Collapsible';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { DemoBadge } from '../../components/DemoBadge';
import { useI18n } from '../../i18n/I18nProvider';
import { assistantFixture } from './assistantFixture';

/**
 * Prototype of the assistant (S9 is not live): a labelled demo with one typed sample
 * answer, collapsed by default so it never competes with real results. No vote, no accept.
 */
export function AssistantPanel() {
  const { t, locale } = useI18n();
  const answer = assistantFixture;
  const lang = answer.lang === locale ? undefined : answer.lang;
  return (
    <section aria-labelledby="assistant-title">
      <VStack gap={3} hAlign="start">
        <HStack gap={2} vAlign="center" wrap="wrap">
          <Heading level={2} id="assistant-title">
            {t('search.assistant.title')}
          </Heading>
          <Badge variant="neutral" label={t('search.assistant.ai')} />
          <DemoBadge />
        </HStack>
        <Text as="p" type="supporting">
          {t('search.assistant.notice')}
        </Text>
        <Collapsible defaultIsOpen={false} trigger={t('search.assistant.example')}>
          <div lang={lang}>
            <VStack gap={3} hAlign="start">
              <Text as="p">{answer.answer}</Text>
              <Heading level={3}>{t('search.assistant.verify')}</Heading>
              <Text as="p">{answer.how_to_verify}</Text>
              <Heading level={3}>{t('search.assistant.sources')}</Heading>
              <VStack as="ul" gap={1} hAlign="start">
                {answer.sources.map((source) => (
                  <li key={source.id}>
                    <Link href={source.url}>{source.title}</Link>
                  </li>
                ))}
              </VStack>
            </VStack>
          </div>
        </Collapsible>
      </VStack>
    </section>
  );
}
