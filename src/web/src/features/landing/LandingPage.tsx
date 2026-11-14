import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { Grid } from '@astryxdesign/core/Grid';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate } from '@tanstack/react-router';
import { useState, type SubmitEvent } from 'react';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { useLatestResources } from './useLatestResources';

const MAX_QUERY = 100;

interface Pillar {
  id: 'learn' | 'connect' | 'solve' | 'discover' | 'build';
  title: MessageKey;
  text: MessageKey;
  /** Only built pillars link somewhere; the others carry a "Bientôt" marker. */
  live?: { to: string; label: MessageKey };
}

const featured: Pillar[] = [
  { id: 'learn', title: 'landing.pillar.learn.title', text: 'landing.pillar.learn.text' },
  { id: 'connect', title: 'landing.pillar.connect.title', text: 'landing.pillar.connect.text' },
];

const others: Pillar[] = [
  { id: 'solve', title: 'landing.pillar.solve.title', text: 'landing.pillar.solve.text' },
  {
    id: 'discover',
    title: 'landing.pillar.discover.title',
    text: 'landing.pillar.discover.text',
    live: { to: '/discover', label: 'landing.pillar.discover.link' },
  },
  { id: 'build', title: 'landing.pillar.build.title', text: 'landing.pillar.build.text' },
];

function PillarCard({ pillar, prominent }: { pillar: Pillar; prominent: boolean }) {
  const { t } = useI18n();
  return (
    <Card
      variant={prominent ? 'default' : 'muted'}
      padding={prominent ? 6 : 4}
      elevation={prominent ? 'low' : 'none'}
    >
      <VStack gap={3} hAlign="start">
        <HStack gap={2} wrap="wrap" vAlign="center">
          <Heading level={3}>{t(pillar.title)}</Heading>
          {pillar.live ? null : <Badge label={t('landing.soon')} />}
        </HStack>
        <Text as="p" type={prominent ? 'large' : 'body'}>
          {t(pillar.text)}
        </Text>
        {pillar.live ? <Link href={pillar.live.to}>{t(pillar.live.label)}</Link> : null}
      </VStack>
    </Card>
  );
}

function SearchBox() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [value, setValue] = useState('');

  function submit(event: SubmitEvent) {
    event.preventDefault();
    const q = value.trim().slice(0, MAX_QUERY);
    if (!q) return;
    void navigate({ to: '/recherche', search: { q } });
  }

  return (
    <form onSubmit={submit} role="search">
      <HStack gap={2} wrap="wrap" vAlign="end">
        <TextInput
          type="search"
          label={t('landing.search.label')}
          placeholder={t('landing.search.placeholder')}
          value={value}
          onChange={setValue}
          enterKeyHint="search"
          width={360}
        />
        <Button type="submit" label={t('landing.search.submit')} />
      </HStack>
    </form>
  );
}

function Catalogue() {
  const { t } = useI18n();
  const latest = useLatestResources();
  if (latest.isPending) return <LoadingState />;
  if (latest.isError)
    return <ErrorState error={latest.error} onRetry={() => void latest.refetch()} />;
  if (latest.data.items.length === 0) return <Text as="p">{t('landing.catalogue.empty')}</Text>;
  return (
    <Grid columns={{ minWidth: 240, max: 3 }} gap={4}>
      {latest.data.items.map((resource) => (
        <Card key={resource.id} variant="muted" padding={4}>
          <VStack gap={2} hAlign="start">
            <Link href={`/discover/${resource.id}`} weight="semibold">
              {resource.name}
            </Link>
            <Text type="supporting">{resource.provider.name}</Text>
            <Text as="p">{resource.summary}</Text>
          </VStack>
        </Card>
      ))}
    </Grid>
  );
}

/** `/` for visitors: promise, honest pillars, one live catalogue section, sign-up. */
export function LandingPage() {
  const { t } = useI18n();
  return (
    <VStack gap={10} maxWidth={1040}>
      <VStack gap={4} hAlign="start" maxWidth={720} paddingBlock={4}>
        <Heading level={1} type="display-2">
          {t('landing.title')}
        </Heading>
        <Text as="p" type="large">
          {t('landing.lede')}
        </Text>
        <Text as="p" type="label">
          {t('landing.positioning')}
        </Text>
        <HStack gap={3} wrap="wrap">
          <Button variant="primary" size="lg" href="/inscription" label={t('landing.cta.signup')} />
          <Button size="lg" href="/discover" label={t('landing.cta.browse')} />
        </HStack>
        <SearchBox />
      </VStack>

      <VStack as="section" gap={4} aria-labelledby="landing-pillars">
        <Heading level={2} id="landing-pillars">
          {t('landing.pillars.title')}
        </Heading>
        <Grid columns={{ minWidth: 280, max: 2 }} gap={4}>
          {featured.map((pillar) => (
            <PillarCard key={pillar.id} pillar={pillar} prominent />
          ))}
        </Grid>
        <Grid columns={{ minWidth: 240, max: 3 }} gap={4}>
          {others.map((pillar) => (
            <PillarCard key={pillar.id} pillar={pillar} prominent={false} />
          ))}
        </Grid>
      </VStack>

      <VStack as="section" gap={4} aria-labelledby="landing-catalogue">
        <Heading level={2} id="landing-catalogue">
          {t('landing.catalogue.title')}
        </Heading>
        <Catalogue />
      </VStack>

      <Card variant="muted" padding={8}>
        <VStack gap={3} hAlign="start">
          <Heading level={2}>{t('landing.end.title')}</Heading>
          <Text as="p">{t('landing.end.text')}</Text>
          <Button variant="primary" href="/inscription" label={t('landing.cta.signup')} />
        </VStack>
      </Card>
    </VStack>
  );
}
