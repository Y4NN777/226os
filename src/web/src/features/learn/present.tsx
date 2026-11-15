import { Avatar } from '@astryxdesign/core/Avatar';
import { Badge } from '@astryxdesign/core/Badge';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { Link } from '@astryxdesign/core/Link';
import { Text } from '@astryxdesign/core/Text';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import type { ReactNode } from 'react';
import type { components } from '../../api/schema';
import { useI18n } from '../../i18n/I18nProvider';
import { countryName, formatDate } from '../discover/present';
import { BookIcon, CodeIcon, FileTextIcon, type SvgIcon } from './icons';

type Author = components['schemas']['AuthorRef'];
type Lang = components['schemas']['Lang'];
type Level = components['schemas']['Level'];

/** What a Learn row shows: an article, a tutorial or a training series ("formation"). */
export type LearnKind = components['schemas']['ArticleType'] | 'series';

const kindIcon: Record<LearnKind, SvgIcon> = {
  article: FileTextIcon,
  tutorial: CodeIcon,
  series: BookIcon,
};

export function TagList({ tags }: { tags: readonly string[] }) {
  return (
    <HStack as="ul" gap={1} wrap="wrap">
      {tags.map((tag) => (
        <li key={tag}>
          <Badge variant="blue" label={tag} />
        </li>
      ))}
    </HStack>
  );
}

/** Tags as the prototype shows them under a lesson: small monospace chips. */
export function CodeTags({ tags, label }: { tags: readonly string[]; label: string }) {
  return (
    <ul className="learn-tags" aria-label={label}>
      {tags.map((tag) => (
        <li key={tag} className="learn-tag">
          {tag}
        </li>
      ))}
    </ul>
  );
}

/** Display name only, as in lists. */
export function Byline({ author }: { author: Author }) {
  const { t } = useI18n();
  return <Text type="supporting">{t('learn.by', { name: author.display_name })}</Text>;
}

/** Display name and a link to the profile by username. */
export function AuthorLine({ author }: { author: Author }) {
  const { t } = useI18n();
  return (
    <HStack gap={2} wrap="wrap" vAlign="center">
      <Text type="supporting">{t('learn.by', { name: author.display_name })}</Text>
      <Link href={`/profils/${author.username}`} type="supporting">
        @{author.username}
      </Link>
    </HStack>
  );
}

/** Byline of a page: avatar, name and profile link, country code, publication date. */
export function AuthorByline({ author, date }: { author: Author; date?: string | null }) {
  const { locale } = useI18n();
  return (
    <div className="learn-by">
      <span className="learn-by-who">
        {/* Decorative: the name follows in text, so it is not read twice. */}
        <Avatar name={author.display_name} size={32} tooltip={false} aria-hidden="true" />
        <AuthorLine author={author} />
      </span>
      {author.country ? (
        <Code code={author.country} name={countryName(author.country, locale)} />
      ) : null}
      {date ? (
        <time className="learn-meta-item" dateTime={date}>
          {formatDate(date, locale)}
        </time>
      ) : null}
    </div>
  );
}

/** A short code on screen (FR, SN) whose full name is what a screen reader reads. */
export function Code({ code, name }: { code: string; name: string }) {
  return (
    <span className="learn-code">
      <span aria-hidden="true">{code.toUpperCase()}</span>
      <VisuallyHidden>{name}</VisuallyHidden>
    </span>
  );
}

export function LangCode({ lang }: { lang: Lang }) {
  const { t } = useI18n();
  return <Code code={lang} name={t(`learn.lang.${lang}`)} />;
}

/** Type badge with its icon and its name: never colour alone. */
export function KindBadge({ kind }: { kind: LearnKind }) {
  const { t } = useI18n();
  return (
    <Badge
      variant={kind === 'series' ? 'green' : 'neutral'}
      icon={<Icon icon={kindIcon[kind]} size="sm" />}
      label={t(`learn.kind.${kind}`)}
    />
  );
}

export function LevelBadge({ level }: { level: Level }) {
  const { t } = useI18n();
  return <Badge label={t(`learn.level.${level}`)} />;
}

/** French puts 0 and 1 in the singular (« 0 leçon »), English only 1. */
export function isSingular(n: number, locale: string): boolean {
  return locale === 'fr' ? n <= 1 : n === 1;
}

/** "25 min", "1 h", "3 h 05": the prototype's duration format, the same in both languages. */
export function formatMinutes(total: number): string {
  const hours = Math.floor(total / 60);
  const minutes = total % 60;
  if (hours === 0) return `${String(total)} min`;
  return minutes === 0
    ? `${String(hours)} h`
    : `${String(hours)} h ${String(minutes).padStart(2, '0')}`;
}

/** Stable tint per item, from the theme tokens only. */
function toneOf(id: string): 0 | 1 | 2 {
  let hash = 0;
  for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  return (hash % 3) as 0 | 1 | 2;
}

export function KindThumb({ kind, id }: { kind: LearnKind; id: string }) {
  const tone = kind === 'series' ? 0 : toneOf(id);
  return (
    <span className={`learn-thumb learn-tone-${String(tone)}`} aria-hidden="true">
      <Icon icon={kindIcon[kind]} size="lg" />
    </span>
  );
}

/**
 * One dense list row (prototype `list > row`): a lead tile, a top line, the title link (the whole
 * row opens it), an excerpt, a meta line and an aside. Content-language parts carry `lang`.
 */
export function ItemRow({
  lead,
  top,
  title,
  href,
  lang,
  excerpt,
  meta,
  aside,
}: {
  lead?: ReactNode;
  top?: ReactNode;
  title: string;
  href: string;
  lang?: Lang | undefined;
  excerpt?: string | undefined;
  meta?: ReactNode;
  aside?: ReactNode;
}) {
  return (
    <li className={lead ? 'learn-row' : 'learn-row learn-row-plain'}>
      {lead ? <div>{lead}</div> : null}
      <div className="learn-row-main">
        {top ? <div className="learn-row-top">{top}</div> : null}
        <h3 className="learn-row-title" lang={lang}>
          <Link href={href} type="inherit" color="inherit">
            {title}
          </Link>
        </h3>
        {excerpt ? (
          <p className="learn-row-excerpt" lang={lang}>
            {excerpt}
          </p>
        ) : null}
        {meta ? <div className="learn-meta">{meta}</div> : null}
      </div>
      {aside ? <div className="learn-row-aside">{aside}</div> : null}
    </li>
  );
}
