import { Avatar } from '@astryxdesign/core/Avatar';
import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { TextInput } from '@astryxdesign/core/TextInput';
import { useNavigate } from '@tanstack/react-router';
import { useState, type SubmitEvent } from 'react';
import type { components } from '../../api/schema';
import { DemoBadge } from '../../components/DemoBadge';
import { ErrorState } from '../../components/ErrorState';
import { LoadingState } from '../../components/LoadingState';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FeedIcon, type FeedIconName } from '../home/FeedIcon';
import { Country, formatDate, StatePill, type PillKind } from '../home/rows';
import { useLatestResources } from './useLatestResources';
import './landing.css';

const MAX_QUERY = 100;

type Resource = components['schemas']['Resource'];

interface Pillar {
  id: 'learn' | 'connect' | 'solve' | 'discover' | 'build';
  icon: FeedIconName;
  title: MessageKey;
  text: MessageKey;
  link: { to: string; label: MessageKey };
}

// Learn and Connect lead (product decision of 2026-11-12); every other pillar stays present.
const pillars: Pillar[] = [
  {
    id: 'learn',
    icon: 'book',
    title: 'landing.pillar.learn.title',
    text: 'landing.pillar.learn.text',
    link: { to: '/apprendre', label: 'landing.pillar.learn.link' },
  },
  {
    id: 'connect',
    icon: 'users',
    title: 'landing.pillar.connect.title',
    text: 'landing.pillar.connect.text',
    link: { to: '/connect', label: 'landing.pillar.connect.link' },
  },
  {
    id: 'solve',
    icon: 'help',
    title: 'landing.pillar.solve.title',
    text: 'landing.pillar.solve.text',
    link: { to: '/solve', label: 'landing.pillar.solve.link' },
  },
  {
    id: 'discover',
    icon: 'compass',
    title: 'landing.pillar.discover.title',
    text: 'landing.pillar.discover.text',
    link: { to: '/discover', label: 'landing.pillar.discover.link' },
  },
  {
    id: 'build',
    icon: 'hammer',
    title: 'landing.pillar.build.title',
    text: 'landing.pillar.build.text',
    link: { to: '/build', label: 'landing.pillar.build.link' },
  },
];

/** A text link followed by an arrow, the prototype's `.lp-link`. */
function ArrowLink({ href, label }: { href: string; label: string }) {
  return (
    <span className="lp-link">
      <Link href={href} weight="semibold">
        {label}
      </Link>
      <FeedIcon name="arrowRight" />
    </span>
  );
}

function PillarCard({ pillar }: { pillar: Pillar }) {
  const { t } = useI18n();
  const lead = pillar.id === 'learn' || pillar.id === 'connect';
  return (
    <li className={lead ? 'lp-pillar lp-pillar-lead' : 'lp-pillar'}>
      <span className="lp-pillar-icon" aria-hidden="true">
        <FeedIcon name={pillar.icon} />
      </span>
      <Heading level={3}>{t(pillar.title)}</Heading>
      <p>{t(pillar.text)}</p>
      <ArrowLink href={pillar.link.to} label={t(pillar.link.label)} />
    </li>
  );
}

function SearchBox() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [value, setValue] = useState('');

  function submit(event: SubmitEvent) {
    event.preventDefault();
    const q = value.trim().slice(0, MAX_QUERY);
    if (!q) return;
    void navigate({ to: '/recherche', search: { q } });
  }

  return (
    <form onSubmit={submit} role="search" className="lp-search">
      <span className="lp-search-field">
        <TextInput
          type="search"
          label={t('landing.search.label')}
          isLabelHidden
          startIcon={<Icon icon="search" />}
          placeholder={t('landing.search.placeholder')}
          value={value}
          onChange={setValue}
          enterKeyHint="search"
          width="100%"
        />
      </span>
      <Button type="submit" label={t('landing.search.submit')} />
    </form>
  );
}

const verificationPill: Record<Resource['verification_status'], PillKind> = {
  verified: 'verified',
  claimed: 'unverified',
  unverified: 'unverified',
  stale: 'stale',
  disputed: 'stale',
};

