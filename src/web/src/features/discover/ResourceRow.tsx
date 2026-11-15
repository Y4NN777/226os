import { Link } from '@astryxdesign/core/Link';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { DemoBadge } from '../../components/DemoBadge';
import { useI18n } from '../../i18n/I18nProvider';
import { categoryLabel } from './categories';
import { CountryCodes } from './CountryCodes';
import { formatDate, showsVerifiedDate, typeKey } from './present';
import type { Resource } from './region';
import { ResourceThumb } from './ResourceThumb';
import { VerificationBadge } from './VerificationBadge';

/** The dot between the parts of a meta line; decorative, the parts are separate words already. */
function Dot() {
  return (
    <span className="discover-dot" aria-hidden="true">
      {' · '}
    </span>
  );
}

/**
 * One catalogue row (prototype `li.row.discover-row`): monogram, title link, provider line,
 * one-line summary, country codes and source count, then the status pill and its date.
 */
export function ResourceRow({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const verifiedAt =
    resource.last_verified_at && showsVerifiedDate(resource.verification_status)
      ? resource.last_verified_at
      : null;
  return (
    <li className="discover-row">
      <ResourceThumb name={resource.name} />
      <div className="discover-row-main">
        <h3 className="discover-row-title">
          <Link href={`/discover/${resource.id}`} weight="semibold">
            {resource.name}
          </Link>
          {resource.provenance === 'demo' ? <DemoBadge /> : null}
        </h3>
        <p className="discover-sub">
          <span>{resource.provider.name}</span>
          <Dot />
          <span>{categoryLabel(resource.category, t)}</span>
          <Dot />
          <span>{t(typeKey[resource.type])}</span>
        </p>
        <p className="discover-excerpt">{resource.summary}</p>
        <p className="discover-meta">
          {resource.countries.length > 0 ? (
            <span className="discover-meta-item">
              <VisuallyHidden>{t('discover.countries.label')}</VisuallyHidden>
              <CountryCodes codes={resource.countries} />
            </span>
          ) : null}
          <span className="discover-meta-item">
            {resource.source_count === 1
              ? t('discover.sources.one')
              : t('discover.sources.count', { n: resource.source_count })}
          </span>
        </p>
      </div>
      <div className="discover-row-aside">
        <VerificationBadge status={resource.verification_status} />
        {verifiedAt ? (
          <time className="discover-when" dateTime={verifiedAt}>
            {t('discover.verifiedOn', { date: formatDate(verifiedAt, locale) })}
          </time>
        ) : null}
      </div>
    </li>
  );
}
