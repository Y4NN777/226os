// search catalogue, merged by src/i18n/messages.ts. Keys start with 'search.'.
// French typography rules are enforced by src/i18n/translate.test.ts.
export const fr = {
  'search.title': 'Recherche',
  'search.label': 'Votre recherche',
  'search.submit': 'Rechercher',
  'search.start.title': 'Que cherchez-vous ?',
  'search.start.body':
    'La recherche couvre le catalogue d’API et d’outils, les questions, les contenus à apprendre et les projets.',
  'search.short.title': 'Il faut un peu plus',
  'search.short.body': 'Saisissez au moins deux caractères.',
  'search.error.title': 'La recherche a échoué',
  'search.group.resources': 'Catalogue',
  'search.group.questions': 'Questions',
  'search.group.learn': 'Apprendre',
  'search.group.projects': 'Projets',
  'search.more': 'D’autres résultats existent. Affinez votre recherche pour les voir.',
  'search.none.title': 'Aucun résultat pour « {q} »',
  'search.none.body': 'Essayez d’autres mots, un terme plus court ou un sujet plus large.',
  'search.assistant.title': 'Assistant',
  'search.assistant.ai': 'IA',
  'search.assistant.notice':
    'L’assistant n’est pas encore en service. Une fois actif, votre question sera envoyée à un fournisseur d’IA pour générer la réponse.',
  'search.assistant.example': 'Voir un exemple de réponse',
  'search.assistant.verify': 'Comment vérifier',
  'search.assistant.sources': 'Sources',
} as const;

export const en: Record<keyof typeof fr, string> = {
  'search.title': 'Search',
  'search.label': 'Your search',
  'search.submit': 'Search',
  'search.start.title': 'What are you looking for?',
  'search.start.body':
    'Search covers the catalogue of APIs and tools, questions, learning content and projects.',
  'search.short.title': 'A little more, please',
  'search.short.body': 'Enter at least two characters.',
  'search.error.title': 'Search failed',
  'search.group.resources': 'Catalogue',
  'search.group.questions': 'Questions',
  'search.group.learn': 'Learn',
  'search.group.projects': 'Projects',
  'search.more': 'There are more results. Refine your search to see them.',
  'search.none.title': 'No results for “{q}”',
  'search.none.body': 'Try other words, a shorter term or a broader topic.',
  'search.assistant.title': 'Assistant',
  'search.assistant.ai': 'AI',
  'search.assistant.notice':
    'The assistant is not live yet. Once it is, your question will be sent to an AI provider to generate the answer.',
  'search.assistant.example': 'See a sample answer',
  'search.assistant.verify': 'How to verify',
  'search.assistant.sources': 'Sources',
};
