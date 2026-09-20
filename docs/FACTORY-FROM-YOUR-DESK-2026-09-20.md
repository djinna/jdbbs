# Running your book through the factory — from your own desk

*jdbb studio · draft 1 · 20 September 2026*

This is for a press or an author who has a **Factory Pass** and would rather work
from a folder on their own computer than from the factory web page. It takes a
Word file and gives you back a typeset print PDF and an EPUB. There is nothing
to install.

---

## 1. What you have been given

When your pass was set up, the studio sent you three things. Have them to hand.

| | What it looks like | What it is for |
|---|---|---|
| **Project id** | a number, e.g. `22` | which book on the factory is yours |
| **Token** | a long string of letters and numbers | your key to that project — treat it as a password |
| **Factory page** | `https://jdbbs.exe.xyz/`*yourname*`/`*book*`/factory/` | the same factory in a browser, if you ever want to look |

The pass itself gives you: **unlimited proofs** (free; the PDF carries a small
*PROOF* line at the foot of every page), a set number of **finals** (the clean
PDF you send a printer), unlimited inspections and uploads, and six months of
storage.

**Keep the token private.** Anyone holding it can build on your project and
download your book. If it leaks, email j@djinna.com and it will be replaced.

---

## 2. Before you start: the manuscript

The factory typesets what you send; it does not edit. It needs:

- **One Word file (`.docx`).** From Word, Pages (*File → Export To → Word*), Google
  Docs (*File → Download → Microsoft Word*) or LibreOffice.
- **Chapter titles as Heading 1, sub-heads as Heading 2.** That is the whole
  contract. The factory finds your front matter, chapters and back matter by
  those headings and their wording ("Preface", "Notes", "Bibliography" …).
- **Special paragraphs marked in the text.** Type `[[quote]]`, `[[verse]]`,
  `[[epigraph]]` or `[[code]]` at the start of a paragraph that needs that
  treatment; the factory applies the style and removes the marker.
- **No title page, copyright page or contents.** Those are generated. If they
  are in the file, they are dropped — Inspect tells you so.
- **Images in the file, one per paragraph, caption in the paragraph after.**
  At least 1100 px wide for a full-width figure. Leave them in colour.

One thing is done once, in the browser: the **transmittal** on your factory
page — book information, trim size, typeface, copyright page. If the studio
filled it in with you, it is done. If not, open your factory page, fill it in
(it autosaves) and mark it **final**. Everything after that happens from your
folder.

---

## 3. Set up the folder — once, two minutes

1. Make a folder for the book. Put your manuscript in it.
2. Download the tool into the same folder: <https://jdbbs.exe.xyz/factory/api/factory.py>
   (right-click → *Save Link As…*, or in the terminal step below:
   `curl -sO https://jdbbs.exe.xyz/factory/api/factory.py`).
3. Open a terminal **in that folder**. On a Mac: open *Terminal*, type `cd `,
   drag the folder onto the window, press return.
4. Run:

   ```
   python3 factory.py setup
   ```

   It asks for the project id and the token, checks them with the factory, and
   saves them in `factory.json` beside the tool. You will not be asked again.
   It also shows your pass: finals left, storage until when.

If `python3` is not found: on a Mac, run `xcode-select --install` once (Apple's
free command-line tools include Python); on Windows, install Python from
python.org and use `python` instead of `python3`.

---

## 4. The loop: inspect → proof → final

All from the same terminal, in the same folder. `manuscript.docx` is whatever
your file is called.

### Inspect — free, as often as you like

```
python3 factory.py inspect manuscript.docx
```

Uploads the file and reads it the way the build will. You get a one-line count
(*worth fixing / worth a look / just noting*), a map of how the book is read —
which headings are front matter, where page 1 starts, what was dropped — and
the full report opens in your browser and is saved in `out/`. Fix what it
flags in Word, save, run it again. The first time it asks for the book's title
and author; it remembers them.

### Proof — free

```
python3 factory.py proof manuscript.docx
```

