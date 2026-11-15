import { Button } from '@astryxdesign/core/Button';
import { Heading } from '@astryxdesign/core/Heading';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { TextInput } from '@astryxdesign/core/TextInput';
import { useNavigate } from '@tanstack/react-router';
import { useState, type SubmitEvent } from 'react';
import type { components } from '../../api/schema';
import { Avatar } from '../../components/Avatar';
import {
  ArrowRightIcon,
  BookIcon,
  ClockIcon,
  CompassIcon,
  EyeIcon,
  HammerIcon,
  HelpIcon,
  NetworkIcon,
  UserPlusIcon,
  type SvgIcon,
} from '../../components/icons';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { List, ListRow } from '../../components/List';
import { LoadingState } from '../../components/LoadingState';
import { Meta, MetaItem } from '../../components/Meta';
import { StatusBadge } from '../../components/StatusBadge';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { Country, formatDate, VerificationStatusBadge } from '../home/rows';
import { useLatestResources } from './useLatestResources';
import { useReveal } from './useReveal';
import './landing.css';

const MAX_QUERY = 100;

interface Shot {
  id: 'article' | 'question' | 'search' | 'mobile';
  src: string;
  width: number;
  height: number;
  alt: MessageKey;
  caption: MessageKey;
}

// Real captures of the live product, anonymous pages with live content only (public/landing).
const shots: Shot[] = [
  {
    id: 'article',
    src: '/landing/article.jpg',
    width: 1440,
    height: 900,
    alt: 'landing.shots.article.alt',
    caption: 'landing.shots.article.caption',
  },
  {
    id: 'question',
    src: '/landing/question.jpg',
    width: 1440,
    height: 900,
    alt: 'landing.shots.question.alt',
    caption: 'landing.shots.question.caption',
  },
  {
    id: 'search',
    src: '/landing/search.jpg',
    width: 1440,
    height: 570,
    alt: 'landing.shots.search.alt',
    caption: 'landing.shots.search.caption',
  },
  {
    id: 'mobile',
    src: '/landing/mobile.jpg',
    width: 390,
    height: 844,
    alt: 'landing.shots.mobile.alt',
    caption: 'landing.shots.mobile.caption',
  },
];

type Resource = components['schemas']['Resource'];

interface Pillar {
  id: 'learn' | 'connect' | 'solve' | 'discover' | 'build';
  icon: SvgIcon;
  title: MessageKey;
  text: MessageKey;
  link: { to: string; label: MessageKey };
}

// Learn and Connect lead (product decision of 2026-11-12); every other pillar stays present.
const pillars: Pillar[] = [
  {
    id: 'learn',
    icon: BookIcon,
    title: 'landing.pillar.learn.title',
    text: 'landing.pillar.learn.text',
    link: { to: '/apprendre', label: 'landing.pillar.learn.link' },
  },
  {
    id: 'connect',
    icon: NetworkIcon,
    title: 'landing.pillar.connect.title',
    text: 'landing.pillar.connect.text',
    link: { to: '/connect', label: 'landing.pillar.connect.link' },
  },
  {
    id: 'solve',
    icon: HelpIcon,
    title: 'landing.pillar.solve.title',
    text: 'landing.pillar.solve.text',
    link: { to: '/solve', label: 'landing.pillar.solve.link' },
  },
  {
    id: 'discover',
    icon: CompassIcon,
    title: 'landing.pillar.discover.title',
    text: 'landing.pillar.discover.text',
    link: { to: '/discover', label: 'landing.pillar.discover.link' },
  },
  {
    id: 'build',
    icon: HammerIcon,
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
      <Icon icon={ArrowRightIcon} size="sm" />
    </span>
  );
}

function PillarCard({ pillar }: { pillar: Pillar }) {
  const { t } = useI18n();
  const lead = pillar.id === 'learn' || pillar.id === 'connect';
  return (
    <li className={lead ? 'lp-pillar lp-pillar-lead' : 'lp-pillar'}>
      <span className="lp-pillar-icon" aria-hidden="true">
        <Icon icon={pillar.icon} color="accent" size="lg" />
      </span>
      <Heading level={3}>{t(pillar.title)}</Heading>
      <p>{t(pillar.text)}</p>
      <ArrowLink href={pillar.link.to} label={t(pillar.link.label)} />
    </li>
  );
}

