import { Badge } from '@astryxdesign/core/Badge';
import { Button } from '@astryxdesign/core/Button';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Selector } from '@astryxdesign/core/Selector';
import { Text } from '@astryxdesign/core/Text';
import { TextArea } from '@astryxdesign/core/TextArea';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { ErrorState } from '../../components/ErrorState';
import { FocusRegion } from '../identity/notices';
import { useFailure } from '../identity/useFailure';
import { useInvalidFocus } from '../identity/useInvalidFocus';
import { errorStatus } from '../identity/validation';
import { violationName } from '../identity/violations';
import type { LinkDraft } from './routeSearch';
import { topicCodes, topicKey, type TopicCode } from './topics';
import { checkBody, checkSummary, checkTags, checkTitle, parseTags } from './validate';

type ArticleCreate = components['schemas']['ArticleCreate'];
type Lang = ArticleCreate['lang'];
type Level = ArticleCreate['level'];
type ArticleType = ArticleCreate['type'];

const articleTypes = ['article', 'tutorial'] as const satisfies readonly ArticleType[];
const levels = ['beginner', 'intermediate', 'advanced'] as const satisfies readonly Level[];

type Field = 'title' | 'summary' | 'body' | 'tags' | 'topic' | 'link';
type ServerErrors = Partial<Record<Field, MessageKey>>;

const keyOfField: Record<Field, MessageKey> = {
  title: 'learn.err.title',
  summary: 'learn.err.summary',
  body: 'learn.err.body',
  tags: 'learn.err.tags',
  topic: 'learn.err.topic',
  link: 'learn.err.link',
};

/** "/links/0" is the link field, "/body_md" the body: any violation reads as its field's message. */
function fieldOfViolation(pointer: string): Field | undefined {
  switch (violationName(pointer).split('/')[0]) {
    case 'title':
      return 'title';
    case 'summary':
      return 'summary';
    case 'body_md':
      return 'body';
    case 'tags':
      return 'tags';
    case 'topic':
      return 'topic';
    case 'links':
      return 'link';
    default:
      return undefined;
  }
}

function serverFieldErrors(error: ApiError): ServerErrors {
  const result: ServerErrors = {};
  for (const violation of error.violations) {
    const field = fieldOfViolation(violation.field);
    if (field && !result[field]) result[field] = keyOfField[field];
  }
  return result;
}

const initialType: ArticleType = 'article';
const initialLevel: Level = 'beginner';

function isTopic(value: string): value is TopicCode {
  return (topicCodes as readonly string[]).includes(value);
}

/**
 * The article form. Articles are always created as published (the contract default), so there is
 * no draft choice. The page owns the API call; the form owns validation, the 422 field mapping and
 * the failure notice.
 */
