import { useState, type ReactNode } from 'react';
import { useNavigate } from '@tanstack/react-router';
import {
  ArrowRightIcon,
  BookIcon,
  CompassIcon,
  EyeIcon,
  FileTextIcon,
  GitBranchIcon,
  HammerIcon,
  HelpIcon,
  LanguagesIcon,
  MapPinIcon,
  MessageIcon,
  SearchIcon,
  ShieldCheckIcon,
  UsersIcon,
  type SvgIcon,
} from '../../components/icons';
import { RouterAnchor } from '../../components/RouterAnchor';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { formatDate } from '../home/rows';
import { CountUp } from './CountUp';
import { Globe } from './Globe';
import { reveal, useReveal } from './reveal';
import { useLandingActivity, useLandingTalents } from './useLandingData';
import { useLandingMotion } from './useLandingMotion';
import './landing.css';

const key = (id: string, part: string) => `landing.${id}.${part}` as MessageKey;

/** The one button of the page: a real link to the sign-up form, never a fake button. */
function Cta({ label }: { label: string }) {
  return (
    <RouterAnchor href="/inscription" className="lp-cta">
      <span>{label}</span>
      <span className="lp-cta-arrow" aria-hidden="true">
        <ArrowRightIcon />
      </span>
    </RouterAnchor>
  );
}

/** Decorative photo of a person, in a circle. Never named, never presented as a member. */
function Circle({ file, className, index }: { file: string; className: string; index: number }) {
  return (
    <span className={`lp-circle ${className}`} {...reveal('zoom-in', index)}>
      <img
        src={`/landing/${file}.jpg`}
        alt=""
        width={96}
        height={96}
        loading="lazy"
        decoding="async"
      />
    </span>
  );
}

/** Decorative scene photo, in a 600 and a 1200 px version. */
function Scene({
  name,
  width,
  height,
  sizes,
  position,
}: {
  name: string;
  width: number;
  height: number;
  sizes: string;
  position?: string;
}) {
  return (
    <img
      className="lp-scene"
      src={`/landing/${name}-1200.jpg`}
      srcSet={`/landing/${name}-600.jpg 600w, /landing/${name}-1200.jpg 1200w`}
      sizes={sizes}
      alt=""
      width={width}
      height={height}
      loading="lazy"
      decoding="async"
      style={position ? { objectPosition: position } : undefined}
    />
  );
}

const SHOTS = {
  phone: { width: 232, height: 502 },
  desktop: { width: 560, height: 350 },
} as const;

/**
 * A device frame in plain CSS around a real capture of the interface (Professor's public pages,
 * dark theme, taken at 2x). The 1x file is the displayed size and the 2x file is for dense
 * screens: the picture is never enlarged, so never soft.
 */
function Device({
  kind,
  name,
  className,
  ...attributes
}: {
  kind: keyof typeof SHOTS;
  name: string;
  className?: string;
} & Partial<ReturnType<typeof reveal>>) {
  const src = `/landing/${name}-1x.jpg`;
  const { width, height } = SHOTS[kind];
  return (
    <div
      className={`lp-device lp-device-${kind}${className ? ` ${className}` : ''}`}
      {...attributes}
    >
      {kind === 'desktop' ? (
        <span className="lp-device-bar" aria-hidden="true">
          <i />
          <i />
          <i />
        </span>
      ) : null}
      <img
        src={src}
        srcSet={`${src} 1x, ${src.replace('-1x', '-2x')} 2x`}
        alt=""
        width={width}
        height={height}
        loading="lazy"
        decoding="async"
      />
    </div>
  );
}

/**
 * The site search, open to visitors: a GET form to `/recherche?q=` that works without script,
 * and navigates in the app when it runs.
 */
