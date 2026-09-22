Continue ProdCal (workshop week; Tue 22 Sep, day 2 live) in /home/exedev/prodcal. Read AGENTS.md, docs/CONTEXT-HYGIENE.md, then only addendum 24 of docs/reviews/SESSION-HANDOFF-2026-09-18.md and scratch/run/CHECKLIST.md (list 5, short).

First action: restart the runpage on this conversation's id (tmux kill-session -t runpage; tmux new-session -d -s runpage -c /home/exedev/prodcal "RUNPAGE_CHAT_CONV=$SHELLEY_CONVERSATION_ID scripts/run-page.sh"). Punch list: https://jdbbs.exe.xyz:8766/. Answer on items via POST localhost:8766/note {"id","text","who":"shelley"}. Jenna prefers to discuss what you see before you build; propose, then act.

Freeze on (hotfix-only through Wed 23; smoke only on mcheck = project 17, never pinstitute = 22). Before every make build && sudo systemctl restart prodcal: sqlite3 db.sqlite3 "select count(*) from books where status='converting'" must be 0. Version shows on the factory strip as v MMDD.hash (/api/version).

Next action: 0.36 — Seapunk before/after review, delegated to a subagent (read-only; write the brief to scratch/seapunk/BRIEF.md modelled on scratch/fotis/BRIEF.md, one-line prompt, tell it to create the report file skeleton first). Project 32 (seapunkstudios/book-001), books 46–55 (sources in the DB blobs, proofs in book_outputs, events in factory_events). Deliverable: docs/reviews/SEAPUNK-REVIEW-2026-09-22.md with (a) suggestions for the factory, (b) a note Jenna can forward to the Seapunk team. Read only its summary; post the gist on item 0.36.

Then: Floor watch (Sam 31, Toby 29, Fotis 30 — can retry after pruning declarations, Ellen 33, a new participant); hotfixes between slices. Thu 24: 5.25 slices 4–7 incl. verse-family coalesce, 5.26, 5.24/5.27, 0.34.

Push with git push git@github.com:djinna/jdbbs.git main. Commit before returning to Jenna; write addendum 25 + this file's successor at ~50–65 % context.
