import { Button } from '@astryxdesign/core/Button';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { SegmentedControl, SegmentedControlItem } from '@astryxdesign/core/SegmentedControl';
import { Text } from '@astryxdesign/core/Text';
import { TextArea } from '@astryxdesign/core/TextArea';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useState, type ReactNode } from 'react';
import { ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { useI18n } from '../../i18n/I18nProvider';
import type { MessageKey } from '../../i18n/messages';
import { FailureNotice } from '../identity/notices';
import { useFailure } from '../identity/useFailure';
import { useInvalidFocus } from '../identity/useInvalidFocus';
import { errorStatus } from '../identity/validation';
import { violationName } from '../identity/violations';
import {
  checkBody,
  checkSnippetCode,
  checkSnippetLanguage,
  checkTags,
  checkTitle,
  isValidLinkId,
  linkTypes,
  parseTags,
  snippetsOf,
  type LinkType,
} from './validate';

type Lang = components['schemas']['Lang'];
type Snippet = components['schemas']['Snippet'];
type LinkRef = components['schemas']['LinkRef'];

/** What the form hands to the page: ready for createQuestion (or the diff for updateQuestion). */
export interface QuestionDraft {
  title: string;
  body_md: string;
  lang: Lang;
  tags: string[];
  snippets: Snippet[];
  links: LinkRef[];
}

export interface QuestionInitial {
  title: string;
  body: string;
  tags: string[];
  lang: Lang;
}

type Field = 'title' | 'body' | 'tags' | 'snippetLanguage' | 'snippetCode' | 'link';
type ServerErrors = Partial<Record<Field, MessageKey>>;

const keyOfField: Record<Field, MessageKey> = {
  title: 'solve.err.title',
  body: 'solve.err.body',
  tags: 'solve.err.tags',
  snippetLanguage: 'solve.err.snippet.language',
  snippetCode: 'solve.err.snippet.code',
  link: 'solve.err.link',
};

/** "/snippets/0/code" is the code field, "/tags" the tags field; any violation reads as that field's message. */
function fieldOfViolation(pointer: string): Field | undefined {
  const [head, , leaf] = violationName(pointer).split('/');
  switch (head) {
    case 'title':
      return 'title';
    case 'body_md':
      return 'body';
    case 'tags':
      return 'tags';
    case 'snippets':
      return leaf === 'language' ? 'snippetLanguage' : 'snippetCode';
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

interface LinkDraft {
  type: LinkType;
  id: string;
}

/**
 * The question form, for asking (with one optional snippet and links) and for editing (title,
 * body and tags). The page owns the API call; the form owns validation, the 422 field mapping
 * and the failure notice. `failureNotice` lets a page answer a specific failure (412).
 */
export function QuestionForm({
  mode,
  initial,
  submitLabel,
  onSubmit,
  onCancel,
  failureNotice,
}: {
  mode: 'create' | 'edit';
  initial: QuestionInitial;
  submitLabel: string;
  onSubmit: (draft: QuestionDraft) => Promise<void>;
  onCancel: () => void;
  failureNotice?: (error: unknown) => ReactNode;
}) {
  const { t } = useI18n();
  const { failure, report, clear } = useFailure();
  const { formRef, focusFirstInvalid } = useInvalidFocus();
  const [serverErrors, setServerErrors] = useState<ServerErrors>({});
  const [links, setLinks] = useState<LinkDraft[]>([]);
  const [linkType, setLinkType] = useState<LinkType>('resource');
  const [linkId, setLinkId] = useState('');
  const [linkInvalid, setLinkInvalid] = useState(false);

  const form = useForm({
    defaultValues: {
      title: initial.title,
      body: initial.body,
      tags: initial.tags.join(', '),
      lang: initial.lang,
      snippetLanguage: '',
      snippetCode: '',
    },
    onSubmitInvalid: focusFirstInvalid,
    onSubmit: async ({ value }) => {
      clear();
      setServerErrors({});
      try {
        await onSubmit({
          title: value.title.trim(),
          body_md: value.body.trim(),
          lang: value.lang,
          tags: parseTags(value.tags),
          snippets: snippetsOf(value.snippetLanguage, value.snippetCode),
          links,
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

  const addLink = () => {
    if (!isValidLinkId(linkId)) {
      setLinkInvalid(true);
      return;
    }
    const id = linkId.trim().toLowerCase();
    setLinks((current) =>
      current.some((link) => link.id === id) ? current : [...current, { type: linkType, id }],
    );
    setLinkId('');
    setLinkInvalid(false);
  };

  const notice = failure ? (failureNotice?.(failure.error) ?? null) : null;

  return (
    <VStack gap={4}>
      {failure ? (notice ?? <FailureNotice key={failure.n} error={failure.error} />) : null}
      <form
        method="post"
        ref={formRef}
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          void form.handleSubmit();
        }}
      >
        <VStack gap={4}>
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
                  label={t('solve.form.title')}
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
                  label={t('solve.form.body')}
                  description={t('solve.form.body.hint')}
                  rows={8}
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
                  label={t('solve.form.tags')}
                  description={t('solve.form.tags.hint')}
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
          {mode === 'create' ? (
            <>
              <form.Field name="lang">
                {(field) => (
                  <SegmentedControl
                    label={t('solve.form.lang')}
                    value={field.state.value}
                    onChange={(value) => {
                      field.handleChange(value as Lang);
                    }}
                  >
                    <SegmentedControlItem value="fr" label={t('solve.lang.fr')} />
                    <SegmentedControlItem value="en" label={t('solve.lang.en')} />
                  </SegmentedControl>
                )}
              </form.Field>
              <VStack gap={2}>
                <Text as="p" type="supporting">
                  {t('solve.form.snippet.hint')}
                </Text>
                <form.Field
                  name="snippetLanguage"
                  validators={{
                    onSubmit: ({ value, fieldApi }) =>
                      checkSnippetLanguage(value, fieldApi.form.getFieldValue('snippetCode')),
                  }}
                >
                  {(field) => {
                    const key = field.state.meta.errors[0] ?? serverErrors.snippetLanguage;
                    return (
                      <TextInput
                        label={t('solve.form.snippet.language')}
                        description={t('solve.form.snippet.language.hint')}
                        isOptional
                        value={field.state.value}
                        htmlName={field.name}
                        autoCapitalize="none"
                        hasAutoCorrect={false}
                        hasSpellCheck={false}
                        onBlur={field.handleBlur}
                        onChange={(value) => {
                          forget('snippetLanguage');
                          field.handleChange(value);
                        }}
                        {...errorStatus(key ? t(key) : undefined)}
                      />
                    );
                  }}
                </form.Field>
                <form.Field
                  name="snippetCode"
                  validators={{ onSubmit: ({ value }) => checkSnippetCode(value) }}
                >
                  {(field) => {
                    const key = field.state.meta.errors[0] ?? serverErrors.snippetCode;
                    return (
                      <TextArea
                        label={t('solve.form.snippet.code')}
                        isOptional
                        rows={6}
                        value={field.state.value}
                        htmlName={field.name}
                        hasSpellCheck={false}
                        onBlur={field.handleBlur}
                        onChange={(value) => {
                          forget('snippetCode');
                          field.handleChange(value);
                        }}
                        {...errorStatus(key ? t(key) : undefined)}
                      />
                    );
                  }}
                </form.Field>
              </VStack>
              <VStack gap={2}>
                <SegmentedControl
                  label={t('solve.form.links.type')}
                  value={linkType}
                  onChange={(value) => {
                    setLinkType(value as LinkType);
                  }}
                >
                  {linkTypes.map((type) => (
                    <SegmentedControlItem key={type} value={type} label={t(`solve.link.${type}`)} />
                  ))}
                </SegmentedControl>
                <HStack gap={2} vAlign="end" wrap="nowrap">
                  <TextInput
                    label={t('solve.form.links.id')}
                    isOptional
                    value={linkId}
                    autoCapitalize="none"
                    hasAutoCorrect={false}
                    hasSpellCheck={false}
                    width="100%"
                    onChange={(value) => {
                      setLinkInvalid(false);
                      setLinkId(value);
                    }}
                    {...errorStatus(
                      linkInvalid || serverErrors.link ? t('solve.err.link') : undefined,
                    )}
                  />
                  <Button label={t('solve.form.links.add')} onClick={addLink} />
                </HStack>
                {links.length > 0 ? (
                  <VStack as="ul" gap={1}>
                    {links.map((link) => (
                      <li key={link.id}>
                        <HStack gap={2} vAlign="center">
                          <Text>{`${t(`solve.link.${link.type}`)} ${link.id}`}</Text>
                          <IconButton
                            size="sm"
                            variant="ghost"
                            label={t('solve.form.links.remove', { name: link.id })}
                            icon={<Icon icon="close" />}
                            onClick={() => {
                              setLinks((current) => current.filter((item) => item.id !== link.id));
                            }}
                          />
                        </HStack>
                      </li>
                    ))}
                  </VStack>
                ) : null}
              </VStack>
            </>
          ) : null}
          <form.Subscribe selector={(state) => state.isSubmitting}>
            {(isSubmitting) => (
              <HStack gap={2}>
                <Button
                  type="submit"
                  variant="primary"
                  label={submitLabel}
                  isLoading={isSubmitting}
                />
                <Button variant="ghost" label={t('solve.form.cancel')} onClick={onCancel} />
              </HStack>
            )}
          </form.Subscribe>
        </VStack>
      </form>
    </VStack>
  );
}
