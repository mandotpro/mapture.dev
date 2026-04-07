---
name: agent-docs
description: Use when adding, updating, or reorganizing project rules, conventions, or architectural decisions that future AI coding agents should follow. Invoke proactively whenever the user says something like "add a rule", "from now on...", "we always/never...", "convention is...", "document this decision", or when a recurring correction should become durable guidance. Keeps AGENTS.md as the single source of truth and ensures tool-specific files (CLAUDE.md, .cursorrules, .github/copilot-instructions.md, etc.) remain symlinks or mirrors to it.
---

# agent-docs

Mapture follows the [agents.md](https://agents.md) convention: [`AGENTS.md`](../../../AGENTS.md) is the **single canonical** file for AI agent guidance. Tool-specific files (`CLAUDE.md`, `.cursorrules`, `.github/copilot-instructions.md`, `.windsurfrules`, etc.) are symlinks to `AGENTS.md` when they exist at all. This skill exists so rules stay organized as they accumulate and so tool-specific files never drift.

## When to invoke this skill

Trigger on any of these signals — don't wait to be asked explicitly:

- The user states a rule or preference that should outlive the current conversation: "from now on", "always", "never", "we prefer", "don't do X here", "the convention is".
- The user corrects something twice. Second correction = durable rule.
- The user documents an architectural decision, a naming choice, or a non-obvious constraint.
- The user asks to "update AGENTS.md", "update CLAUDE.md", "add this to the agent rules", "document this for future agents".
- A new agent tool (Cursor, Copilot, Windsurf, Codex, etc.) is being onboarded and needs its own instructions file.

Do **not** invoke for ephemeral in-conversation context, for one-off corrections to the current task, or for anything already derivable from the code itself (file paths, existing APIs, git history). Those are not rules — they are transient.

## What counts as a rule

A rule is a short, actionable statement an agent can follow without re-deriving reasoning. Good rules answer one of:

- **What to do** — "Cite PRD sections in doc comments."
- **What not to do** — "Don't put business logic in `cmd/`."
- **How to choose** — "When a feature belongs to a later milestone, defer it."

Avoid:

- Restating obvious software-engineering practices (use meaningful names, write tests, etc.).
- Explaining code that can be read directly.
- Pasting large rationales. Keep the rule one line; link to PRD section or commit if context is needed.

## How to add a rule

1. **Always edit [`AGENTS.md`](../../../AGENTS.md), never `CLAUDE.md` directly.** `CLAUDE.md` is a symlink — editing it works, but the convention is to treat `AGENTS.md` as the name of record. Verify before editing with `ls -la CLAUDE.md`; if it is a regular file rather than a symlink, stop and flag it to the user (the symlink has been clobbered and needs restoring).
2. **Find the `## Project rules` section** in `AGENTS.md`, terminated by the `<!-- agent-docs:rules:end -->` marker.
3. **Check for duplicates or near-duplicates.** If the new rule refines an existing one, edit in place — do not append a second rule that says almost the same thing.
4. **Write the rule as a bullet** in the form `**Short imperative title.** One-sentence body. PRD §X if relevant.` Keep it under ~200 characters.
5. **Insert before the `<!-- agent-docs:rules:end -->` marker** so the marker always stays at the end.
6. **Confirm with the user**: briefly show the rule you added. Don't be verbose.

## How to update or remove a rule

- **Update**: edit the existing bullet in place. If the rule is being softened or narrowed, update the wording to reflect the new scope. Don't add "previously we said X, now Y" — just state the current rule.
- **Remove**: delete the bullet. If removing because the rule became obsolete (e.g. a milestone passed), also check whether the PRD section it cites needs an update and flag that to the user.
- **Reorganize**: if the rules section grows past ~15 bullets, propose splitting into sub-headings (e.g. `### Code conventions`, `### Process`, `### Architecture`). Do not split proactively before then.

## Tool-specific files

`AGENTS.md` is canonical. Every tool-specific file in the repo must either be a symlink to `AGENTS.md` or not exist at all. Known targets:

| File | Tool | Default handling |
|---|---|---|
| `CLAUDE.md` | Claude Code | Symlink to `AGENTS.md` (already in place) |
| `.cursorrules` | Cursor (legacy) | Symlink when user adopts Cursor |
| `.cursor/rules/mapture.mdc` | Cursor (modern) | Cursor's modern format uses frontmatter — create as a real file that *includes* AGENTS.md content if the user opts in |
| `.github/copilot-instructions.md` | GitHub Copilot | Symlink when user adopts Copilot |
| `.windsurfrules` | Windsurf | Symlink when user adopts Windsurf |
| `.aider.conf.yml` / `.aider.conf.md` | Aider | Symlink when user adopts Aider |

**Rules for tool files:**

1. **Do not create tool files preemptively.** Only create one when the user says they are using that tool, or explicitly asks for it. Stale files in the repo are worse than missing ones.
2. **Prefer `ln -s AGENTS.md <target>`** over duplicating content. One file, zero drift.
3. **If a tool requires frontmatter or a different format** (e.g. Cursor `.mdc` rules), create a real file whose body is a short pointer like `See AGENTS.md at the repo root for the canonical project guidance.` rather than copy-pasting the full content. Copies drift; pointers don't.
4. **If any tool-specific file ever stops being a symlink** (e.g. someone committed a regular file by mistake), treat `AGENTS.md` as authoritative and restore the symlink. Never merge divergent content back into `AGENTS.md` without user review.
5. **Never commit `.claude/settings.local.json`** — it's user-local and already gitignored.

## Boundaries

- This skill edits `AGENTS.md` and the symlinks/pointers above. It does not edit Go code, YAML catalogs, or the PRD.
- If the user is actually asking for a PRD change (not an agent rule), redirect them to edit `_docs/mapture-dev-prd-v1.md` directly — the PRD is the product spec, `AGENTS.md` is agent guidance about how to work in the repo.
- If a proposed rule contradicts the PRD, surface the conflict to the user before adding it. Don't silently encode contradictions.
