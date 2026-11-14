-- Free Learn series (S4 content script, B5 DECIDED, H19). Plain SQL, one-time, PR-reviewed; neither a
-- seed nor a migration. Run it only through scripts/content/run-free-series.sh, which asks for an explicit
-- confirmation and a backup first (staging-test, then production-live).
--
-- What it creates, in ONE statement (a DO block, so all or nothing):
--   one training_series (origin 'platform', topic 'ai') and 4 tutorial learn_articles referenced in order by
--   training_lessons.article_id, all provenance 'live', status 'published', language 'fr'.
-- Attribution: the owner's real account, looked up by username from the session setting
-- content.owner_username (never a hard-coded id). The account must exist and be active.
-- Refuses (nothing written) when the setting is missing, the account is unknown or inactive, or the series
-- already exists, so a second run changes nothing.
-- Text: original team text, written for 226os; review it in the PR before the first run.

DO $script$
DECLARE
  v_owner_name text := current_setting('content.owner_username', true);
  v_owner    uuid;
  v_series   uuid;
  v_article  uuid;
  v_tags     uuid[];
  lesson     record;
BEGIN
  IF v_owner_name IS NULL OR v_owner_name = '' THEN
    RAISE EXCEPTION 'content.owner_username is not set';
  END IF;
  SELECT u.id INTO v_owner
  FROM profiles p JOIN users u ON u.id = p.user_id
  WHERE p.username = v_owner_name AND u.status = 'active';
  IF v_owner IS NULL THEN
    RAISE EXCEPTION 'no active account with username %', v_owner_name;
  END IF;
  IF EXISTS (SELECT 1 FROM training_series WHERE origin = 'platform' AND title = 'Bien démarrer avec l''IA générative') THEN
    RAISE EXCEPTION 'the free series already exists; nothing changed';
  END IF;

  INSERT INTO tags (slug, label, created_by)
  SELECT s, s, v_owner FROM unnest(ARRAY['ia', 'llm', 'prompt', 'debutant']) AS s
  ON CONFLICT (slug) DO NOTHING;
  SELECT array_agg(id) INTO v_tags FROM tags WHERE slug IN ('ia', 'llm', 'prompt', 'debutant');

  INSERT INTO training_series (author_id, origin, title, summary, language, level, topic_code, provenance)
  VALUES (v_owner, 'platform', 'Bien démarrer avec l''IA générative',
          'Quatre leçons courtes et gratuites pour comprendre ce qu''est un modèle de langage, bien formuler une demande, vérifier une réponse et coder avec l''IA sans perdre la main.',
          'fr', 'beginner', 'ai', 'live')
  RETURNING id INTO v_series;

  FOR lesson IN
    SELECT * FROM (VALUES
      (1, 'Qu''est-ce qu''un grand modèle de langage ?',
       'Ce que fait vraiment un LLM, ce qu''il ne fait pas, et pourquoi il peut se tromper avec assurance.',
       $l1$# Qu'est-ce qu'un grand modèle de langage ?

Un grand modèle de langage (LLM, *large language model*) est un programme entraîné sur une très grande quantité de textes. À partir d'un début de texte, il prédit la suite la plus probable, mot après mot.

## Ce qu'il sait faire

- reformuler, résumer, traduire un texte ;
- expliquer une notion avec d'autres mots ;
- proposer un premier jet de code, de courriel ou de plan.

## Ce qu'il ne fait pas

- Il ne « sait » pas au sens humain : il produit un texte plausible, pas une vérité vérifiée.
- Il ne connaît pas les événements postérieurs à son entraînement, sauf si on lui fournit les informations.
- Il ne voit pas vos fichiers ni votre écran, sauf ce que vous lui donnez.

## À retenir

Une réponse bien écrite n'est pas forcément juste. Le modèle peut inventer une fonction, une loi ou une référence avec beaucoup d'assurance : on parle d'**hallucination**. La leçon 3 montre comment vérifier.
$l1$),
      (2, 'Bien formuler une demande',
       'Contexte, tâche, format attendu : une méthode simple pour obtenir des réponses utiles du premier coup.',
       $l2$# Bien formuler une demande

Le modèle ne connaît que ce que vous écrivez. Une demande vague donne une réponse vague.

## Une méthode en trois parties

1. **Le contexte** : qui vous êtes, sur quoi vous travaillez. *« Je développe une application mobile de paiement au Burkina Faso, en Kotlin. »*
2. **La tâche** : ce que vous voulez, précisément. *« Explique comment valider un numéro de téléphone local avant un paiement. »*
3. **Le format** : la forme de la réponse. *« Réponds en cinq étapes maximum, avec un exemple de code. »*

## Améliorer en dialoguant

La première réponse est un brouillon. Demandez une correction ciblée : *« Ton exemple ne gère pas les espaces dans le numéro, corrige-le. »*

## À éviter

Ne collez jamais un mot de passe, une clé d'API ou des données personnelles de vos utilisateurs dans une demande.
$l2$),
      (3, 'Vérifier une réponse de l''IA',
       'Des réflexes simples pour repérer une erreur avant qu''elle ne se retrouve dans votre travail.',
       $l3$# Vérifier une réponse de l'IA

Traitez chaque réponse comme le travail d'un stagiaire motivé : utile, mais à relire.

## Quatre réflexes

- **Remonter à la source** : pour une API, une loi ou un chiffre, cherchez la documentation officielle ou le texte d'origine.
- **Tester** : un extrait de code se vérifie en l'exécutant, avec un cas normal et un cas limite.
- **Poser la question autrement** : si deux formulations donnent deux réponses opposées, aucune n'est fiable telle quelle.
- **Demander les limites** : *« Dans quels cas ta réponse serait-elle fausse ? »* fait souvent apparaître les hypothèses cachées.

## Un signal d'alerte

Une référence précise (nom de fonction, numéro d'article, lien) que vous ne retrouvez nulle part est probablement inventée.
$l3$),
      (4, 'Coder avec l''IA sans perdre la main',
       'Utiliser un assistant pour aller plus vite tout en continuant à comprendre et à progresser.',
       $l4$# Coder avec l'IA sans perdre la main

Un assistant de code fait gagner du temps sur ce que vous savez déjà faire. Il devient un risque quand il écrit ce que vous ne comprenez pas.

## Une règle simple

Ne gardez que le code que vous pourriez expliquer ligne par ligne à un collègue.

## Une façon de travailler

1. Écrivez d'abord, en une phrase, ce que le code doit faire.
2. Demandez une proposition, puis lisez-la avant de l'exécuter.
3. Écrivez ou demandez un test qui échoue si le code est faux.
4. Demandez « pourquoi ? » sur chaque partie qui vous surprend.

## Pour continuer à progresser

Essayez régulièrement de résoudre un petit problème seul avant de demander. Comparez ensuite votre solution à celle de l'assistant : c'est là que l'on apprend le plus.
$l4$)
    ) AS t(position, title, summary, body)
    ORDER BY position
  LOOP
    INSERT INTO learn_articles (author_id, kind, title, summary, body_md, language, level, topic_code, status, provenance, published_at)
    VALUES (v_owner, 'tutorial', lesson.title, lesson.summary, lesson.body, 'fr', 'beginner', 'ai', 'published', 'live', now())
    RETURNING id INTO v_article;
    INSERT INTO learn_article_tags (article_id, tag_id) SELECT v_article, unnest(v_tags);
    INSERT INTO training_lessons (series_id, position, title, language, article_id)
    VALUES (v_series, lesson.position, lesson.title, 'fr', v_article);
  END LOOP;

  RAISE NOTICE 'free series % created for % with 4 lessons', v_series, v_owner_name;
END
$script$;
