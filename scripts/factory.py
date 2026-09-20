#!/usr/bin/env python3
"""factory.py — run your book through the jdbb book factory from a folder on your own computer.

No installation. Needs only python3 (already on every Mac; on Windows, python.org).
Keep this file in a folder with your manuscript. Then, in a terminal, in that folder:

  python3 factory.py setup                      once: project id + token, checked and saved here
  python3 factory.py status                     your pass: finals left, storage until, what you've built
  python3 factory.py inspect  manuscript.docx   free — what the factory sees in your file; report opens in your browser
  python3 factory.py proof    manuscript.docx   free — PDF (with a PROOF line) + EPUB land in ./out
  python3 factory.py final    manuscript.docx   uses one final — the clean PDF you send a printer, + EPUB
  python3 factory.py download                   fetch the newest files again

Title and author are asked for the first time and remembered (override: --title, --author).
Everything this script does is six ordinary web calls; see https://jdbbs.exe.xyz/factory/api
"""
import argparse, datetime, json, os, re, sys, time, urllib.error, urllib.request, uuid, webbrowser

CONFIG = "factory.json"          # lives next to this script's working folder
BASE = "https://jdbbs.exe.xyz"
OUT = "out"


# ───────────────────────── plain-English errors ─────────────────────────

class FactoryError(Exception):
    pass


EXPLAIN = {
    401: "The token was not accepted. Check factory.json — copy and paste the token from the message you received. Run `python3 factory.py setup` to re-enter it.",
    402: "No finals left on this pass. Proofs still work (`proof`). To add finals, use the factory page or email j@djinna.com.",
    403: "This project has no live Factory Pass, so it cannot build. Email j@djinna.com.",
    409: "A build is already running on this project. Wait for it to finish (a minute or two) and try again.",
    429: "Proof limit reached: 30 proofs in 24 hours on this project. Try again later, or export a final.",
}


def die(msg):
    print("\n" + msg, file=sys.stderr)
    sys.exit(1)


# ───────────────────────── the six calls ─────────────────────────

class Factory:
    def __init__(self, base, project, token):
        self.base, self.project, self.token = base.rstrip("/"), int(project), token

    def call(self, method, path, body=None, headers=None, timeout=120):
        """One authenticated request. Returns parsed JSON, or raw bytes for files."""
        h = {"Authorization": "Bearer " + self.token, **(headers or {})}
        if isinstance(body, (dict, list)):
            body, h["Content-Type"] = json.dumps(body).encode(), "application/json"
        req = urllib.request.Request(self.base + path, data=body, method=method, headers=h)
        try:
            with urllib.request.urlopen(req, timeout=timeout) as r:
                data, ctype = r.read(), r.headers.get("Content-Type", "")
        except urllib.error.HTTPError as e:
            text = e.read().decode(errors="replace")
            try:
                text = json.loads(text).get("error", text)
            except ValueError:
                pass
            raise FactoryError(EXPLAIN.get(e.code, f"The factory answered {e.code}: {text}")) from None
        except urllib.error.URLError as e:
            raise FactoryError(f"Could not reach {self.base} ({e.reason}). Are you online?") from None
        if ctype.startswith("application/json"):
            return json.loads(data)
        return data

    def pass_status(self):
        return self.call("GET", f"/api/projects/{self.project}/pass")

    def books(self):
        return self.call("GET", f"/api/projects/{self.project}/books")

    def upload(self, path, title, author):
        """multipart/form-data, built by hand: fields file, title, author, project_id."""
        if not path.lower().endswith(".docx"):
            raise FactoryError(f"{path}: the factory takes Word files (.docx). Save As → Word Document, or export .docx from Pages / Google Docs / LibreOffice.")
        try:
            data = open(path, "rb").read()
        except OSError as e:
            raise FactoryError(f"Cannot read {path}: {e.strerror}") from None
        b = uuid.uuid4().hex
        part = lambda name, val: f'--{b}\r\nContent-Disposition: form-data; name="{name}"\r\n\r\n{val}\r\n'.encode()
        body = part("title", title) + part("author", author) + part("project_id", self.project)
        body += (f'--{b}\r\nContent-Disposition: form-data; name="file"; filename="{os.path.basename(path)}"\r\n'
                 f'Content-Type: application/vnd.openxmlformats-officedocument.wordprocessingml.document\r\n\r\n').encode()
        body += data + f"\r\n--{b}--\r\n".encode()
        return self.call("POST", "/api/books/upload", body, {"Content-Type": f"multipart/form-data; boundary={b}"})["id"]

    def inspect(self, book_id):
        return self.call("POST", f"/api/projects/{self.project}/preflight", {"book_id": book_id})

    def inspect_report(self, book_id):
        return self.call("GET", f"/api/projects/{self.project}/preflight/report?book_id={book_id}")

    def convert(self, book_id, kind):
        return self.call("POST", f"/api/books/{book_id}/convert", {"format": "both", "kind": kind})

    def status(self, book_id):
        return self.call("GET", f"/api/books/{book_id}")

    def wait(self, book_id, every=3):
        print("  building", end="", flush=True)
        while True:
            st = self.status(book_id)
            if st["status"] in ("ready", "error"):
                print()
                return st
            print(".", end="", flush=True)
            time.sleep(every)


