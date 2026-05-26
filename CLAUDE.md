# gstack

Use the `/browse` skill from gstack for all web browsing. Never use `mcp__claude-in-chrome__*` tools.

Available skills: `/office-hours`, `/plan-ceo-review`, `/plan-eng-review`, `/plan-design-review`, `/design-consultation`, `/review`, `/ship`, `/land-and-deploy`, `/canary`, `/benchmark`, `/browse`, `/qa`, `/qa-only`, `/design-review`, `/setup-browser-cookies`, `/setup-deploy`, `/retro`, `/investigate`, `/document-release`, `/codex`, `/careful`, `/freeze`, `/guard`, `/unfreeze`, `/gstack-upgrade`.

If gstack skills are not working, run `cd .claude/skills/gstack && ./setup` to build the binary and register skills.

# Scratch / throwaway files

For throwaway intermediate files (assembled DOCX for testing, downloaded PDFs for diffing, etc.) use `./scratch/` at the repo root. It's gitignored. Don't use `/tmp` — macOS Finder hides it by default, which is friction when the user needs to interact with the file via UI (file picker, etc.). Don't scatter `tmp/` subdirs across the source tree (e.g. inside `manuscripts/`) — they pollute the source areas.
