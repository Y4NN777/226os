"""Tests for .claude/hooks/domain_skills.py.

Run from the repository root:
    python3 -m unittest discover -s .claude/hooks/tests -v
"""

import json
import os
import pathlib
import subprocess
import sys
import tempfile
import unittest

REPO = pathlib.Path(__file__).resolve().parents[3]
HOOK = REPO / ".claude" / "hooks" / "domain_skills.py"
REAL_MAP = REPO / ".claude" / "hooks" / "skill-map.json"

FAKE_MAP = {
    "skills": [
        {"skill": "engineering-principles", "paths": ["src/**", "db/**", "api/**"]},
        {"skill": "go-backend", "paths": ["src/api/**"]},
        {"skill": "tdd-unit", "paths": ["**/*_test.go", "**/*.test.ts"]},
        {"skill": "ci-cd-delivery", "paths": ["Makefile", ".github/workflows/**"]},
    ]
}


class HookHarness(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        root = pathlib.Path(self.tmp.name)
        self.project = root / "project"
        self.state = root / "state"
        (self.project / ".claude" / "hooks").mkdir(parents=True)
        (self.project / ".claude" / "hooks" / "skill-map.json").write_text(json.dumps(FAKE_MAP))
        for entry in FAKE_MAP["skills"]:
            d = self.project / ".claude" / "skills" / entry["skill"]
            d.mkdir(parents=True)
            (d / "SKILL.md").write_text(
                f"---\nname: {entry['skill']}\ndescription: x\n---\n# BODY-{entry['skill']}\nrule\n"
            )

    def tearDown(self):
        self.tmp.cleanup()

    def run_hook(self, mode, payload, raw=None):
        env = dict(os.environ, CLAUDE_PROJECT_DIR=str(self.project), CLAUDE_SKILL_STATE_DIR=str(self.state))
        proc = subprocess.run(
            [sys.executable, str(HOOK), mode],
            input=raw if raw is not None else json.dumps(payload),
            capture_output=True, text=True, env=env, timeout=20,
        )
        self.assertEqual(proc.returncode, 0, proc.stderr)
        return json.loads(proc.stdout) if proc.stdout.strip() else None

    def pre(self, tool, rel, session="s1", **extra):
        tool_input = {"notebook_path" if tool == "NotebookEdit" else "file_path": str(self.project / rel)}
        payload = {"session_id": session, "hook_event_name": "PreToolUse", "cwd": str(self.project),
                   "tool_name": tool, "tool_input": tool_input, **extra}
        return self.run_hook("pre", payload)


class PreToolUseTests(HookHarness):
    def load(self, *skills, session="s1", **extra):
        for name in skills:
            self.run_hook("post-skill", {"session_id": session, "tool_name": "Skill",
                                         "tool_input": {"skill": name}, **extra})

    def reason(self, out):
        hso = out["hookSpecificOutput"]
        self.assertEqual(hso["permissionDecision"], "deny")
        return hso["permissionDecisionReason"]

    def test_non_domain_path_is_silent(self):
        self.assertIsNone(self.pre("Edit", "README.md"))

    def test_first_write_is_denied_naming_missing_skills_without_bodies(self):
        reason = self.reason(self.pre("Write", "src/api/server.go"))
        self.assertIn("go-backend", reason)
        self.assertIn("engineering-principles", reason)
        self.assertIn("Skill tool", reason)
        self.assertNotIn("tdd-unit", reason)
        self.assertNotIn("BODY-", reason, "bodies are loaded by the Skill tool, not inlined")
        self.assertLess(len(reason), 1500, "must stay under the hook-output preview limit")

    def test_write_stays_denied_until_skills_are_loaded(self):
        self.pre("Edit", "src/api/server.go")
        self.reason(self.pre("Edit", "src/api/server.go"))
        self.load("go-backend")
        self.assertIn("engineering-principles", self.reason(self.pre("Edit", "src/api/server.go")))
        self.load("engineering-principles")
        self.assertIsNone(self.pre("Edit", "src/api/server.go"))

    def test_loaded_skills_are_cached_and_only_new_ones_requested(self):
        self.load("go-backend", "engineering-principles")
        self.assertIsNone(self.pre("Edit", "src/api/other.go"))
        reason = self.reason(self.pre("Edit", "src/api/server_test.go"))
        self.assertIn("tdd-unit", reason)
        self.assertNotIn("go-backend", reason)

    def test_other_session_is_not_cached(self):
        self.load("go-backend", "engineering-principles", session="s1")
        self.reason(self.pre("Edit", "src/api/server.go", session="s2"))

    def test_subagent_has_its_own_cache(self):
        self.load("go-backend", "engineering-principles")
        self.reason(self.pre("Edit", "src/api/server.go", agent_id="sub-1", agent_type="scout"))

    def test_read_adds_short_directive_once_and_does_not_unlock_writes(self):
        hso = self.pre("Read", "src/api/server.go")["hookSpecificOutput"]
        self.assertNotIn("permissionDecision", hso)
        self.assertIn("go-backend", hso["additionalContext"])
        self.assertIn("Skill tool", hso["additionalContext"])
        self.assertNotIn("BODY-", hso["additionalContext"])
        self.assertIsNone(self.pre("Read", "src/api/other.go"), "directive is not repeated")
        self.reason(self.pre("Write", "src/api/server.go"))

    def test_root_level_globs(self):
        self.assertIn("tdd-unit", self.reason(self.pre("Edit", "x_test.go")))
        self.assertIn("ci-cd-delivery", self.reason(self.pre("Edit", "Makefile", session="s9")))

    def test_multiedit_and_notebook_are_guarded(self):
        self.reason(self.pre("MultiEdit", "db/x.sql"))
        self.reason(self.pre("NotebookEdit", "src/web/a.ipynb", session="s3"))

    def test_path_outside_project_is_ignored(self):
        payload = {"session_id": "s1", "tool_name": "Edit", "tool_input": {"file_path": "/etc/src/api/x.go"}}
        self.assertIsNone(self.run_hook("pre", payload))

    def test_relative_path_resolves_against_cwd(self):
        payload = {"session_id": "s1", "cwd": str(self.project), "tool_name": "Edit",
                   "tool_input": {"file_path": "src/api/x.go"}}
        self.reason(self.run_hook("pre", payload))

    def test_malformed_input_fails_open_silently(self):
        self.assertIsNone(self.run_hook("pre", None, raw="not json"))

    def test_missing_skill_file_blocks_writes(self):
        (self.project / ".claude" / "skills" / "go-backend" / "SKILL.md").unlink()
        reason = self.reason(self.pre("Edit", "src/api/server.go"))
        self.assertIn("go-backend", reason)
        self.assertIn("missing", reason.lower())


class FailClosedAndPreloadTests(HookHarness):
    def test_broken_map_blocks_writes(self):
        (self.project / ".claude" / "hooks" / "skill-map.json").write_text("{not json")
        hso = self.pre("Edit", "src/api/server.go")["hookSpecificOutput"]
        self.assertEqual(hso["permissionDecision"], "deny")
        self.assertIn("skill-map.json", hso["permissionDecisionReason"])

    def test_broken_map_stays_editable(self):
        (self.project / ".claude" / "hooks" / "skill-map.json").write_text("{bad")
        self.assertIsNone(self.pre("Edit", ".claude/hooks/skill-map.json"))
        self.assertIsNone(self.pre("Write", ".claude/skills/go-backend/SKILL.md"))

    def test_unexpected_input_types_fail_open_without_crash(self):
        self.assertIsNone(self.run_hook("post-skill", {"session_id": "s1", "tool_name": "Skill", "tool_input": "go-backend"}))
        self.assertIsNone(self.run_hook("pre", {"session_id": "s1", "tool_name": "Edit", "tool_input": {"file_path": 123}}))

    def test_broken_map_does_not_block_reads(self):
        (self.project / ".claude" / "hooks" / "skill-map.json").unlink()
        self.assertIsNone(self.pre("Read", "src/api/server.go"))

    def test_non_dict_tool_input_fails_open(self):
        payload = {"session_id": "s1", "tool_name": "Edit", "tool_input": "oops"}
        self.assertIsNone(self.run_hook("pre", payload))

    def test_subagent_preloaded_skills_count_as_loaded(self):
        agents = self.project / ".claude" / "agents"
        agents.mkdir()
        (agents / "go-implementer.md").write_text(
            "---\nname: go-implementer\nskills:\n  - engineering-principles\n  - go-backend\n---\nbody\n")
        self.assertIsNone(self.pre("Edit", "src/api/server.go", agent_id="sub-1", agent_type="go-implementer"))
        hso = self.pre("Edit", "src/api/server_test.go", agent_id="sub-1", agent_type="go-implementer")["hookSpecificOutput"]
        self.assertIn("tdd-unit", hso["permissionDecisionReason"])
        self.assertNotIn("go-backend", hso["permissionDecisionReason"])


class SkillAndSessionTests(HookHarness):
    def test_slash_form_skill_name_is_normalised(self):
        self.run_hook("post-skill", {"session_id": "s1", "tool_name": "Skill", "tool_input": {"skill": "/go-backend extra"}})
        self.run_hook("post-skill", {"session_id": "s1", "tool_name": "Skill", "tool_input": {"skill": "engineering-principles"}})
        self.assertIsNone(self.pre("Edit", "src/api/server.go"))

    def test_compact_clears_cache(self):
        PreToolUseTests.load(self, "go-backend", "engineering-principles")
        self.run_hook("session-start", {"session_id": "s1", "source": "compact"})
        self.assertEqual(self.pre("Edit", "src/api/server.go")["hookSpecificOutput"]["permissionDecision"], "deny")

    def test_resume_keeps_cache(self):
        PreToolUseTests.load(self, "go-backend", "engineering-principles")
        self.run_hook("session-start", {"session_id": "s1", "source": "resume"})
        self.assertIsNone(self.pre("Edit", "src/api/server.go"))


class RealConfigTests(unittest.TestCase):
    """Consistency of the committed map, skills, agents and settings."""

    def setUp(self):
        self.map = json.loads(REAL_MAP.read_text())

    def test_every_mapped_skill_exists_with_matching_name(self):
        for entry in self.map["skills"]:
            path = REPO / ".claude" / "skills" / entry["skill"] / "SKILL.md"
            self.assertTrue(path.is_file(), f"missing {path}")
            self.assertIn(f"\nname: {entry['skill']}\n", "\n" + path.read_text().split("---")[1])

    def test_domain_skills_are_not_path_scoped(self):
        # A `paths:` frontmatter makes Claude Code hide the skill until a matching file is
        # touched ("Unknown skill" from the Skill tool, observed live on 2.1.322), which also
        # breaks subagent preloads. Path routing belongs to skill-map.json + the hook only.
        for entry in self.map["skills"]:
            front = (REPO / ".claude" / "skills" / entry["skill"] / "SKILL.md").read_text().split("---")[1]
            self.assertNotIn("\npaths:", "\n" + front, entry["skill"])

    def test_glob_translation_on_real_map(self):
        sys.path.insert(0, str(HOOK.parent))
        import domain_skills
        cases = {
            "src/api/internal/identity/session.go": {"engineering-principles", "go-backend", "security-authz"},
            "src/api/internal/identity/session_test.go": {"engineering-principles", "go-backend", "security-authz", "tdd-unit"},
            "db/migrations/0001_init.sql": {"engineering-principles", "postgres-data"},
            "api/openapi.yaml": {"engineering-principles", "api-contract"},
            "src/web/src/features/solve/Question.test.tsx": {"engineering-principles", "web-frontend", "tdd-unit"},
            "tests/e2e/solve.feature": {"bdd-acceptance"},
            ".github/workflows/ci.yml": {"ci-cd-delivery"},
            "Makefile": {"ci-cd-delivery"},
            "README.md": set(),
            "docs/SPEC.md": set(),
        }
        for rel, expected in cases.items():
            self.assertEqual(set(domain_skills.skills_for(rel, REPO)), expected, rel)

    def test_agent_preloaded_skills_exist(self):
        expected = {"go-implementer", "postgres-expert", "web-implementer", "skeptic", "lead"}
        seen = set()
        for agent in (REPO / ".claude" / "agents").glob("*.md"):
            lines = agent.read_text().split("---")[1].splitlines()
            names, in_skills = [], False
            for line in lines:
                if line.startswith("skills:"):
                    in_skills = True
                    continue
                if in_skills and line.startswith("  - "):
                    names.append(line[4:].strip())
                elif in_skills:
                    in_skills = False
            if names:
                seen.add(agent.stem)
            for name in names:
                skill = REPO / ".claude" / "skills" / name / "SKILL.md"
                self.assertTrue(skill.is_file(), f"{agent.name}: {name}")
                self.assertNotIn("disable-model-invocation: true", skill.read_text(), f"{name} cannot be preloaded")
        self.assertEqual(seen, expected)

    def test_writing_agents_can_load_skills(self):
        for agent in (REPO / ".claude" / "agents").glob("*.md"):
            front = agent.read_text().split("---")[1]
            tools = next((l.split(":", 1)[1] for l in front.splitlines() if l.startswith("tools:")), "")
            names = {t.strip() for t in tools.split(",")}
            if names & {"Edit", "Write"}:
                self.assertIn("Skill", names, agent.name)

    def test_settings_register_all_three_hooks(self):
        hooks = json.loads((REPO / ".claude" / "settings.json").read_text())["hooks"]
        commands = json.dumps(hooks)
        for mode in ("pre", "post-skill", "session-start"):
            self.assertIn(f"domain_skills.py {mode}", commands)


if __name__ == "__main__":
    unittest.main()
