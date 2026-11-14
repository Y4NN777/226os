-- Demo fixtures 226os: learn (articles, article tags, series, lessons).
-- Contract: db/seeds/README.md. Deterministic, idempotent, local-dev only.
-- No BEGIN/COMMIT: make seed wraps every file in one psql -1 transaction.

-- 4. learn_articles
INSERT INTO learn_articles (id, author_id, kind, title, summary, body_md, language, level, topic_code, status, provenance, published_at, removed_at, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000701', '01900000-0000-7000-8000-000000000003', 'tutorial', 'Comprendre un webhook',
   'Un webhook est un appel HTTP envoyé par un service lorsqu''un événement se produit. Cette leçon explique le principe, le corps JSON, l''accusé de réception et la livraison au moins une fois.',
   $md$
Un webhook est un appel HTTP qu'un service distant envoie à votre application lorsqu'un événement se produit. Au lieu d'interroger l'API toutes les minutes pour savoir si quelque chose a changé, vous exposez une URL et le service vous prévient au bon moment.

## Le principe

Vous déclarez une URL de réception auprès du fournisseur, par exemple `https://example.org/votre-service/hooks/paiement`. Lorsqu'un événement survient (paiement reçu, message livré), le fournisseur envoie une requête `POST` vers cette URL avec un corps JSON qui décrit l'événement.

Un corps typique ressemble à ceci :

```json
{
  "id": "evt_0001",
  "type": "paiement.reussi",
  "created_at": "2026-11-10T09:00:00Z",
  "data": { "montant": 5000, "devise": "XOF" }
}
```

## Ce que votre point d'entrée doit faire

1. Lire le corps brut sans le modifier.
2. Vérifier que la requête vient bien du fournisseur (leçon 2).
3. Accuser réception rapidement avec un code `2xx`.
4. Traiter l'événement en arrière-plan, sans bloquer la réponse.

Le fournisseur ne lit pas votre logique métier : il regarde surtout le code de statut. Une réponse trop lente ou en erreur déclenche souvent une nouvelle tentative.

## Livraison au moins une fois

La plupart des services garantissent qu'un événement arrive au moins une fois, et non exactement une fois. Le même événement peut donc être reçu deux fois, par exemple après un délai d'attente réseau. Gardez cela en tête : la leçon 3 montre comment rendre le traitement idempotent.

## Pourquoi c'est utile

Sur une connexion instable, le webhook évite de multiplier les appels inutiles. Votre application réagit quand il faut, et le fournisseur gère les tentatives. En contrepartie, votre point d'entrée doit être robuste, vérifiable et observable.
$md$,
   'fr', 'beginner', 'api-integration', 'published', 'demo', '2026-11-10T09:00:00Z', NULL, '2026-11-10T09:00:00Z', '2026-11-10T09:00:00Z'),
  ('01900000-0000-7000-8000-000000000702', '01900000-0000-7000-8000-000000000003', 'tutorial', 'Vérifier une signature',
   'La signature prouve qu''un webhook vient bien du fournisseur et n''a pas été modifié en route. Cette leçon montre HMAC, la conservation du corps brut et la comparaison à temps constant.',
   $md$
Recevoir un webhook ne suffit pas : n'importe qui connaissant votre URL peut envoyer une requête. La signature est le mécanisme qui prouve que le message vient bien du fournisseur et qu'il n'a pas été modifié en route.

## Une signature, c'est un code d'authentification

Le fournisseur et vous partagez un secret, une longue chaîne aléatoire que vous stockez côté serveur et qui ne circule jamais dans le navigateur. Pour chaque envoi, le fournisseur calcule une empreinte du corps du message avec ce secret, par exemple avec HMAC-SHA256. Il place le résultat dans un en-tête, comme `X-Signature`.

À la réception, vous recalculez la même empreinte avec votre copie du secret. Si les deux correspondent, le message est authentique.

## Le piège du corps brut

La signature porte sur les octets exacts du corps. Si votre framework analyse le JSON puis le réécrit, l'empreinte change et la vérification échoue. Il faut donc conserver le corps brut avant toute transformation, puis le passer tel quel au calcul.

```text
signature_attendue = HMAC_SHA256(secret, corps_brut)
```

## Comparer sans fuite

N'utilisez pas une comparaison naïve caractère par caractère : elle peut révéler, par son temps d'exécution, combien de caractères sont corrects. Utilisez la fonction de comparaison à temps constant fournie par votre langage.

## Que faire en cas d'échec

Si la signature est absente ou invalide, répondez `401` et n'allez pas plus loin. Journalisez l'événement sans exposer le secret ni le corps sensible.

## Renouveler le secret

Prévoyez une procédure de rotation : acceptez temporairement deux secrets, basculez le fournisseur, puis retirez l'ancien. Cela évite une coupure lors du changement.
$md$,
   'fr', 'intermediate', 'api-integration', 'published', 'demo', '2026-11-11T09:00:00Z', NULL, '2026-11-11T09:00:00Z', '2026-11-11T09:00:00Z'),
  ('01900000-0000-7000-8000-000000000703', '01900000-0000-7000-8000-000000000003', 'tutorial', 'Rendre le traitement idempotent',
   'Un même événement peut arriver plusieurs fois. Cette leçon explique la clé d''idempotence, la table de déduplication et l''atomicité nécessaires pour éviter les doublons.',
   $md$
Un webhook peut être livré plusieurs fois. Un paiement ne doit pourtant être enregistré qu'une seule fois. Rendre le traitement idempotent signifie que rejouer le même événement produit le même résultat qu'une exécution unique.

## Le problème concret

Imaginez un événement `paiement.reussi` reçu une première fois, traité, puis renvoyé par le fournisseur parce que votre réponse a mis trop de temps. Sans précaution, vous créditez le compte deux fois.

## La clé de l'événement

Chaque événement possède un identifiant stable, souvent un champ `id` unique. C'est votre clé d'idempotence. Avant de traiter, vérifiez si cet identifiant a déjà été enregistré.

Une table de déduplication suffit :

```sql
CREATE TABLE evenements_traites (
  id text PRIMARY KEY,
  recu_at timestamptz NOT NULL
);
```

Tentez d'insérer l'identifiant dans la même transaction que l'effet métier. Si l'insertion viole la clé primaire, l'événement a déjà été traité : ignorez-le et renvoyez `2xx`.

## La transaction comme garde-fou

Le point clé est l'atomicité : l'insertion de l'identifiant et la mise à jour métier doivent réussir ou échouer ensemble. Une base relationnelle le garantit si les deux opérations partagent la même transaction.

## Attention aux effets externes

Si le traitement envoie un SMS ou un courriel, l'idempotence en base ne suffit pas. Enregistrez l'état avant l'envoi, et rendez l'envoi lui-même conditionnel à une transition d'état.

## Résumé

- Identifiant d'événement stable comme clé.
- Déduplication dans la même transaction que l'effet.
- Réponse `2xx` même pour un doublon, afin d'arrêter les tentatives.

Ainsi, une nouvelle livraison ne change plus rien.
$md$,
   'fr', 'intermediate', 'api-integration', 'published', 'demo', '2026-11-12T09:00:00Z', NULL, '2026-11-12T09:00:00Z', '2026-11-12T09:00:00Z'),
  ('01900000-0000-7000-8000-000000000704', '01900000-0000-7000-8000-000000000003', 'tutorial', 'Tester et rejouer',
   'Tester une intégration webhook demande un récepteur local, des fixtures versionnées et une file de rejeu. Cette leçon montre comment vérifier et rejouer sans créer de doublon.',
   $md$
Une intégration webhook se vérifie en rejouant des événements. Tester et rejouer permet de corriger un traitement sans attendre un nouvel incident réel.

## Un récepteur local

Commencez par une petite application locale qui expose une route de réception. Elle journalise chaque requête avec son en-tête de signature et le corps brut. Vous pouvez ainsi observer ce que le fournisseur envoie réellement.

## Des fixtures fidèles

Conservez quelques corps d'événements réels, anonymisés, dans des fichiers de test. Rejouer ces fixtures donne des tests déterministes : même entrée, même résultat attendu. Rangez-les à côté du code et versionnez-les.

## Tester la signature

Un bon test vérifie trois cas : une signature valide est acceptée, une signature absente est refusée, une signature altérée est refusée. Un quatrième cas utile modifie un octet du corps en gardant la signature, pour confirmer que la vérification porte bien sur le contenu.

## Rejouer après un incident

Lorsqu'un traitement échoue, gardez le corps brut en file. Une fois le correctif déployé, rejouez la file. Comme le traitement est idempotent, rejouer un événement déjà traité ne provoque pas de doublon.

## Ce qu'il faut mesurer

- le temps de réponse à l'accusé de réception ;
- le nombre de doublons détectés ;
- les signatures invalides, qui peuvent signaler une mauvaise configuration.

## En résumé

Un récepteur local, des fixtures versionnées et une file de rejeu transforment une intégration fragile en composant testable. C'est la dernière étape de la série.
$md$,
   'fr', 'intermediate', 'api-integration', 'published', 'demo', '2026-11-13T09:00:00Z', NULL, '2026-11-13T09:00:00Z', '2026-11-13T09:00:00Z')
