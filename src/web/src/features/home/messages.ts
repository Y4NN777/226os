// home catalogue, merged by src/i18n/messages.ts. Keys start with 'feed.'.
// French typography rules are enforced by src/i18n/translate.test.ts.
export const fr = {
  'feed.title': 'Accueil',
  'feed.hello': 'Bonjour, {name}',
  'feed.search.label': 'Rechercher dans 226os',
  'feed.search.submit': 'Rechercher',
  'feed.demo.note':
    'Exemples fictifs. Les contenus réels arriveront à l’ouverture de cette section.',
  'feed.by': 'par {name}',
  'feed.level.beginner': 'Débutant',
  'feed.level.intermediate': 'Intermédiaire',
  'feed.level.advanced': 'Avancé',
  'feed.learn.title': 'Apprendre',
  'feed.connect.title': 'Connect',
  'feed.connect.soon': 'Bientôt : l’annuaire des développeurs.',
  'feed.questions.title': 'Questions à résoudre',
  'feed.catalog.title': 'API du catalogue',
  'feed.catalog.provider': 'Fournisseur : {name}',
  'feed.catalog.more': 'Parcourir le catalogue',
  'feed.catalog.empty': 'Le catalogue est vide pour l’instant.',
  'feed.projects.title': 'Projets',
} as const;

export const en: Record<keyof typeof fr, string> = {
  'feed.title': 'Home',
  'feed.hello': 'Hello, {name}',
  'feed.search.label': 'Search 226os',
  'feed.search.submit': 'Search',
  'feed.demo.note': 'Made-up examples. Real content will appear when this section opens.',
  'feed.by': 'by {name}',
  'feed.level.beginner': 'Beginner',
  'feed.level.intermediate': 'Intermediate',
  'feed.level.advanced': 'Advanced',
  'feed.learn.title': 'Learn',
  'feed.connect.title': 'Connect',
  'feed.connect.soon': 'Soon: the developer directory.',
  'feed.questions.title': 'Questions to solve',
  'feed.catalog.title': 'Catalogue APIs',
  'feed.catalog.provider': 'Provider: {name}',
  'feed.catalog.more': 'Browse the catalogue',
  'feed.catalog.empty': 'The catalogue is empty for now.',
  'feed.projects.title': 'Projects',
};
