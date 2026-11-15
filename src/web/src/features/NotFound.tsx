import { Link } from '@tanstack/react-router';
import { EmptyState } from '../components/EmptyState';
import { CompassIcon } from '../components/icons';
import { useI18n } from '../i18n/I18nProvider';

export function NotFound() {
  const { t } = useI18n();
  return (
    // No route matched, so no layout route rendered: this page brings its own frame.
    <main className="auth-layout">
      <div className="not-found">
        <EmptyState
          headingLevel={1}
          icon={CompassIcon}
          title={t('notFound.title')}
          description={t('notFound.body')}
          actions={
            <Link to="/" className="app-nav-link">
              {t('notFound.back')}
            </Link>
          }
        />
      </div>
    </main>
  );
}
