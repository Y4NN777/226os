//go:build integration

package pgtest

// Fixed identifiers (UUIDv7-shaped literals) for the minimal valid graph below.
// This is a test fixture, not seed data: it lives only inside rolled-back
// transactions or throwaway databases.
const idp = "018f0000-0000-7000-8000-"

const (
	uAlice = idp + "000000000001"
	uBob   = idp + "000000000002"
	uMod   = idp + "000000000003"

	tGo      = idp + "000000000011"
	tWebhook = idp + "000000000012"

	sAlice    = idp + "000000000021"
	evBob     = idp + "000000000022"
	resetAlic = idp + "000000000023"

	qAlice  = idp + "000000000031"
	qBob    = idp + "000000000032"
	qAccent = idp + "000000000033"
	qPlain  = idp + "000000000034"

	aBob   = idp + "000000000041" // answer by Bob on Alice's question
	aAlice = idp + "000000000042" // answer by Alice on Bob's question

	qsAlice = idp + "000000000051"
	asBob   = idp + "000000000052"

	pKal = idp + "000000000061"

	rKal   = idp + "000000000071"
	rSahel = idp + "000000000072"

	rsKal   = idp + "000000000081"
	rsSahel = idp + "000000000082"

	sdkKal = idp + "000000000091"

	erKal = idp + "000000000101"

	prAlice = idp + "000000000111"
	prBob   = idp + "000000000112"

	nAlice = idp + "000000000121"

	apBob = idp + "000000000131"

	arAlice = idp + "000000000141"

	seAlice = idp + "000000000151"

	leOne = idp + "000000000161"
	leTwo = idp + "000000000162"

	crOne = idp + "000000000171"

	uTmp = idp + "000000000901"
	qTmp = idp + "000000000902"
)

