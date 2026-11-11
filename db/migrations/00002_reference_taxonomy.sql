-- Migration: 00002_reference_taxonomy
-- Purpose: Create the reference and taxonomy tables and load the P0 reference rows (countries, learn topics, resource categories).
-- Tables/indexes: countries, tags (+ tags_slug_trgm_idx), learn_topics, resource_categories. Reference rows are REFERENCE DATA, not demo data, all owner-approved (owner decision, 2026-11-11): countries = the 54 African states (ISO 3166-1 alpha-2, French and English names); resource_categories = 15 rows (sort_order step 10) and learn_topics = 9 rows, lists researched by the lead agent and approved by the owner. Later additions or changes are new forward-only migrations.
-- Lock risk: none (empty database, new tables only).
-- Data volume: 54 countries, 9 learn topics, 15 resource categories; tags starts empty.
-- Transaction: single transaction (goose default).
-- Compatibility window: none (new tables). tags.created_by gets its FK to users in 00003.
-- Restore plan: forward-only. On an empty database, recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.1. Forward-only: no Down section.
-- +goose Up
CREATE TABLE countries (
  code    text PRIMARY KEY,
  name_fr text NOT NULL,
  name_en text NOT NULL,
  CONSTRAINT countries_code_ck CHECK (code ~ '^[A-Z]{2}$'),
  CONSTRAINT countries_name_fr_ck CHECK (char_length(name_fr) BETWEEN 1 AND 100),
  CONSTRAINT countries_name_en_ck CHECK (char_length(name_en) BETWEEN 1 AND 100)
);

CREATE TABLE tags (
  id         uuid PRIMARY KEY DEFAULT uuidv7(),
  slug       text NOT NULL,
  label      text NOT NULL,
  status     text NOT NULL DEFAULT 'active',
  provenance text NOT NULL DEFAULT 'live',
  created_by uuid,                             -- FK to users added in 00003 (users is defined later)
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT tags_slug_uq UNIQUE (slug),
  CONSTRAINT tags_slug_ck CHECK (slug ~ '^[a-z0-9][a-z0-9.+#-]{0,31}$'),
  CONSTRAINT tags_label_ck CHECK (char_length(label) BETWEEN 1 AND 60),
  CONSTRAINT tags_status_ck CHECK (status IN ('active','hidden')),
  CONSTRAINT tags_provenance_ck CHECK (provenance IN ('live','demo'))
);
CREATE INDEX tags_slug_trgm_idx ON tags USING gin (slug gin_trgm_ops);   -- tag autocomplete (bounded, slug <= 32)

CREATE TABLE learn_topics (
  code    text PRIMARY KEY,
  name_fr text NOT NULL,
  name_en text NOT NULL,
  CONSTRAINT learn_topics_code_ck CHECK (code ~ '^[a-z][a-z0-9-]{1,29}$')
);

CREATE TABLE resource_categories (
  code       text PRIMARY KEY,
  name_fr    text NOT NULL,
  name_en    text NOT NULL,
  sort_order smallint NOT NULL DEFAULT 0,
  CONSTRAINT resource_categories_code_ck CHECK (code ~ '^[a-z][a-z0-9-]{1,29}$')
);

