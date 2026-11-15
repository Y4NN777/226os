import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import { Avatar } from '../../components/Avatar';
import { CountryCode } from '../../components/CountryCode';
import { ListRow } from '../../components/List';
import { Meta, MetaItem } from '../../components/Meta';
import { StatusBadge } from '../../components/StatusBadge';
import { useI18n } from '../../i18n/I18nProvider';
import { categoryLabel } from './categories';
import { formatDate, showsVerifiedDate, statusKey, statusKind, typeKey } from './present';
import type { Resource } from './region';

/** The dot between the parts of the provider line; decorative, the parts are separate words. */
function Dot() {
  return <span aria-hidden="true">{' · '}</span>;
}

/**
 * One catalogue row (prototype `li.row.discover-row`) on the shared ListRow: monogram tile, title
 * link, provider line, one-line summary, country codes and source count, then the status pill and
 * its date.
 */
export function ResourceRow({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const verifiedAt =
    resource.last_verified_at && showsVerifiedDate(resource.verification_status)
      ? resource.last_verified_at
      : null;
  return (
    <ListRow
      title={resource.name}
      href={`/discover/${resource.id}`}
      lead={<Avatar name={resource.name} variant="tile" />}
      subtitle={
        <>
          <span>{resource.provider.name}</span>
          <Dot />
          <span>{categoryLabel(resource.category, t)}</span>
          <Dot />
          <span>{t(typeKey[resource.type])}</span>
        </>
      }
      excerpt={resource.summary}
      meta={
        <Meta>
          {resource.countries.length > 0 ? (
            <MetaItem>
              <VisuallyHidden>{t('discover.countries.label')}</VisuallyHidden>
              <span className="ui-countries">
                {resource.countries.map((code) => (
                  <CountryCode key={code} code={code} />
                ))}
              </span>
            </MetaItem>
          ) : null}
          <MetaItem>
            {resource.source_count === 1
              ? t('discover.sources.one')
              : t('discover.sources.count', { n: resource.source_count })}
          </MetaItem>
        </Meta>
      }
      aside={
        <>
          <StatusBadge
            kind={statusKind[resource.verification_status]}
            label={t(statusKey[resource.verification_status])}
          />
          {resource.provenance === 'demo' ? <StatusBadge kind="demo" /> : null}
          {verifiedAt ? (
            <time dateTime={verifiedAt}>
              {t('discover.verifiedOn', { date: formatDate(verifiedAt, locale) })}
            </time>
          ) : null}
        </>
      }
    />
  );
}
