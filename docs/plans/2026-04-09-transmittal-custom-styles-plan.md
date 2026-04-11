# Transmittal Custom Styles → Spec → Word Template Implementation Plan

> For Hermes: Use subagent-driven-development skill to implement this plan task-by-task.

Goal: Add a structured custom-style path from the Prodcal Manuscript Transmittal UI into book_specs.custom_styles so real projects can request custom styles, export a Word template with those styles present, and then exercise EPUB/Typst conversion against actual styled DOCX input.

Architecture: Keep the first pass narrow and explicit. Add a small structured custom-styles array to the transmittal data model/UI, map it directly into spec.custom_styles in prodcal, and rely on the already-existing generate-word-template.py support for custom_styles. Do not try to solve arbitrary EPUB/Typst rendering semantics in the same pass; this phase should only guarantee structured capture and template export.

Tech Stack: Go backend (Prodcal), vanilla JS transmittal SPA, SQLite-backed JSON blobs, Python docx template generator in book-production.

---

## Current verified state

What already exists:
- Prodcal transmittal data includes free-text editing fields:
  - editing.special_characters
  - editing.math_formulas
  - editing.instructions
- Prodcal book spec defaults include:
  - custom_styles: []
- book-production/scripts/generate-word-template.py already reads spec.custom_styles and creates paragraph/character styles in the generated .docx.
- Existing docs anticipate this flow:
  - book-production/docs/TRANSMITTAL_TO_TEMPLATE_FLOW.md
  - book-production/templates/manuscript-transmittal-form.md
  - book-production/docs/AUTHOR_GUIDELINES.md

What is not yet implemented:
- No structured custom-style array in live Prodcal transmittal data/UI.
- Pull Transmittal → Spec does not currently populate spec.custom_styles.
- No regression tests for transmittal custom styles reaching the spec.
- No end-to-end proof yet that a generated Word template carries the requested custom styles.

Non-goals for this phase:
- Do not fully solve EPUB/Typst custom-style rendering semantics yet.
- Do not implement automatic NLP parsing of free-text instructions into styles.
- Do not redesign the entire transmittal form.

---

## Data model to add

Add a structured array on transmittal data:

```json
"custom_styles": [
  {
    "name": "vgr-tweet",
    "type": "paragraph",
    "description": "Tweet / short social post"
  },
  {
    "name": "vgr-handle",
    "type": "character",
    "description": "Social media handle"
  }
]
```

Rules:
- name: required, lowercase-ish free text accepted by UI but should be trimmed; user is already operating with shortcode-first naming convention.
- type: one of paragraph | character
- description: required enough for editorial clarity, but can be simple text.

Spec mapping target:

```json
"custom_styles": [
  {
    "name": "vgr-tweet",
    "word_style": "vgr-tweet",
    "type": "paragraph",
    "description": "Tweet / short social post"
  }
]
```

Rationale:
- generate-word-template.py already prefers word_style or name.
- using both name and word_style makes intent explicit and avoids later ambiguity.

---

## Files expected to change

Prodcal:
- Modify: `/home/exedev/prodcal/srv/transmittal.go`
- Modify: `/home/exedev/prodcal/srv/static/transmittal.js`
- Modify: `/home/exedev/prodcal/srv/bookspecs.go`
- Create or modify tests: `/home/exedev/prodcal/srv/transmittal_test.go`
- Create or modify tests: `/home/exedev/prodcal/srv/bookspecs_test.go` or nearest existing spec-related test file

Possible docs:
- Create: `/home/exedev/prodcal/docs/plans/2026-04-09-transmittal-custom-styles-plan.md` (this file already created)

Book-production verification only (likely no code change in first pass):
- Read/verify: `/home/exedev/book-production/scripts/generate-word-template.py`

---

## Task 1: Add failing backend test for transmittal defaults

Objective: Prove the default transmittal payload includes a structured custom_styles array.

Files:
- Modify: `/home/exedev/prodcal/srv/transmittal_test.go`

Step 1: Write failing test

Add a test like:

```go
func TestTransmittalDefaultsIncludeCustomStylesArray(t *testing.T) {
    _, ts, cleanup := testServer(t)
    defer cleanup()

    resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
        "name": "Custom Styles Defaults",
        "start_date": "2026-04-09",
        "client_slug": "vgr",
        "project_slug": "custom-style-defaults",
    })
    if resp.StatusCode != 201 {
        t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
    }
    var project map[string]any
    decodeJSON(t, resp, &project)
    pid := itoa(int64(project["ID"].(float64)))

    resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/transmittal", nil)
    if resp.StatusCode != 200 {
        t.Fatalf("get transmittal: expected 200, got %d", resp.StatusCode)
    }
    var tx map[string]any
    decodeJSON(t, resp, &tx)
    data := tx["data"].(map[string]any)

    styles, ok := data["custom_styles"].([]any)
    if !ok {
        t.Fatalf("expected custom_styles array in defaults, got %T", data["custom_styles"])
    }
    if len(styles) != 0 {
        t.Fatalf("expected empty custom_styles defaults, got %d", len(styles))
    }
}
```

Step 2: Run test to verify failure

Run:
`go test ./srv -run 'TestTransmittalDefaultsIncludeCustomStylesArray' -count=1`

Expected: FAIL because custom_styles is missing from transmittal defaults.

Step 3: Minimal implementation
- In `srv/transmittal.go`, add:

```json
"custom_styles": []
```

into `defaultTransmittalData()`.

Step 4: Run test to verify pass

Run:
`go test ./srv -run 'TestTransmittalDefaultsIncludeCustomStylesArray' -count=1`

Expected: PASS

Step 5: Commit

```bash
git add srv/transmittal.go srv/transmittal_test.go
git commit -m "test(transmittal): add custom styles defaults coverage"
```

---

## Task 2: Add structured custom styles UI to transmittal page

Objective: Let editors enter custom styles directly in the Manuscript Transmittal UI.

Files:
- Modify: `/home/exedev/prodcal/srv/static/transmittal.js`

UI shape:
- New section under editorial/instructions area or adjacent to editing section.
- Repeating rows with fields:
  - Style name
  - Type (paragraph/character)
  - Purpose / description
- Buttons:
  - Add custom style
  - Remove row

Recommended minimal rendering shape:

```js
function renderCustomStylesSection() {
  const styles = state.transmittal.data.custom_styles || [];
  return h('div', { className: 'tx-section' },
    h('div', { className: 'tx-section-header' }, 'Custom Styles'),
    h('div', { className: 'tx-help' }, 'Add any project-specific Word styles needed for this manuscript.'),
    ...styles.map((style, i) =>
      h('div', { className: 'tx-row-3' },
        textField('Style name', `custom_styles.${i}.name`),
        selectField('Type', `custom_styles.${i}.type`, [
          ['paragraph', 'Paragraph'],
          ['character', 'Character'],
        ]),
        textField('Purpose', `custom_styles.${i}.description`)
      )
    ),
    h('button', { className: 'btn btn-sm', onClick: addCustomStyle }, '+ Add custom style')
  );
}
```

Important implementation note:
- The current `setField()` helper supports nested object paths, not array row creation/removal ergonomics by itself.
- For add/remove actions, mutate `state.transmittal.data.custom_styles` array directly, call `setField('custom_styles', styles)`, then `render()`.

Step 1: Write/extend a test first if practical at the server JSON level, not browser level.
- If no frontend automation exists, skip new UI test and rely on manual verification after backend coverage.

Step 2: Implement minimal UI.

Step 3: Verify no JS syntax errors by loading the page and checking console in manual QA later.

Step 4: Commit

```bash
git add srv/static/transmittal.js
git commit -m "feat(transmittal): add structured custom style fields"
```

---

## Task 3: Map transmittal custom styles into book spec

Objective: Ensure Pull Transmittal → Spec copies structured custom styles into spec.custom_styles.

Files:
- Modify: `/home/exedev/prodcal/srv/bookspecs.go`
- Create/modify tests: `/home/exedev/prodcal/srv/bookspecs_test.go` (or nearest appropriate test file)

Step 1: Write failing test

Add a test that:
- creates a project
- writes transmittal data with custom_styles
- POSTs to `/api/projects/{id}/book-spec/pull-transmittal`
- asserts response `data.custom_styles` contains the expected mapped rows

Expected mapped row shape:

```json
{
  "name": "vgr-tweet",
  "word_style": "vgr-tweet",
  "type": "paragraph",
  "description": "Tweet / short social post"
}
```

Step 2: Run test to verify failure

Run:
`go test ./srv -run 'TestPullTransmittalMapsCustomStyles' -count=1`

Expected: FAIL because bookspec mapping does not yet populate custom_styles.

Step 3: Minimal implementation

In `srv/bookspecs.go`, after current metadata/front/back matter mapping, add something like:

```go
if styles, ok := tx["custom_styles"].([]any); ok {
    var mapped []any
    for _, item := range styles {
        m, ok := item.(map[string]any)
        if !ok {
            continue
        }
        name, _ := m["name"].(string)
        styleType, _ := m["type"].(string)
        desc, _ := m["description"].(string)
        name = strings.TrimSpace(name)
        styleType = strings.TrimSpace(styleType)
        desc = strings.TrimSpace(desc)
        if name == "" {
            continue
        }
        if styleType == "" {
            styleType = "paragraph"
        }
        mapped = append(mapped, map[string]any{
            "name": name,
            "word_style": name,
            "type": styleType,
            "description": desc,
        })
    }
    specData["custom_styles"] = mapped
}
```

Also ensure `strings` is imported if needed.

Step 4: Run test to verify pass

Run:
`go test ./srv -run 'TestPullTransmittalMapsCustomStyles' -count=1`

Expected: PASS

Step 5: Run related suite

Run:
`go test ./srv -run 'Transmittal|BookSpec|Client' -count=1`

Step 6: Commit

```bash
git add srv/bookspecs.go srv/bookspecs_test.go
git commit -m "feat(bookspec): map transmittal custom styles into spec"
```

---

## Task 4: Verify Word template generation consumes mapped custom styles

Objective: Prove the existing Python template generator works with the spec shape produced by Prodcal.

Files:
- Read/verify: `/home/exedev/book-production/scripts/generate-word-template.py`
- Optional create: a temporary JSON fixture under `/home/exedev/book-production/test-cases/`

Step 1: Create a minimal spec fixture containing:

```json
{
  "metadata": {"title": "Test", "author": "Tester"},
  "typography": {},
  "headings": {},
  "elements": {},
  "custom_styles": [
    {"name": "vgr-tweet", "word_style": "vgr-tweet", "type": "paragraph", "description": "Tweet block"},
    {"name": "vgr-handle", "word_style": "vgr-handle", "type": "character", "description": "Handle inline"}
  ]
}
```

Step 2: Run generator

Run:
`python3 /home/exedev/book-production/scripts/generate-word-template.py --spec-file /path/to/spec.json --output /tmp/custom-style-template.docx`

Expected: file created successfully.

Step 3: Manual verification
- Open the .docx in Word and confirm both styles exist.
- If Word inspection is not available in-session, note this as a manual verification step for the user.

Step 4: Do not change generate-word-template.py unless this verification actually fails.

---

## Task 5: Manual QA checklist for the next real-project test

Objective: Define the exact test sequence for your real manuscript.

Manual test sequence:
1. Open live Manuscript Transmittal for the real test project.
2. Add two custom styles, e.g.:
   - `vgr-tweet` — paragraph — Tweet / short social post
   - `vgr-handle` — character — Social media handle
3. Save/reload transmittal; confirm persistence.
4. In admin, pull transmittal into spec.
5. Generate the Word template.
6. Open the Word template; confirm those styles exist.
7. Move representative text into the template and apply those styles.
8. Run EPUB path first.
9. Run edge-case detector against that DOCX and verify any stray local formatting is flagged.
10. Only after EPUB review, move to the Typst/print path.

Expected pass criteria for this phase:
- styles can be entered structurally in transmittal
- styles persist after reload
- styles appear in pulled spec.custom_styles
- generated Word template includes them

Not yet required for this phase:
- guaranteed custom semantic rendering in EPUB/Typst output

---

## Task 6: Save/ship and checkpoint

Objective: Land the structured custom-style capture feature cleanly before broader EPUB testing.

Suggested commit after implementation is complete:

```bash
git add srv/transmittal.go srv/static/transmittal.js srv/bookspecs.go srv/transmittal_test.go srv/bookspecs_test.go docs/plans/2026-04-09-transmittal-custom-styles-plan.md
git commit -m "feat(transmittal): add structured custom style capture"
```

Suggested checkpoint after live verification:

```bash
git tag -a checkpoint-2026-04-09-transmittal-custom-styles -m "Checkpoint: transmittal custom styles now map into book specs and are ready for Word template verification"
git push origin checkpoint-2026-04-09-transmittal-custom-styles
```

---

## Final recommendation

Implement this phase before returning to the broader EPUB workflow. Without it, the real-project test you want to run will be partly manual and partly speculative. With it, the next EPUB test cycle becomes concrete and much more meaningful.