-- Reference rows (deterministic, not demo data).
-- The 54 African states (ISO 3166-1 alpha-2), ordered by code. Western Sahara (EH)
-- is not a UN member state and is not included.
INSERT INTO countries (code, name_fr, name_en) VALUES
  ('AO', 'Angola', 'Angola'),
  ('BF', 'Burkina Faso', 'Burkina Faso'),
  ('BI', 'Burundi', 'Burundi'),
  ('BJ', 'Bénin', 'Benin'),
  ('BW', 'Botswana', 'Botswana'),
  ('CD', 'République démocratique du Congo', 'Democratic Republic of the Congo'),
  ('CF', 'République centrafricaine', 'Central African Republic'),
  ('CG', 'République du Congo', 'Republic of the Congo'),
  ('CI', 'Côte d’Ivoire', 'Côte d’Ivoire'),
  ('CM', 'Cameroun', 'Cameroon'),
  ('CV', 'Cap-Vert', 'Cabo Verde'),
  ('DJ', 'Djibouti', 'Djibouti'),
  ('DZ', 'Algérie', 'Algeria'),
  ('EG', 'Égypte', 'Egypt'),
  ('ER', 'Érythrée', 'Eritrea'),
  ('ET', 'Éthiopie', 'Ethiopia'),
  ('GA', 'Gabon', 'Gabon'),
  ('GH', 'Ghana', 'Ghana'),
  ('GM', 'Gambie', 'Gambia'),
  ('GN', 'Guinée', 'Guinea'),
  ('GQ', 'Guinée équatoriale', 'Equatorial Guinea'),
  ('GW', 'Guinée-Bissau', 'Guinea-Bissau'),
  ('KE', 'Kenya', 'Kenya'),
  ('KM', 'Comores', 'Comoros'),
  ('LR', 'Libéria', 'Liberia'),
  ('LS', 'Lesotho', 'Lesotho'),
  ('LY', 'Libye', 'Libya'),
  ('MA', 'Maroc', 'Morocco'),
  ('MG', 'Madagascar', 'Madagascar'),
  ('ML', 'Mali', 'Mali'),
  ('MR', 'Mauritanie', 'Mauritania'),
  ('MU', 'Maurice', 'Mauritius'),
  ('MW', 'Malawi', 'Malawi'),
  ('MZ', 'Mozambique', 'Mozambique'),
  ('NA', 'Namibie', 'Namibia'),
  ('NE', 'Niger', 'Niger'),
  ('NG', 'Nigéria', 'Nigeria'),
  ('RW', 'Rwanda', 'Rwanda'),
  ('SC', 'Seychelles', 'Seychelles'),
  ('SD', 'Soudan', 'Sudan'),
  ('SL', 'Sierra Leone', 'Sierra Leone'),
  ('SN', 'Sénégal', 'Senegal'),
  ('SO', 'Somalie', 'Somalia'),
  ('SS', 'Soudan du Sud', 'South Sudan'),
  ('ST', 'Sao Tomé-et-Principe', 'São Tomé and Príncipe'),
  ('SZ', 'Eswatini', 'Eswatini'),
  ('TD', 'Tchad', 'Chad'),
  ('TG', 'Togo', 'Togo'),
  ('TN', 'Tunisie', 'Tunisia'),
  ('TZ', 'Tanzanie', 'Tanzania'),
  ('UG', 'Ouganda', 'Uganda'),
  ('ZA', 'Afrique du Sud', 'South Africa'),
  ('ZM', 'Zambie', 'Zambia'),
  ('ZW', 'Zimbabwe', 'Zimbabwe');

-- Owner-approved list (owner decision, 2026-11-11; researched by the lead agent).
INSERT INTO learn_topics (code, name_fr, name_en) VALUES
  ('ai', 'IA appliquée et langues africaines', 'Applied AI & African languages'),
  ('payments', 'Paiements et Mobile Money', 'Payments & mobile money'),
  ('ussd-messaging', 'USSD, SMS et WhatsApp', 'USSD, SMS & WhatsApp'),
  ('mobile', 'Mobile, hors connexion et faible débit', 'Mobile, offline & low bandwidth'),
  ('api-integration', 'Intégration d’API et webhooks', 'API integration & webhooks'),
  ('security-privacy', 'Sécurité, fraude et données personnelles', 'Security, fraud & personal data'),
  ('cloud-devops', 'Déploiement, cloud et souveraineté des données', 'Deployment, cloud & data residency'),
  ('geo-data', 'Données, géospatial et open data', 'Data, geospatial & open data'),
  ('open-source', 'Open source et communs numériques', 'Open source & digital public goods');

-- Owner-approved list (owner decision, 2026-11-11; researched by the lead agent); sort_order step 10 in list order.
INSERT INTO resource_categories (code, name_fr, name_en, sort_order) VALUES
  ('mobile-money', 'Paiements Mobile Money', 'Mobile money payments', 10),
  ('online-checkout', 'Paiement en ligne et cartes', 'Online & card checkout', 20),
  ('cross-border', 'Transferts transfrontaliers et change', 'Cross-border payouts & FX', 30),
  ('open-finance', 'Open banking, comptabilité et crédit', 'Open banking, ledgers & lending', 40),
  ('e-invoicing', 'Facturation électronique et fiscalité', 'E-invoicing & tax', 50),
  ('identity-kyc', 'Identité, KYC et anti-fraude', 'Identity, KYC & fraud prevention', 60),
  ('sms-ussd', 'SMS, USSD et recharges', 'SMS, USSD & airtime', 70),
  ('chat-messaging', 'WhatsApp et messageries', 'WhatsApp & chat apps', 80),
  ('maps-addressing', 'Cartes, adressage et géodonnées', 'Maps, addressing & geodata', 90),
  ('logistics', 'Logistique et livraison', 'Logistics & delivery', 100),
  ('open-gov', 'Données et services publics', 'Open data & public services', 110),
  ('energy-climate', 'Énergie, climat et agriculture', 'Energy, climate & agriculture', 120),
  ('ai-language', 'IA et langues africaines', 'AI & African languages', 130),
  ('cloud-hosting', 'Hébergement et cloud en Afrique', 'Hosting & cloud in Africa', 140),
  ('dev-tools', 'Autres outils pour développeurs', 'Other developer tools', 150);
