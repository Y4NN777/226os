import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { Card } from '@astryxdesign/core/Card';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Text } from '@astryxdesign/core/Text';
import { VStack } from '@astryxdesign/core/VStack';
import { ApiError } from '../../api/http';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FocusRegion } from '../identity/notices';
import { Banner } from '@astryxdesign/core/Banner';
import { PersonLink } from './present';
import { useApplications, useDecideApplication, type Application } from './queries';

const statusKey = {
  pending: 'project.apps.status.pending',
  accepted: 'project.apps.status.accepted',
  declined: 'project.apps.status.declined',
  withdrawn: 'project.apps.status.withdrawn',
} as const satisfies Record<Application['status'], MessageKey>;

const statusVariant = {
  pending: 'neutral',
  accepted: 'success',
  declined: 'neutral',
  withdrawn: 'neutral',
} as const;

/**
 * The owner's inbox. It is only mounted for the owner, so a third party never calls
 * listProjectApplications and no count or row can leak to them.
 */
export function Applications({ projectId }: { projectId: string }) {
  const { t } = useI18n();
  const applications = useApplications(projectId);
  const decide = useDecideApplication(projectId);
  const items = applications.data?.pages.flatMap((page) => page.items) ?? [];
  const alreadyDecided = decide.error instanceof ApiError && decide.error.code === 'STATE_CONFLICT';

  return (
    <section aria-label={t('project.apps.title')}>
      <VStack gap={3}>
        <Heading level={2}>{t('project.apps.title')}</Heading>
        {decide.isError ? (
          alreadyDecided ? (
            <FocusRegion role="alert">
              <Banner status="warning" title={t('project.err.decided')} collapsible={false} />
            </FocusRegion>
          ) : (
            <ErrorState error={decide.error} />
          )
        ) : null}
        {applications.isPending ? <LoadingState /> : null}
        {applications.isError && !applications.data ? (
          <ErrorState error={applications.error} onRetry={() => void applications.refetch()} />
        ) : null}
        {applications.data && items.length === 0 ? (
          <Text as="p">{t('project.apps.none')}</Text>
        ) : null}
        {items.length > 0 ? (
          <VStack as="ul" gap={3}>
            {items.map((application) => (
              <li key={application.id}>
                <Card variant="muted">
                  <VStack gap={2}>
                    <HStack gap={2} wrap="wrap" vAlign="center">
                      <PersonLink
                        username={application.applicant.username}
                        name={application.applicant.display_name}
                      />
                      <Badge
                        variant={statusVariant[application.status]}
                        label={t(statusKey[application.status])}
                      />
                      {application.status === 'pending' ? (
                        <HStack gap={1}>
                          <IconButton
                            size="sm"
                            variant="ghost"
                            label={t('project.apps.accept', {
                              name: application.applicant.display_name,
                            })}
                            icon={<Icon icon="check" />}
                            isDisabled={decide.isPending}
                            onClick={() => {
                              decide.mutate({ id: application.id, status: 'accepted' });
                            }}
                          />
                          <IconButton
                            size="sm"
                            variant="ghost"
                            label={t('project.apps.decline', {
                              name: application.applicant.display_name,
                            })}
                            icon={<Icon icon="close" />}
                            isDisabled={decide.isPending}
                            onClick={() => {
                              decide.mutate({ id: application.id, status: 'declined' });
                            }}
                          />
                        </HStack>
                      ) : null}
                    </HStack>
                    <Text as="p">{application.message}</Text>
                  </VStack>
                </Card>
              </li>
            ))}
          </VStack>
        ) : null}
        {applications.hasNextPage ? (
          <HStack hAlign="center">
            <Button
              label={t('build.more')}
              isLoading={applications.isFetchingNextPage}
              onClick={() => void applications.fetchNextPage()}
            />
          </HStack>
        ) : null}
      </VStack>
    </section>
  );
}