function HeroSearch() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const [value, setValue] = useState('');
  return (
    <form
      role="search"
      aria-label={t('landing.search.label')}
      action="/recherche"
      method="get"
      className="lp-search"
      onSubmit={(event) => {
        event.preventDefault();
        const q = value.trim();
        void navigate({ to: '/recherche', search: q ? { q } : {} });
      }}
    >
      <label htmlFor="landing-search" className="lp-sr">
        {t('landing.search.label')}
      </label>
      <SearchIcon className="lp-search-icon" aria-hidden="true" />
      <input
        id="landing-search"
        className="lp-search-input"
        type="search"
        name="q"
        value={value}
        onChange={(event) => {
          setValue(event.target.value);
        }}
        placeholder={t('landing.search.placeholder')}
        enterKeyHint="search"
        autoComplete="off"
      />
      <button type="submit" className="lp-btn lp-search-submit">
        {t('landing.search.submit')}
      </button>
    </form>
  );
}

// The five spaces, Découvrir and Connect first: one icon, one name, one line and a route each.
const spaces: readonly { id: string; href: string; icon: SvgIcon; wide?: boolean }[] = [
  { id: 'discover', href: '/discover', icon: CompassIcon, wide: true },
  { id: 'connect', href: '/connect', icon: UsersIcon, wide: true },
  { id: 'solve', href: '/solve', icon: HelpIcon },
  { id: 'learn', href: '/apprendre', icon: BookIcon },
  { id: 'build', href: '/build', icon: HammerIcon },
];

function Hero() {
  const { t } = useI18n();
  const motion = useLandingMotion();
  return (
    <section ref={motion} id="accueil" className="lp-hero lp-dark" aria-labelledby="landing-title">
      <Globe />
      <div className="lp-hero-copy">
        <h1 id="landing-title" className="lp-h1">
          {t('landing.hero.a')} <span className="lp-accent">{t('landing.hero.em')}</span>
        </h1>
        <p className="lp-lede">{t('landing.lede')}</p>
        <HeroSearch />
        <nav className="lp-quick" aria-label={t('landing.quick.label')}>
          <ul>
            {spaces.map(({ id, href, icon: Icon }) => (
              <li key={id}>
                <RouterAnchor href={href} className="lp-quick-link">
                  <Icon aria-hidden="true" />
                  {t(key(`pillar.${id}`, 'title'))}
                </RouterAnchor>
              </li>
            ))}
          </ul>
        </nav>
        <Cta label={t('landing.cta.signup')} />
      </div>
    </section>
  );
}

const advantages: readonly { id: string; icon: SvgIcon }[] = [
  { id: 'verified', icon: ShieldCheckIcon },
  { id: 'local', icon: MapPinIcon },
  { id: 'lang', icon: LanguagesIcon },
  { id: 'profile', icon: EyeIcon },
];

