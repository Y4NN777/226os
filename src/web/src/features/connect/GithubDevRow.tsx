import { Link } from '@astryxdesign/core/Link';
import { useState } from 'react';
import { Avatar } from '../../components/Avatar';
import { useI18n } from '../../i18n/I18nProvider';
import { githubProfileUrl } from '../profile/github';
import { inCountry } from './place';
import type { ExternalDeveloper } from './queries';

const AVATAR_HOST = 'https://avatars.githubusercontent.com/';

/** The GitHub avatar from its own host only; the initials when it is elsewhere or fails. */
function ExternalAvatar({ dev }: { dev: ExternalDeveloper }) {
  const [failed, setFailed] = useState(false);
  if (failed || !dev.avatar_url.startsWith(AVATAR_HOST)) {
    return <Avatar name={dev.name} size="md" isDecorative />;
  }
  return (
    <img
      src={dev.avatar_url}
      alt=""
      width={36}
      height={36}
      loading="lazy"
      decoding="async"
      referrerPolicy="no-referrer"
      style={{ borderRadius: '50%', objectFit: 'cover', flex: 'none' }}
      onError={() => {
        setFailed(true);
      }}
    />
  );
}

/**
 * A public GitHub developer ranked by committers.top (ADR-0016), in the row layout of `PersonRow`. It is
 * not an account: the name opens GitHub in a new tab, the label says where the entry comes from, and a
 * link invites the person to join.
 */
export function GithubDevRow({ dev }: { dev: ExternalDeveloper }) {
  const { t, locale } = useI18n();
  const url = githubProfileUrl(dev.login);
  return (
    <li className="ui-row has-lead">
      <div className="ui-row-lead">
        <ExternalAvatar dev={dev} />
      </div>
      <div className="ui-row-main">
        <h3 className="ui-row-title">
          {url ? (
            <Link href={url} isExternalLink newTabLabel={t('discover.newTab')}>
              {dev.name}
            </Link>
          ) : (
            dev.name
          )}
        </h3>
        <p className="ui-row-subtitle">
          {t('connect.external.label', { rank: dev.rank, place: inCountry(dev.country, locale) })}
        </p>
        <p className="ui-row-subtitle">
          <Link href="/inscription">{t('connect.external.join')}</Link>
        </p>
      </div>
    </li>
  );
}
