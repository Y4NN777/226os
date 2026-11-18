#!/usr/bin/env python3
"""Build the local-dev Connect seed from committers.top (db/seeds/local-dev/00001_connect_committers.sql).

Source: https://github.com/ashkulz/committers.top (MIT), branch gh-pages, _data/locations/<country>.yml. Only the
`users` list (all contributions) is read. Offline and deterministic: the YAML files are downloaded beforehand and passed
as arguments, so the seed itself never touches the network. Real public GitHub data, so the output is a local-dev seed
only (db/seeds/README.md); `make seed` loads db/seeds/local-dev/ for APP_ENV=local-dev and never elsewhere.

Usage: committers-seed.py CC=path/to/country.yml:LIMIT [...] > db/seeds/local-dev/00001_connect_committers.sql
  CC     ISO 3166-1 alpha-2 code of an African country of the countries table
  LIMIT  top N users of that country (1 to 100)
"""
import re
import sys

COUNTRIES = {"BF", "CI", "SN", "ML", "NE", "BJ", "TG", "GN", "GH", "NG"}
USERNAME = re.compile(r"[a-z0-9][a-z0-9_-]{2,29}")  # profiles_username_ck
GITHUB = re.compile(r"[A-Za-z0-9][A-Za-z0-9-]{0,38}")  # profiles_github_ck
RESERVED = {"demo-aicha", "demo-kofi", "demo-nadia", "demo-y4nn", "jury-cadev"}
ID_BASE = 0xA01  # block A01-AFF of db/seeds/README.md
ID_LAST = 0xAFF
STAMP = "2026-11-15T12:00:00Z"
KEYS = ("rank", "name", "login", "avatarUrl", "contributions", "company", "organizations")


def die(msg):
    print(f"committers-seed: {msg}", file=sys.stderr)
    sys.exit(2)


def scalar(raw):
    raw = raw.strip()
    if raw.startswith('"') and raw.endswith('"') and len(raw) >= 2:
        out, i, body = [], 0, raw[1:-1]
        while i < len(body):
            c = body[i]
            if c == "\\" and body[i + 1 : i + 2] == "u" and re.fullmatch(r"[0-9a-fA-F]{4}", body[i + 2 : i + 6]):
                out.append(chr(int(body[i + 2 : i + 6], 16)))
                i += 6
                continue
            if c == "\\" and i + 1 < len(body):
                out.append({"n": "\n", "t": "\t"}.get(body[i + 1], body[i + 1]))
                i += 2
                continue
            out.append(c)
            i += 1
        return "".join(out)
    return raw


def parse(path):
    """The `users` list and `generated` value of a committers.top location file (its fixed, flat shape)."""
    try:
        lines = open(path, encoding="utf-8").read().splitlines()
    except OSError as e:
        die(f"cannot read {path}: {e.strerror}")
    users, generated, section, cur = [], "", None, None
    for line in lines:
        top = re.fullmatch(r"([a-z_]+):(.*)", line)
        if top:
            section = top.group(1)
            if section == "generated":
                generated = top.group(2).strip()
            cur = None
            continue
        if section != "users":
            continue
        item = re.fullmatch(r"\s+- ([A-Za-z]+): ?(.*)", line)
        field = re.fullmatch(r"\s+([A-Za-z]+): ?(.*)", line)
        if item:
            cur = {}
            users.append(cur)
            field = item
        if field and cur is not None and field.group(1) in KEYS:
            cur[field.group(1)] = scalar(field.group(2))
    return users, generated


def q(s):
    return "'" + s.replace("'", "''") + "'"


