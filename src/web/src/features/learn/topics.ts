import type { MessageKey } from '../../i18n/messages';

/**
 * The 14 learn topics (learn_topics.code), as seeded by migrations 00002 and 00011. The API has
 * no topic-list endpoint yet, so the codes live here; the server still rejects an unknown one (422).
 */
export const topicCodes = [
  'ai',
  'payments',
  'ussd-messaging',
  'mobile',
  'api-integration',
  'security-privacy',
  'cloud-devops',
  'geo-data',
  'open-source',
  'learn-to-learn',
  'ai-assisted-dev',
  'ai-agents',
  'open-models',
  'african-languages',
] as const;

export type TopicCode = (typeof topicCodes)[number];

export function topicKey(code: TopicCode): MessageKey {
  return `learn.topic.${code}`;
}
