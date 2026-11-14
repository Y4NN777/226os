import type { components } from '../../api/schema';

type Schemas = components['schemas'];

/**
 * Typed client-side samples for areas whose endpoints still answer 501 (ROADMAP C15).
 * Every list is wrapped with `source: 'fixture'` so it can never be mistaken for API data.
 * Authors are `demo-` accounts; no vote or answer counts are ever displayed.
 */
export interface Fixture<T> {
  source: 'fixture';
  items: readonly T[];
}

const at = '2026-11-01T09:00:00Z';

function author(name: string, id: string): Schemas['AuthorRef'] {
  return { id, username: `demo-${name}`, display_name: `demo-${name}` };
}

const kofi = author('kofi', '00000000-0000-4000-8000-00000000d001');
const aminata = author('aminata', '00000000-0000-4000-8000-00000000d002');

export const articleFixture: Fixture<Schemas['ArticleSummary']> = {
  source: 'fixture',
  items: [
    {
      id: '00000000-0000-4000-8000-0000000a0001',
      type: 'tutorial',
      title: 'Recevoir un webhook de paiement sans le traiter deux fois',
      summary: 'Vérifier la signature, enregistrer l’événement, répondre vite : le pas à pas.',
      lang: 'fr',
      level: 'intermediate',
      tags: ['webhooks', 'paiement'],
      author: kofi,
      published_at: at,
    },
    {
      id: '00000000-0000-4000-8000-0000000a0002',
      type: 'article',
      title: 'Écrire une API qui tient sur une connexion lente',
      summary: 'Pagination, compression et réponses courtes pour les réseaux mobiles.',
      lang: 'fr',
      level: 'beginner',
      tags: ['api', 'performance'],
      author: aminata,
      published_at: at,
    },
  ],
};

export const questionFixture: Fixture<Schemas['Question']> = {
  source: 'fixture',
  items: [
    {
      id: '00000000-0000-4000-8000-0000000b0001',
      title: 'Pourquoi mon appel CORS échoue-t-il uniquement en production ?',
      body_md: 'Exemple fictif.',
      lang: 'fr',
      tags: ['cors', 'go'],
      snippets: [],
      state: 'open',
      author: aminata,
      answer_count: 0,
      vote_score: 0,
      viewer_vote: 0,
      accepted_answer_id: null,
      related: [],
      created_at: at,
      updated_at: at,
    },
    {
      id: '00000000-0000-4000-8000-0000000b0002',
      title: 'Comment tester une migration sans toucher aux données réelles ?',
      body_md: 'Exemple fictif.',
      lang: 'fr',
      tags: ['postgresql', 'tests'],
      snippets: [],
      state: 'open',
      author: kofi,
      answer_count: 0,
      vote_score: 0,
      viewer_vote: 0,
      accepted_answer_id: null,
      related: [],
      created_at: at,
      updated_at: at,
    },
  ],
};

export const projectFixture: Fixture<Schemas['Project']> = {
  source: 'fixture',
  items: [
    {
      id: '00000000-0000-4000-8000-0000000c0001',
      name: 'demo-carnet-de-caisse',
      github_url: 'https://github.com/demo-226os/carnet-de-caisse',
      description: 'Application de suivi de ventes pour petits commerces, utilisable hors ligne.',
      lang: 'fr',
      technologies: ['react', 'typescript'],
      visibility: 'public',
      owner: kofi,
      open_needs_count: 0,
      related: [],
      created_at: at,
      updated_at: at,
    },
    {
      id: '00000000-0000-4000-8000-0000000c0002',
      name: 'demo-annuaire-marches',
      github_url: 'https://github.com/demo-226os/annuaire-marches',
      description: 'Carte des marchés et de leurs jours d’ouverture, avec une API ouverte.',
      lang: 'fr',
      technologies: ['go', 'postgresql'],
      visibility: 'public',
      owner: aminata,
      open_needs_count: 0,
      related: [],
      created_at: at,
      updated_at: at,
    },
  ],
};
