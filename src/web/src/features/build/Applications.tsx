import { Avatar } from '@astryxdesign/core/Avatar';
import { Button } from '@astryxdesign/core/Button';
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
import { DateText, PersonLink, StatePill, type PillTone } from './present';
import { useApplications, useDecideApplication, type Application } from './queries';

const statusKey = {
  pending: 'project.apps.status.pending',
  accepted: 'project.apps.status.accepted',
  declined: 'project.apps.status.declined',
  withdrawn: 'project.apps.status.withdrawn',
} as const satisfies Record<Application['status'], MessageKey>;

const statusTone: Record<Application['status'], PillTone> = {
  pending: 'pending',
  accepted: 'open',
  declined: 'muted',
  withdrawn: 'muted',
};

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
          <ul className="build-apps">
            {items.map((application) => (
              <li key={application.id} className="build-app">
                <span aria-hidden="true">
                  <Avatar name={application.applicant.display_name} size="md" />
                </span>
                <div className="build-app-main">
                  <PersonLink
                    username={application.applicant.username}
                    name={application.applicant.display_name}
                  />
                  <p className="build-app-quote">{application.message}</p>
                  <div className="build-meta">
                    <DateText iso={application.created_at} />
                  </div>
                </div>
                <div className="build-app-aside">
                  <StatePill
                    tone={statusTone[application.status]}
                    label={t(statusKey[application.status])}
                  />
                  {application.status === 'pending' ? (
                    <div className="build-app-actions">
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
                    </div>
                  ) : null}
                </div>
              </li>
            ))}
          </ul>
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