ON CONFLICT (id) DO NOTHING;

-- 5. learn_article_tags
INSERT INTO learn_article_tags (article_id, tag_id) VALUES
  ('01900000-0000-7000-8000-000000000701', '01900000-0000-7000-8000-000000000108'),
  ('01900000-0000-7000-8000-000000000701', '01900000-0000-7000-8000-000000000107'),
  ('01900000-0000-7000-8000-000000000702', '01900000-0000-7000-8000-000000000108'),
  ('01900000-0000-7000-8000-000000000702', '01900000-0000-7000-8000-000000000109'),
  ('01900000-0000-7000-8000-000000000703', '01900000-0000-7000-8000-000000000108'),
  ('01900000-0000-7000-8000-000000000703', '01900000-0000-7000-8000-000000000109'),
  ('01900000-0000-7000-8000-000000000703', '01900000-0000-7000-8000-000000000107'),
  ('01900000-0000-7000-8000-000000000704', '01900000-0000-7000-8000-000000000108'),
  ('01900000-0000-7000-8000-000000000704', '01900000-0000-7000-8000-000000000107')
ON CONFLICT DO NOTHING;

-- 6. training_series
INSERT INTO training_series (id, author_id, origin, title, summary, language, level, topic_code, status, provenance, published_at, removed_at, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000801', '01900000-0000-7000-8000-000000000003', 'platform', 'Premiers pas avec les webhooks',
   'Quatre leçons pour intégrer sereinement les webhooks : comprendre l''événement, vérifier la signature, rendre le traitement idempotent, puis tester et rejouer. Un parcours court pour les développeuses et développeurs qui branchent une API de paiement ou de messagerie sur une connexion instable.',
   'fr', 'beginner', 'api-integration', 'published', 'demo', '2026-11-13T18:00:00Z', NULL, '2026-11-13T18:00:00Z', '2026-11-13T18:00:00Z')