function ShotFigure({ shot }: { shot: Shot }) {
  const { t } = useI18n();
  const reveal = useReveal();
  return (
    <figure ref={reveal} className={`lp-shot lp-shot-${shot.id}`}>
      <img
        className="lp-shot-img"
        src={shot.src}
        width={shot.width}
        height={shot.height}
        alt={t(shot.alt)}
        loading="lazy"
        decoding="async"
      />
      <figcaption>{t(shot.caption)}</figcaption>
    </figure>
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

/** The prototype's catalogue card as a shared grid row: name, provider, state, summary, date. */
function ResourceCard({ resource }: { resource: Resource }) {
  const { t, locale } = useI18n();
  const checked = resource.verification_status === 'verified' ? resource.last_verified_at : null;
  return (
    <ListRow
      title={resource.name}
      href={`/discover/${resource.id}`}
      lead={<Avatar name={resource.name} size="lg" />}
      excerpt={resource.summary}
      meta={
        <Meta>
          <MetaItem>{resource.provider.name}</MetaItem>
          {checked ? (
            <MetaItem icon={ClockIcon}>
              <time dateTime={checked}>
                {t('landing.catalogue.checked', { date: formatDate(checked, locale) })}
              </time>
            </MetaItem>
          ) : null}
          {resource.countries.slice(0, 6).map((code) => (
            <Country key={code} code={code} />
          ))}
        </Meta>
      }
      aside={
        <>
          <VerificationStatusBadge status={resource.verification_status} />
          {resource.provenance === 'demo' ? <StatusBadge kind="demo" /> : null}
        </>
      }
    />
  );
}

function Catalogue() {
  const { t } = useI18n();
  const latest = useLatestResources();
  if (latest.isPending) return <LoadingState />;
  if (latest.isError)
    return <ErrorState error={latest.error} onRetry={() => void latest.refetch()} />;
  if (latest.data.items.length === 0)
    return <EmptyState icon={CompassIcon} title={t('landing.catalogue.empty')} headingLevel={3} />;
  return (
    <List label={t('landing.catalogue.title')} isGrid>
      {latest.data.items.map((resource) => (
        <ResourceCard key={resource.id} resource={resource} />
      ))}
    </List>
  );
}

/** `/` for visitors: promise, honest pillars, one live catalogue section, sign-up. */
export function LandingPage() {
  const { t } = useI18n();
  const revealShots = useReveal();
  const revealPillars = useReveal();
  const revealCatalogue = useReveal();
  const revealEnd = useReveal();
  return (
    <div className="lp">
      <section className="lp-hero" aria-labelledby="landing-title">
        <h1 id="landing-title" className="lp-h1">
          {t('landing.hero.a')} <span className="lp-h1-em">{t('landing.hero.b')}</span>
        </h1>
        <p className="lp-lede">{t('landing.lede')}</p>
        <div className="lp-actions">
          <Button
            variant="primary"
            size="lg"
            href="/inscription"
            icon={<Icon icon={UserPlusIcon} color="inherit" />}
            label={t('landing.cta.signup')}
          />
          <ArrowLink href="/discover" label={t('landing.cta.browse')} />
        </div>
        <p className="lp-note">
          <Icon icon={EyeIcon} size="sm" color="inherit" />
          {t('landing.note')}
        </p>
        <SearchBox />
      </section>

      <section ref={revealShots} className="lp-section" aria-labelledby="landing-shots">
        <div className="lp-head">
          <h2 id="landing-shots" className="lp-h2">
            {t('landing.shots.title')}
          </h2>
        </div>
        <div className="lp-shots">
          {shots.map((shot) => (
            <ShotFigure key={shot.id} shot={shot} />
          ))}
        </div>
      </section>

      <section ref={revealPillars} className="lp-section" aria-labelledby="landing-pillars">
        <div className="lp-head">
          <h2 id="landing-pillars" className="lp-h2">
            {t('landing.pillars.title')}
          </h2>
        </div>
        <ul className="lp-pillars">
          {pillars.map((pillar) => (
            <PillarCard key={pillar.id} pillar={pillar} />
          ))}
        </ul>
      </section>

      <section ref={revealCatalogue} className="lp-section" aria-labelledby="landing-catalogue">
        <div className="lp-head lp-head-row">
          <div>
            <h2 id="landing-catalogue" className="lp-h2">
              {t('landing.catalogue.title')}
            </h2>
          </div>
          <ArrowLink href="/discover" label={t('landing.catalogue.more')} />
        </div>
        <Catalogue />
      </section>

      <section ref={revealEnd} className="lp-final" aria-labelledby="landing-end">
        <h2 id="landing-end" className="lp-final-h">
          {t('landing.end.title')}
        </h2>
        <p className="lp-final-sub">{t('landing.end.text')}</p>
        <Button
          variant="primary"
          size="lg"
          href="/inscription"
          icon={<Icon icon={UserPlusIcon} color="inherit" />}
          label={t('landing.cta.signup')}
        />
      </section>
    </div>
  );
}