# ───────────────────────── the folder: config + outputs ─────────────────────────

def load_config(args):
    cfg = {}
    if os.path.exists(CONFIG):
        try:
            cfg = json.load(open(CONFIG))
        except ValueError:
            die(f"{CONFIG} is not valid JSON — delete it and run `python3 factory.py setup` again.")
    if args.project:
        cfg["project"] = args.project
    if args.token:
        cfg["token"] = args.token
    if args.base:
        cfg["base"] = args.base
    cfg.setdefault("base", BASE)
    return cfg


def save_config(cfg):
    with open(CONFIG, "w") as f:
        json.dump(cfg, f, indent=2)
        f.write("\n")
    try:
        os.chmod(CONFIG, 0o600)  # the token is a password
    except OSError:
        pass


def need_factory(cfg):
    if not cfg.get("project") or not cfg.get("token"):
        die(f"No project id / token yet. Run `python3 factory.py setup` in this folder first.")
    return Factory(cfg["base"], cfg["project"], cfg["token"])


def ask(prompt, default=None):
    s = input(f"{prompt}{' [' + default + ']' if default else ''}: ").strip()
    return s or (default or "")


def title_author(cfg, args, path):
    title = args.title or cfg.get("title")
    author = args.author or cfg.get("author")
    if not title:
        title = ask("Book title", os.path.splitext(os.path.basename(path))[0])
    if not author:
        author = ask("Author")
    if not author:
        die("The factory needs an author name (it goes on the generated title page). Use --author.")
    if title != cfg.get("title") or author != cfg.get("author"):
        cfg["title"], cfg["author"] = title, author
        save_config(cfg)
    return title, author


def slug(s):
    return re.sub(r"[^A-Za-z0-9]+", "-", s).strip("-")[:60] or "book"


def stamp():
    return datetime.datetime.now().strftime("%Y-%m-%d-%H%M")


def fmt_date(iso):
    try:
        d = datetime.datetime.fromisoformat(iso.replace("Z", "+00:00"))
        return f"{d.day} {d:%b %Y}"
    except (ValueError, AttributeError):
        return iso or ""


def save_outputs(f, st, title, kind):
    os.makedirs(OUT, exist_ok=True)
    got = []
    for o in st.get("outputs", []):
        name = f"{slug(title)}-{kind}-{stamp()}.{o['format']}"
        dest = os.path.join(OUT, name)
        open(dest, "wb").write(f.call("GET", o["download_url"]))
        got.append(dest)
        print(f"  {o['format'].upper():4}  {o['size_bytes']:>10,} bytes  →  {dest}")
    return got


# ───────────────────────── commands ─────────────────────────

def cmd_setup(cfg, args):
    print("Enter the two things you were given for this book. Nothing is sent anywhere except to the factory.")
    project = ask("Project id (a number)", str(cfg.get("project", "")) or None)
    token = ask("Token", None) or cfg.get("token", "")
    if not project.isdigit() or not token:
        die("Need a numeric project id and a token.")
    f = Factory(cfg["base"], project, token)
    p = f.pass_status()          # 401 here = wrong token; the error text says so
    cfg.update(project=int(project), token=token)
    save_config(cfg)
    print(f"\nSaved to ./{CONFIG} (keep this file private — the token is a password).")
    show_pass(p)
    print("\nNext: put your manuscript (.docx) in this folder and run  python3 factory.py inspect manuscript.docx")


def show_pass(p):
    if not p.get("exists"):
        print("This project has no Factory Pass yet — nothing can be built until it has one.")
        return
    left, total = p["credits_remaining"], p.get("builds_included", 0) + p.get("builds_extra", 0)
    print(f"Factory Pass for {p.get('customer_name', '')}: {p['status']}"
          f" · {left} of {total} finals left · proofs free · storage until {fmt_date(p.get('expires_at'))}")


def cmd_status(cfg, args):
    f = need_factory(cfg)
    show_pass(f.pass_status())
    books = f.books()
    if not books:
        print("No manuscripts uploaded yet.")
        return
    print(f"\n{len(books)} upload(s) on this project, newest first:")
    for b in sorted(books, key=lambda b: b.get("CreatedAt", ""), reverse=True)[:10]:
        kind = (" " + b["BuildKind"]) if b.get("BuildKind") else ""
        print(f"  #{b['ID']:<4} {b.get('SourceFilename', '')[:48]:<48} {b['Status']}{kind}  {fmt_date(b.get('CreatedAt', ''))}")