function Why() {
  const { t } = useI18n();
  return (
    <section id="avantages" className="lp-section lp-why" aria-labelledby="landing-why">
      <div className="lp-wrap lp-split">
        <div className="lp-media lp-media-device" {...reveal('left')}>
          <Device kind="phone" name="mock-phone-question" />
          <Circle file="portrait-3" className="lp-circle-a" index={1} />
          <Circle file="portrait-6" className="lp-circle-b" index={2} />
          <Circle file="portrait-5" className="lp-circle-c" index={3} />
        </div>
        <div className="lp-body">
          <h2 id="landing-why" className="lp-h2" {...reveal('up')}>
            {t('landing.why.title')}
          </h2>
          <ul className="lp-list">
            {advantages.map(({ id, icon: Icon }, i) => (
              <li key={id} className="lp-point" {...reveal('right', i)}>
                <span className="lp-tile" aria-hidden="true">
                  <Icon />
                </span>
                <div>
                  <h3 className="lp-h3">{t(key(`why.${id}`, 'title'))}</h3>
                  <p>{t(key(`why.${id}`, 'text'))}</p>
                </div>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}

function Spaces() {
  const { t } = useI18n();
  return (
    <section id="espaces" className="lp-section lp-spaces" aria-labelledby="landing-spaces">
      <div className="lp-wrap lp-center">
        <h2 id="landing-spaces" className="lp-h2" {...reveal('down')}>
          {t('landing.spaces.title')}
        </h2>
        <ul className="lp-grid">
          {spaces.map(({ id, href, icon: Icon, wide }, i) => (
            <li
              key={id}
              className={`lp-space${wide ? ' lp-space-wide' : ''}`}
              {...reveal('zoom-in', i)}
            >
              <RouterAnchor href={href} className="lp-space-link">
                <span className="lp-tile" aria-hidden="true">
                  <Icon />
                </span>
                <h3 className="lp-h3">{t(key(`pillar.${id}`, 'title'))}</h3>
                <p>{t(key(`pillar.${id}`, 'text'))}</p>
                <span className="lp-space-more">{t('landing.spaces.explore')}</span>
              </RouterAnchor>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}

/**
 * Public GitHub ranking for Burkina Faso (committers.top). Each entry opens the GitHub profile in
 * a new tab. No data, or a failed request, renders no section.
 */
function Talents() {
  const { t, locale } = useI18n();
  const developers = useLandingTalents().data ?? [];
  if (developers.length === 0) return null;
  const number = new Intl.NumberFormat(locale);
  return (
    <section id="talents" className="lp-section lp-talents" aria-labelledby="landing-talents">
      <div className="lp-wrap lp-center">
        <h2 id="landing-talents" className="lp-h2" {...reveal('up')}>
          {t('landing.talents.title')}
        </h2>
        <ul className="lp-talent-list" aria-label={t('landing.talents.list')}>
          {developers.map((dev, i) => (
            <li key={dev.login} {...reveal('up', i)}>
              <a
                className="lp-talent"
                href={`https://github.com/${encodeURIComponent(dev.login)}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                {dev.avatar_url ? (
                  <img
                    src={dev.avatar_url}
                    alt=""
                    width={56}
                    height={56}
                    loading="lazy"
                    decoding="async"
                    referrerPolicy="no-referrer"
                  />
                ) : (
                  <span className="lp-talent-avatar" aria-hidden="true" />
                )}
                <span className="lp-talent-rank">
                  {t('landing.talents.rank', { rank: String(dev.rank) })}
                </span>
                <span className="lp-talent-name">{dev.name || dev.login}</span>
                <span className="lp-talent-meta">
                  {t('landing.talents.contributions', {
                    count: number.format(dev.contributions),
                  })}
                </span>
              </a>
            </li>
          ))}
        </ul>
        <div className="lp-talent-actions" {...reveal('up', developers.length)}>
          <RouterAnchor href="/connect?source=github&country=BF" className="lp-btn">
            {t('landing.talents.all')}
          </RouterAnchor>
          <RouterAnchor href="/inscription" className="lp-btn lp-btn-ghost">
            {t('landing.talents.join')}
          </RouterAnchor>
        </div>
        <p className="lp-talent-source">{t('landing.talents.source')}</p>
      </div>
    </section>
  );
}

type Noun = 'questions' | 'articles' | 'resources' | 'projects';
const nouns: readonly Noun[] = ['questions', 'articles', 'resources', 'projects'];

/** Exact totals read from the API. A zero, a missing or a failed one is not rendered. */
function Totals({ totals }: { totals: Record<Noun, number | undefined> }) {
  const { t, locale } = useI18n();
  const shown = nouns.filter((noun) => Boolean(totals[noun]));
  if (shown.length === 0) return null;
  return (
    <ul className="lp-proof" aria-label={t('landing.proof.label')}>
      {shown.map((noun, i) => {
        const total = totals[noun] ?? 0;
        const plural = new Intl.PluralRules(locale).select(total) === 'one' ? 'one' : 'other';
        return (
          <li key={noun} {...reveal('up', i)}>
            <strong className="lp-num">
              <CountUp value={total} />
            </strong>
            <span>{t(`landing.proof.${noun}.${plural}` as MessageKey)}</span>
          </li>
        );
      })}
    </ul>
  );
}

function LiveCard({
  icon: Icon,
  label,
  title,
  by,
  date,
  index,
}: {
  icon: SvgIcon;
  label: string;
  title: string;
  by: string;
  date: string;
  index: number;
}) {
  const { locale } = useI18n();
  return (
    <li className="lp-live-card" {...reveal('up', index)}>
      <span className="lp-live-label">
        <Icon aria-hidden="true" />
        {label}
      </span>
      <span className="lp-live-title">{title}</span>
      <span className="lp-live-meta">
        {by} · <time dateTime={date}>{formatDate(date, locale)}</time>
      </span>
    </li>
  );
}

/** Real content, shown without letting a visitor into the app: plain cards, not links. */
function Latest() {
  const { t } = useI18n();
  const { latest } = useLandingActivity();
  const cards: ReactNode[] = [];
  const { question, article, project } = latest;
  if (question) {
    cards.push(
      <LiveCard
        key="q"
        icon={MessageIcon}
        label={t('landing.live.question')}
        title={question.title}
        by={t('landing.live.by', { name: question.author.display_name })}
        date={question.created_at}
        index={cards.length}
      />,
    );
  }
  if (article) {
    cards.push(
      <LiveCard
        key="a"
        icon={FileTextIcon}
        label={t('landing.live.article')}
        title={article.title}
        by={t('landing.live.by', { name: article.author.display_name })}
        date={article.published_at}
        index={cards.length}
      />,
    );
  }
  if (project) {
    cards.push(
      <LiveCard
        key="p"
        icon={GitBranchIcon}
        label={t('landing.live.project')}
        title={project.name}
        by={t('landing.live.by', { name: project.owner.display_name })}
        date={project.created_at}
        index={cards.length}
      />,
    );
  }
  if (cards.length === 0) return null;
  return (
    <ul className="lp-live" aria-label={t('landing.live.label')}>
      {cards}
    </ul>
  );
}

/** Real totals and latest posts. No data, no section. */
function Activity() {
  const { t } = useI18n();
  const { totals, hasActivity } = useLandingActivity();
  if (!hasActivity) return null;
  return (
    <section id="activite" className="lp-section lp-activity" aria-labelledby="landing-activity">
      <div className="lp-wrap lp-center">
        <h2 id="landing-activity" className="lp-h2" {...reveal('up')}>
          {t('landing.activity.title')}
        </h2>
        <Totals totals={totals} />
        <div className="lp-stage">
          <Device kind="desktop" name="mock-desktop-article" {...reveal('zoom-out')} />
          <Device
            kind="phone"
            name="mock-phone-article"
            className="lp-stage-phone"
            {...reveal('right', 1)}
          />
        </div>
        <Latest />
      </div>
    </section>
  );
}

const steps = ['step1', 'step2', 'step3'] as const;

function Join() {
  const { t } = useI18n();
  return (
    <section id="rejoindre" className="lp-section lp-join lp-dark" aria-labelledby="landing-join">
      <div className="lp-halo" aria-hidden="true" />
      <div className="lp-wrap lp-split lp-split-flip">
        <div className="lp-body">
          <h2 id="landing-join" className="lp-h2" {...reveal('up')}>
            {t('landing.join.title')}
          </h2>
          <ol className="lp-steps" aria-label={t('landing.join.steps')}>
            {steps.map((id, i) => (
              <li key={id} className="lp-step" {...reveal('left', i)}>
                <span className="lp-step-n" aria-hidden="true">
                  {i + 1}
                </span>
                <div>
                  <h3 className="lp-h3">{t(key(`join.${id}`, 'title'))}</h3>
                  <p>{t(key(`join.${id}`, 'text'))}</p>
                </div>
              </li>
            ))}
          </ol>
          <div {...reveal('zoom-in', steps.length)}>
            <Cta label={t('landing.cta.signup')} />
          </div>
        </div>
        <div className="lp-media" {...reveal('right')}>
          <Scene
            name="scene-community"
            width={1200}
            height={800}
            sizes="(min-width: 1024px) 560px, 92vw"
          />
          <Circle file="portrait-1" className="lp-circle-d" index={1} />
          <Circle file="portrait-8" className="lp-circle-e" index={2} />
          <Circle file="portrait-4" className="lp-circle-f" index={3} />
          <Circle file="portrait-2" className="lp-circle-g" index={4} />
        </div>
      </div>
    </section>
  );
}

/**
 * `/` for visitors: promise with search and direct links to the five spaces, the spaces as
 * entries, the public GitHub talents, advantages, real activity, three steps to sign up.
 */
export function LandingPage() {
  const revealRoot = useReveal();
  return (
    <div ref={revealRoot} className="lp-page">
      <Hero />
      <Spaces />
      <Talents />
      <Why />
      <Activity />
      <Join />
    </div>
  );
}
