import { Link } from '@astryxdesign/core/Link';
import { useSession } from '../../app/session';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import { DisplayNameForm } from './DisplayNameForm';
import { useMe } from './useMe';

/** Loads the account name, then hands it to the form as its default. */
export function DisplayName() {
  const { t } = useI18n();
  const session = useSession();
  const me = useMe();

  if (session.isPending) return null;
  if (session.data?.authenticated !== true) {
    return (
      <Link href="/connexion?next=%2Fparametres" isStandalone>
        {t('settings.name.signIn')}
      </Link>
    );
  }
  if (me.isPending) return <LoadingState />;
  if (me.isError) return <ErrorState error={me.error} onRetry={() => void me.refetch()} />;
  return <DisplayNameForm initialName={me.data.profile.display_name} />;
}
