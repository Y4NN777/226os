import { Card } from '@astryxdesign/core/Card';
import { CheckboxInput } from '@astryxdesign/core/CheckboxInput';
import { Heading } from '@astryxdesign/core/Heading';
import { HStack } from '@astryxdesign/core/HStack';
import { Icon } from '@astryxdesign/core/Icon';
import { IconButton } from '@astryxdesign/core/IconButton';
import { ProgressBar } from '@astryxdesign/core/ProgressBar';
import { RadioList, RadioListItem } from '@astryxdesign/core/RadioList';
import { Text } from '@astryxdesign/core/Text';
import { TextArea } from '@astryxdesign/core/TextArea';
import { TextInput } from '@astryxdesign/core/TextInput';
import { VStack } from '@astryxdesign/core/VStack';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { ApiError } from '../../api/http';
import { useI18n } from '../../i18n/I18nProvider';
import { FailureNotice } from '../identity/notices';
import { errorStatus } from '../identity/validation';
import { violationName } from '../identity/violations';
import { useMe } from '../settings/useMe';
import { useSaveProfileField } from './queries';
import { bodyOf, checkItem, checklistItems, intents, type ChecklistItem } from './validate';

function ItemForm({ item, username }: { item: ChecklistItem; username: string }) {
  const { t } = useI18n();
  const save = useSaveProfileField(username);
  const [serverInvalid, setServerInvalid] = useState(false);
  const [failure, setFailure] = useState<{ n: number; error: unknown } | null>(null);
  const label = t(`profile.item.${item}`);

  const form = useForm({
    defaultValues: { value: '' },
    onSubmit: async ({ value }) => {
      setServerInvalid(false);
      setFailure(null);
      try {
        await save.mutateAsync(bodyOf(item, value.value));
      } catch (error) {
        const onField =
          error instanceof ApiError &&
          error.violations.some(
            (violation) => violationName(violation.field).split('/')[0] === item,
          );
        if (onField) setServerInvalid(true);
        else setFailure((previous) => ({ n: (previous?.n ?? 0) + 1, error }));
      }
    },
  });

  const submit = (
    <form.Subscribe selector={(state) => state.isSubmitting}>
      {(isSubmitting) => (
        <IconButton
          type="submit"
          label={t('profile.item.save', { field: label })}
          icon={<Icon icon="check" />}
          isDisabled={isSubmitting}
        />
      )}
    </form.Subscribe>
  );

  return (
    <form
      noValidate
      onSubmit={(event) => {
        event.preventDefault();
        void form.handleSubmit();
      }}
    >
      <VStack gap={2}>
        <Text as="p" type="supporting">
          {t(`profile.item.${item}.value`)}
        </Text>
        <form.Field name="value" validators={{ onSubmit: ({ value }) => checkItem(item, value) }}>
          {(field) => {
            const key =
              field.state.meta.errors[0] ??
              (serverInvalid ? (`profile.err.${item}` as const) : undefined);
            const status = errorStatus(key ? t(key) : undefined);
            const change = (value: string) => {
              setServerInvalid(false);
              field.handleChange(value);
            };
            if (item === 'intent') {
              return (
                <HStack gap={2} vAlign="end" wrap="wrap">
                  <RadioList label={label} value={field.state.value} onChange={change}>
                    {intents.map((intent) => (
                      <RadioListItem
                        key={intent}
                        value={intent}
                        label={t(`profile.intent.${intent}`)}
                      />
                    ))}
                  </RadioList>
                  {submit}
                </HStack>
              );
            }
            if (item === 'open_to_collab') {
              return (
                <HStack gap={2} vAlign="end">
                  <CheckboxInput
                    label={label}
                    description={t('profile.item.open_to_collab.hint')}
                    value={field.state.value === 'true'}
                    onChange={(checked) => change(String(checked))}
                    {...status}
                  />
                  {submit}
                </HStack>
              );
            }
            if (item === 'bio') {
              return (
                <HStack gap={2} vAlign="end">
                  <TextArea
                    label={label}
                    description={t('profile.item.bio.hint')}
                    width="100%"
                    rows={3}
                    value={field.state.value}
                    htmlName={field.name}
                    onChange={change}
                    {...status}
                  />
                  {submit}
                </HStack>
              );
            }
            return (
              <HStack gap={2} vAlign="end">
                <TextInput
                  label={label}
                  description={t(`profile.item.${item}.hint`)}
                  width="100%"
                  value={field.state.value}
                  htmlName={field.name}
                  autoCapitalize="none"
                  hasAutoCorrect={false}
                  hasSpellCheck={false}
                  onChange={change}
                  {...status}
                />
                {submit}
              </HStack>
            );
          }}
        </form.Field>
        {failure ? <FailureNotice key={failure.n} error={failure.error} /> : null}
      </VStack>
    </form>
  );
}

/**
 * The member's own completion checklist: the server says which fields are still empty
 * (`checklist_missing`), the meter follows, each item is completed in place, and the card
 * is gone at 100 %. Mounted by the profile page only when the profile is the caller's own.
 */
export function Checklist({ username }: { username: string }) {
  const { t } = useI18n();
  const me = useMe();
  if (!me.data) return null;
  const missing = checklistItems.filter((item) => me.data.profile.checklist_missing.includes(item));
  if (missing.length === 0) return null;
  const percent = Math.round(
    ((checklistItems.length - missing.length) / checklistItems.length) * 100,
  );

  return (
    <section aria-label={t('profile.checklist.title')}>
      <Card variant="muted">
        <VStack gap={4}>
          <Heading level={2}>{t('profile.checklist.title')}</Heading>
          <ProgressBar label={t('profile.checklist.progress')} value={percent} hasValueLabel />
          <VStack as="ul" gap={4}>
            {missing.map((item) => (
              <li key={item}>
                <ItemForm item={item} username={username} />
              </li>
            ))}
          </VStack>
        </VStack>
      </Card>
    </section>
  );
}
