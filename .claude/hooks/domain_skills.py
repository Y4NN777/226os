#!/usr/bin/env python3
"""226os domain-skill guard for Claude Code hooks (stdlib only).

Modes (argv[1]):
  pre            PreToolUse on Read|Edit|Write|MultiEdit|NotebookEdit.
                 Maps the file path to domain skills (.claude/hooks/skill-map.json).
                 Read  -> allow; once per skill, add a short directive to load the missing
                          skills with the Skill tool (additionalContext).
                 Write -> deny until every mapped skill is loaded in this context.
                 Skill bodies are never inlined: Claude Code truncates long hook output to
                 a ~2 KB preview, while the Skill tool loads the full skill reliably.
  post-skill     PostToolUse on Skill: records the invoked skill as loaded.
  session-start  SessionStart: clears the session cache on compact/clear.

Cache: one JSON file per (session_id, agent_id) under CLAUDE_SKILL_STATE_DIR
(default: $TMPDIR/226os-domain-skills). A skill counts as loaded once
the Skill tool ran for it in that context (or the subagent preloads it).
Malformed hook input fails open (exit 0, no output). For writes, an unreadable or
invalid skill-map.json, or a mapped skill whose SKILL.md is missing, is denied
so a broken map cannot silently disable the guard. Subagent `skills:` preloads
(.claude/agents/<agent_type>.md) count as already loaded.
"""

import hashlib
import json
import os
import pathlib
import re
import sys
import tempfile

READ_TOOLS = {"Read"}
WRITE_TOOLS = {"Edit", "Write", "MultiEdit", "NotebookEdit"}
RESET_SOURCES = {"compact", "clear"}
NOTIFIED = "notified:"


def project_dir(payload):
    return pathlib.Path(os.environ.get("CLAUDE_PROJECT_DIR") or payload.get("cwd") or os.getcwd()).resolve()


def state_file(payload):
    base = pathlib.Path(os.environ.get("CLAUDE_SKILL_STATE_DIR") or pathlib.Path(tempfile.gettempdir()) / "226os-domain-skills")
    key = f"{payload.get('session_id', 'nosession')}::{payload.get('agent_id', 'main')}"
    return base / (hashlib.sha256(key.encode()).hexdigest()[:32] + ".json")


def load_state(path):
    try:
        return set(json.loads(path.read_text()))
    except (OSError, ValueError):
        return set()


def save_state(path, loaded):
    path.parent.mkdir(parents=True, exist_ok=True)
    tmp = path.with_suffix(".tmp")
    tmp.write_text(json.dumps(sorted(loaded)))
    tmp.replace(path)


def glob_to_regex(glob):
    """Translate a repo-relative glob: ** = zero or more directories, * and ? stay within one segment."""
    parts = glob.split("/")
    out = []
    for i, part in enumerate(parts):
        last = i == len(parts) - 1
        if part == "**":
            out.append(".*" if last else "(?:[^/]+/)*")
            continue
        seg = ""
        for ch in part:
            seg += "[^/]*" if ch == "*" else "[^/]" if ch == "?" else re.escape(ch)
        out.append(seg if last else seg + "/")
    return re.compile("^" + "".join(out) + "$")


def skills_for(rel, project):
    data = json.loads((project / ".claude" / "hooks" / "skill-map.json").read_text())
    return [e["skill"] for e in data["skills"] if any(glob_to_regex(g).match(rel) for g in e["paths"])]


def target_path(payload):
    tool_input = payload.get("tool_input") or {}
    raw = tool_input.get("file_path") or tool_input.get("notebook_path") or tool_input.get("path")
    if not isinstance(raw, str) or not raw:
        return None
    p = pathlib.Path(raw)
    if not p.is_absolute():
        p = pathlib.Path(payload.get("cwd") or os.getcwd()) / p
    return p.resolve()


def emit(obj):
    sys.stdout.write(json.dumps(obj))