// NOTE: the fixture references reference codes learn_topics (payments, ai) and
// resource_categories (mobile-money, sms-ussd); update them if those lists change.
// K25: the series titled "Introduction à l'IA" stays on topic ai (Understanding LLMs).
// It is not African-language content, so it does not move to african-languages.
//
// fixtureGraph inserts one valid row in every P0 table (the three reference
// tables are filled by the migrations themselves).
const fixtureGraph = `
INSERT INTO users (id, email, email_verified_at, role) VALUES
  ('` + uAlice + `', 'alice@example.org', now(), 'member'),
  ('` + uBob + `', 'bob@example.org', now(), 'member'),
  ('` + uMod + `', 'mod@example.org', now(), 'moderator');
INSERT INTO profiles (user_id, username, display_name, intent, country_code, languages, github_username, bio, bio_language) VALUES
  ('` + uAlice + `', 'alice', 'Alice', 'help', 'BF', '{fr,en}', 'alice-gh', 'Backend dev', 'en'),
  ('` + uBob + `', 'bob', 'Bob', 'ask', 'GH', '{en}', NULL, NULL, NULL);
INSERT INTO user_credentials (user_id, password_hash) VALUES
  ('` + uAlice + `', '$argon2id$v=19$m=19456,t=2,p=1$c29tZXNhbHQ$aGFzaGhhc2g');
INSERT INTO tags (id, slug, label, created_by) VALUES
  ('` + tGo + `', 'go', 'Go', '` + uAlice + `'),
  ('` + tWebhook + `', 'webhook', 'Webhook', '` + uBob + `');
INSERT INTO user_skills (user_id, tag_id) VALUES ('` + uAlice + `', '` + tGo + `');
INSERT INTO sessions (id, user_id, token_hash, csrf_token_hash, idle_expires_at, absolute_expires_at, ip_prefix, user_agent_summary) VALUES
  ('` + sAlice + `', '` + uAlice + `', decode(repeat('ab', 32), 'hex'), decode(repeat('cd', 32), 'hex'),
   now() + interval '2 hours', now() + interval '7 days', '192.0.2.0/24', 'Firefox on Linux');
INSERT INTO email_verifications (id, user_id, email, token_hash, expires_at) VALUES
  ('` + evBob + `', '` + uBob + `', 'bob@example.org', decode(repeat('01', 32), 'hex'), now() + interval '24 hours');
INSERT INTO password_resets (id, user_id, token_hash, expires_at) VALUES
  ('` + resetAlic + `', '` + uAlice + `', decode(repeat('02', 32), 'hex'), now() + interval '30 minutes');
INSERT INTO rate_limit_buckets (scope, key_hash, window_start) VALUES
  ('signin:ip', decode(repeat('03', 32), 'hex'), date_trunc('minute', now()));

INSERT INTO questions (id, author_id, title, body_md, language, country_code) VALUES
  ('` + qAlice + `', '` + uAlice + `', 'Webhook signature fails', 'HMAC mismatch on retry', 'en', 'BF'),
  ('` + qBob + `', '` + uBob + `', 'Timeout USSD', 'La session expire', 'fr', 'GH');
INSERT INTO answers (id, question_id, author_id, body_md, language) VALUES
  ('` + aBob + `', '` + qAlice + `', '` + uBob + `', 'Compare against the raw body', 'en'),
  ('` + aAlice + `', '` + qBob + `', '` + uAlice + `', 'Augmenter le délai', 'fr');
INSERT INTO votes (user_id, target_type, question_id, value) VALUES ('` + uBob + `', 'question', '` + qAlice + `', 1);
INSERT INTO question_tags (question_id, tag_id) VALUES ('` + qAlice + `', '` + tGo + `');
INSERT INTO question_sources (id, question_id, title, url) VALUES ('` + qsAlice + `', '` + qAlice + `', 'Spec', 'https://example.org/spec');
INSERT INTO answer_sources (id, answer_id, title, url) VALUES ('` + asBob + `', '` + aBob + `', 'Doc', 'https://example.org/doc');

INSERT INTO providers (id, slug, name, website_url, created_by) VALUES
  ('` + pKal + `', 'kalimba-pay', 'Kalimba Pay (demo)', 'https://example.org', '` + uAlice + `');
INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, doc_languages, created_by) VALUES
  ('` + rKal + `', '` + pKal + `', 'mobile-money', 'api', 'kalimba-api', 'Kalimba API', 'Mobile money API', 'en', '{en,fr}', '` + uAlice + `'),
  ('` + rSahel + `', NULL, 'sms-ussd', 'sdk', 'sahel-sms', 'Sahel SMS', 'SMS gateway', 'en', '{}', '` + uAlice + `');
INSERT INTO resource_sources (id, resource_id, kind, title, url, checked_at, checked_by) VALUES
  ('` + rsKal + `', '` + rKal + `', 'official_docs', 'Docs', 'https://example.org/kalimba', now(), '` + uMod + `'),
  ('` + rsSahel + `', '` + rSahel + `', 'official_docs', 'Docs', 'https://example.org/sahel', now(), '` + uMod + `');
INSERT INTO resource_countries (resource_id, country_code) VALUES ('` + rKal + `', 'BF');
INSERT INTO resource_tags (resource_id, tag_id) VALUES ('` + rKal + `', '` + tGo + `');
INSERT INTO resource_sdk_languages (id, resource_id, language, name) VALUES ('` + sdkKal + `', '` + rKal + `', 'go', 'kalimba-go');
INSERT INTO error_reports (id, resource_id, reporter_id, kind) VALUES ('` + erKal + `', '` + rKal + `', '` + uBob + `', 'broken_link');

INSERT INTO projects (id, owner_id, name, summary, language, github_url, stage, country_code) VALUES
  ('` + prAlice + `', '` + uAlice + `', 'Webhook Kit', 'Verify webhook signatures', 'en', 'https://github.com/example/webhook-kit', 'active', 'BF'),
  ('` + prBob + `', '` + uBob + `', 'Offline drafts', 'Idea', 'fr', NULL, 'idea', NULL);
INSERT INTO project_maintainers (project_id, user_id, added_by) VALUES ('` + prAlice + `', '` + uBob + `', '` + uAlice + `');
INSERT INTO project_tags (project_id, tag_id) VALUES ('` + prAlice + `', '` + tGo + `');
INSERT INTO contribution_needs (id, project_id, title, description_md, language, created_by) VALUES
  ('` + nAlice + `', '` + prAlice + `', 'Docs', 'Write the quickstart', 'en', '` + uAlice + `');
INSERT INTO need_skills (need_id, tag_id) VALUES ('` + nAlice + `', '` + tGo + `');
INSERT INTO applications (id, contribution_need_id, user_id, message, language) VALUES
  ('` + apBob + `', '` + nAlice + `', '` + uBob + `', 'I can help', 'en');

INSERT INTO learn_articles (id, author_id, kind, title, summary, body_md, language, level, topic_code, status, published_at) VALUES
  ('` + arAlice + `', '` + uAlice + `', 'tutorial', 'Idempotence des webhooks', 'Résumé', 'Corps du tutoriel', 'fr', 'intermediate', 'payments', 'published', now());
INSERT INTO learn_article_tags (article_id, tag_id) VALUES ('` + arAlice + `', '` + tWebhook + `');
INSERT INTO training_series (id, author_id, origin, title, summary, language, level, topic_code) VALUES
  ('` + seAlice + `', '` + uAlice + `', 'platform', 'Introduction à l''IA', 'Série gratuite', 'fr', 'beginner', 'ai');
INSERT INTO training_lessons (id, series_id, position, title, body_md, language) VALUES
  ('` + leOne + `', '` + seAlice + `', 1, 'Leçon 1', 'Contenu', 'fr');

INSERT INTO content_relations (id, source_type, src_question_id, target_type, tgt_resource_id, created_by) VALUES
  ('` + crOne + `', 'question', '` + qAlice + `', 'resource', '` + rKal + `', '` + uAlice + `');

INSERT INTO audit_log (actor_id, actor_role, action, target_type, target_id, reason) VALUES
  ('` + uMod + `', 'moderator', 'error_report.resolve', 'report', '` + erKal + `', 'fixture');
INSERT INTO idempotency_keys (user_id, idempotency_key, method, route, request_hash, expires_at) VALUES
  ('` + uBob + `', 'key-0001-abcdef', 'POST', '/api/v1/needs/{id}/applications', decode(repeat('04', 32), 'hex'), now() + interval '24 hours');
`
