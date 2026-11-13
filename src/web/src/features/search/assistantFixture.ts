import type { components } from '../../api/schema';

/**
 * Sample reply shown by the assistant panel until the assistant is live (askAssistant
 * answers 501). Original text, generic advice only: no claim about any real provider.
 * Sources point at routes that exist in this app.
 */
export const assistantFixture: components['schemas']['AssistantAnswer'] = {
  outcome: 'answered',
  generated_by: 'ai',
  lang: 'fr',
  answer:
    'Avant d’intégrer une API de paiement, utilisez d’abord son environnement de test, s’il existe. Il permet d’essayer les appels et les notifications sans engager de vrais fonds. Vérifiez aussi dans la documentation du fournisseur quels paramètres sont obligatoires et comment signer ou authentifier les requêtes.',
  how_to_verify:
    'Lisez la documentation officielle du fournisseur et rejouez l’exemple dans l’environnement de test avant tout appel réel. Comparez cette réponse aux contenus cités plus bas, puis confirmez avec les personnes qui utilisent déjà cette API.',
  sources: [
    {
      type: 'resource',
      id: '7a1c2a34-0000-4000-8000-00000000f001',
      title: 'Catalogue des API',
      url: '/discover',
      published_at: null,
    },
    {
      type: 'article',
      id: '7a1c2a34-0000-4000-8000-00000000f002',
      title: 'Recherche : sandbox de paiement',
      url: '/recherche?q=sandbox%20paiement',
      published_at: null,
    },
  ],
};