def preloaded_skills(project, agent_type):
    """Skills a subagent already has in context via its `skills:` frontmatter list."""
    if not agent_type or not re.fullmatch(r"[A-Za-z0-9_:-]+", agent_type):
        return set()
    try:
        text = (project / ".claude" / "agents" / f"{agent_type}.md").read_text()
    except OSError:
        return set()
    front = text.split("---")[1] if text.startswith("---") else ""
    names, in_list = set(), False
    for line in front.splitlines():
        if line.startswith("skills:"):
            in_list = True
        elif in_list and line.startswith("  - "):
            names.add(line[4:].strip().strip("\"'"))
        elif in_list:
            in_list = False
    return names


def mode_pre(payload):
    tool = payload.get("tool_name")
    if tool not in READ_TOOLS | WRITE_TOOLS or not isinstance(payload.get("tool_input"), dict):
        return
    project = project_dir(payload)
    target = target_path(payload)
    if target is None:
        return
    try:
        rel = target.relative_to(project).as_posix()
    except ValueError:
        return
    if rel.startswith(".claude/"):  # harness files are never domain-mapped; keeps a broken map repairable
        return
    try:
        wanted = skills_for(rel, project)
    except Exception as exc:  # broken map must not silently disable the guard for writes
        if tool in WRITE_TOOLS:
            emit({"hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": (
                    f"226os domain-skill guard: .claude/hooks/skill-map.json is unreadable or invalid "
                    f"({type(exc).__name__}: {exc}). Fix it before editing {rel}."),
            }})
        return
    if not wanted:
        return
    sfile = state_file(payload)
    raw = load_state(sfile)
    loaded = {n for n in raw if not n.startswith(NOTIFIED)} | preloaded_skills(project, payload.get("agent_type"))
    missing = [s for s in wanted if s not in loaded]
    if not missing:
        return

    broken = [n for n in missing if not (project / ".claude" / "skills" / n / "SKILL.md").is_file()]
    if broken:
        if tool in WRITE_TOOLS:
            emit({"hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": (
                    f"226os domain-skill guard: mapped skill(s) {', '.join(broken)} are missing "
                    f"(.claude/skills/<name>/SKILL.md). Fix .claude/hooks/skill-map.json or restore the skill before editing {rel}."),
            }})
        return

    names = ", ".join(missing)
    directive = (f"226os domain-skill guard: `{rel}` is governed by project skill(s) {names}. "
                 f"Load each one with the Skill tool (skill: {names}) and apply its rules; "
                 "these are binding project rules for this domain.")

    if tool in READ_TOOLS:
        notes = {NOTIFIED + n for n in missing}
        if notes <= raw:
            return
        save_state(sfile, raw | notes)
        emit({"hookSpecificOutput": {"hookEventName": "PreToolUse", "additionalContext": directive}})
    else:
        emit({"hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": directive + " Edits here stay blocked until they are loaded in this context; "
                                                    "then re-check your planned change against them and retry the same tool call.",
        }})


def mode_post_skill(payload):
    tool_input = payload.get("tool_input")
    if not isinstance(tool_input, dict):
        return
    name = tool_input.get("skill") or tool_input.get("name") or tool_input.get("command")
    if not name:
        return
    name = str(name).lstrip("/").split()[0]
    sfile = state_file(payload)
    save_state(sfile, load_state(sfile) | {name})


def mode_session_start(payload):
    if payload.get("source") in RESET_SOURCES:
        try:
            state_file(payload).unlink()
        except FileNotFoundError:
            pass


def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else ""
    try:
        payload = json.loads(sys.stdin.read() or "{}")
        if not isinstance(payload, dict):
            return 0
        {"pre": mode_pre, "post-skill": mode_post_skill, "session-start": mode_session_start}[mode](payload)
    except Exception as exc:  # never crash: a hook error must not leak a traceback; writes are guarded in mode_pre
        print(f"domain_skills: {type(exc).__name__}: {exc}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
