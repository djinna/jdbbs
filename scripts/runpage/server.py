#!/usr/bin/env python3
"""Shared checklist page (:8766). Source of truth is scratch/run/CHECKLIST.md.

GET  /           page (markdown rendered by pandoc, plus the JS below)
GET  /fragment   rendered body only (the page polls this)
GET  /notes      notes.json  {item_id: [{ts, who, text}]}
POST /note       {id, text, who}  -> append a note
POST /tick       {id, state}  state in " ", "x", "~"  -> edits CHECKLIST.md in place
POST /add        {text}  -> appends "- [ ] 0.N text" under "## 0 · Inbox" (created if missing)
POST /upload     raw image body (Content-Type image/png|jpeg|gif|webp) -> {url: "/img/<name>"}; saved in scratch/run/img/
GET  /img/<name> serves an uploaded image. Notes may carry "images": [url, …] (pasted screenshots).
Items are lines like "- [ ] 1.2 …" (id = the leading token). Notes are appended, never overwritten;
Shelley reads notes.json and answers with who=shelley. If RUNPAGE_CHAT_CONV is set (a Shelley
conversation id), Jenna's notes and inbox adds are also pushed into that chat as they arrive."""
import json, os, re, subprocess, time, http.server, threading

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "scratch", "run"))
MD = os.path.join(ROOT, "CHECKLIST.md")
NOTES = os.path.join(ROOT, "notes.json")
IMG = os.path.join(ROOT, "img")
IMG_TYPES = {"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp"}
CHAT_CONV = os.environ.get("RUNPAGE_CHAT_CONV", "")  # Shelley conversation to push Jenna's notes into

def push_to_chat(msg):
    """Pull → push: forward a note into the live Shelley conversation so it is
    seen when written, not when the list is next opened. Best-effort."""
    if not CHAT_CONV: return
    try:
        subprocess.Popen(["shelley", "client", "chat", "-c", CHAT_CONV, "-p", msg],
                         stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, start_new_session=True)
    except Exception:
        pass
HERE = os.path.dirname(os.path.abspath(__file__))
LOCK = threading.Lock()
ITEM = re.compile(r"^(\s*- \[)([ x~])(\] )(\d+\.\d+[a-z]?)\b", re.M)

def read(p, default):
    try:
        with open(p) as f: return f.read() if default is None else json.load(f)
    except Exception: return default

def render():
    md = read(MD, None) or ""
    html = subprocess.run(["pandoc", "-f", "markdown-task_lists", "-t", "html"], input=md, capture_output=True, text=True).stdout
    # tag each item li with its id + state, and turn the [ ] into a clickable box
    def sub(m):
        st = m.group(1)
        sym = {"x": "☑", "~": "◐", " ": "☐"}[st]
        return f'<span class="box st-{st.strip() or "o"}" data-tick>{sym}</span> <b class="id">{m.group(2)}</b>'
    html = re.sub(r"\[([ x~])\] (\d+\.\d+[a-z]?)\b", sub, html)
    return html

