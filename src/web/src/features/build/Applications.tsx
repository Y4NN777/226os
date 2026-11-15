import { Banner } from '@astryxdesign/core/Banner';
import { Button } from '@astryxdesign/core/Button';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { Text } from '@astryxdesign/core/Text';
import { ApiError } from '../../api/http';
import { Avatar } from '../../components/Avatar';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { Meta } from '../../components/Meta';
import { Section } from '../../components/Section';
import { StatusBadge, type StatusKind } from '../../components/StatusBadge';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FocusRegion } from '../identity/notices';
import { DateText, PersonLink } from './present';
import { useApplications, useDecideApplication, type Application } from './queries';

const statusKey = {
  pending: 'project.apps.status.pending',
  accepted: 'project.apps.status.accepted',
  declined: 'project.apps.status.declined',
  withdrawn: 'project.apps.status.withdrawn',
} as const satisfies Record<Application['status'], MessageKey>;

const statusKind: Record<Application['status'], StatusKind> = {
  pending: 'pending',
  accepted: 'accepted',
  declined: 'closed',
  withdrawn: 'closed',
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
    <Section title={t('project.apps.title')}>
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
              <Avatar name={application.applicant.display_name} />
              <div className="build-app-main">
                <PersonLink
                  username={application.applicant.username}
                  name={application.applicant.display_name}
                />
                <p className="build-app-quote">{application.message}</p>
                <Meta>
                  <DateText iso={application.created_at} />
                </Meta>
              </div>
              <div className="build-app-aside">
                <StatusBadge
                  kind={statusKind[application.status]}
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
    </Section>
  );
}
