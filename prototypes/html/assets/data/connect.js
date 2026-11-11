/* Identity + Connect lane mock data (FICTIONAL). Read-only inputs: MOCK.people/questions/projects (shared).
   Bios and Connect entries are USER CONTENT: original language kept (`lang`), never translated by the UI switch.
   Counts/metrics are deliberately absent (no reputation, no endorsements). */
window.MOCK = window.MOCK || {};
window.MOCK.connect = {
  selfId: 'p-aminata',                       /* "own profile" when pages/profile.html has no ?id= */
  /* Per-field visibility (AUTHORIZATION_MATRIX: GET /profiles/{id} = visibility-dependent; email never public). */
  visibility: { name: 'public', handle: 'public', country: 'public', langs: 'public', skills: 'public', bio: 'public', github: 'public', openTo: 'public', since: 'public', city: 'members', email: 'self' },
  profiles: {
    'p-aminata': { bioLang: 'fr', bio: 'Je construis des services de paiement en Go. J’aime écrire de la documentation claire et accompagner les personnes qui débutent avec les webhooks.', github: 'https://example.org/gh/demo-aminata', email: 'aminata@example.org' },
    'p-kofi': { bioLang: 'en', bio: 'Full-stack engineer integrating mobile money providers. I write small, well-tested libraries and enjoy pairing on hard bugs.', github: 'https://example.org/gh/demo-kofi', email: 'kofi@example.org' },
    'p-fatou': { bioLang: 'fr', bio: 'Je conçois des applications mobiles pour réseaux instables, avec SMS et USSD comme canaux de secours.', github: 'https://example.org/gh/demo-fatou', email: 'fatou@example.org' },
    'p-chinedu': { bioLang: 'en', bio: 'Backend developer focused on KYC flows. Happy to review API designs and explain verification trade-offs.', github: 'https://example.org/gh/demo-chinedu', email: 'chinedu@example.org' },
    'p-wanjiru': { bioLang: 'en', bio: 'I maintain developer tooling and documentation. Looking for contributors who like writing and CI.', github: 'https://example.org/gh/demo-wanjiru', email: 'wanjiru@example.org' },
    'p-yves': { bioLang: 'fr', bio: 'Développeur logistique : cartographie, adresses informelles et PostGIS. Je partage mes modèles de données.', github: 'https://example.org/gh/demo-yves', email: 'yves@example.org' },
    'p-salif': { bioLang: 'fr', bio: 'Développeur PHP/Laravel, intégrations bancaires et Mobile Money en Côte d’Ivoire.', github: 'https://example.org/gh/demo-salif', email: 'salif@example.org' },
    'p-imane': { bioLang: 'fr', bio: 'Ingénieure DevOps : observabilité, durcissement et déploiements reproductibles avec Docker.', github: 'https://example.org/gh/demo-imane', email: 'imane@example.org' }
  },
  /* Activity links (ids from common.js). `answers` = questions the person answered; `contributions` = projects joined but not owned. */
  answers: {
    'p-aminata': ['q-idempotency', 'q-ussd-timeout'], 'p-kofi': ['q-webhook-sig'], 'p-fatou': ['q-sms-encoding'], 'p-chinedu': ['q-idempotency'],
    'p-wanjiru': ['q-docker-traefik'], 'p-yves': ['q-webhook-sig'], 'p-salif': [], 'p-imane': ['q-docker-traefik', 'q-webhook-sig']
  },
  contributions: {
    'p-aminata': ['pr-ussd-flow'], 'p-kofi': ['pr-webhook-kit'], 'p-fatou': ['pr-docs-fr'], 'p-chinedu': [], 'p-wanjiru': ['pr-geo-informal'],
    'p-yves': ['pr-webhook-kit'], 'p-salif': [], 'p-imane': ['pr-webhook-kit', 'pr-offline-drafts']
  },
  /* Connect preview (PRD 5.5): mentorship requests/offers, CV review, interview prep, study groups, opportunities.
     kind keys are UI strings in connect.js. Text = user content in its own language. */
  entries: [
    { id: 'c-1', kind: 'mentor-offer', person: 'p-aminata', lang: 'fr', title: 'Accompagnement sur les webhooks et la vérification de signatures', skills: ['go', 'webhook', 'docker'], country: 'BF', speaks: ['fr', 'en'] },
    { id: 'c-2', kind: 'mentor-request', person: 'p-salif', lang: 'fr', title: 'Je cherche un mentor pour structurer une API Laravel de paiement', skills: ['php', 'laravel', 'mobile-money'], country: 'CI', speaks: ['fr'] },
    { id: 'c-3', kind: 'cv-review', person: 'p-chinedu', lang: 'en', title: 'Offering CV and portfolio reviews for backend roles', skills: ['python', 'django', 'kyc'], country: 'NG', speaks: ['en'] },
    { id: 'c-4', kind: 'interview', person: 'p-wanjiru', lang: 'en', title: 'Mock technical interviews for open source contributors', skills: ['rust', 'go', 'ci'], country: 'KE', speaks: ['en'] },
    { id: 'c-5', kind: 'study-group', person: 'p-yves', lang: 'fr', title: 'Groupe d’étude PostgreSQL et PostGIS, sessions en français', skills: ['postgresql', 'postgis'], country: 'CM', speaks: ['fr', 'en'] },
    { id: 'c-6', kind: 'mentor-request', person: 'p-fatou', lang: 'fr', title: 'Besoin d’un retour sur l’architecture hors ligne d’une application Flutter', skills: ['flutter', 'sms', 'ussd'], country: 'SN', speaks: ['fr', 'en'] },
    { id: 'c-7', kind: 'mission', person: 'p-imane', lang: 'fr', title: 'Mission à distance : audit d’observabilité pour une petite équipe', skills: ['docker', 'prometheus', 'security'], country: 'MA', speaks: ['fr', 'en'] },
    { id: 'c-8', kind: 'mentor-offer', person: 'p-kofi', lang: 'en', title: 'Mentoring on React and TypeScript for mobile money dashboards', skills: ['react', 'typescript', 'node'], country: 'GH', speaks: ['en'] }
  ],
  /* Community spaces (PRD 5.4: P2) – names only, no threads, no member counts. */
  spaces: [
    { id: 'by-country', icon: 'globe' }, { id: 'by-tech', icon: 'code' }, { id: 'by-interest', icon: 'users' }
  ]
};
