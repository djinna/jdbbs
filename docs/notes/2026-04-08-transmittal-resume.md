# Prodcal resume notes — 2026-04-08

Current goal in progress
- Patch Manuscript Transmittal checklist/backmatter to support an explicit 3-state status model:
  - included now
  - coming later
  - not in book
- Also fold in only closely related low-distraction items now:
  - add ISBN (EPUB) field to Book Information
  - add clearer autosave note/status copy if easy while touching transmittal UI

User-confirmed product decisions
- New project calendars should seed from the standard book-production workflow by default (not blank) — queued, not being implemented in this patch.
- Calendar/task dated history is desirable but likely next major revision unless it turns out cheap.
- Checklist UX should be clearer for non-production users.

Manual QA state before interruption
- Public site verified by user:
  - landing page client-code flow works
  - VGR portal works
  - new project creation worked
  - fresh-start project landed on transmittal first; user said that seemed correct
  - transmittal edits persisted after reload
- Gap found:
  - fresh project calendar is blank because current client-project creation only creates the project row and does not seed tasks

Queued items currently held
- add third ISBN field for EPUB in Book Information
- add visible autosave/realtime-save note in transmittal
- seed standard workflow when creating a new project calendar
- add dated history/versioning for calendar/tasks
- checklist/backmatter explicit “not in book” state and clearer terminology

Code investigation findings
- Current checklist/backmatter data model only has:
  - here_now
  - to_come_when
- No explicit negative state exists in defaults or UI.
- Relevant files:
  - srv/transmittal.go
  - srv/static/transmittal.js
- Fresh project creation path that leaves calendar blank:
  - srv/client.go handleClientCreateProject only calls CreateProject and does not seed tasks.

TDD progress already started
1. Added new untracked test file:
   - srv/transmittal_test.go

2. Tests added there:
   - TestTransmittalDefaultsIncludeEPUBISBNAndChecklistStatus
   - TestDuplicateTransmittalClearsChecklistStatusesAndDates

3. RED state already confirmed with:
   go test ./srv -run 'TestTransmittalDefaultsIncludeEPUBISBNAndChecklistStatus|TestDuplicateTransmittalClearsChecklistStatusesAndDates' -count=1

4. Observed failures from RED run:
   - default book data does not include isbn_epub
   - duplicate test initially got 400 because test used source project id in /api/transmittals/{id}/duplicate path, but that route expects project id as an integer path param and the test needed string conversion cleanup

Important current file state
- git status showed:
  - ?? srv/transmittal_test.go
- No source patches had been applied yet to transmittal.go or transmittal.js at the time these notes were written.

Planned implementation once test file is fixed
1. srv/transmittal.go
- In defaultTransmittalData():
  - add book.isbn_epub
  - add status:"" to each checklist item
  - add status:"" to each backmatter item
- In duplicate transmittal reset logic:
  - clear isbn_epub
  - clear checklist item status
  - clear backmatter item status

2. srv/static/transmittal.js
- In renderBookSection():
  - add textField('ISBN (EPUB)', 'book.isbn_epub')
- In calcCompletion():
  - count checklist/backmatter item as filled when status is set
  - preserve backward compatibility for older saved rows that only have here_now / to_come_when
- In renderChecklistSection():
  - replace current Here + To Come-only row model with explicit status control per row
  - recommended status values:
    - included
    - later
    - not_in_book
  - if status=included:
    - set here_now=true
    - clear to_come_when
  - if status=later:
    - set here_now=false
    - enable date field
  - if status=not_in_book:
    - set here_now=false
    - clear to_come_when
    - disable date field
  - preserve compatibility with old data by deriving temporary status on render if status missing:
    - here_now => included
    - to_come_when => later
    - else blank
- Apply same pattern to backmatter rows.
- While touching header, consider adding a small explicit autosave note near tx-save-status if easy.

Suggested plain-language labels for current test
- Section header can stay “Manuscript Checklist” for now
- Row status labels should be plain:
  - In manuscript now
  - Coming later
  - Not included
- Date field label/placeholder:
  - Expected date

Verification after GREEN
- targeted tests:
  go test ./srv -run 'TestTransmittalDefaultsIncludeEPUBISBNAndChecklistStatus|TestDuplicateTransmittalClearsChecklistStatusesAndDates' -count=1
- broader transmittal/client tests:
  go test ./srv -run 'Transmittal|Client' -count=1
- if source changed substantially, run full package:
  go test ./srv -count=1

Manual smoke to ask user for after deploy
- open current test book transmittal
- verify new ISBN EPUB field is present
- try checklist rows with:
  - In manuscript now
  - Coming later + date
  - Not included
- confirm date disables/clears correctly where expected
- reload page and verify persistence

Relevant note about deploy environment
- Host has duplicate systemd units:
  - srv.service
  - prodcal.service
- Both point at same binary/port 8000; restarts can race.
- Current known-safe reality during previous smoke pass:
  - prodcal.service was the live listener
  - srv.service was flapping with bind: address already in use
- Be careful at deploy time.