class H(http.server.BaseHTTPRequestHandler):
    def _send(self, code, body, ctype="application/json"):
        data = body if isinstance(body, bytes) else json.dumps(body).encode()
        self.send_response(code); self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data))); self.send_header("Cache-Control", "no-store")
        self.end_headers(); self.wfile.write(data)
    def do_GET(self):
        if self.path == "/":
            page = read(os.path.join(HERE, "page.html"), None)
            self._send(200, page.replace("{{BODY}}", render()).encode(), "text/html; charset=utf-8")
        elif self.path.startswith("/fragment"):
            self.send_response(200); self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Cache-Control", "no-store")
            # Page script version: the page reloads itself when page.html changed under it.
            self.send_header("X-Page-Version", str(int(os.stat(os.path.join(HERE, "page.html")).st_mtime)))
            body = render().encode(); self.send_header("Content-Length", str(len(body))); self.end_headers(); self.wfile.write(body)
        elif self.path.startswith("/notes"):
            self._send(200, read(NOTES, {}))
        elif self.path.startswith("/img/"):
            name = os.path.basename(self.path[5:].split("?")[0])
            fp = os.path.join(IMG, name)
            ext = os.path.splitext(name)[1].lower()
            ctype = {v: k for k, v in IMG_TYPES.items()}.get(ext) or ({".pdf": "application/pdf"}.get(ext))  # PDFs the agent drops in for review
            if not ctype or not os.path.isfile(fp): return self._send(404, {"error": "not found"})
            with open(fp, "rb") as f: data = f.read()
            self.send_response(200); self.send_header("Content-Type", ctype)
            self.send_header("Content-Length", str(len(data))); self.send_header("Cache-Control", "public, max-age=86400")
            self.end_headers(); self.wfile.write(data)
        else:
            self._send(404, {"error": "not found"})
    def do_POST(self):
        n = int(self.headers.get("Content-Length", 0))
        if self.path == "/upload":
            # Pasted screenshot: raw bytes, typed by Content-Type. 12 MB cap.
            ext = IMG_TYPES.get((self.headers.get("Content-Type") or "").split(";")[0].strip())
            if not ext: return self._send(415, {"error": "png, jpeg, gif or webp only"})
            if n > 12 * 1024 * 1024: return self._send(413, {"error": "too big"})
            data = self.rfile.read(n)
            os.makedirs(IMG, exist_ok=True)
            name = time.strftime("%Y%m%d-%H%M%S", time.gmtime()) + f"-{int(time.time()*1000) % 1000:03d}" + ext
            with open(os.path.join(IMG, name), "wb") as f: f.write(data)
            return self._send(200, {"url": "/img/" + name})
        body = json.loads(self.rfile.read(n) or b"{}")
        with LOCK:
            if self.path == "/note":
                iid, text = body.get("id"), (body.get("text") or "").strip()
                images = [u for u in (body.get("images") or []) if isinstance(u, str) and u.startswith("/img/")]
                if not iid or not (text or images): return self._send(400, {"error": "id and text required"})
                notes = read(NOTES, {})
                note = {"ts": time.strftime("%Y-%m-%d %H:%M UTC", time.gmtime()), "who": body.get("who", "jenna"), "text": text}
                if images: note["images"] = images
                notes.setdefault(iid, []).append(note)
                with open(NOTES, "w") as f: json.dump(notes, f, indent=1, ensure_ascii=False)
                if body.get("who", "jenna") != "shelley":
                    # Images go into chat as VM paths so the agent can open them with read_image.
                    paths = "".join(f" [screenshot: {os.path.join(IMG, os.path.basename(u))}]" for u in images)
                    push_to_chat(f"Punch-list note from Jenna on {iid}: {text}{paths}")
                return self._send(200, {"ok": True})
            if self.path == "/tick":
                iid, st = body.get("id"), body.get("state", " ")
                if st not in (" ", "x", "~"): return self._send(400, {"error": "bad state"})
                md = read(MD, None) or ""
                new, k = ITEM.subn(lambda m: m.group(1) + st + m.group(3) + m.group(4) if m.group(4) == iid else m.group(0), md)
                if md == new and not re.search(rf"- \[[ x~]\] {re.escape(iid)}\b", md): return self._send(404, {"error": "item not found"})
                with open(MD, "w") as f: f.write(new)
                return self._send(200, {"ok": True})
            if self.path == "/add":
                text = (body.get("text") or "").strip()
                if not text: return self._send(400, {"error": "text required"})
                push_to_chat("Punch-list inbox item added by Jenna: " + re.sub(r"!\[[^\]]*\]\(/img/([^)]+)\)", lambda m: f"[screenshot: {os.path.join(IMG, m.group(1))}]", text))
                md = read(MD, None) or ""
                if "## 0 · Inbox" not in md:
                    md = md.rstrip("\n") + "\n\n## 0 · Inbox — new items, untriaged (Shelley moves them into a section)\n\n"
                nums = [int(x) for x in re.findall(r"- \[[ x~]\] 0\.(\d+)", md)]
                line = f"- [ ] 0.{max(nums, default=0) + 1} {text}  ·  _added {time.strftime('%a %H:%M UTC', time.gmtime())}_\n"
                head, sep, tail = md.partition("## 0 · Inbox")
                # append at the end of the inbox section, wherever it sits in the file
                m = re.search(r"\n## ", tail)
                if m:
                    sec, rest = tail[:m.start()], tail[m.start():]
                    md = head + sep + sec.rstrip("\n") + "\n" + line + rest
                else:
                    md = head + sep + tail.rstrip("\n") + "\n" + line
                with open(MD, "w") as f: f.write(md)
                return self._send(200, {"ok": True})
        self._send(404, {"error": "not found"})
    def log_message(self, *a): pass

if __name__ == "__main__":
    http.server.ThreadingHTTPServer(("127.0.0.1", 8766), H).serve_forever()