Uploads, builds, waits (a minute or two — you can watch the dots), and puts two
files in `out/`: `…-proof-….pdf` and `…-proof-….epub`. Read the **EPUB first** —
it is the quickest way to see how the factory understood your file — then the
PDF. Change the manuscript, run it again. Proofs do not count against anything.

### Final — uses one of your finals

```
python3 factory.py final manuscript.docx
```

The same build, clean, no PROOF line. It asks you to confirm, tells you how
many finals you have left, and saves `…-final-….pdf` and `…-final-….epub` in
`out/`. The PDF is the one you send a printer. A final that fails is not
counted.

### Status and download

```
python3 factory.py status        # pass, finals left, storage until, your uploads
python3 factory.py download      # fetch the newest build's files again
```

**Keep your own copies.** The factory stores your files for six months from the
day the pass started; after that the project goes read-only, then is deleted.

---

## 5. When something goes wrong

The tool says what happened in plain words. The ones you might see:

| It says | What to do |
|---|---|
| *The token was not accepted* | Copy and paste the token from the message you were sent; run `setup` again. |
| *A build is already running on this project* | Wait a minute or two, run it again. One build at a time per project. |
| *No finals left on this pass* | Proofs still work. For more finals: the factory page, or j@djinna.com. |
| *Proof limit reached* | 30 proofs in 24 hours on one project. Take a break, or export a final. |
| *The build failed:* followed by an explanation | Read it — it says what in the file caused it and, where it can, quotes the text nearby (*Near: “…”*) so you can search for it in Word. Fix, run again. Nothing was counted. |
| *Could not reach jdbbs.exe.xyz* | You're offline, or we are. Try again in a few minutes. |

---

## 6. If you'd rather click

Everything above is also on your factory page, in the same order, under the
same names. The two are interchangeable — a proof you build in the terminal
shows up on the page, and the other way round.

| From the folder | On the factory page |
|---|---|
| `setup` | *Password* box — paste the token |
| `status` | the line under the title: *n of m finals left · storage until …* |
| `inspect` | **2 · Upload** then **3 · Inspect** → *Open the full report* |
| `proof` | **5 · Build** → *Build proof — free* |
| `final` | **5 · Build** → *Export final — uses 1 of m* |
| `download` | **6 · Download** |

---

## 7. For presses with their own systems

`factory.py` is deliberately small and readable — one file, no dependencies —
because it is also the worked example. Everything it does is **six ordinary web
calls** with the token in an `Authorization: Bearer` header:

```
GET  /api/projects/{id}/pass                      your pass and credits
POST /api/books/upload                            the .docx (multipart: file, title, author, project_id)
POST /api/projects/{id}/preflight  {book_id}      inspect
POST /api/books/{book}/convert     {format, kind} build — kind "proof" or "final"
GET  /api/books/{book}                            poll until status is ready or error
GET  /api/books/{book}/download/pdf  | /epub      the files
```

So a lit mag's deadline script, a CMS hook, or a folder-watcher can drive the
factory directly: upload when a manuscript is approved, inspect, proof, and
collect the files — no one at a keyboard. The full reference, with every
request and response shown as run against the live server, is at
<https://jdbbs.exe.xyz/factory/api>. If you want a use supported that isn't
there yet, say so: j@djinna.com.

---

## 8. Ground rules

- **One build at a time** per project. Polling every few seconds is fine.
- **Proofs are free; finals are counted** when the build is accepted, and
  refunded if it fails.
- **Storage is six months** from the day the pass was set up. Downloads keep
  working 30 days beyond that; then the manuscript and builds are deleted and
  only the record (title, dates, how many builds) is kept. Your unpublished
  book is yours; the hard expiry is deliberate.
- **Rights are yours to clear.** The factory typesets what it is sent and does
  not check permissions.
- **Help:** live help during the workshop is included; outside that the
  Inspect report and this page are the self-serve path, and a pair of eyes on
  it is USD 100/hr, booked in advance. Terms: <https://jdbbs.exe.xyz/factory/terms>.
