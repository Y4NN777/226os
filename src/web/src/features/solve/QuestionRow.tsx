import { Icon } from '@astryxdesign/core/Icon';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import type { ReactNode } from 'react';
import { CheckCircleIcon, MessageIcon } from '../../components/icons';
import { LangCode } from '../../components/LangCode';
import { ListRow } from '../../components/List';
import { Meta } from '../../components/Meta';
import { TagList } from '../../components/Tag';
import { useI18n } from '../../i18n/I18nProvider';
import { DateText, excerptOf, Person, QuestionState } from './present';
import type { Question } from './queries';

/** A count with its glyph; the number is shown, the full label is what a screen reader reads. */
function Stat({
  icon,
  value,
  label,
  tone,
}: {
  icon: ReactNode;
  value: number;
  label: string;
  tone?: 'accepted' | 'none' | undefined;
}) {
  return (
    <span className={tone ? `solve-stat solve-stat-${tone}` : 'solve-stat'}>
      {icon}
      <strong aria-hidden="true">{value}</strong>
      <VisuallyHidden>{label}</VisuallyHidden>
    </span>
  );
}

/**
 * One dense list row (prototype `.list > .row`) on the shared ListRow: score and answer count in
 * the lead, the title link covering the row, a plain-text excerpt, the state pill and tags, then
 * author, country and date; the original language sits at the end. Only contract fields.
 */
export function QuestionRow({ question }: { question: Question }) {
  const { t } = useI18n();
  const resolved = question.state === 'resolved';
  const answerTone = resolved ? 'accepted' : question.answer_count === 0 ? 'none' : undefined;
  const excerpt = excerptOf(question.body_md);
  return (
    <ListRow
      title={question.title}
      href={`/questions/${question.id}`}
      headingLevel={2}
      titleLang={question.lang}
      lead={
        <div className="solve-stats">
          <Stat
            icon={<Icon icon="arrowUp" size="sm" />}
            value={question.vote_score}
            label={t('solve.score', { n: question.vote_score })}
          />
          <Stat
            icon={<Icon icon={resolved ? CheckCircleIcon : MessageIcon} size="sm" />}
            value={question.answer_count}
            label={t('solve.answers.count', { n: question.answer_count })}
            tone={answerTone}
          />
        </div>
      }
      {...(excerpt ? { excerpt } : {})}
      {...(question.state !== 'answered' || question.tags.length > 0
        ? {
            tags: (
              <div className="solve-pill-tags">
                <QuestionState state={question.state} />
                {question.tags.length > 0 ? <TagList tags={question.tags} /> : null}
              </div>
            ),
          }
        : {})}
      meta={
        <Meta>
          <Person author={question.author} />
          <DateText iso={question.created_at} />
        </Meta>
      }
      aside={<LangCode lang={question.lang} />}
      asideCorner
    />
  );
}