ON CONFLICT (id) DO NOTHING;

-- 7. training_lessons
INSERT INTO training_lessons (id, series_id, position, title, body_md, language, article_id, status, removed_at, created_at, updated_at) VALUES
  ('01900000-0000-7000-8000-000000000811', '01900000-0000-7000-8000-000000000801', 1, 'Comprendre un webhook', NULL, 'fr', '01900000-0000-7000-8000-000000000701', 'published', NULL, '2026-11-13T18:10:00Z', '2026-11-13T18:10:00Z'),
  ('01900000-0000-7000-8000-000000000812', '01900000-0000-7000-8000-000000000801', 2, 'Vérifier une signature', NULL, 'fr', '01900000-0000-7000-8000-000000000702', 'published', NULL, '2026-11-13T18:11:00Z', '2026-11-13T18:11:00Z'),
  ('01900000-0000-7000-8000-000000000813', '01900000-0000-7000-8000-000000000801', 3, 'Rendre le traitement idempotent', NULL, 'fr', '01900000-0000-7000-8000-000000000703', 'published', NULL, '2026-11-13T18:12:00Z', '2026-11-13T18:12:00Z'),
  ('01900000-0000-7000-8000-000000000814', '01900000-0000-7000-8000-000000000801', 4, 'Tester et rejouer', NULL, 'fr', '01900000-0000-7000-8000-000000000704', 'published', NULL, '2026-11-13T18:13:00Z', '2026-11-13T18:13:00Z')
ON CONFLICT (id) DO NOTHING;
