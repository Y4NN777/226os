import { Link } from '@astryxdesign/core/Link';
import { CountryCode } from '../../components/CountryCode';
import { GitBranchIcon, LanguagesIcon, MapPinIcon } from '../../components/icons';
import { ListRow } from '../../components/List';
import { Meta, MetaItem } from '../../components/Meta';
import { ProfilePhoto } from '../../components/ProfilePhoto';
import { TagList } from '../../components/Tag';
import { useI18n } from '../../i18n/I18nProvider';
import { languageName } from '../discover/present';
import { githubProfileUrl } from '../profile/github';
import type { ProfileSummary } from './queries';

/**
 * A directory row (prototype `list > row`) on the shared ListRow: avatar, name linking to the
 * profile, handle, a meta line with what the caller may see, skills, and the GitHub link at the
 * end. Fields the server hid are absent, so nothing is drawn for them.
 */
export function PersonRow({ person }: { person: ProfileSummary }) {
  const { t, locale } = useI18n();
  const githubUrl = person.github ? githubProfileUrl(person.github) : null;
  const languages =
    person.languages && person.languages.length > 0
      ? person.languages.map((code) => languageName(code, locale)).join(', ')
      : null;
  const hasMeta = Boolean(person.country) || languages !== null;
  return (
    <ListRow
      title={person.display_name}
      href={`/profils/${person.username}`}
      subtitle={`@${person.username}`}
      lead={<ProfilePhoto name={person.display_name} github={person.github} />}
      {...(hasMeta
        ? {
            meta: (
              <Meta>
                {person.country ? (
                  <MetaItem icon={MapPinIcon}>
                    <CountryCode code={person.country} withName />
                  </MetaItem>
                ) : null}
                {languages ? <MetaItem icon={LanguagesIcon}>{languages}</MetaItem> : null}
              </Meta>
            ),
          }
        : {})}
      {...(person.skills && person.skills.length > 0
        ? { tags: <TagList tags={person.skills} /> }
        : {})}
      {...(githubUrl
        ? {
            aside: (
              <Link
                href={githubUrl}
                aria-label={t('connect.github', { name: person.display_name })}
                isStandalone
              >
                <MetaItem icon={GitBranchIcon}>GitHub</MetaItem>
              </Link>
            ),
          }
        : {})}
    />
  );
}
