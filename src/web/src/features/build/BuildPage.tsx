import { Button } from '@astryxdesign/core/Button';
import { Badge } from '@astryxdesign/core/Badge';
import { Card } from '@astryxdesign/core/Card';
import { EmptyState } from '@astryxdesign/core/EmptyState';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { signInHref } from '../solve/signIn';
import { OwnerByline, Technologies } from './present';
import { ProjectForm } from './ProjectForm';
import { useCreateProject, useProjectList, type Project } from './queries';

export function needsLabelKey(count: number) {
  if (count === 0) return 'build.needs.none' as const;
  return count === 1 ? ('build.needs.one' as const) : ('build.needs.many' as const);
}

function ProjectCard({ project }: { project: Project }) {
  const { t } = useI18n();
  return (
    <Card variant="muted">
      <VStack gap={2}>
        <HStack gap={2} wrap="wrap" vAlign="center">
          <Link href={`/build/${project.id}`} type="large" weight="semibold">
            {project.name}
          </Link>
          {project.visibility === 'private' ? (
            <Badge variant="neutral" label={t('build.private')} />
          ) : null}
        </HStack>
        {project.description ? <Text as="p">{project.description}</Text> : null}
        <Technologies items={project.technologies} />
        <HStack gap={3} wrap="wrap" vAlign="center">
          <OwnerByline username={project.owner.username} name={project.owner.display_name} />
          <Text type="supporting">
            {t(needsLabelKey(project.open_needs_count), { n: project.open_needs_count })}
          </Text>
        </HStack>
      </VStack>
    </Card>
  );
}

function ProposeForm({ onClose }: { onClose: () => void }) {
  const navigate = useNavigate();
  const create = useCreateProject();
  return (
    <ProjectForm
      onCancel={onClose}
      onSubmit={async (body) => {
        const created = await create.mutateAsync(body);
        await navigate({ to: '/build/$projectId', params: { projectId: created.id } });
      }}
    />
  );
}

export function BuildPage() {
  const { t } = useI18n();
  const session = useSession();
  const [proposing, setProposing] = useState(false);
  const list = useProjectList();
  const items = list.data?.pages.flatMap((page) => page.items) ?? [];
  const signedIn = session.data?.authenticated === true;

  return (
    <VStack gap={5}>
      <VStack gap={1}>
        <Heading level={1}>{t('build.title')}</Heading>
        <Text as="p">{t('build.lede')}</Text>
      </VStack>
      {proposing && signedIn ? (
        <ProposeForm
          onClose={() => {
            setProposing(false);
          }}
        />
      ) : (
        <HStack>
          {signedIn ? (
            <Button
              variant="primary"
              label={t('build.propose')}
              onClick={() => {
                setProposing(true);
              }}
            />
          ) : session.isPending ? null : (
            <Button
              variant="primary"
              label={t('build.propose.signIn')}
              href={signInHref('/build')}
            />
          )}
        </HStack>
      )}
      <section aria-label={t('build.results.label')}>
        {list.isPending ? <LoadingState /> : null}
        {list.isError && !list.data ? (
          <ErrorState error={list.error} onRetry={() => void list.refetch()} />
        ) : null}
        {list.data && items.length === 0 ? (
          <EmptyState
            headingLevel={2}
            title={t('build.empty.title')}
            description={t('build.empty.body')}
          />
        ) : null}
        {items.length > 0 ? (
          <VStack gap={3}>
            <VStack as="ul" gap={3}>
              {items.map((project) => (
                <li key={project.id}>
                  <ProjectCard project={project} />
                </li>
              ))}
            </VStack>
            {list.isError ? (
              <ErrorState error={list.error} onRetry={() => void list.fetchNextPage()} />
            ) : null}
            {list.hasNextPage ? (
              <HStack hAlign="center">
                <Button
                  label={t('build.more')}
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