def main(argv):
    if not argv:
        die("usage: committers-seed.py CC=file.yml:LIMIT [...]")
    rows, sources, seen = [], [], set(RESERVED)
    for arg in argv:
        m = re.fullmatch(r"([A-Z]{2})=(.+):([0-9]{1,3})", arg)
        if not m:
            die(f"bad argument {arg!r} (CC=file.yml:LIMIT)")
        cc, path, limit = m.group(1), m.group(2), int(m.group(3))
        if cc not in COUNTRIES:
            die(f"unknown country code {cc}")
        if not 1 <= limit <= 100:
            die(f"limit for {cc} must be 1 to 100")
        users, generated = parse(path)
        if not users:
            die(f"no users in {path}")
        sources.append(f"{cc} generated {generated or 'unknown'}")
        kept = 0
        for u in users:
            if kept >= limit:
                break
            login = u.get("login", "")
            if not GITHUB.fullmatch(login):
                continue
            username = login.lower()
            if not USERNAME.fullmatch(username) or username in seen:
                continue
            name = (u.get("name") or "").strip() or login
            name = name[:80]
            seen.add(username)
            rows.append((cc, username, name, login))
            kept += 1
    if ID_BASE + len(rows) - 1 > ID_LAST:
        die(f"{len(rows)} users exceed the id block A01-AFF")

    out = [
        "-- Local-dev Connect seed: real public GitHub developers ranked by committers.top. GENERATED, do not edit.",
        "-- Source: committers.top (https://github.com/ashkulz/committers.top, MIT), _data/locations/<country>.yml, list `users`.",
        "-- " + "; ".join(sources) + ".",
        "-- Regenerate with scripts/content/committers-seed.py. Local-dev only (db/seeds/README.md): make seed loads",
        "-- db/seeds/local-dev/ for APP_ENV=local-dev only. Demo provenance, no credentials. No BEGIN/COMMIT (make seed wraps it).",
        "",
        "-- A username, GitHub login or email already taken (a live account, an earlier seed) is skipped, never an error.",
        "CREATE TEMP TABLE seed_committers (id uuid, username text, display_name text, country_code text, github text) ON COMMIT DROP;",
        "INSERT INTO seed_committers (id, username, display_name, country_code, github) VALUES",
    ]
    ids = [f"01900000-0000-7000-8000-000000000{ID_BASE + i:03x}" for i in range(len(rows))]
    out.append(",\n".join(f"  ({q(i)}, {q(r[1])}, {q(r[2])}, {q(r[0])}, {q(r[3])})" for i, r in zip(ids, rows)) + ";")
    out.append("")
    out.append("INSERT INTO users (id, email, email_verified_at, role, status, provenance, created_at, updated_at, deleted_at)")
    out.append(f"SELECT s.id, s.username || '@demo.226os.invalid', {q(STAMP)}, 'member', 'active', 'demo', {q(STAMP)}, {q(STAMP)}, NULL")
    out.append("FROM seed_committers s")
    out.append("WHERE NOT EXISTS (SELECT 1 FROM profiles p WHERE p.username = s.username OR lower(p.github_username) = lower(s.github))")
    # users_email_uq: a leftover row with this email under another id would abort the whole seed (#330 notes).
    out.append("AND NOT EXISTS (SELECT 1 FROM users x WHERE x.email = s.username || '@demo.226os.invalid')")
    out.append("ON CONFLICT (id) DO NOTHING;")
    out.append("")
    out.append("INSERT INTO profiles (user_id, username, display_name, intent, country_code, languages, github_username, bio, "
               "bio_language, open_to_collab, country_visibility, languages_visibility, skills_visibility, github_visibility, "
               "bio_visibility, open_to_collab_visibility, created_at, updated_at)")
    # No invented intent for a real person (owner 2026-11-18): intent NULL; open_to_collab true so the local-dev
    # Connect directory lists them (the purpose of this seed).
    out.append(f"SELECT s.id, s.username, s.display_name, NULL, s.country_code, '{{}}', s.github, NULL, NULL, true, 'public', "
               f"'public', 'public', 'public', 'public', 'public', {q(STAMP)}, {q(STAMP)}")
    out.append("FROM seed_committers s")
    out.append("JOIN users u ON u.id = s.id AND u.provenance = 'demo'")
    out.append("WHERE NOT EXISTS (SELECT 1 FROM profiles p WHERE p.username = s.username OR lower(p.github_username) = lower(s.github))")
    out.append("ON CONFLICT (user_id) DO NOTHING;")
    sys.stdout.write("\n".join(out) + "\n")


if __name__ == "__main__":
    main(sys.argv[1:])
