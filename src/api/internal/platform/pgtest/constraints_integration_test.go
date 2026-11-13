//go:build integration

package pgtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

// SQLSTATE classes used below.
const (
	notNull    = "23502"
	fkViol     = "23503"
	uniqueViol = "23505"
	checkViol  = "23514"
	// ON DELETE RESTRICT reports restrict_violation, not foreign_key_violation:
	// repositories must map both 23001 and 23503.
	restrictViol = "23001"
)

type violation struct {
	name string
	// setup runs after the fixture graph, inside the same rolled-back transaction.
	setup []string
	// stmt must fail with code (and constraint, when not empty).
	stmt       string
	code       string
	constraint string
}

func TestConstraintViolations(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)

	cases := []violation{
		// identity
		{name: "duplicate email", stmt: `INSERT INTO users (email) VALUES ('alice@example.org')`, code: uniqueViol, constraint: "users_email_uq"},
		{name: "email with uppercase", stmt: `INSERT INTO users (email) VALUES ('Mixed@Example.org')`, code: checkViol, constraint: "users_email_normalized_ck"},
		{name: "email with padding", stmt: `INSERT INTO users (email) VALUES (' pad@example.org')`, code: checkViol, constraint: "users_email_normalized_ck"},
		{name: "email without at sign", stmt: `INSERT INTO users (email) VALUES ('nobody')`, code: checkViol, constraint: "users_email_normalized_ck"},
		{name: "unknown role", stmt: `INSERT INTO users (email, role) VALUES ('r@example.org', 'admin')`, code: checkViol, constraint: "users_role_ck"},
		{name: "unknown status", stmt: `INSERT INTO users (email, status) VALUES ('s@example.org', 'banned')`, code: checkViol, constraint: "users_status_ck"},
		{name: "deleted without deleted_at", stmt: `INSERT INTO users (email, status) VALUES ('d@example.org', 'deleted')`, code: checkViol, constraint: "users_deleted_ck"},
		{name: "unknown provenance", stmt: `INSERT INTO users (email, provenance) VALUES ('p@example.org', 'fake')`, code: checkViol, constraint: "users_provenance_ck"},
		{name: "credentials not argon2id", stmt: `INSERT INTO user_credentials (user_id, password_hash) VALUES ('` + uBob + `', 'plaintext')`, code: checkViol, constraint: "user_credentials_hash_ck"},
		{name: "credentials for unknown user", stmt: `INSERT INTO user_credentials (user_id, password_hash) VALUES ('` + idp + `0000000000ff', '$argon2id$x')`, code: fkViol, constraint: "user_credentials_user_fk"},
		{name: "duplicate username", stmt: `INSERT INTO profiles (user_id, username, display_name) VALUES ('` + uMod + `', 'alice', 'Mod')`, code: uniqueViol, constraint: "profiles_username_uq"},
		{name: "second profile for a user", stmt: `INSERT INTO profiles (user_id, username, display_name) VALUES ('` + uAlice + `', 'alice2', 'Alice')`, code: uniqueViol, constraint: "profiles_pkey"},
		{name: "unknown country", stmt: `INSERT INTO profiles (user_id, username, display_name, country_code) VALUES ('` + uMod + `', 'mod', 'Mod', 'ZZ')`, code: fkViol, constraint: "profiles_country_fk"},
		{name: "bad username", stmt: `INSERT INTO profiles (user_id, username, display_name) VALUES ('` + uMod + `', 'A!', 'Mod')`, code: checkViol, constraint: "profiles_username_ck"},
		{name: "bio without language", stmt: `INSERT INTO profiles (user_id, username, display_name, bio) VALUES ('` + uMod + `', 'mod', 'Mod', 'hi')`, code: checkViol, constraint: "profiles_bio_language_ck"},
		{name: "bad profile visibility", stmt: `INSERT INTO profiles (user_id, username, display_name, bio_visibility) VALUES ('` + uMod + `', 'mod', 'Mod', 'friends')`, code: checkViol, constraint: "profiles_visibility_ck"},
		{name: "unknown profile visibility", stmt: `INSERT INTO profiles (user_id, username, display_name, profile_visibility) VALUES ('` + uMod + `', 'mod', 'Mod', 'friends')`, code: checkViol, constraint: "profiles_profile_visibility_ck"},
		{name: "duplicate session token hash", stmt: `INSERT INTO sessions (user_id, token_hash, csrf_token_hash, idle_expires_at, absolute_expires_at)
			VALUES ('` + uBob + `', decode(repeat('ab', 32), 'hex'), decode(repeat('ee', 32), 'hex'), now() + interval '1 hour', now() + interval '1 day')`, code: uniqueViol, constraint: "sessions_token_hash_uq"},
		{name: "session hash of wrong length", stmt: `INSERT INTO sessions (user_id, token_hash, csrf_token_hash, idle_expires_at, absolute_expires_at)
			VALUES ('` + uBob + `', decode('abcd', 'hex'), decode(repeat('ee', 32), 'hex'), now() + interval '1 hour', now() + interval '1 day')`, code: checkViol, constraint: "sessions_token_hash_ck"},
		{name: "session idle after absolute expiry", stmt: `INSERT INTO sessions (user_id, token_hash, csrf_token_hash, idle_expires_at, absolute_expires_at)
			VALUES ('` + uBob + `', decode(repeat('ef', 32), 'hex'), decode(repeat('ee', 32), 'hex'), now() + interval '2 days', now() + interval '1 day')`, code: checkViol, constraint: "sessions_expiry_ck"},
		{name: "session revoked without reason", stmt: `UPDATE sessions SET revoked_at = now() WHERE id = '` + sAlice + `'`, code: checkViol, constraint: "sessions_revoked_ck"},
		{name: "session unknown revoke reason", stmt: `UPDATE sessions SET revoked_at = now(), revoked_reason = 'bored' WHERE id = '` + sAlice + `'`, code: checkViol, constraint: "sessions_revoked_reason_ck"},
		{name: "duplicate email verification token", stmt: `INSERT INTO email_verifications (user_id, email, token_hash, expires_at)
			VALUES ('` + uAlice + `', 'alice@example.org', decode(repeat('01', 32), 'hex'), now() + interval '1 day')`, code: uniqueViol, constraint: "email_verifications_token_uq"},
		{name: "password reset expiring before creation", stmt: `INSERT INTO password_resets (user_id, token_hash, expires_at)
			VALUES ('` + uBob + `', decode(repeat('05', 32), 'hex'), now() - interval '1 minute')`, code: checkViol, constraint: "password_resets_expiry_ck"},
		{name: "rate limit hits must be positive", stmt: `INSERT INTO rate_limit_buckets (scope, key_hash, window_start, hits) VALUES ('reset:ip', decode(repeat('06', 32), 'hex'), now(), 0)`, code: checkViol, constraint: "rate_limit_buckets_hits_ck"},
		{name: "rate limit bucket is unique per window", stmt: `INSERT INTO rate_limit_buckets (scope, key_hash, window_start) VALUES ('signin:ip', decode(repeat('03', 32), 'hex'), date_trunc('minute', now()))`, code: uniqueViol, constraint: "rate_limit_buckets_pk"},

		// taxonomy
		{name: "duplicate tag slug", stmt: `INSERT INTO tags (slug, label) VALUES ('go', 'Go again')`, code: uniqueViol, constraint: "tags_slug_uq"},
		{name: "uppercase tag slug", stmt: `INSERT INTO tags (slug, label) VALUES ('Go', 'Go')`, code: checkViol, constraint: "tags_slug_ck"},
		{name: "tag slug starting with a dash", stmt: `INSERT INTO tags (slug, label) VALUES ('-go', 'Go')`, code: checkViol, constraint: "tags_slug_ck"},
		{name: "tag slug over 32 characters", stmt: `INSERT INTO tags (slug, label) VALUES (repeat('a', 33), 'Long')`, code: checkViol, constraint: "tags_slug_ck"},
		{name: "tag created by unknown user", stmt: `INSERT INTO tags (slug, label, created_by) VALUES ('rust', 'Rust', '` + idp + `0000000000ff')`, code: fkViol, constraint: "tags_created_by_fk"},
		{name: "country code must be two capitals", stmt: `INSERT INTO countries (code, name_fr, name_en) VALUES ('fra', 'France', 'France')`, code: checkViol, constraint: "countries_code_ck"},

		// solve
		{name: "question language", stmt: `INSERT INTO questions (author_id, title, body_md, language) VALUES ('` + uAlice + `', 't', 'b', 'de')`, code: checkViol, constraint: "questions_language_ck"},
		{name: "question title empty", stmt: `INSERT INTO questions (author_id, title, body_md, language) VALUES ('` + uAlice + `', '', 'b', 'fr')`, code: checkViol, constraint: "questions_title_ck"},
		{name: "question body over 20000", stmt: `INSERT INTO questions (author_id, title, body_md, language) VALUES ('` + uAlice + `', 't', repeat('x', 20001), 'fr')`, code: checkViol, constraint: "questions_body_ck"},
		{name: "question removed without removed_at", stmt: `UPDATE questions SET status = 'removed' WHERE id = '` + qAlice + `'`, code: checkViol, constraint: "questions_removed_ck"},
		{name: "question status", stmt: `UPDATE questions SET status = 'solved' WHERE id = '` + qAlice + `'`, code: checkViol, constraint: "questions_status_ck"},
		{name: "question without author", stmt: `INSERT INTO questions (author_id, title, body_md, language) VALUES (NULL, 't', 'b', 'fr')`, code: notNull},
		{name: "question by unknown author", stmt: `INSERT INTO questions (author_id, title, body_md, language) VALUES ('` + idp + `0000000000ff', 't', 'b', 'fr')`, code: fkViol, constraint: "questions_author_fk"},
		{name: "question unknown country", stmt: `UPDATE questions SET country_code = 'ZZ' WHERE id = '` + qAlice + `'`, code: fkViol, constraint: "questions_country_fk"},
		{name: "negative answer count", stmt: `UPDATE questions SET answer_count = -1 WHERE id = '` + qAlice + `'`, code: checkViol, constraint: "questions_counters_ck"},
		{name: "accepted answer of another question", stmt: `UPDATE questions SET accepted_answer_id = '` + aAlice + `' WHERE id = '` + qAlice + `'`, code: fkViol, constraint: "questions_accepted_answer_fk"},
		{name: "accepted answer that does not exist", stmt: `UPDATE questions SET accepted_answer_id = '` + idp + `0000000000ff' WHERE id = '` + qAlice + `'`, code: fkViol, constraint: "questions_accepted_answer_fk"},
		{name: "answer on unknown question", stmt: `INSERT INTO answers (question_id, author_id, body_md, language) VALUES ('` + idp + `0000000000ff', '` + uBob + `', 'b', 'fr')`, code: fkViol, constraint: "answers_question_fk"},
		{name: "answer language", stmt: `UPDATE answers SET language = 'es' WHERE id = '` + aBob + `'`, code: checkViol, constraint: "answers_language_ck"},
		{name: "duplicate vote", stmt: `INSERT INTO votes (user_id, target_type, question_id, value) VALUES ('` + uBob + `', 'question', '` + qAlice + `', -1)`, code: uniqueViol, constraint: "votes_pk"},
		{name: "vote with two targets", stmt: `INSERT INTO votes (user_id, target_type, question_id, answer_id) VALUES ('` + uMod + `', 'question', '` + qAlice + `', '` + aBob + `')`, code: checkViol, constraint: "votes_target_ck"},
		{name: "vote target type mismatch", stmt: `INSERT INTO votes (user_id, target_type, question_id) VALUES ('` + uMod + `', 'answer', '` + qAlice + `')`, code: checkViol, constraint: "votes_target_ck"},
		{name: "vote without target", stmt: `INSERT INTO votes (user_id, target_type) VALUES ('` + uMod + `', 'question')`, code: notNull},
		{name: "vote on unknown question", stmt: `INSERT INTO votes (user_id, target_type, question_id) VALUES ('` + uMod + `', 'question', '` + idp + `0000000000ff')`, code: fkViol, constraint: "votes_question_fk"},
		{name: "vote value", stmt: `INSERT INTO votes (user_id, target_type, answer_id, value) VALUES ('` + uMod + `', 'answer', '` + aBob + `', 2)`, code: checkViol, constraint: "votes_value_ck"},
		{name: "duplicate question tag", stmt: `INSERT INTO question_tags (question_id, tag_id) VALUES ('` + qAlice + `', '` + tGo + `')`, code: uniqueViol, constraint: "question_tags_pk"},
		{name: "question tag on unknown tag", stmt: `INSERT INTO question_tags (question_id, tag_id) VALUES ('` + qAlice + `', '` + idp + `0000000000ff')`, code: fkViol, constraint: "question_tags_tag_fk"},
		{name: "question source must be http(s)", stmt: `INSERT INTO question_sources (question_id, url) VALUES ('` + qAlice + `', 'javascript:alert(1)')`, code: checkViol, constraint: "question_sources_url_ck"},
		{name: "duplicate question source", stmt: `INSERT INTO question_sources (question_id, url) VALUES ('` + qAlice + `', 'https://example.org/spec')`, code: uniqueViol, constraint: "question_sources_url_uq"},
		{name: "answer source must be http(s)", stmt: `INSERT INTO answer_sources (answer_id, url) VALUES ('` + aBob + `', 'ftp://example.org/x')`, code: checkViol, constraint: "answer_sources_url_ck"},

		// discover
		{name: "duplicate provider slug", stmt: `INSERT INTO providers (slug, name, created_by) VALUES ('kalimba-pay', 'Other', '` + uAlice + `')`, code: uniqueViol, constraint: "providers_slug_uq"},
		{name: "provider website must be https", stmt: `UPDATE providers SET website_url = 'http://example.org' WHERE id = '` + pKal + `'`, code: checkViol, constraint: "providers_website_ck"},
		{name: "provider summary without language", stmt: `UPDATE providers SET summary = 'x' WHERE id = '` + pKal + `'`, code: checkViol, constraint: "providers_summary_language_ck"},
		{name: "duplicate resource slug", stmt: `INSERT INTO resources (category_code, type, slug, name, summary, summary_language, created_by) VALUES ('mobile-money', 'api', 'kalimba-api', 'n', 's', 'en', '` + uAlice + `')`, code: uniqueViol, constraint: "resources_slug_uq"},
		{name: "resource type", stmt: `UPDATE resources SET type = 'plugin' WHERE id = '` + rKal + `'`, code: checkViol, constraint: "resources_type_ck"},
		{name: "resource unknown category", stmt: `UPDATE resources SET category_code = 'nope' WHERE id = '` + rKal + `'`, code: fkViol, constraint: "resources_category_fk"},
		{name: "resource doc language outside fr/en", stmt: `UPDATE resources SET doc_languages = '{en,de}' WHERE id = '` + rKal + `'`, code: checkViol, constraint: "resources_doc_languages_ck"},
		{name: "resource url must be https", stmt: `UPDATE resources SET docs_url = 'http://example.org' WHERE id = '` + rKal + `'`, code: checkViol, constraint: "resources_urls_ck"},
		{name: "verified resource needs a source", stmt: `UPDATE resources SET verification_status = 'verified' WHERE id = '` + rKal + `'`, code: checkViol, constraint: "resources_verified_ck"},
		{name: "verified source of another resource", stmt: `UPDATE resources SET verification_status = 'verified', verified_at = now(), verified_source_id = '` + rsSahel + `' WHERE id = '` + rKal + `'`, code: fkViol, constraint: "resources_verified_source_fk"},
		{name: "source needs a check date", stmt: `INSERT INTO resource_sources (resource_id, kind, title, url) VALUES ('` + rKal + `', 'official_docs', 't', 'https://example.org/x')`, code: checkViol, constraint: "resource_sources_checked_ck"},
		{name: "source url must be https", stmt: `INSERT INTO resource_sources (resource_id, kind, title, url, checked_at) VALUES ('` + rKal + `', 'official_docs', 't', 'http://example.org/x', now())`, code: checkViol, constraint: "resource_sources_url_ck"},
		{name: "source kind", stmt: `INSERT INTO resource_sources (resource_id, kind, title, url, checked_at) VALUES ('` + rKal + `', 'rumour', 't', 'https://example.org/x', now())`, code: checkViol, constraint: "resource_sources_kind_ck"},
		{name: "duplicate source url per resource", stmt: `INSERT INTO resource_sources (resource_id, kind, title, url, checked_at) VALUES ('` + rKal + `', 'changelog', 't', 'https://example.org/kalimba', now())`, code: uniqueViol, constraint: "resource_sources_url_uq"},
		{name: "resource on unknown country", stmt: `INSERT INTO resource_countries (resource_id, country_code) VALUES ('` + rKal + `', 'ZZ')`, code: fkViol, constraint: "resource_countries_country_fk"},
		{name: "duplicate resource country", stmt: `INSERT INTO resource_countries (resource_id, country_code) VALUES ('` + rKal + `', 'BF')`, code: uniqueViol, constraint: "resource_countries_pk"},
		{name: "duplicate sdk language entry", stmt: `INSERT INTO resource_sdk_languages (resource_id, language, name) VALUES ('` + rKal + `', 'go', 'kalimba-go')`, code: uniqueViol, constraint: "resource_sdk_languages_uq"},
		{name: "sdk repo url must be https", stmt: `UPDATE resource_sdk_languages SET repo_url = 'git://example.org/x' WHERE id = '` + sdkKal + `'`, code: checkViol, constraint: "resource_sdk_languages_urls_ck"},
		{name: "second open error report", stmt: `INSERT INTO error_reports (resource_id, reporter_id, kind) VALUES ('` + rKal + `', '` + uBob + `', 'broken_link')`, code: uniqueViol, constraint: "error_reports_open_uq"},
		{name: "second triaged error report", setup: []string{`UPDATE error_reports SET status = 'triaged'`},
			stmt: `INSERT INTO error_reports (resource_id, reporter_id, kind) VALUES ('` + rKal + `', '` + uBob + `', 'broken_link')`, code: uniqueViol, constraint: "error_reports_open_uq"},
		{name: "error report resolved without handler", stmt: `UPDATE error_reports SET status = 'resolved' WHERE id = '` + erKal + `'`, code: checkViol, constraint: "error_reports_handled_ck"},
		{name: "error report on a source of another resource", stmt: `INSERT INTO error_reports (resource_id, reporter_id, kind, source_id) VALUES ('` + rKal + `', '` + uBob + `', 'outdated', '` + rsSahel + `')`, code: fkViol, constraint: "error_reports_source_fk"},
		{name: "error report kind", stmt: `UPDATE error_reports SET kind = 'spam' WHERE id = '` + erKal + `'`, code: checkViol, constraint: "error_reports_kind_ck"},

		// build
		{name: "duplicate github url among published projects", stmt: `INSERT INTO projects (owner_id, name, summary, language, github_url, stage) VALUES ('` + uBob + `', 'Fork', 's', 'en', 'https://github.com/example/webhook-kit', 'active')`, code: uniqueViol, constraint: "projects_github_url_uq"},
		{name: "duplicate github url against a private project", setup: []string{`UPDATE projects SET visibility = 'private' WHERE id = '` + prAlice + `'`},
			stmt: `INSERT INTO projects (owner_id, name, summary, language, github_url, stage) VALUES ('` + uBob + `', 'Fork', 's', 'en', 'https://github.com/example/webhook-kit', 'active')`, code: uniqueViol, constraint: "projects_github_url_uq"},
		{name: "github url with uppercase", stmt: `UPDATE projects SET github_url = 'https://github.com/Example/webhook-kit' WHERE id = '` + prAlice + `'`, code: checkViol, constraint: "projects_github_url_ck"},
		{name: "github url ending in .git", stmt: `UPDATE projects SET github_url = 'https://github.com/example/webhook-kit.git' WHERE id = '` + prAlice + `'`, code: checkViol, constraint: "projects_github_url_ck"},
		{name: "github url with trailing slash", stmt: `UPDATE projects SET github_url = 'https://github.com/example/webhook-kit/' WHERE id = '` + prAlice + `'`, code: checkViol, constraint: "projects_github_url_ck"},
		{name: "github url on another host", stmt: `UPDATE projects SET github_url = 'https://gitlab.com/example/webhook-kit' WHERE id = '` + prAlice + `'`, code: checkViol, constraint: "projects_github_url_ck"},
		{name: "non-idea project needs a github url", stmt: `UPDATE projects SET github_url = NULL WHERE id = '` + prAlice + `'`, code: checkViol, constraint: "projects_github_required_ck"},
		{name: "project visibility", stmt: `UPDATE projects SET visibility = 'secret' WHERE id = '` + prAlice + `'`, code: checkViol, constraint: "projects_visibility_ck"},
		{name: "project stage", stmt: `UPDATE projects SET stage = 'dead' WHERE id = '` + prAlice + `'`, code: checkViol, constraint: "projects_stage_ck"},
		{name: "project owner required", stmt: `INSERT INTO projects (owner_id, name, summary, language) VALUES (NULL, 'n', 's', 'en')`, code: notNull},
		{name: "project owner must exist", stmt: `UPDATE projects SET owner_id = '` + idp + `0000000000ff' WHERE id = '` + prAlice + `'`, code: fkViol, constraint: "projects_owner_fk"},
		{name: "duplicate maintainer", stmt: `INSERT INTO project_maintainers (project_id, user_id, added_by) VALUES ('` + prAlice + `', '` + uBob + `', '` + uAlice + `')`, code: uniqueViol, constraint: "project_maintainers_pk"},
		{name: "need on unknown project", stmt: `INSERT INTO contribution_needs (project_id, title, description_md, language, created_by) VALUES ('` + idp + `0000000000ff', 't', 'd', 'en', '` + uAlice + `')`, code: fkViol, constraint: "contribution_needs_project_fk"},
		{name: "need status", stmt: `UPDATE contribution_needs SET status = 'archived' WHERE id = '` + nAlice + `'`, code: checkViol, constraint: "contribution_needs_status_ck"},
		{name: "closed need needs closed_at", stmt: `UPDATE contribution_needs SET status = 'closed' WHERE id = '` + nAlice + `'`, code: checkViol, constraint: "contribution_needs_closed_ck"},
		{name: "need level", stmt: `UPDATE contribution_needs SET level = 'guru' WHERE id = '` + nAlice + `'`, code: checkViol, constraint: "contribution_needs_level_ck"},
		{name: "duplicate need skill", stmt: `INSERT INTO need_skills (need_id, tag_id) VALUES ('` + nAlice + `', '` + tGo + `')`, code: uniqueViol, constraint: "need_skills_pk"},
		{name: "duplicate application for user and need", stmt: `INSERT INTO applications (contribution_need_id, user_id, message, language) VALUES ('` + nAlice + `', '` + uBob + `', 'again', 'en')`, code: uniqueViol, constraint: "applications_user_need_uq"},
		{name: "accepted application needs a decision", stmt: `UPDATE applications SET status = 'accepted' WHERE id = '` + apBob + `'`, code: checkViol, constraint: "applications_decision_ck"},
		{name: "withdrawn application needs withdrawn_at", stmt: `UPDATE applications SET status = 'withdrawn' WHERE id = '` + apBob + `'`, code: checkViol, constraint: "applications_withdrawn_ck"},
		{name: "application status", stmt: `UPDATE applications SET status = 'maybe' WHERE id = '` + apBob + `'`, code: checkViol, constraint: "applications_status_ck"},
		{name: "application message empty", stmt: `UPDATE applications SET message = '' WHERE id = '` + apBob + `'`, code: checkViol, constraint: "applications_message_ck"},

		// learn
		{name: "article level", stmt: `UPDATE learn_articles SET level = 'expert' WHERE id = '` + arAlice + `'`, code: checkViol, constraint: "learn_articles_level_ck"},
		{name: "article language", stmt: `UPDATE learn_articles SET language = 'es' WHERE id = '` + arAlice + `'`, code: checkViol, constraint: "learn_articles_language_ck"},
		{name: "article kind", stmt: `UPDATE learn_articles SET kind = 'video' WHERE id = '` + arAlice + `'`, code: checkViol, constraint: "learn_articles_kind_ck"},
		{name: "published article needs published_at", stmt: `UPDATE learn_articles SET published_at = NULL WHERE id = '` + arAlice + `'`, code: checkViol, constraint: "learn_articles_published_ck"},
		{name: "article unknown topic", stmt: `UPDATE learn_articles SET topic_code = 'nope' WHERE id = '` + arAlice + `'`, code: fkViol, constraint: "learn_articles_topic_fk"},
		{name: "article status", stmt: `UPDATE learn_articles SET status = 'pending' WHERE id = '` + arAlice + `'`, code: checkViol, constraint: "learn_articles_status_ck"},
		{name: "series origin", stmt: `UPDATE training_series SET origin = 'bot' WHERE id = '` + seAlice + `'`, code: checkViol, constraint: "training_series_origin_ck"},
		{name: "lesson without content", stmt: `INSERT INTO training_lessons (series_id, position, title, language) VALUES ('` + seAlice + `', 2, 't', 'fr')`, code: checkViol, constraint: "training_lessons_content_ck"},
		{name: "lesson position out of range", stmt: `INSERT INTO training_lessons (series_id, position, title, body_md, language) VALUES ('` + seAlice + `', 0, 't', 'b', 'fr')`, code: checkViol, constraint: "training_lessons_position_ck"},
		{name: "lesson on unknown series", stmt: `INSERT INTO training_lessons (series_id, position, title, body_md, language) VALUES ('` + idp + `0000000000ff', 1, 't', 'b', 'fr')`, code: fkViol, constraint: "training_lessons_series_fk"},
		{name: "duplicate lesson position (deferred, checked at once)",
			setup: []string{`INSERT INTO training_lessons (series_id, position, title, body_md, language) VALUES ('` + seAlice + `', 1, 'dup', 'b', 'fr')`},
			stmt:  `SET CONSTRAINTS training_lessons_position_uq IMMEDIATE`, code: uniqueViol, constraint: "training_lessons_position_uq"},
		{name: "duplicate article tag", stmt: `INSERT INTO learn_article_tags (article_id, tag_id) VALUES ('` + arAlice + `', '` + tWebhook + `')`, code: uniqueViol, constraint: "learn_article_tags_pk"},

		// relations
		{name: "relation in the non-canonical direction", stmt: `INSERT INTO content_relations (source_type, src_resource_id, target_type, tgt_question_id, created_by) VALUES ('resource', '` + rKal + `', 'question', '` + qAlice + `', '` + uAlice + `')`, code: checkViol, constraint: "content_relations_pair_ck"},
		{name: "relation between two items of the same type", stmt: `INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_question_id, created_by) VALUES ('question', '` + qAlice + `', 'question', '` + qBob + `', '` + uAlice + `')`, code: checkViol, constraint: "content_relations_pair_ck"},
		{name: "duplicate relation", stmt: `INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_resource_id, created_by) VALUES ('question', '` + qAlice + `', 'resource', '` + rKal + `', '` + uBob + `')`, code: uniqueViol, constraint: "content_relations_uq"},
		{name: "relation with two source columns", stmt: `INSERT INTO content_relations (source_type, src_question_id, src_project_id, target_type, tgt_resource_id, created_by) VALUES ('question', '` + qBob + `', '` + prAlice + `', 'resource', '` + rKal + `', '` + uAlice + `')`, code: checkViol, constraint: "content_relations_source_ck"},
		{name: "relation target column disagrees with type", stmt: `INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_resource_id, created_by) VALUES ('question', '` + qBob + `', 'project', '` + rKal + `', '` + uAlice + `')`, code: checkViol, constraint: "content_relations_target_ck"},
		{name: "relation without target", stmt: `INSERT INTO content_relations (source_type, src_question_id, target_type, created_by) VALUES ('question', '` + qBob + `', 'resource', '` + uAlice + `')`, code: notNull},
		{name: "relation to a dangling resource", stmt: `INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_resource_id, created_by) VALUES ('question', '` + qBob + `', 'resource', '` + idp + `0000000000ff', '` + uAlice + `')`, code: fkViol, constraint: "content_relations_tgt_resource_fk"},
		{name: "relation from a dangling question", stmt: `INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_project_id, created_by) VALUES ('question', '` + idp + `0000000000ff', 'project', '` + prAlice + `', '` + uAlice + `')`, code: fkViol, constraint: "content_relations_src_question_fk"},
		{name: "relation kind", stmt: `UPDATE content_relations SET relation_kind = 'hates' WHERE id = '` + crOne + `'`, code: checkViol, constraint: "content_relations_kind_ck"},
		{name: "relation status", stmt: `UPDATE content_relations SET status = 'deleted' WHERE id = '` + crOne + `'`, code: checkViol, constraint: "content_relations_status_ck"},

		// platform
		{name: "audit action format", stmt: `INSERT INTO audit_log (action) VALUES ('BadAction')`, code: checkViol, constraint: "audit_log_action_ck"},
		{name: "audit target type", stmt: `INSERT INTO audit_log (action, target_type) VALUES ('a.b', 'planet')`, code: checkViol, constraint: "audit_log_target_type_ck"},
		{name: "audit outcome", stmt: `INSERT INTO audit_log (action, outcome) VALUES ('a.b', 'maybe')`, code: checkViol, constraint: "audit_log_outcome_ck"},
		{name: "audit details must be an object", stmt: `INSERT INTO audit_log (action, details) VALUES ('a.b', '[]'::jsonb)`, code: checkViol, constraint: "audit_log_details_ck"},
		{name: "audit details over 4096 bytes", stmt: `INSERT INTO audit_log (action, details) VALUES ('a.b', jsonb_build_object('k', repeat('x', 5000)))`, code: checkViol, constraint: "audit_log_details_ck"},
		{name: "audit actor must exist", stmt: `INSERT INTO audit_log (action, actor_id) VALUES ('a.b', '` + idp + `0000000000ff')`, code: fkViol, constraint: "audit_log_actor_fk"},
		{name: "audit id cannot be supplied", stmt: `INSERT INTO audit_log (id, action) VALUES (1, 'a.b')`, code: "428C9"},
		{name: "idempotency key reuse for the same user", stmt: `INSERT INTO idempotency_keys (user_id, idempotency_key, method, route, request_hash, expires_at) VALUES ('` + uBob + `', 'key-0001-abcdef', 'POST', '/x', decode(repeat('04', 32), 'hex'), now() + interval '1 day')`, code: uniqueViol, constraint: "idempotency_keys_pk"},
		{name: "idempotency method", stmt: `UPDATE idempotency_keys SET method = 'GET'`, code: checkViol, constraint: "idempotency_keys_method_ck"},
		{name: "idempotency completed needs a response", stmt: `UPDATE idempotency_keys SET state = 'completed'`, code: checkViol, constraint: "idempotency_keys_completed_ck"},
		{name: "idempotency hash length", stmt: `UPDATE idempotency_keys SET request_hash = decode('abcd', 'hex')`, code: checkViol, constraint: "idempotency_keys_hash_ck"},

		// ON DELETE RESTRICT on user-owned content
		{name: "a user with content cannot be deleted", setup: []string{
			`INSERT INTO users (id, email) VALUES ('` + uTmp + `', 'tmp@example.org')`,
			`INSERT INTO questions (id, author_id, title, body_md, language) VALUES ('` + qTmp + `', '` + uTmp + `', 't', 'b', 'fr')`},
			stmt: `DELETE FROM users WHERE id = '` + uTmp + `'`, code: restrictViol, constraint: "questions_author_fk"},
		{name: "a question with answers cannot be deleted", stmt: `DELETE FROM questions WHERE id = '` + qBob + `'`, code: restrictViol, constraint: "answers_question_fk"},
		{name: "an accepted answer cannot be deleted", setup: []string{`UPDATE questions SET accepted_answer_id = '` + aBob + `' WHERE id = '` + qAlice + `'`},
			stmt: `DELETE FROM answers WHERE id = '` + aBob + `'`, code: restrictViol, constraint: "questions_accepted_answer_fk"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tx := begin(t, conn)
			mustExec(t, tx, fixtureGraph)
			for _, s := range tc.setup {
				mustExec(t, tx, s)
			}
			_, err := tx.Exec(ctx, tc.stmt)
			e := pgErr(err)
			if e == nil {
				t.Fatalf("expected SQLSTATE %s (%s), got error %v", tc.code, tc.constraint, err)
			}
			if e.Code != tc.code {
				t.Fatalf("SQLSTATE = %s (%s), want %s (%s): %s", e.Code, e.ConstraintName, tc.code, tc.constraint, e.Message)
			}
			if tc.constraint != "" && e.ConstraintName != tc.constraint {
				t.Fatalf("constraint = %q, want %q: %s", e.ConstraintName, tc.constraint, e.Message)
			}
		})
	}
}

// begin opens a transaction that is always rolled back with the subtest.
func begin(t *testing.T, conn *pgx.Conn) pgx.Tx {
	t.Helper()
	tx, err := conn.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	return tx
}

func mustExec(t *testing.T, tx pgx.Tx, sql string) {
	t.Helper()
	if _, err := tx.Exec(t.Context(), sql); err != nil {
		t.Fatalf("setup failed: %v\nSQL: %s", err, sql)
	}
}
