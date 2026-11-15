import { Icon } from '@astryxdesign/core/Icon';
import { VisuallyHidden } from '@astryxdesign/core/VisuallyHidden';
import type { ReactNode } from 'react';
import { RouterAnchor } from '../../components/RouterAnchor';
import { useI18n } from '../../i18n/I18nProvider';
import { ChatGlyph, CheckCircleGlyph } from './icons';
import { DateText, excerptOf, LangMark, Person, StatePill, TagList } from './present';
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
 * One dense list row (prototype `.list > .row`): score and answer count in the lead, the title
 * link covering the row, a plain-text excerpt, the state pill and tags, then author, country and
 * date; the original language sits at the end. Only contract fields are shown.
 */
export function QuestionRow({ question }: { question: Question }) {
  const { t } = useI18n();
  const resolved = question.state === 'resolved';
  const answerTone = resolved ? 'accepted' : question.answer_count === 0 ? 'none' : undefined;
  const excerpt = excerptOf(question.body_md);
  return (
    <li className="solve-row">
      <div className="solve-row-lead">
        <Stat
          icon={<Icon icon="arrowUp" size="sm" />}
          value={question.vote_score}
          label={t('solve.score', { n: question.vote_score })}
        />
        <Stat
          icon={<Icon icon={resolved ? CheckCircleGlyph : ChatGlyph} size="sm" />}
          value={question.answer_count}
          label={t('solve.answers.count', { n: question.answer_count })}
          tone={answerTone}
        />
      </div>
      <div className="solve-row-main">
        <h2 className="solve-row-title" lang={question.lang}>
          <RouterAnchor href={`/questions/${question.id}`}>{question.title}</RouterAnchor>
        </h2>
        {excerpt ? (
          <p className="solve-row-excerpt" lang={question.lang}>
            {excerpt}
          </p>
        ) : null}
        {question.state !== 'answered' || question.tags.length > 0 ? (
          <div className="solve-meta">
            <StatePill state={question.state} />
            <TagList tags={question.tags} />
          </div>
        ) : null}
        <div className="solve-meta">
          <Person author={question.author} />
          <DateText iso={question.created_at} />
        </div>
      </div>
      <div className="solve-row-aside">
        <LangMark lang={question.lang} />
      </div>
    </li>
  );
}
