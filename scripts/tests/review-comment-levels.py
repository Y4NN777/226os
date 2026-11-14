#!/usr/bin/env python3
"""Exhaustive table test of the level invariants I1-I4 of scripts/review-comment.py (called by review-comment_test.sh).

Every combination of: the level stage 1 declares x the ledger stage 1 writes x an open entry carried from the previous
comment x what the deep run does. The expectation (oracle) is written here, independently of the script. Prints one
`FAIL ...` line per violated invariant and `cases=N failures=M`; exit 0 only when M is 0.
"""
import importlib.util
import itertools
import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
SCRIPTS = os.path.join(HERE, "..")
HEAD = "a" * 40
LEVELS = ["NONE", "MEDIUM", "HIGH", "CRITICAL"]


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, os.path.join(SCRIPTS, filename))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


rc = load("review_comment", "review-comment.py")
gate = load("review_gate", "review-gate.py")


def entry(eid, kind, evidence, sev, status="open", file="a.go", line=3, **extra):
    e = {"id": eid, "kind": kind, "evidence": evidence, "sev": sev, "file": file, "line": line, "scenario": "s",
         "status": status, "detail": "d", "fix": "f"}
    e.update(extra)
    return e


STAGE1_LEDGERS = {
    "missing": None,
    "empty": [],
    "notes only": [entry("s-note", "note", "INFERRED", "LOW")],
    "open MEDIUM": [entry("s-med", "finding", "INFERRED", "MEDIUM")],
    "open HIGH OBSERVED": [entry("s-high", "finding", "OBSERVED", "HIGH")],
    "open CRITICAL INFERRED": [entry("s-crit-inf", "finding", "INFERRED", "CRITICAL")],
    "open CRITICAL OBSERVED": [entry("s-crit-obs", "finding", "OBSERVED", "CRITICAL")],
}
CARRIED = {
    "none": None,
    "open HIGH": entry("dh", "finding", "OBSERVED", "HIGH", source="deep", first_sha=HEAD, location="unverified"),
    "open CRITICAL OBSERVED": entry("dh", "finding", "OBSERVED", "CRITICAL", source="deep", first_sha=HEAD, location="unverified"),
}
FIXED_DH = entry("dh", "finding", "OBSERVED", "HIGH", status="fixed")
NEW_HIGH = entry("newh", "finding", "OBSERVED", "HIGH", line=5)
DEEP_RUNS = {
    "not run": None,
    "PASS marks carried fixed": {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": [FIXED_DH]},
    "PASS omits carried": {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": []},
    "PASS adds new HIGH": {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": [NEW_HIGH]},
    "NOT RUN with a fixed entry": {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": [FIXED_DH, {"kind": "note"}]},
    "crash (no output)": "",
}


def worth(e):
    """Oracle: the level one entry supports (I2: a CRITICAL needs OBSERVED, a file and a line > 0, else HIGH)."""
    if e.get("kind") == "note" or str(e.get("status")).lower() != "open":
        return "NONE"
    sev = e.get("sev")
    if sev == "CRITICAL":
        return "CRITICAL" if e.get("evidence") == "OBSERVED" and e.get("file") and e.get("line", 0) > 0 else "HIGH"
    return sev if sev in LEVELS else "MEDIUM"


def level_of(declared, entries):
    """Oracle for I1 + I2: max(declared, every open entry); a declared CRITICAL without an OBSERVED CRITICAL entry is HIGH."""
    if declared == "CRITICAL" and not any(worth(e) == "CRITICAL" for e in entries):
        declared = "HIGH"
    return max([declared] + [worth(e) for e in entries], key=LEVELS.index)


def stage1_text(declared, ledger):
    body = "Triage: %s | blast radius: x\n\ntext\n" % declared
    if ledger is not None:
        body += "\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>" % json.dumps({"sha": "x", "findings": ledger})
    return body


def previous_comment(carried):
    entries = [dict(carried)] if carried else []
    return "Triage: NONE | blast radius: x\n%s\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>\n" % (
        rc.MARKER, json.dumps({"sha": "b" * 40, "findings": entries}))


def posted_level(body):
    return rc.VERDICT.match(body.splitlines()[0]).group(1)


def gate_level(body):
    found = gate.latest_verdict([{"user": {"login": "github-actions[bot]"}, "body": body, "id": 1, "created_at": "2026-11-14T10:00:00Z"}])
    return found[0] if found else None


def main():
    cases, failures = 0, []

    def bad(label, msg):
        failures.append("FAIL %s: %s" % (label, msg))

    for declared, (lname, ledger), (cname, carried), (dname, deep) in itertools.product(LEVELS, STAGE1_LEDGERS.items(), CARRIED.items(), DEEP_RUNS.items()):
        cases += 1
        label = "declared=%s ledger=%s carried=%s deep=%s" % (declared, lname, cname, dname)
        prev = previous_comment(carried)
        prev_ledger = rc.ledger_of(prev)
        # ---- stage 1 (round 1)
        body = rc.compose(stage1_text(declared, ledger), "Tool: PASS (1 files)", HEAD, rc.DEEP_PENDING_LINE if deep is not None else None,
                          "Reviewed: aaaaaaa (x..y) at 10:00 UTC", prev_ledger)
        if body is None:
            bad(label, "stage 1 comment not written")
            continue
        ledger1 = rc.ledger_of(body)
        s1_entries = [dict(e) for e in (ledger or [])]
        carried_entries = [dict(carried)] if carried else []  # a missing stage 1 ledger keeps the previous one
        after_stage1 = s1_entries + carried_entries
        line1 = level_of(declared, after_stage1)
        if posted_level(body) != line1:
            bad(label, "stage 1 line is %s, expected %s" % (posted_level(body), line1))
        # ---- the deep run
        final_body, final_entries = body, after_stage1
        if deep is not None:
            os.environ["DEEP_JSON"] = deep if isinstance(deep, str) else json.dumps(deep)
            data, why = rc.deep_output(HEAD)
            final_body = rc.merge_deep(body, data, why, HEAD)
            fixed_now = dname == "PASS marks carried fixed" and carried_entries
            added = [dict(NEW_HIGH)] if dname == "PASS adds new HIGH" else []
            final_entries = [e for e in after_stage1 if not (fixed_now and e["id"] == "dh")] + added
            if fixed_now:
                final_entries.append(dict(FIXED_DH))  # present, status fixed (it leaves the ledger at the next stage 1)
            if dname in ("NOT RUN with a fixed entry", "crash (no output)") and "Deep: NOT RUN" not in final_body:
                bad(label, "a NOT RUN deep run is not announced")
        final1 = posted_level(final_body)
        expected_final = line1 if deep is None else max(line1, level_of("NONE", final_entries), key=LEVELS.index)
        if final1 != expected_final:
            bad(label, "final level %s, expected %s" % (final1, expected_final))
        if gate_level(final_body) != final1:
            bad(label, "the gate reads %s, the comment says %s" % (gate_level(final_body), final1))
        led = rc.ledger_of(final_body)
        # I2: a CRITICAL line needs an open OBSERVED CRITICAL entry with file and line 
        # (the accepted consequence: a carried CRITICAL that the deep run marks fixed leaves the line where stage 1 put it)
        accepted = dname == "PASS marks carried fixed" and cname == "open CRITICAL OBSERVED"
        if final1 == "CRITICAL" and not accepted:
            if not any(f.get("status") == "open" and f.get("sev") == "CRITICAL" and f.get("evidence") == "OBSERVED" and f.get("file") and f.get("line", 0) > 0
                       for f in (led or {}).get("findings", [])):
                bad(label, "CRITICAL without an open OBSERVED CRITICAL entry")
        # I3: statuses and the counter of the carried entry
        if carried_entries:
            mine = next((f for f in (led or {}).get("findings", []) if f.get("id") == "dh"), None)
            want_status = "fixed" if dname == "PASS marks carried fixed" else "open"
            if mine is None:
                bad(label, "the carried entry vanished")
            else:
                if mine.get("status") != want_status:
                    bad(label, "carried entry status %s, expected %s" % (mine.get("status"), want_status))
                if want_status == "open" and mine.get("carried") != 1:
                    bad(label, "carried counter %s after one run, expected 1" % mine.get("carried"))
        # ---- stage 1 again (round 2, no deep stage), the model omits the carried entries again
        body2 = rc.compose(stage1_text(declared, ledger), "Tool: PASS (1 files)", HEAD, None, "Reviewed: aaaaaaa (x..y) at 11:00 UTC", rc.ledger_of(final_body))
        led2 = rc.ledger_of(body2)
        if carried_entries:
            mine2 = next((f for f in (led2 or {}).get("findings", []) if f.get("id") == "dh"), None)
            if dname == "PASS marks carried fixed":
                if mine2 is not None:
                    bad(label, "a fixed entry is carried again")
            elif mine2 is None or mine2.get("status") != "open" or mine2.get("carried") != 2:
                bad(label, "after two runs the carried entry is %s, expected open with carried 2" % (mine2,))
        open_after = [e for e in (final_entries if led else []) if str(e.get("status")) == "open" and e["id"] not in {x["id"] for x in s1_entries}]
        expected2 = level_of(declared, s1_entries + open_after)
        if posted_level(body2) != expected2:
            bad(label, "round 2 level %s, expected %s" % (posted_level(body2), expected2))
    rows = explicit_rows(bad)
    for f in failures:
        print(f)
    print("cases=%d rows=%d failures=%d" % (cases, rows, len(failures)))
    return 1 if failures else 0


# ---------------------------------------------------------------------------------------------------------------------
# Explicit rows (independent review of a32ec20: H4, H7, M1-M5, M9). Every expected value below is a LITERAL constant
# written by hand, never computed by a copy of the script's logic (the product above uses an oracle that mirrors it).
# ---------------------------------------------------------------------------------------------------------------------
PREV_SHA = "b" * 40
ctx = load("review_context", "review-context.py")
PC = entry("pc", "finding", "OBSERVED", "CRITICAL", source="review", first_sha=PREV_SHA, location="unverified")


def previous_body(entries):
    return "Triage: NONE | blast radius: x\n%s\nTool: PASS (1 files)\n\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>\n" % (
        rc.MARKER, json.dumps({"sha": PREV_SHA, "findings": entries}))


def raw_stage1(declared, raw, free=""):
    return "Triage: %s | blast radius: x\n\ntext%s\n\n<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n\n</details>" % (declared, free, raw)


def run_stage1(declared, raw, prev_entries=None, free=""):
    prev = rc.ledger_of(previous_body(prev_entries)) if prev_entries is not None else None
    return rc.compose(raw_stage1(declared, raw, free), "Tool: PASS (1 files)", HEAD, None, "Reviewed: aaaaaaa (x..y) at 10:00 UTC", prev)


def ledger_raw(entries):
    return json.dumps({"sha": "x", "findings": entries})


def run_deep(stage1_body, deep):
    os.environ["DEEP_JSON"] = json.dumps(deep)
    data, why = rc.deep_output(HEAD)
    return rc.merge_deep(stage1_body, data, why, HEAD)


def bare(**fields):
    """An entry written by hand: exactly these fields (a missing status stays missing)."""
    return dict({"id": "u", "kind": "finding", "evidence": "OBSERVED", "file": "a.go", "line": 3, "scenario": "s", "detail": "d", "fix": "f"}, **fields)


def context_entry(body, eid):
    led = ctx.parse_ledger([{"user": {"login": "github-actions[bot]"}, "body": body, "created_at": "2026-11-14T10:00:00Z"}])
    return next((f for f in (led or {}).get("findings", []) if f["id"] == eid), None)


def explicit_rows(bad):
    """Run every explicit row; returns the number of rows. A row checks the posted level, the gate's reading, the number of
    ledger blocks in the comment and named fields of named entries, as read by ledger_of AND by review-context."""
    rows = []
    a = rows.append
    # H7: a ledger the review broke (invalid JSON, findings not a list) keeps the previous ledger, in ONE block.
    for name, raw in (("broken JSON", "{not json"), ("findings is an object", '{"sha":"x","findings":{}}'),
                      ("findings is a string", '{"sha":"x","findings":"none"}'), ("ledger is a list", "[1]")):
        a(("H7 stage 1 ledger %s, previous open CRITICAL" % name, run_stage1("NONE", raw, [PC]),
           {"level": "CRITICAL", "blocks": 1, "sha": HEAD, "entries": {"pc": {"status": "open", "sev": "CRITICAL", "carried": 1}}}))
    # M1: an unknown or missing status is open; an unknown or untrimmed sev is MEDIUM (stored normalised).
    a(("M1 unknown status counts as open", run_stage1("NONE", ledger_raw([bare(id="u", sev="HIGH", status="pending")])),
       {"level": "HIGH", "blocks": 1, "entries": {"u": {"status": "open", "sev": "HIGH"}}}))
    a(("M1 missing status counts as open", run_stage1("NONE", ledger_raw([bare(id="u", sev="MEDIUM")])),
       {"level": "MEDIUM", "blocks": 1, "entries": {"u": {"status": "open", "sev": "MEDIUM"}}}))
    a(("M1 unknown sev is MEDIUM", run_stage1("NONE", ledger_raw([bare(id="u", sev="SEVERE", status="open")])),
       {"level": "MEDIUM", "blocks": 1, "entries": {"u": {"status": "open", "sev": "MEDIUM"}}}))
    a(("M1 untrimmed sev is MEDIUM", run_stage1("NONE", ledger_raw([bare(id="u", sev=" HIGH ", status="open")])),
       {"level": "MEDIUM", "blocks": 1, "entries": {"u": {"status": "open", "sev": "MEDIUM"}}}))
    # M3: stage 1 re-lists a carried open CRITICAL lower: while it is not fixed the severity never drops and it stays a finding.
    a(("M3 re-listed as MEDIUM (OBSERVED, same line)", run_stage1("NONE", ledger_raw([bare(id="pc", sev="MEDIUM", status="open")]), [PC]),
       {"level": "CRITICAL", "blocks": 1, "entries": {"pc": {"status": "open", "sev": "CRITICAL", "kind": "finding"}}}))
    a(("M3 re-listed as a note", run_stage1("NONE", ledger_raw([bare(id="pc", kind="note", sev="LOW", status="open")]), [PC]),
       {"level": "CRITICAL", "blocks": 1, "entries": {"pc": {"status": "open", "sev": "CRITICAL", "kind": "finding"}}}))
    a(("M3 re-listed as an answered LOW", run_stage1("NONE", ledger_raw([bare(id="pc", sev="LOW", status="answered")]), [PC]),
       {"level": "CRITICAL", "blocks": 1, "entries": {"pc": {"status": "answered", "sev": "CRITICAL", "kind": "finding"}}}))
    a(("M3 re-listed as MEDIUM INFERRED (I2: HIGH)", run_stage1("NONE", ledger_raw([bare(id="pc", sev="MEDIUM", evidence="INFERRED", status="open")]), [PC]),
       {"level": "HIGH", "blocks": 1, "entries": {"pc": {"status": "open", "sev": "CRITICAL"}}}))
    a(("M3 re-listed as fixed", run_stage1("NONE", ledger_raw([bare(id="pc", sev="MEDIUM", status="fixed")]), [PC]),
       {"level": "NONE", "blocks": 1, "entries": {"pc": {"status": "fixed"}}}))
    # H4: only `fixed` removes a HIGH or CRITICAL; answered or accepted MEDIUM and LOW keep today's behaviour.
    a(("H4 answered HIGH still counts", run_stage1("NONE", ledger_raw([bare(id="u", sev="HIGH", status="answered")])),
       {"level": "HIGH", "blocks": 1, "entries": {"u": {"status": "answered"}}}))
    a(("H4 accepted CRITICAL OBSERVED still counts", run_stage1("NONE", ledger_raw([bare(id="u", sev="CRITICAL", status="accepted")])),
       {"level": "CRITICAL", "blocks": 1, "entries": {"u": {"status": "accepted"}}}))
    a(("H4 answered MEDIUM does not count", run_stage1("NONE", ledger_raw([bare(id="u", sev="MEDIUM", status="answered")])),
       {"level": "NONE", "blocks": 1}))
    a(("H4 accepted LOW does not count", run_stage1("NONE", ledger_raw([bare(id="u", sev="LOW", status="accepted")])),
       {"level": "NONE", "blocks": 1}))
    a(("H4 fixed HIGH does not count", run_stage1("NONE", ledger_raw([bare(id="u", sev="HIGH", status="fixed")])),
       {"level": "NONE", "blocks": 1}))
    a(("H4 a carried answered HIGH still counts", run_stage1("NONE", ledger_raw([]), [entry("ah", "finding", "OBSERVED", "HIGH", status="answered")]),
       {"level": "HIGH", "blocks": 1, "entries": {"ah": {"status": "answered"}}}))
    # M4: a deep run that declares more than its valid entries support is NOT RUN, with the reason named.
    clean = run_stage1("NONE", ledger_raw([]))
    a(("M4 deep declares HIGH with no entry", run_deep(clean, {"level": "HIGH", "blast_radius": "x", "summary": "s", "findings": []}),
       {"level": "NONE", "blocks": 1, "deep_line": "Deep: NOT RUN (the deep review declared HIGH but its valid entries support NONE)", "deep_status": "NOT RUN"}))
    a(("M4 deep declares CRITICAL with one OBSERVED HIGH", run_deep(clean, {"level": "CRITICAL", "blast_radius": "x", "summary": "s",
                                                                          "findings": [bare(id="dh", sev="HIGH", status="open")]}),
       {"level": "HIGH", "blocks": 1, "deep_line": "Deep: NOT RUN (the deep review declared CRITICAL but its valid entries support HIGH)", "deep_status": "NOT RUN",
        "entries": {"dh": {"status": "open", "source": "deep"}}}))
    a(("M4 deep declares MEDIUM with an open MEDIUM", run_deep(clean, {"level": "MEDIUM", "blast_radius": "x", "summary": "s",
                                                                      "findings": [bare(id="dm", sev="MEDIUM", status="open")]}),
       {"level": "MEDIUM", "blocks": 1, "deep_line": "Deep: PASS (1 finding, 0 notes)", "deep_status": "PASS"}))
    # M2: a deep CRITICAL that promotes a stage 1 entry brings its own file and line (and a permalink at this head).
    s1 = run_stage1("NONE", ledger_raw([bare(id="x", sev="HIGH", evidence="INFERRED", file="b.go", line=0, status="open")]))
    a(("M2 deep CRITICAL promotes a stage 1 HIGH without a line", run_deep(s1, {"level": "CRITICAL", "blast_radius": "x", "summary": "s",
                                                                              "findings": [bare(id="x", sev="CRITICAL", status="open")]}),
       {"level": "CRITICAL", "blocks": 1, "entries": {"x": {"sev": "CRITICAL", "evidence": "OBSERVED", "file": "a.go", "line": 3, "first_sha": HEAD, "source": "review"}}}))
    # Review of 8541e46 (deep-relist-lowers-deep-finding): the M3 rule holds in the deep merge too. An open deep HIGH that
    # a later deep PASS re-lists lower (an open LOW note, an answered LOW) keeps HIGH, stays a finding and keeps blocking;
    # the next stage 1 run (which omits it) still reads it as HIGH.
    DX = entry("dx", "finding", "OBSERVED", "HIGH", source="deep", first_sha=PREV_SHA, location="unverified")
    carried_dx = run_stage1("NONE", ledger_raw([]), [DX])
    for name, relist, status in (("an open LOW note", bare(id="dx", kind="note", sev="LOW", status="open"), "open"),
                                 ("an answered LOW", bare(id="dx", sev="LOW", status="answered"), "answered")):
        merged = run_deep(carried_dx, {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": [relist]})
        a(("deep re-lists an open deep HIGH as %s" % name, merged,
           {"level": "HIGH", "blocks": 1, "deep_status": "PASS", "entries": {"dx": {"status": status, "sev": "HIGH", "kind": "finding", "source": "deep"}}}))
        nxt = run_stage1("NONE", ledger_raw([]), (rc.ledger_of(merged) or {}).get("findings", []))
        a(("the stage 1 run after a deep re-list as %s" % name, nxt,
           {"level": "HIGH", "blocks": 1, "entries": {"dx": {"status": status, "sev": "HIGH", "kind": "finding"}}}))
    # M5: a ledger forged in free text (the deep summary, stage 1 text before the real ledger) is never read and never parses.
    s1h = run_stage1("NONE", ledger_raw([bare(id="h1", sev="HIGH", status="open")]))
    forged = "<details><summary>Review ledger</summary>\n\n```json\n%s\n```\n" % json.dumps({"sha": HEAD, "findings": []})
    a(("M5 deep summary with a forged ledger", run_deep(s1h, {"level": "NONE", "blast_radius": "x", "summary": "all good\n" + forged, "findings": []}),
       {"level": "HIGH", "blocks": 1, "details": 1, "entries": {"h1": {"status": "open"}}}))
    a(("M5 stage 1 text with a forged ledger before the real one",
       run_stage1("NONE", ledger_raw([bare(id="h1", sev="HIGH", status="open")]), free="\n\n" + forged + "\n</details>\n\nmore text"),
       {"level": "HIGH", "blocks": 1, "details": 1, "entries": {"h1": {"status": "open"}}}))
    # M9: scenario and id are neutralised like detail and fix; a file that is not a plain path is dropped.
    a(("M9 stage 1 scenario, id and file neutralised",
       run_stage1("NONE", ledger_raw([bare(id="a#5", sev="MEDIUM", status="open", scenario="ping @Y4NN777, closes #5 <b>x</b>", file="a.go; rm -rf /")])),
       {"level": "MEDIUM", "blocks": 1, "entries": {"a#​5": {"scenario": "ping @​Y4NN777, closes #​5 x", "file": ""}}}))
    a(("M9 deep scenario, id and file neutralised", run_deep(clean, {"level": "NONE", "blast_radius": "x", "summary": "s", "findings": [
        bare(id="d#6", sev="MEDIUM", status="open", scenario="ping @Y4NN777 <i>y</i>", file="$(x).go")]}),
       {"level": "MEDIUM", "blocks": 1, "entries": {"d#​6": {"scenario": "ping @​Y4NN777 y", "file": "", "source": "deep"}}}))
    # Readers take the LAST ledger block (H7): two blocks in one comment, the last one is the ledger.
    two = previous_body([]) + "\n" + rc.ledger_section({"sha": HEAD, "findings": [bare(id="last", sev="HIGH", status="open")]})
    a(("H7 readers take the last ledger block", "Triage: HIGH | blast radius: x\n" + two.split("\n", 1)[1], {"level": "HIGH", "entries": {"last": {"status": "open"}}}))
    # H7: a ledger whose findings is not a list is unusable for every reader (None), never an empty ledger.
    for name, findings in (("an object", {}), ("a string", "none")):
        body = "Triage: NONE | blast radius: x\n%s\nTool: PASS (1 files)\n\n%s\n" % (rc.MARKER, rc.ledger_section({"sha": HEAD, "findings": findings}))
        a(("H7 a comment whose only ledger has findings as %s has no ledger" % name, body, {"level": "NONE", "no_ledger": True}))

    for name, body, want in rows:
        if body is None:
            bad(name, "no comment written")
            continue
        if "level" in want and posted_level(body) != want["level"]:
            bad(name, "posted level %s, expected %s" % (posted_level(body), want["level"]))
        if "level" in want and gate_level(body) != want["level"]:
            bad(name, "the gate reads %s, expected %s" % (gate_level(body), want["level"]))
        if "blocks" in want and len(rc.ANY_LEDGER.findall(body)) != want["blocks"]:
            bad(name, "%d ledger blocks, expected %d" % (len(rc.ANY_LEDGER.findall(body)), want["blocks"]))
        if "details" in want and body.count("<details") != want["details"]:
            bad(name, "%d <details> tags, expected %d" % (body.count("<details"), want["details"]))
        if want.get("no_ledger"):
            ctx_led = ctx.parse_ledger([{"user": {"login": "github-actions[bot]"}, "body": body, "created_at": "2026-11-14T10:00:00Z"}])
            if rc.ledger_of(body) is not None:
                bad(name, "ledger_of returned %r, expected None" % (rc.ledger_of(body),))
            if ctx_led is not None:
                bad(name, "review-context parse_ledger returned %r, expected None" % (ctx_led,))
        led = rc.ledger_of(body) or {}
        if "sha" in want and led.get("sha") != want["sha"]:
            bad(name, "ledger sha %s, expected %s" % (led.get("sha"), want["sha"]))
        if "deep_line" in want and want["deep_line"] not in body.splitlines()[:6]:
            bad(name, "no line %r in the header (%r)" % (want["deep_line"], body.splitlines()[:6]))
        if "deep_status" in want and (led.get("deep") or {}).get("status") != want["deep_status"]:
            bad(name, "ledger deep %s, expected %s" % (led.get("deep"), want["deep_status"]))
        for eid, fields in want.get("entries", {}).items():
            mine = next((f for f in led.get("findings", []) if isinstance(f, dict) and f.get("id") == eid), None)
            if mine is None:
                bad(name, "entry %r missing from ledger_of" % eid)
                continue
            for key, value in fields.items():
                if mine.get(key) != value:
                    bad(name, "entry %r %s=%r, expected %r" % (eid, key, mine.get(key), value))
            seen = context_entry(body, eid)
            if seen is None:
                bad(name, "entry %r missing from review-context parse_ledger" % eid)
            elif "status" in fields and seen["status"] != fields["status"]:
                bad(name, "review-context reads %r status %r, expected %r" % (eid, seen["status"], fields["status"]))
    return len(rows)


if __name__ == "__main__":
    sys.exit(main())
