import { CodeBlock } from '@astryxdesign/core/CodeBlock';
import { Markdown } from '@astryxdesign/core/Markdown';
import { VStack } from '@astryxdesign/core/VStack';
import type { components } from '../../api/schema';

type Snippet = components['schemas']['Snippet'];

/**
 * User Markdown rendered by Astryx: raw HTML is never turned into elements and unsafe link
 * destinations become plain text. Snippets are shown as text in a code block with a copy
 * button, never executed.
 */
export function RichText({ markdown, snippets = [] }: { markdown: string; snippets?: Snippet[] }) {
  return (
    <VStack gap={3}>
      <Markdown headingLevelStart={3} contentWidth="100%">
        {markdown}
      </Markdown>
      {snippets.map((snippet, index) => (
        <CodeBlock
          key={index}
          code={snippet.code}
          language={snippet.language}
          hasCopyButton
          isWrapped={false}
        />
      ))}
    </VStack>
  );
}
