#!/usr/bin/env python3
"""factory-cli.py — drive one jdbb book-factory project from a terminal.

Stdlib only; copy it into your own tooling. See docs/API-CLI-RECIPE-2026-09-19.md.

  factory-cli.py --base https://jdbbs.exe.xyz --project 14 --token XXX build  ms.docx [--out dir] [--format both|pdf|epub]
  factory-cli.py --base https://jdbbs.exe.xyz --project 14 --token XXX inspect ms.docx

build   = upload -> convert (a proof is free; --kind final uses one credit) -> poll until ready/error -> download outputs
inspect = upload -> preflight (free) -> print the report JSON
"""
import argparse, json, os, sys, time, urllib.request, urllib.error, uuid


class Factory:
    def __init__(self, base, project, token):
        self.base, self.project, self.token = base.rstrip("/"), project, token

    def call(self, method, path, body=None, headers=None):
        """One authenticated request. JSON in/out; raw bytes when the reply isn't JSON."""
        h = {"Authorization": "Bearer " + self.token, **(headers or {})}
        if isinstance(body, (dict, list)):
            body, h["Content-Type"] = json.dumps(body).encode(), "application/json"
        req = urllib.request.Request(self.base + path, data=body, method=method, headers=h)
        try:
            with urllib.request.urlopen(req, timeout=120) as r:
                data, ctype = r.read(), r.headers.get("Content-Type", "")
        except urllib.error.HTTPError as e:
            sys.exit(f"{method} {path} -> HTTP {e.code}: {e.read().decode(errors='replace')}")
        return json.loads(data) if ctype.startswith("application/json") or data[:1] == b"{" else data

    def upload(self, path, title, author):
        """multipart/form-data by hand: fields file, title, author, project_id."""
        b = uuid.uuid4().hex
        part = lambda name, val: f'--{b}\r\nContent-Disposition: form-data; name="{name}"\r\n\r\n{val}\r\n'.encode()
        body = part("title", title) + part("author", author) + part("project_id", self.project)
        body += (f'--{b}\r\nContent-Disposition: form-data; name="file"; filename="{os.path.basename(path)}"\r\n'
                 f'Content-Type: application/vnd.openxmlformats-officedocument.wordprocessingml.document\r\n\r\n').encode()
        body += open(path, "rb").read() + f"\r\n--{b}--\r\n".encode()
        return self.call("POST", "/api/books/upload", body, {"Content-Type": f"multipart/form-data; boundary={b}"})["id"]

    def inspect(self, book_id):
        return self.call("POST", f"/api/projects/{self.project}/preflight", {"book_id": book_id})

    def build(self, book_id, fmt="both", kind="proof", every=3):
        # kind "proof": free, PROOF line on the PDF. "final": one credit, clean PDF.
        self.call("POST", f"/api/books/{book_id}/convert", {"format": fmt, "kind": kind})  # 402 = no finals left, 409 = build running, 429 = too many proofs today
        while True:
            st = self.call("GET", f"/api/books/{book_id}")
            if st["status"] in ("ready", "error"):
                return st
            time.sleep(every)

    def download(self, st, out):
        os.makedirs(out, exist_ok=True)
        for o in st["outputs"]:  # each: {id, format, size_bytes, download_url}
            dest = os.path.join(out, f"book-{st['book_id']}.{o['format']}")
            open(dest, "wb").write(self.call("GET", o["download_url"]))
            print(f"  {o['format']:5} {o['size_bytes']:>9} bytes -> {dest}")


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--base", default="https://jdbbs.exe.xyz")
    ap.add_argument("--project", required=True, type=int, help="numeric project id")
    ap.add_argument("--token", default=os.environ.get("FACTORY_TOKEN"), help="project password (or $FACTORY_TOKEN)")
    ap.add_argument("cmd", choices=["build", "inspect"])
    ap.add_argument("docx")
    ap.add_argument("--title", help="default: file name")
    ap.add_argument("--author", default="Unknown", help="required by the API; default: Unknown")
    ap.add_argument("--format", default="both", choices=["both", "pdf", "epub"])
    ap.add_argument("--kind", default="proof", choices=["proof", "final"], help="proof (free, stamped PDF) or final (one credit, clean PDF)")
    ap.add_argument("--out", default=".", help="where downloads go (build)")
    a = ap.parse_args()
    if not a.token:
        sys.exit("need --token or $FACTORY_TOKEN")

    f = Factory(a.base, a.project, a.token)
    book = f.upload(a.docx, a.title or os.path.splitext(os.path.basename(a.docx))[0], a.author)
    print(f"uploaded book {book}", file=sys.stderr)
    if a.cmd == "inspect":
        print(json.dumps(f.inspect(book), indent=2))
        return
    st = f.build(book, a.format, a.kind)
    if st["status"] != "ready":
        sys.exit(f"build failed:\n{st.get('error')}")
    print(f"build ready ({len(st['outputs'])} outputs)")
    f.download(st, a.out)


if __name__ == "__main__":
    main()
