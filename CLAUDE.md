# gstack

Use the `/browse` skill from gstack for all web browsing. Never use `mcp__claude-in-chrome__*` tools.

Available skills: `/office-hours`, `/plan-ceo-review`, `/plan-eng-review`, `/plan-design-review`, `/design-consultation`, `/review`, `/ship`, `/land-and-deploy`, `/canary`, `/benchmark`, `/browse`, `/qa`, `/qa-only`, `/design-review`, `/setup-browser-cookies`, `/setup-deploy`, `/retro`, `/investigate`, `/document-release`, `/codex`, `/careful`, `/freeze`, `/guard`, `/unfreeze`, `/gstack-upgrade`.

If gstack skills are not working, run `cd .claude/skills/gstack && ./setup` to build the binary and register skills.

# VM ssh — user-run, not assistant-run

The harness's auto-mode classifier blocks `ssh exedev@jdbbs.exe.xyz '...'` from the assistant (treats prod VM access as needing explicit approval). Don't try to run ssh yourself. Instead, write a single concatenated shell command (one line, `&&`-chained or `;`-chained, multi-step pre-flights and compiles wrapped in one ssh invocation) and ask the user to paste the output back. The user has shell access and runs it themselves; you stay read-only on local files and operate on what they paste in.

Examples of commands to give the user:
- Pre-flight: `ssh exedev@jdbbs.exe.xyz 'systemctl is-active prodcal && git -C /home/exedev/prodcal log --oneline -1 && typst --version'`
- Compile + scp back: `ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && typst compile --root . --font-path typesetting/fonts manuscripts/ghosts/main.typ /tmp/ghosts-typst-v2.pdf && pdfinfo /tmp/ghosts-typst-v2.pdf | grep -E "Pages|Page size"' && scp exedev@jdbbs.exe.xyz:/tmp/ghosts-typst-v2.pdf scratch/`

# Scratch / throwaway files

For throwaway intermediate files (assembled DOCX for testing, downloaded PDFs for diffing, etc.) use `./scratch/` at the repo root. It's gitignored. Don't use `/tmp` — macOS Finder hides it by default, which is friction when the user needs to interact with the file via UI (file picker, etc.). Don't scatter `tmp/` subdirs across the source tree (e.g. inside `manuscripts/`) — they pollute the source areas.