export function ArticleForm({
  initialLink,
  onSubmit,
}: {
  /** A link pre-filled by the page that sent the author here; shown as a removable chip. */
  initialLink?: LinkDraft | undefined;
  onSubmit: (draft: ArticleCreate) => Promise<void>;
}) {
  const { t, locale } = useI18n();
  const { failure, report, clear } = useFailure();
  const { formRef, focusFirstInvalid } = useInvalidFocus();
  const [serverErrors, setServerErrors] = useState<ServerErrors>({});
  const [links, setLinks] = useState<LinkDraft[]>(initialLink ? [initialLink] : []);

  const form = useForm({
    defaultValues: {
      type: initialType,
      title: '',
      summary: '',
      body: '',
      tags: '',
      lang: locale,
      level: initialLevel,
      topic: null as TopicCode | null,
    },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      clear();
      setServerErrors({});
      const summary = value.summary.trim();
      try {
        await onSubmit({
          type: value.type,
          title: value.title.trim(),
          ...(summary ? { summary } : {}),
          body_md: value.body.trim(),
          lang: value.lang,
          level: value.level,
          ...(value.topic ? { topic: value.topic } : {}),
          tags: parseTags(value.tags),
          status: 'published',
          ...(links.length > 0 ? { links } : {}),
        });
      } catch (error) {
        const fields = error instanceof ApiError ? serverFieldErrors(error) : {};
        if (Object.keys(fields).length > 0) {
          setServerErrors(fields);
          focusFirstInvalid();
        } else {
          report(error);
        }
      }
    },
  });

  const forget = (field: Field) => {
    setServerErrors((current) => ({ ...current, [field]: undefined }));
  };

  return (
    <VStack gap={4}>
      {failure ? (
        <FocusRegion key={failure.n}>
          <ErrorState error={failure.error} title="learn.error.title" />
        </FocusRegion>
      ) : null}
      <form
        method="post"
        ref={formRef}
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          // A blur error on one field makes handleSubmit stop early: check every field first, so
          // all the problems are shown at once.
          void form.validateAllFields('submit').then(() => form.handleSubmit());
        }}
      >
        <VStack gap={4}>
          <form.Field name="type">
            {(field) => (
              <SegmentedControl
                label={t('learn.form.type')}
                value={field.state.value}
                onChange={(value) => {
                  field.handleChange(value as ArticleType);
                }}
              >
                {articleTypes.map((option) => (
                  <SegmentedControlItem
                    key={option}
                    value={option}
                    label={t(`learn.type.${option}`)}
                  />
                ))}
              </SegmentedControl>
            )}
          </form.Field>
          <form.Field
            name="title"
            validators={{
              onBlur: ({ value }) => checkTitle(value),
              onSubmit: ({ value }) => checkTitle(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.title;
              return (
                <TextInput
                  label={t('learn.form.title')}
                  description={t('learn.form.title.hint')}
                  value={field.state.value}
                  htmlName={field.name}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('title');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="summary"
            validators={{
              onBlur: ({ value }) => checkSummary(value),
              onSubmit: ({ value }) => checkSummary(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.summary;
              return (
                <TextInput
                  label={t('learn.form.summary')}
                  description={t('learn.form.summary.hint')}
                  isOptional
                  value={field.state.value}
                  htmlName={field.name}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('summary');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="body"
            validators={{
              onBlur: ({ value }) => checkBody(value),
              onSubmit: ({ value }) => checkBody(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.body;
              return (
                <TextArea
                  label={t('learn.form.body')}
                  description={t('learn.form.body.hint')}
                  rows={14}
                  hasSpellCheck
                  value={field.state.value}
                  htmlName={field.name}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('body');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field
            name="tags"
            validators={{
              onBlur: ({ value }) => checkTags(value),
              onSubmit: ({ value }) => checkTags(value),
            }}
          >
            {(field) => {
              const key = field.state.meta.errors[0] ?? serverErrors.tags;
              return (
                <TextInput
                  label={t('learn.form.tags')}
                  description={t('learn.form.tags.hint')}
                  value={field.state.value}
                  htmlName={field.name}
                  autoCapitalize="none"
                  hasAutoCorrect={false}
                  hasSpellCheck={false}
                  onBlur={field.handleBlur}
                  onChange={(value) => {
                    forget('tags');
                    field.handleChange(value);
                  }}
                  {...errorStatus(key ? t(key) : undefined)}
                />
              );
            }}
          </form.Field>
          <form.Field name="topic">
            {(field) => {
              const key = serverErrors.topic;
              return (
                <Selector
                  label={t('learn.form.topic')}
                  isOptional
                  placeholder={t('learn.form.topic.placeholder')}
                  hasClear
                  options={topicCodes.map((code) => ({ value: code, label: t(topicKey(code)) }))}
                  value={field.state.value}
                  onChange={(value) => {
                    forget('topic');
                    field.handleChange(value && isTopic(value) ? value : null);
                  }}
                  {...(key ? { status: { type: 'error' as const, message: t(key) } } : {})}
                />
              );
            }}
          </form.Field>
          <HStack gap={4} wrap="wrap">
            <form.Field name="lang">
              {(field) => (
                <SegmentedControl
                  label={t('learn.form.lang')}
                  value={field.state.value}
                  onChange={(value) => {
                    field.handleChange(value as Lang);
                  }}
                >
                  <SegmentedControlItem value="fr" label={t('learn.lang.fr')} />
                  <SegmentedControlItem value="en" label={t('learn.lang.en')} />
                </SegmentedControl>
              )}
            </form.Field>
            <form.Field name="level">
              {(field) => (
                <SegmentedControl
                  label={t('learn.form.level')}
                  value={field.state.value}
                  onChange={(value) => {
                    field.handleChange(value as Level);
                  }}
                >
                  {levels.map((option) => (
                    <SegmentedControlItem
                      key={option}
                      value={option}
                      label={t(`learn.level.${option}`)}
                    />
                  ))}
                </SegmentedControl>
              )}
            </form.Field>
          </HStack>
          {links.length > 0 ? (
            <VStack gap={1}>
              <Text type="label" color="secondary">
                {t('learn.form.links.label')}
              </Text>
              <HStack gap={1} wrap="wrap">
                {links.map((link) => (
                  <HStack key={link.id} gap={0.5} vAlign="center">
                    <Badge variant="neutral" label={t(`learn.link.${link.type}`)} />
                    <IconButton
                      size="sm"
                      variant="ghost"
                      label={t('learn.form.links.remove', { name: t(`learn.link.${link.type}`) })}
                      icon={<Icon icon="close" />}
                      onClick={() => {
                        forget('link');
                        setLinks((current) => current.filter((item) => item.id !== link.id));
                      }}
                    />
                  </HStack>
                ))}
              </HStack>
              {serverErrors.link ? <Text type="supporting">{t('learn.err.link')}</Text> : null}
            </VStack>
          ) : null}
          <Text as="p" type="supporting">
            {t('learn.form.publishNote')}
          </Text>
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <HStack gap={2}>
                <Button
                  type="submit"
                  variant="primary"
                  label={t('learn.form.publish')}
                  isLoading={isSubmitting}
                />
              </HStack>
            )}
          </form.Subscribe>
        </VStack>
      </form>
    </VStack>
  );
}