/** The prototype's catalogue card: thumb, name, provider, state, summary, checked date, countries. */
function ResourceCard({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const status = resource.verification_status;
  const checked = status === 'verified' ? resource.last_verified_at : null;
  return (
    <li className="lp-card">
      <div className="lp-card-head">
        <Avatar name={resource.name} shape="rounded" size={48} aria-hidden="true" />
        <div className="lp-card-id">
          <h3 className="lp-card-name">
            <Link href={`/discover/${resource.id}`} color="primary" weight="semibold">
              {resource.name}
            </Link>
          </h3>
          <span className="lp-muted">{resource.provider.name}</span>
        </div>
      </div>
      <div className="lp-card-pills">
        <StatePill kind={verificationPill[status]} label={t(`feed.verification.${status}`)} />
        {resource.provenance === 'demo' ? <DemoBadge /> : null}
      </div>
      <p className="lp-card-text">{resource.summary}</p>
      <div className="lp-card-foot">
        {checked ? (
          <span className="lp-muted lp-checked">
            <FeedIcon name="clock" />
            <time dateTime={checked}>
              {t('landing.catalogue.checked', { date: formatDate(checked, locale) })}
            </time>
          </span>
        ) : null}
        {resource.countries.length > 0 ? (
          <span className="lp-countries">
            {resource.countries.slice(0, 6).map((code) => (
              <Country key={code} code={code} />
            ))}
          </span>
        ) : null}
      </div>
    </li>
  );
}

function Catalogue() {
  const { t } = useI18n();
  const latest = useLatestResources();
  if (latest.isPending) return <LoadingState />;
  if (latest.isError)
    return <ErrorState error={latest.error} onRetry={() => void latest.refetch()} />;
  if (latest.data.items.length === 0)
    return <p className="lp-empty">{t('landing.catalogue.empty')}</p>;
  return (
    <ul className="lp-cards">
      {latest.data.items.map((resource) => (
        <ResourceCard key={resource.id} resource={resource} />
      ))}
    </ul>
  );
}

/** `/` for visitors: promise, honest pillars, one live catalogue section, sign-up. */
export function LandingPage() {
  const { t } = useI18n();
  return (
    <div className="lp">
      <section className="lp-hero" aria-labelledby="landing-title">
        <p className="lp-eyebrow">
          <FeedIcon name="sparkles" />
          {t('landing.positioning')}
        </p>
        <h1 id="landing-title" className="lp-h1">
          {t('landing.hero.a')} <span className="lp-h1-em">{t('landing.hero.b')}</span>
        </h1>
        <p className="lp-lede">{t('landing.lede')}</p>
        <div className="lp-actions">
          <Button
            variant="primary"
            size="lg"
            href="/inscription"
            icon={<FeedIcon name="userPlus" />}
            label={t('landing.cta.signup')}
          />
          <ArrowLink href="/discover" label={t('landing.cta.browse')} />
        </div>
        <p className="lp-note">
          <FeedIcon name="eye" />
          {t('landing.note')}
        </p>
        <SearchBox />
      </section>

      <section className="lp-section" aria-labelledby="landing-pillars">
        <header className="lp-head">
          <h2 id="landing-pillars" className="lp-h2">
            {t('landing.pillars.title')}
          </h2>
          <p className="lp-sub">{t('landing.pillars.lede')}</p>
        </header>
        <ul className="lp-pillars">
          {pillars.map((pillar) => (
            <PillarCard key={pillar.id} pillar={pillar} />
          ))}
        </ul>
      </section>

      <section className="lp-section" aria-labelledby="landing-catalogue">
        <header className="lp-head lp-head-row">
          <div>
            <h2 id="landing-catalogue" className="lp-h2">
              {t('landing.catalogue.title')}
            </h2>
            <p className="lp-sub">{t('landing.catalogue.lede')}</p>
          </div>
          <ArrowLink href="/discover" label={t('landing.catalogue.more')} />
        </header>
        <Catalogue />
      </section>

      <section className="lp-final" aria-labelledby="landing-end">
        <h2 id="landing-end" className="lp-final-h">
          {t('landing.end.title')}
        </h2>
        <p className="lp-final-sub">{t('landing.end.text')}</p>
        <Button
          variant="primary"
          size="lg"
          href="/inscription"
          icon={<FeedIcon name="userPlus" />}
          label={t('landing.cta.signup')}
        />
      </section>
    </div>
  );
}