def cmd_inspect(cfg, args):
    f = need_factory(cfg)
    title, author = title_author(cfg, args, args.file)
    print(f"Uploading {args.file} …")
    book = f.upload(args.file, title, author)
    print(f"Inspecting (free) …")
    rep = f.inspect(book)
    s = rep.get("summary", {})
    print(f"\n{s.get('total', 0)} things to look at — {s.get('high', 0)} worth fixing, "
          f"{s.get('medium', 0)} worth a look, {s.get('low', 0)} just noting"
          + (f", {s['preserved']} kept as typed." if s.get("preserved") else "."))
    bm = rep.get("book_map") or {}
    secs = bm.get("sections") or []
    if secs:
        print("\nHow the build reads your file:")
        for sec in secs[:40]:
            print(f"  {sec.get('kind', ''):<10} {sec.get('title', '')[:70]}")
        if len(secs) > 40:
            print(f"  … and {len(secs) - 40} more")
    for w in (bm.get("warnings") or [])[:10]:
        print(f"  ! {w}")
    os.makedirs(OUT, exist_ok=True)
    dest = os.path.join(OUT, f"{slug(title)}-inspect-{stamp()}.html")
    open(dest, "wb").write(f.inspect_report(book))
    print(f"\nFull report: {dest}")
    if not args.no_open:
        webbrowser.open("file://" + os.path.abspath(dest))
    print(f"\nWhen it looks right:  python3 factory.py proof {args.file}")


def build(cfg, args, kind):
    f = need_factory(cfg)
    title, author = title_author(cfg, args, args.file)
    if kind == "final":
        p = f.pass_status()
        left = p.get("credits_remaining", 0)
        if p.get("finals_gate") != "off" and left <= 0:
            die(EXPLAIN[402])
        if not args.yes:
            ok = ask(f"Export a FINAL of “{title}”? Uses 1 of your finals ({left} left now). y/N", "n")
            if ok.lower() not in ("y", "yes"):
                print("Not built.")
                return
    print(f"Uploading {args.file} …")
    book = f.upload(args.file, title, author)
    print(f"Starting {kind} build (PDF + EPUB){' — free' if kind == 'proof' else ''} …")
    f.convert(book, kind)
    st = f.wait(book)
    if st["status"] != "ready":
        die("The build failed" + (" — the final was not counted" if kind == "final" else "") + ":\n\n" + (st.get("error") or "(no detail)"))
    print(f"Done. Saving files to ./{OUT}/")
    save_outputs(f, st, title, kind)
    if kind == "proof":
        print("\nRead the EPUB first — it shows quickest how the factory understood your file — then the PDF."
              f"\nThe proof PDF has a PROOF line on every page. When it's right:  python3 factory.py final {args.file}")
    else:
        print("\nThis is the clean print PDF — the one to send a printer. Keep your own copy.")


def cmd_proof(cfg, args):
    build(cfg, args, "proof")


def cmd_final(cfg, args):
    build(cfg, args, "final")


def cmd_download(cfg, args):
    f = need_factory(cfg)
    books = [b for b in f.books() if b.get("Status") == "ready"]
    if not books:
        die("Nothing built yet on this project.")
    b = sorted(books, key=lambda b: b.get("UpdatedAt", ""), reverse=True)[0]
    st = f.status(b["ID"])
    kind = b.get("BuildKind") or "build"
    print(f"Newest build: #{b['ID']} “{b.get('Title', '')}” ({kind}, {fmt_date(b.get('UpdatedAt', ''))})")
    save_outputs(f, st, b.get("Title") or cfg.get("title") or "book", kind)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--project", type=int, help="project id (else from factory.json)")
    ap.add_argument("--token", default=os.environ.get("FACTORY_TOKEN"), help="token (else from factory.json or $FACTORY_TOKEN)")
    ap.add_argument("--base", help=argparse.SUPPRESS)
    ap.add_argument("--title", help="book title (remembered)")
    ap.add_argument("--author", help="author name (remembered)")
    ap.add_argument("-y", "--yes", action="store_true", help="final: don't ask for confirmation")
    ap.add_argument("--no-open", action="store_true", help="inspect: don't open the report in a browser")
    ap.add_argument("cmd", choices=["setup", "status", "inspect", "proof", "final", "download"])
    ap.add_argument("file", nargs="?", help="your manuscript (.docx) — for inspect / proof / final")
    args = ap.parse_args()
    if args.cmd in ("inspect", "proof", "final") and not args.file:
        ap.error(f"{args.cmd} needs your manuscript, e.g.  python3 factory.py {args.cmd} manuscript.docx")
    cfg = load_config(args)
    try:
        {"setup": cmd_setup, "status": cmd_status, "inspect": cmd_inspect,
         "proof": cmd_proof, "final": cmd_final, "download": cmd_download}[args.cmd](cfg, args)
    except FactoryError as e:
        die(str(e))
    except KeyboardInterrupt:
        print("\nStopped. (A build already started will finish on the factory; `python3 factory.py download` fetches it.)")
        sys.exit(130)


if __name__ == "__main__":
    main()
