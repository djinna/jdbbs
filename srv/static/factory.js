// factory.js — the customer-facing Factory page for a Factory Pass.
//
// Served at /{client}/{project}/factory/ (factory.html; assets under that path
// resolve to srv/static/). Contract: docs/specs/FACTORY-PASS-API-2026-09-03.md.
//
// Shape of this file: the HTML shell is static and readable with JS off/broken;
// this script only fills slots and wires the five steps. Every network failure
// is written into a visible element (#fx-banner or the step's .fx-status) —
// there is no path that leaves a blank page.
//
// Audience: non-technical authors, live in a workshop. Copy says "build",
// never "convert"; "inspect", never "preflight" (except where we name the
// report). Nothing destructive is one click away.

(function () {
'use strict';

// ─── tiny helpers ──────────────────────────────────────────────────────────
var $ = function (id) { return document.getElementById(id); };
var esc = function (s) {
  return String(s == null ? '' : s).replace(/[&<>"']/g, function (c) {
    return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
  });
};
var show = function (el, on) { if (el) el.hidden = !on; };
var setText = function (el, t) { if (el) el.textContent = t; };

function ApiError(message, status, data) {
  this.name = 'ApiError';
  this.message = message;
  this.status = status || 0;
  this.data = data || {};
}
ApiError.prototype = Object.create(Error.prototype);

// api(): JSON fetch. Throws ApiError with .status so callers can branch on
// 401 (password) / 402 (no builds) / 409 (build running).
async function api(url, opts) {
  opts = opts || {};
  var r;
  try {
    r = await fetch(url, {
      method: opts.method || 'GET',
      headers: opts.body && !opts.raw ? { 'Content-Type': 'application/json' } : undefined,
      body: opts.body,
    });
  } catch (e) {
    throw new ApiError("Couldn't reach the server. Check your connection and try again.", 0);
  }
  var data = null;
  var text = await r.text().catch(function () { return ''; });
  if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
  if (!r.ok) {
    var msg = (data && (data.error || data.message)) || (r.status + ' ' + r.statusText);
    throw new ApiError(msg, r.status, data || {});
  }
  return data;
}

// "3 Mar 2027"
var MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
function fmtDate(v) {
  if (!v) return '';
  var d = new Date(v);
  if (isNaN(d.getTime())) return String(v);
  return d.getDate() + ' ' + MONTHS[d.getMonth()] + ' ' + d.getFullYear();
}
function fmtWhen(v) {
  if (!v) return '';
  var d = new Date(v);
  if (isNaN(d.getTime())) return String(v);
  var hh = String(d.getHours()).padStart(2, '0');
  var mm = String(d.getMinutes()).padStart(2, '0');
  return fmtDate(v) + ' ' + hh + ':' + mm;
}
function fmtSize(bytes) {
  var n = Number(bytes || 0);
  if (!n) return '';
  if (n < 1024 * 1024) return Math.max(1, Math.round(n / 1024)) + ' KB';
  return (n / (1024 * 1024)).toFixed(1) + ' MB';
}
function plural(n, one, many) { return n === 1 ? one : many; }

// ─── state ─────────────────────────────────────────────────────────────────
var S = {
  clientSlug: null,
  projectSlug: null,
  projectId: null,
  project: null,
  pass: null,          // null = unknown, {exists:false} or full pass
  books: [],           // newest first
  current: null,       // newest book = "current manuscript"
  preflight: null,     // preflightResponse for the current book
  outputs: [],         // book_outputs rows for the current book
  transmittalStatus: null,
  polling: null,
  pollTicks: 0,
  pendingFile: null,
  uploading: false,
  building: false,
  buildingKind: null,  // 'proof' | 'final' while a build we started is in flight (0.28)
  inspecting: false,
  contactEmail: 'j@djinna.com',
  retry: null,         // re-run after the password gate clears
  passUnavailable: false,
  store: null,         // /api/public/store/config when the store is on
  index: null,         // GET /api/books/{id}/index: {status, entitled, index, error} (5.13)
  indexRows: null,     // working copy of index.entries while reviewing
  indexDirty: false,
  indexSaving: false,
  indexPolling: null,
  includeIndex: true,  // "Include the index in the next build"
};

// Books come back as Go `ListBooksRow` values with no json tags, so the keys
// are PascalCase (ID/Title/Status/…) and ProjectID is a sql.NullInt64
// ({Int64,Valid}). Normalize both that and any future snake_case shape so the
// page doesn't care which one the server sends.
function normBook(b) {
  if (!b) return null;
  var pid = b.ProjectID != null ? b.ProjectID : b.project_id;
  if (pid && typeof pid === 'object') pid = pid.Valid ? pid.Int64 : null;
  return {
    id: b.ID != null ? b.ID : b.id,
    title: b.Title != null ? b.Title : b.title,
    author: b.Author != null ? b.Author : b.author,
    series: b.Series != null ? b.Series : b.series,
    filename: b.SourceFilename != null ? b.SourceFilename : b.source_filename,
    status: b.Status != null ? b.Status : b.status,
    errorMsg: b.ErrorMsg != null ? b.ErrorMsg : b.error_msg,
    buildKind: b.BuildKind != null ? b.BuildKind : (b.build_kind || 'final'),  // kind of the in-flight / last build (0.28)
    projectId: pid,
    createdAt: b.CreatedAt != null ? b.CreatedAt : b.created_at,
    updatedAt: b.UpdatedAt != null ? b.UpdatedAt : b.updated_at,
  };
}

// Book status vocabulary: the pipeline writes 'uploaded' → 'converting' →
// 'ready' | 'error'. The contract calls the success state 'done', so accept
// both spellings.
function isBuilding(b) { return !!b && (b.status === 'converting' || b.status === 'building'); }
function isBuilt(b) { return !!b && (b.status === 'ready' || b.status === 'done'); }
function isFailed(b) { return !!b && b.status === 'error'; }

// ─── banner (page-level failures) ──────────────────────────────────────────
function banner(html) {
  var el = $('fx-banner');
  if (!el) return;
  if (!html) { el.hidden = true; el.innerHTML = ''; return; }
  el.innerHTML = html;
  el.hidden = false;
}

// Something went wrong before we could read anything: say so in the banner AND
// stop the pass badge from spinning forever.
function bail(html) {
  S.passUnavailable = true;
  banner(html);
  renderAll();
}

// ─── entitlement ───────────────────────────────────────────────────────────
// One place decides whether the action buttons work, and why not.
function passState() {
  if (!S.pass) return { ok: false, reason: '' };                 // still loading
  if (S.pass.exists === false) {
    return {
      ok: false,
      reason: 'This project isn\u2019t running on a Factory Pass, so uploading and building are turned off here. You can still read the transmittal and download anything already built.',
    };
  }
  if (S.pass.live === false || S.pass.status === 'expired') {
    return {
      ok: false,
      reason: 'Your pass ended on ' + fmtDate(S.pass.expires_at) + '. Downloads still work; uploading and building are turned off. Email ' +
        S.contactEmail + ' if you need more time.',
    };
  }
  if (S.pass.status === 'revoked' || S.pass.status === 'purged') {
    return { ok: false, reason: 'This pass is no longer active. Email ' + S.contactEmail + ' and we\u2019ll sort it out.' };
  }
  return { ok: true, reason: '' };
}
function creditsLeft() {
  if (!S.pass || S.pass.exists === false) return 0;
  if (typeof S.pass.credits_remaining === 'number') return S.pass.credits_remaining;
  var inc = Number(S.pass.builds_included || 0) + Number(S.pass.builds_extra || 0);
  return Math.max(0, inc - Number(S.pass.builds_used || 0));
}
function creditsTotal() {
  if (!S.pass || S.pass.exists === false) return 0;
  return Number(S.pass.builds_included || 0) + Number(S.pass.builds_extra || 0);
}

// ─── render: header + pass badge ───────────────────────────────────────────
function renderHeader() {
  var name = (S.project && S.project.Name) || S.projectSlug || 'Your book';
  setText($('fx-title'), name);
  document.title = name + ' — Factory';

  var author = S.current && S.current.author;
  setText($('fx-byline'), author ? 'by ' + author : '');

  var passEl = $('fx-pass');
  var noteEl = $('fx-pass-note');
  if (!S.pass) {
    passEl.innerHTML = S.passUnavailable
      ? '<span class="fx-pass-dim">Pass details unavailable</span>'
      : 'Checking your pass\u2026';
    return;
  }

  if (S.pass.exists === false) {
    passEl.innerHTML = '<span class="fx-pass-dim">No Factory Pass on this project \u00b7 read-only</span>';
  } else {
    var left = creditsLeft();
    var total = creditsTotal();
    var cls = left === 0 ? 'fx-pass-out' : '';
    var builds = '<span class="' + cls + '">' + left + ' of ' + total + ' ' + plural(total, 'final', 'finals') + ' left</span>';
    // Past tense once the pass is over — "storage until <a date last month>"
    // reads like a bug.
    var ended = S.pass.live === false || S.pass.status === 'expired';
    var until = S.pass.expires_at
      ? '<span class="fx-pass-dim"> \u00b7 storage ' + (ended ? 'ended ' : 'until ') +
        esc(fmtDate(S.pass.expires_at)) + '</span>'
      : '';
    passEl.innerHTML = builds + until;
  }

  var st = passState();
  if (st.ok) { show(noteEl, false); setText(noteEl, ''); }
  else {
    noteEl.className = 'fx-note' + (S.pass.exists === false ? '' : ' warn');
    setText(noteEl, st.reason);
    show(noteEl, true);
  }

  var expiry = $('fx-foot-expiry');
  if (expiry) setText(expiry, S.pass && S.pass.expires_at ? ' (' + fmtDate(S.pass.expires_at) + ')' : '');

  var contact = $('fx-foot-contact');
  if (contact) {
    var canBuy = S.store && S.store.enabled && S.pass && S.pass.exists !== false &&
      S.pass.status !== 'revoked' && S.pass.status !== 'purged';
    if (canBuy) {
      var b3 = S.store.items['builds-3'], s6 = S.store.items['storage-6mo'];
      contact.innerHTML = 'Need more? ' +
        '<button type="button" class="fx-link" data-addon="builds-3">+3 finals' + (b3 ? ' (' + esc(b3.display) + ')' : '') + '</button> \u00b7 ' +
        '<button type="button" class="fx-link" data-addon="storage-6mo">+6 months storage' + (s6 ? ' (' + esc(s6.display) + ')' : '') + '</button>' +
        ' \u2014 card checkout through Stripe. A pair of eyes on it? Email <a href="mailto:' +
        esc(S.contactEmail) + '">' + esc(S.contactEmail) + '</a>.';
    } else {
      contact.innerHTML = 'Need more finals, more time, or a pair of eyes on it? Email <a href="mailto:' +
        esc(S.contactEmail) + '">' + esc(S.contactEmail) + '</a>.';
    }
  }
}

// ─── store: add-ons against this pass ───────────────────────────────────────
// One click → Stripe Checkout for that add-on → back here with ?order=cs_…,
// which we confirm with the server so the badge updates without waiting for
// the email. All of it is dormant unless /api/public/store/config says on.
function buyAddon(key) {
  var btns = document.querySelectorAll('[data-addon]');
  btns.forEach(function (b) { b.disabled = true; });
  fetch('/api/public/store/checkout', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, credentials: 'same-origin',
    body: JSON.stringify({ kind: 'addon', project_id: S.projectId, items: [{ key: key, qty: 1 }] })
  }).then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
    .then(function (r) {
      if (r.ok && r.d.url) { location.href = r.d.url; return; }
      throw new Error((r.d && r.d.error) || 'Checkout is unavailable right now.');
    })
    .catch(function (e) {
      btns.forEach(function (b) { b.disabled = false; });
      banner('<b>Couldn\u2019t start checkout.</b> ' + esc(e.message));
    });
}

function confirmOrderFromURL() {
  var order = new URLSearchParams(location.search).get('order');
  if (!order) return;
  // Drop the parameter so a reload doesn't re-confirm.
  var u = new URL(location.href); u.searchParams.delete('order'); history.replaceState(null, '', u);
  banner('Confirming your add-on with Stripe\u2026');
  fetch('/api/public/store/session?session_id=' + encodeURIComponent(order))
    .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
    .then(function (r) {
      if (r.ok && r.d.status === 'paid') {
        banner('<b>Added to your pass.</b> ' + esc(r.d.credits_remaining) + ' finals remaining; storage until ' + esc(fmtDate(r.d.expires_at || '')) + '. A confirmation email is on its way.');
        return loadPass().then(renderHeader);
      }
      if (r.ok && r.d.status === 'pending') { banner('Stripe hasn\u2019t confirmed the payment yet. If you were charged, it will land within a few minutes \u2014 reload then.'); return; }
      throw new Error((r.d && r.d.error) || 'Could not confirm the order.');
    })
    .catch(function (e) { banner('<b>Order not confirmed yet.</b> ' + esc(e.message)); });
}

// ─── render: step strip ────────────────────────────────────────────────────
// Which step are you on? Cheap heuristic, deliberately generous: the point is
// to answer "what do I do next", not to police anything.
// Step 4 (Index) is optional, so it is only "current" while a draft is being
// written or waits for review; otherwise the strip skips over it.
function currentStep() {
  if (S.transmittalStatus && S.transmittalStatus !== 'final') return 1;
  if (!S.books.length) return 2;
  if (isBuilding(S.current)) return 5;
  if (indexStatus() === 'drafting') return 4;
  if (hasDownloads()) return 6;
  if (!S.preflight || S.preflight.exists === false) return 3;
  if (indexStatus() === 'draft') return 4;
  return 5;
}
function hasDownloads() {
  return S.outputs.length > 0 || isBuilt(S.current);
}
function renderSteps() {
  var cur = currentStep();
  var done = {
    1: S.transmittalStatus === 'final',
    2: S.books.length > 0,
    3: !!(S.preflight && S.preflight.exists !== false),
    4: indexStatus() === 'reviewed',
    5: hasDownloads(),
    6: false,
  };
  for (var i = 1; i <= 6; i++) {
    var el = $('fx-step-' + i);
    if (!el) continue;
    var cls = 'fx-step';
    if (i === cur) cls += ' current';
    else if (done[i]) cls += ' done';
    el.className = cls;
    el.setAttribute('aria-current', i === cur ? 'step' : 'false');
  }
}

// ─── render: upload + book list ────────────────────────────────────────────
function renderUpload() {
  var st = passState();
  var drop = $('fx-drop');
  if (drop) {
    drop.classList.toggle('disabled', !st.ok);
    drop.setAttribute('aria-disabled', st.ok ? 'false' : 'true');
    var main = $('fx-drop-main');
    // When uploading is off, the target must not look like a live control.
    if (main && !st.ok && !S.pendingFile) {
      main.textContent = (S.pass || S.passUnavailable)
        ? 'Uploading is turned off for this project.'
        : 'Checking\u2026';
    } else if (main && st.ok && !S.pendingFile && main.dataset.picked !== 'true') {
      main.innerHTML = 'Drop your .docx here, or <span class="fx-drop-link">choose a file</span>';
    }
  }

  var meta = $('fx-upload-meta');
  if (meta) setText(meta, S.books.length ? S.books.length + ' ' + plural(S.books.length, 'file', 'files') + ' uploaded' : '');

  var btn = $('fx-upload-btn');
  if (btn) {
    var titleOK = ($('fx-book-title').value || '').trim() !== '';
    var authorOK = ($('fx-book-author').value || '').trim() !== '';
    btn.disabled = !st.ok || !S.pendingFile || !titleOK || !authorOK || S.uploading;
    btn.textContent = S.uploading ? 'Uploading\u2026' : 'Upload manuscript';
  }

  var list = $('fx-books');
  if (!list) return;
  if (!S.books.length) {
    list.innerHTML = '';
    return;
  }
  var html = '<div class="fx-list-head">Your manuscript files</div>';
  S.books.forEach(function (b, i) {
    var tagCls = isBuilt(b) ? 'ok' : isFailed(b) ? 'err' : isBuilding(b) ? 'warn' : '';
    var tagTxt = isBuilt(b) ? '[BUILT]' : isFailed(b) ? '[BUILD FAILED]' : isBuilding(b) ? '[BUILDING\u2026]' : '[NOT BUILT]';
    html += '<div class="fx-row' + (i === 0 ? ' current' : '') + '">' +
      '<div><div class="fx-row-main">' + esc(b.filename || b.title || 'manuscript.docx') +
      (i === 0 ? ' <span class="tag">\u00b7 current</span>' : '') + '</div>' +
      '<div class="fx-row-sub">' + esc(b.title || '') + (b.author ? ' \u00b7 ' + esc(b.author) : '') +
      (b.createdAt ? ' \u00b7 uploaded ' + esc(fmtWhen(b.createdAt)) : '') + '</div>' +
      (isFailed(b) && b.errorMsg ? '<div class="fx-row-sub err">' + esc(shortErr(b.errorMsg)) + '</div>' : '') +
      (isBuilt(b) && b.errorMsg ? '<div class="fx-row-sub warn">Warning: ' + esc(shortErr(b.errorMsg)) + '</div>' : '') +
      '</div>' +
      '<div class="fx-row-right"><span class="tag ' + tagCls + '">' + tagTxt + '</span></div>' +
      '</div>';
  });
  if (S.books.length > 1) {
    html += '<p class="fx-fine">The newest file is the one we inspect and build. Older uploads stay listed so you can see what changed when.</p>';
  }
  list.innerHTML = html;
}

// Pipeline errors are pandoc/typst dumps. Show the first useful line and let
// the rest live in the log; a wall of stderr helps nobody in a workshop.
function shortErr(msg) {
  var first = String(msg || '').split('\n').filter(function (l) { return l.trim(); })[0] || '';
  return first.length > 300 ? first.slice(0, 297) + '\u2026' : first;
}

// Build failures are stored as up to three lines: explanation, "Near: “…”"
// (text quoted from the document so the author can find it in their editor), and
// "Technical detail: …". Render the first as the message and the rest folded.
function errDetailHTML(msg) {
  var lines = String(msg || '').split('\n').filter(function (l) { return l.trim(); }).slice(1);
  if (!lines.length) return '';
  var near = lines.filter(function (l) { return l.indexOf('Near:') === 0; })
    .map(function (l) { return '<div class="fx-err-near">' + esc(l.replace(/^Near:\s*/, 'Near: ')) + ' <span class="muted">(search for this in your manuscript)</span></div>'; }).join('');
  var tech = lines.filter(function (l) { return l.indexOf('Technical detail:') === 0; })
    .map(function (l) { return '<details class="fx-err-detail"><summary>Technical detail</summary><code>' + esc(l.replace(/^Technical detail:\s*/, '')) + '</code></details>'; }).join('');
  return near + tech;
}

// ─── render: inspect ───────────────────────────────────────────────────────
var TYPE_LABELS = {
  undeclared_custom_style: 'Custom styles not in your transmittal',
  declared_custom_style_used: 'Declared styles found in the file',
  observed_style: 'Styles seen in the file',
  heading_lookalike: 'Headings without a Heading style',
  manual_break: 'Scene breaks done by hand',
  stray_quote_marker: 'Stray ">" marks from a pasted email',
  manual_formatting: 'Bold/italic applied by hand',
  manual_list: 'Lists typed by hand',
  direct_spacing: 'Spacing/indent/alignment applied by hand',
  colored_text: 'Colored text',
  unusual_font: 'Unusual fonts (ignored by the factory)',
  style_marker: 'Marked styles ([[style]] markers)',
  mixed_formatting: 'Mixed fonts or sizes in one paragraph',
  image_inventory: 'Images',
  low_resolution_image: 'Images too low-resolution for print',
  manual_page_break: 'Manual page breaks',
  empty_paragraph: 'Empty paragraphs',
  tracked_change: 'Tracked changes still in the file',
  comment: 'Comments still in the file',
  footnote: 'Footnotes',
  table: 'Tables',
  hyperlink: 'Links',
  book_map_warning: 'Book map: front / body / back matter',
  book_map_note: 'Book map: what the build drops or renames',
};

// Book map (P4): how the build reads the file — front / body / back matter
// from Heading 1 text + position. Rendered under the counts on the factory page.
function renderBookMap(bm) {
  if (!bm || !bm.sections) return '';
  var kinds = { front: 'front matter', body: 'body', back: 'back matter', title: 'dropped (title page is generated)', toc: 'dropped (contents are generated)' };
  var html = '<div class="fx-bookmap"><div class="fx-bookmap-head">How the build reads your file</div>';
  html += '<p class="fx-bookmap-summary">' + esc(bm.summary || '') + '</p>';
  var rows = [];
  if (bm.title) rows.push(['Title style: ' + bm.title + (bm.subtitle ? ' \u2014 ' + bm.subtitle : ''), kinds.title]);
  (bm.untitled_front || []).forEach(function (u) {
    rows.push(['\u201c' + u.preview + '\u201d', u.drop ? 'dropped (generated from the transmittal)' : u.name + ' (untitled front matter)']);
  });
  var body = bm.sections.filter(function (sec) { return sec.kind === 'body'; });
  var bodyDone = false;
  bm.sections.forEach(function (sec) {
    if (sec.kind !== 'body') { rows.push([sec.title, kinds[sec.kind] || sec.kind]); return; }
    if (bodyDone) return;
    bodyDone = true;
    var label = body.length === 1 ? '\u201c' + body[0].title + '\u201d' :
      '\u201c' + body[0].title + '\u201d \u2026 \u201c' + body[body.length - 1].title + '\u201d';
    rows.push([label, 'body, ' + body.length + ' ' + plural(body.length, 'section', 'sections') + ', page 1 starts here']);
  });
  if (rows.length) {
    html += '<div class="fx-types">';
    rows.forEach(function (r) {
      html += '<div class="fx-types-row"><span>' + esc(r[0]) + '</span><span>' + esc(r[1]) + '</span></div>';
    });
    html += '</div>';
  }
  if (bm.warnings && bm.warnings.length) {
    html += '<ul class="fx-bookmap-warnings">';
    bm.warnings.forEach(function (w) { html += '<li>' + esc(w) + '</li>'; });
    html += '</ul>';
  }
  if (bm.notes && bm.notes.length) {
    html += '<ul class="fx-bookmap-notes">';
    bm.notes.forEach(function (n) { html += '<li>' + esc(n) + '</li>'; });
    html += '</ul>';
  }
  html += '<p class="fx-fine">Every section head is Heading 1; the build places it by its text. Front matter gets roman folios, the first chapter gets page 1. To move a section, rename or reorder its heading.</p></div>';
  return html;
}
function typeLabel(t) {
  return TYPE_LABELS[t] || String(t || '').replace(/_/g, ' ').replace(/^./, function (c) { return c.toUpperCase(); });
}

function renderInspect() {
  var st = passState();
  var btn = $('fx-inspect-btn');
  if (btn) {
    btn.disabled = !st.ok || !S.current || S.inspecting;
    btn.textContent = S.inspecting ? 'Inspecting\u2026' : (S.preflight && S.preflight.exists !== false ? 'Inspect again' : 'Inspect manuscript');
  }

  var meta = $('fx-preflight-meta');
  var link = $('fx-report-link');
  var box = $('fx-preflight-summary');
  var pf = S.preflight;

  if (!S.current) {
    if (meta) { meta.className = 'fx-sec-meta'; setText(meta, 'Upload a file first'); }
    show(link, false);
    if (box) box.innerHTML = '';
    return;
  }
  if (!pf || pf.exists === false) {
    if (meta) { meta.className = 'fx-sec-meta'; setText(meta, 'Not inspected yet'); }
    show(link, false);
    if (box) box.innerHTML = '';
    return;
  }

  if (meta) {
    meta.className = 'fx-sec-meta';
    setText(meta, pf.updated_at ? 'Last inspected ' + fmtWhen(pf.updated_at) : '');
  }

  if (link) {
    // report_url comes back from the server (it carries book_id/preflight_id);
    // fall back to building it ourselves if it's absent.
    var url = pf.report_url || ('/api/projects/' + S.projectId + '/preflight/report?book_id=' + (pf.book_id || S.current.id));
    link.setAttribute('href', url);
    show(link, true);
  }

  if (!box) return;

  if (pf.status === 'error') {
    box.innerHTML = '<p class="fx-status err">The inspection couldn\u2019t read that file: ' +
      esc(shortErr(pf.error || 'unknown error')) +
      '</p><p class="fx-fine">Re-export it as .docx from your editor and upload again. Nothing was charged.</p>';
    return;
  }

  var sum = pf.summary || {};
  var total = Number(sum.total || 0);
  var html = '<div class="fx-counts">';
  if (total === 0) {
    html += '<div class="fx-count clean"><span class="fx-count-label">Things to look at</span>' +
      '<span class="fx-count-value">0</span></div>';
  } else {
    html += '<div class="fx-count"><span class="fx-count-label">Things to look at</span>' +
      '<span class="fx-count-value">' + total + '</span></div>' +
      '<div class="fx-count high"><span class="fx-count-label">Worth fixing</span>' +
      '<span class="fx-count-value">' + Number(sum.high || 0) + '</span></div>' +
      '<div class="fx-count medium"><span class="fx-count-label">Worth a look</span>' +
      '<span class="fx-count-value">' + Number(sum.medium || 0) + '</span></div>' +
      '<div class="fx-count low"><span class="fx-count-label">Just noting</span>' +
      '<span class="fx-count-value">' + Number(sum.low || 0) + '</span></div>';
    if (Number(sum.preserved || 0) > 0) {
      html += '<div class="fx-count kept"><span class="fx-count-label">Carried through automatically</span>' +
        '<span class="fx-count-value">' + Number(sum.preserved) + '</span></div>';
    }
  }
  html += '</div>';

  var byType = sum.by_type || {};
  var keys = Object.keys(byType).sort(function (a, b) { return byType[b] - byType[a]; }).slice(0, 6);
  if (keys.length) {
    html += '<div class="fx-types">';
    keys.forEach(function (k) {
      html += '<div class="fx-types-row"><span>' + esc(typeLabel(k)) + '</span><span>' + Number(byType[k]) + '</span></div>';
    });
    html += '</div>';
  }

  if (total === 0) {
    html += '<p class="fx-fine">Nothing flagged. Go ahead and build.</p>';
  } else if (Number(sum.high || 0) > 0) {
    html += '<p class="fx-fine">You can build anyway \u2014 nothing here blocks it. But the report explains each item, and fixing the \u201cworth fixing\u201d ones in your manuscript before you export a final usually saves you one. Proofs are free, so build one and read it.</p>';
  } else {
    html += '<p class="fx-fine">Nothing serious. The report has the detail if you\u2019re curious.</p>';
  }

  if (pf.images && pf.images.length) {
    var nColour = pf.images.filter(function (im) { return im.colour === true; }).length;
    html += '<p class="fx-fine">' + pf.images.length + ' ' + plural(pf.images.length, 'image', 'images') +
      ' found \u2014 the report lists each one\u2019s size in print.' +
      (nColour ? ' ' + nColour + ' ' + (nColour === 1 ? 'is' : 'are') + ' colour: kept in the EPUB, converted to grey for the print PDF. Already-grey images are left alone; place your own grey version in the manuscript to override.' : '') +
      '</p>';
  }

  html += renderBookMap(pf.book_map);

  box.innerHTML = html;
}

// ─── render: build ─────────────────────────────────────────────────────────
function renderBuild() {
  var st = passState();
  var left = creditsLeft();
  var total = creditsTotal();
  var proofBtn = $('fx-proof-btn');
  var btn = $('fx-build-btn');
  var meta = $('fx-build-meta');
  var busy = S.building || isBuilding(S.current);
  var noPass = !S.pass || S.pass.exists === false;

  if (meta) {
    meta.className = 'fx-sec-meta' + (left === 0 && S.pass && S.pass.exists !== false ? ' err' : '');
    if (!S.pass) setText(meta, '');
    else if (S.pass.exists === false) setText(meta, 'Read-only');
    else setText(meta, left + ' of ' + total + ' ' + plural(total, 'final', 'finals') + ' left \u00b7 proofs free');
  }

  // Two actions (0.28): a proof is free and unlimited; a final uses a credit.
  if (proofBtn) {
    if (noPass) {
      proofBtn.textContent = 'Build proof';
      proofBtn.disabled = true;
    } else if (busy) {
      proofBtn.textContent = (S.buildingKind || (S.current && S.current.buildKind)) === 'proof' ? 'Building proof\u2026' : 'Build proof';
      proofBtn.disabled = true;
    } else {
      proofBtn.textContent = 'Build proof \u2014 free';
      proofBtn.disabled = !st.ok || !S.current;
    }
  }
  if (btn) {
    if (noPass) {
      btn.textContent = 'Export final';
      btn.disabled = true;
    } else if (busy) {
      btn.textContent = (S.buildingKind || (S.current && S.current.buildKind)) === 'final' ? 'Exporting final\u2026' : 'Export final';
      btn.disabled = true;
    } else if (left <= 0 && S.pass && S.pass.finals_gate !== 'off') {
      btn.textContent = 'No finals left';
      btn.disabled = true;
    } else {
      btn.textContent = 'Export final \u2014 uses 1 of ' + total;
      btn.disabled = !st.ok || !S.current;
    }
  }

  var status = $('fx-build-status');
  if (!status) return;

  var kind = (S.buildingKind || (S.current && S.current.buildKind) || 'final');
  // Only own this line when we're not mid-flight with our own message.
  if (busy) {
    status.className = 'fx-status busy';
    status.textContent = buildingText(kind);
  } else if (isFailed(S.current)) {
    status.className = 'fx-status err';
    status.innerHTML = 'That ' + (kind === 'proof' ? 'proof' : 'final') + ' failed: ' + esc(shortErr(S.current.errorMsg || 'unknown error')) +
      errDetailHTML(S.current.errorMsg) +
      '<span class="fx-status-more">' + (kind === 'proof'
        ? 'Proofs are free, so nothing was used.'
        : 'Failed finals are not counted \u2014 you still have ' + left + ' ' + plural(left, 'final', 'finals') + '.') +
      ' Stuck? Email ' + esc(S.contactEmail) + ' with the message above.</span>';
  } else if (isBuilt(S.current) && S.current.errorMsg) {
    status.className = 'fx-status warn';
    status.innerHTML = 'Build warning: ' + esc(shortErr(S.current.errorMsg)) +
      '<span class="fx-status-more">Your print PDF is still ready below.' +
      (kind === 'final' ? ' This final is counted because a deliverable was produced.' : '') + '</span>';
  } else if (left <= 0 && S.pass && S.pass.exists !== false && S.pass.finals_gate !== 'off') {
    status.className = 'fx-status';
    status.innerHTML = 'You\u2019ve used all ' + total + ' ' + plural(total, 'final', 'finals') + ' on this pass. Proofs still work.' +
      '<span class="fx-status-more">Need more finals? Add +3 above, or email <a href="mailto:' + esc(S.contactEmail) + '">' +
      esc(S.contactEmail) + '</a>. Everything you\u2019ve already built stays downloadable below.</span>';
  } else if (!S.current && S.pass) {
    status.className = 'fx-status';
    status.textContent = 'Upload your manuscript first (step 2).';
  } else if (indexIncludable() && S.includeIndex) {
    status.className = 'fx-status';
    status.textContent = 'The index (step 4) goes in at the back of the print PDF.';
  } else {
    status.className = 'fx-status';
    status.textContent = '';
  }
}

// ─── render: index (step 4, add-on 5.13) ───────────────────────────────────
// S.index is GET /api/books/{id}/index: {status, entitled, index, error}.
// entitled is the server's word (admin, or index_included on the pass).
function indexStatus() { return (S.index && S.index.status) || 'off'; }
function indexEntitled() {
  if (S.index && S.index.entitled) return true;
  return !!(S.pass && S.pass.index_included);
}
function indexIncludable() { var st = indexStatus(); return st === 'draft' || st === 'reviewed'; }
function indexPrice() {
  var it = S.store && S.store.items && S.store.items['index'];
  return it ? it.display : '$100';
}

async function loadIndex() {
  if (!S.current) { S.index = null; S.indexRows = null; return; }
  try {
    var was = S.index;
    S.index = await api('/api/books/' + S.current.id + '/index');
    // Keep unsaved edits unless the document itself changed underneath.
    var changed = !was || !was.index || !S.index.index || was.status !== S.index.status ||
      (was.index.generated !== S.index.index.generated);
    if (changed || !S.indexRows) {
      S.indexRows = S.index.index ? (S.index.index.entries || []).map(rowFromEntry) : null;
      S.indexDirty = false;
    }
  } catch (e) {
    if (e.status === 401) throw e;
    S.index = { status: 'off', entitled: false, _error: e.message };
    S.indexRows = null;
  }
}

function rowFromEntry(e) {
  return {
    heading: e.heading || '', subheading: e.subheading || '', see: e.see || '',
    seeAlso: (e.see_also || []).join('; '),
    anchors: (e.anchors || []).map(function (a) { return { chapter: a.chapter || 0, text: a.text || '' }; }),
  };
}
function entryFromRow(r) {
  return {
    heading: r.heading.trim(), subheading: r.subheading.trim(), see: r.see.trim(),
    see_also: r.seeAlso.split(';').map(function (x) { return x.trim(); }).filter(Boolean),
    anchors: r.anchors,
  };
}
// Unmatched anchors, keyed "heading|subheading" → count (server dry-run).
function unmatchedMap() {
  var m = {};
  var u = (S.index && S.index.index && S.index.index.unmatched) || [];
  u.forEach(function (x) { var k = (x.heading || '') + '|' + (x.subheading || ''); m[k] = (m[k] || 0) + 1; });
  return m;
}

function renderIndex() {
  var meta = $('fx-index-meta');
  var actions = $('fx-index-actions');
  var status = $('fx-index-status');
  var review = $('fx-index-review');
  if (!meta || !actions || !status || !review) return;
  var st = indexStatus();
  var noPass = !S.pass || S.pass.exists === false;
  var entitled = indexEntitled();
  var doc = S.index && S.index.index;
  var n = S.indexRows ? S.indexRows.length : (doc && doc.entries ? doc.entries.length : 0);

  // meta
  meta.className = 'fx-sec-meta' + (st === 'error' ? ' err' : (st === 'reviewed' ? ' ok' : ''));
  if (noPass) setText(meta, 'Read-only');
  else if (!entitled) setText(meta, indexPrice() + ' add-on');
  else if (st === 'drafting') setText(meta, 'Drafting\u2026');
  else if (st === 'draft') setText(meta, n + ' ' + plural(n, 'entry', 'entries') + ' \u00b7 draft, please review');
  else if (st === 'reviewed') setText(meta, n + ' ' + plural(n, 'entry', 'entries') + ' \u00b7 reviewed');
  else if (st === 'error') setText(meta, 'Draft failed');
  else setText(meta, 'Included with your pass');

  // actions
  if (noPass) {
    actions.innerHTML = '';
    status.className = 'fx-status'; status.textContent = '';
    review.innerHTML = '';
    return;
  }
  if (!entitled) {
    var canBuy = S.store && S.store.enabled && S.pass.status !== 'revoked' && S.pass.status !== 'purged';
    actions.innerHTML = '<span class="fx-index-offer"><b>Back-of-book index</b> \u00b7 ' + esc(indexPrice()) + ' add-on</span>' +
      (canBuy ? '<button type="button" class="btn-line" data-addon="index">Add the index \u2014 ' + esc(indexPrice()) + '</button>'
              : '<span class="fx-fine">To add it, email <a href="mailto:' + esc(S.contactEmail) + '">' + esc(S.contactEmail) + '</a>.</span>');
    status.className = 'fx-status'; status.textContent = '';
    review.innerHTML = '';
    return;
  }

  var busy = st === 'drafting';
  var hasDraft = indexIncludable();
  var canDraft = !!S.current && !busy && !isBuilding(S.current) && !S.uploading;
  var label = hasDraft || st === 'error' ? 'Draft the index again' : 'Draft the index';
  actions.innerHTML =
    '<button type="button" class="btn-line" id="fx-index-btn"' + (canDraft ? '' : ' disabled') + '>' + (busy ? 'Drafting\u2026' : esc(label)) + '</button>' +
    (hasDraft ? '<button type="button" class="link-action accent" id="fx-index-save"' + (S.indexDirty && !S.indexSaving ? '' : ' disabled') + '>' +
      (S.indexSaving ? 'Saving\u2026' : (S.indexDirty ? 'Save the review' : (st === 'reviewed' ? 'Saved' : 'Save the review'))) + '</button>' : '');

  // status
  if (busy) {
    status.className = 'fx-status busy';
    status.textContent = 'Drafting \u2014 the factory is reading the book chapter by chapter. It takes a few minutes; this page keeps checking.';
  } else if (st === 'error') {
    status.className = 'fx-status err';
    status.innerHTML = 'The draft didn\u2019t finish: ' + esc(shortErr((S.index && S.index.error) || 'unknown error')) +
      '<span class="fx-status-more">Nothing was charged. Try again, or email ' + esc(S.contactEmail) + '.</span>';
  } else if (!S.current) {
    status.className = 'fx-status';
    status.textContent = 'Upload your manuscript first (step 2).';
  } else if (!hasDraft) {
    status.className = 'fx-status';
    status.textContent = 'Drafting reads the whole book once and takes a few minutes. You review the result here before it goes into a build.';
  } else if (S.indexDirty) {
    status.className = 'fx-status warn';
    status.textContent = 'Unsaved edits \u2014 save the review before you build.';
  } else if (st === 'draft') {
    status.className = 'fx-status';
    status.textContent = 'Read it as a reader would: merge near-duplicates, delete what nobody would look up, fix wording. Then save.';
  } else {
    status.className = 'fx-status ok';
    status.textContent = 'Reviewed. Tick the box below and the next build sets it at the back of the print PDF.';
  }

  renderIndexReview(review, hasDraft);
}

function renderIndexReview(review, hasDraft) {
  if (!hasDraft || !S.indexRows) { review.innerHTML = ''; return; }
  var um = unmatchedMap();
  var rows = S.indexRows;
  // Sorted view: by heading, then subheading; the heading input sits on the
  // first row of each group and renames the whole group.
  var order = rows.map(function (_, i) { return i; }).sort(function (a, b) {
    var ka = sortKey(rows[a].heading), kb = sortKey(rows[b].heading);
    if (ka !== kb) return ka < kb ? -1 : 1;
    var sa = sortKey(rows[a].subheading), sb = sortKey(rows[b].subheading);
    return sa < sb ? -1 : (sa > sb ? 1 : 0);
  });
  var headings = {};
  rows.forEach(function (r) { headings[r.heading.trim().toLowerCase()] = true; });
  var totalUnmatched = 0;
  Object.keys(um).forEach(function (k) { totalUnmatched += um[k]; });

  var h = '<div class="fx-index-head"><span>' + rows.length + ' ' + plural(rows.length, 'line', 'lines') + ', ' +
    Object.keys(headings).length + ' headings</span>' +
    (totalUnmatched ? '<span class="warn">' + totalUnmatched + ' ' + plural(totalUnmatched, 'anchor', 'anchors') + ' not found in the manuscript \u2014 those page numbers will be missing</span>' : '') +
    '</div>';
  h += '<table class="fx-index-table"><thead><tr>' +
    '<th>Heading</th><th>Subentry</th><th><i>See</i></th><th><i>See also</i></th><th class="num">Pages</th><th></th></tr></thead><tbody>';
  var prev = null;
  order.forEach(function (i) {
    var r = rows[i];
    var key = r.heading.trim().toLowerCase();
    var first = key !== prev;
    prev = key;
    var k = r.heading + '|' + r.subheading;
    var bad = um[k] || 0;
    h += '<tr data-row="' + i + '"' + (first ? ' class="first"' : '') + '>';
    h += '<td data-l="Heading">' + (first
      ? '<input class="input-bare" data-f="heading" value="' + esc(r.heading) + '" aria-label="Heading">'
      : '<span class="fx-index-same" title="' + esc(r.heading) + '">\u3003</span>') + '</td>';
    h += '<td data-l="Subentry"><input class="input-bare" data-f="subheading" value="' + esc(r.subheading) + '" placeholder="\u2014" aria-label="Subentry"></td>';
    var seeBad = r.see && !headings[r.see.trim().toLowerCase()];
    h += '<td data-l="See"><input class="input-bare' + (seeBad ? ' bad' : '') + '" data-f="see" value="' + esc(r.see) + '" placeholder="\u2014" aria-label="See"' +
      (seeBad ? ' title="No heading with this name \u2014 it will be dropped on save"' : '') + '></td>';
    h += '<td data-l="See also"><input class="input-bare" data-f="seeAlso" value="' + esc(r.seeAlso) + '" placeholder="\u2014" aria-label="See also"></td>';
    h += '<td class="num" data-l="Pages">' + (r.anchors.length || (r.see ? '\u2014' : '0')) +
      (bad ? ' <span class="fx-index-flag" title="' + bad + ' of the anchor sentences could not be found in the manuscript">\u26a0 ' + bad + '</span>' : '') + '</td>';
    h += '<td class="acts"><button type="button" class="fx-link" data-act="merge" title="Move this line under another heading">merge</button> ' +
      '<button type="button" class="fx-link" data-act="delete">delete</button></td>';
    h += '</tr>';
  });
  h += '</tbody></table>';
  h += '<p class="fx-fine fx-index-foot"><button type="button" class="fx-link" data-act="add">Add an entry</button> \u00b7 ' +
    'Pages counts the places in the text the entry points at; the numbers themselves are set at build time. ' +
    '<i>See also</i> takes several headings separated by semicolons.</p>';
  h += '<p class="fx-index-include"><label><input type="checkbox" id="fx-index-include"' + (S.includeIndex ? ' checked' : '') + '> ' +
    'Include the index in the next build</label></p>';
  review.innerHTML = h;
}

function sortKey(s) {
  return (s || '').toLowerCase().replace(/^(the|a|an) /, '').replace(/[^a-z0-9 ]/g, '').trim();
}

// Draft (async; the page polls) ──
async function doIndexDraft() {
  if (!S.current || indexStatus() === 'drafting') return;
  if (indexIncludable() || S.indexDirty) {
    if (!window.confirm('Draft the index again?\n\nThis replaces the current draft and any edits you have made to it.')) return;
  }
  var status = $('fx-index-status');
  try {
    await api('/api/books/' + S.current.id + '/index/draft', { method: 'POST', body: '{}' });
    S.index = Object.assign({}, S.index || {}, { status: 'drafting', entitled: true });
    S.indexRows = null; S.indexDirty = false;
    renderAll();
    startIndexPolling();
  } catch (e) {
    if (e.status === 401) { showAuth(function () { return refresh(); }); return; }
    if (e.status === 403) { await handleForbidden(status); return; }
    if (e.status === 402) { await loadIndex(); renderAll(); return; }
    status.className = 'fx-status err';
    status.innerHTML = (e.status === 409 ? 'Something else is running for this book \u2014 wait for it to finish, then try again.' :
      'The draft didn\u2019t start: ' + esc(shortErr(e.message))) +
      '<span class="fx-status-more">Nothing was charged.</span>';
  }
}

var INDEX_POLL_MS = 5000;
var INDEX_POLL_MAX = 240; // 20 minutes
function startIndexPolling() {
  if (S.indexPolling) return;
  S.indexPollTicks = 0;
  S.indexPolling = setInterval(indexPollTick, INDEX_POLL_MS);
}
function stopIndexPolling() {
  if (S.indexPolling) clearInterval(S.indexPolling);
  S.indexPolling = null;
}
async function indexPollTick() {
  S.indexPollTicks++;
  if (S.indexPollTicks > INDEX_POLL_MAX) {
    stopIndexPolling();
    var st = $('fx-index-status');
    if (st) { st.className = 'fx-status err'; st.textContent = 'The draft is taking much longer than usual. Reload the page to check on it.'; }
    return;
  }
  try { await loadIndex(); } catch (e) { stopIndexPolling(); if (e.status === 401) showAuth(function () { return refresh(); }); return; }
  if (indexStatus() === 'drafting') return;
  stopIndexPolling();
  S.includeIndex = true;
  renderAll();
  if (indexIncludable()) {
    var sec = $('index');
    if (sec && sec.scrollIntoView) sec.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
}

// Review edits ──
function indexRowEdit(i, field, value) {
  var rows = S.indexRows;
  if (!rows || !rows[i]) return;
  if (field === 'heading') {
    // Renames the whole group (the input sits on the group's first line).
    var old = rows[i].heading.trim().toLowerCase();
    rows.forEach(function (r) { if (r.heading.trim().toLowerCase() === old) r.heading = value; });
  } else {
    rows[i][field] = value;
  }
  S.indexDirty = true;
}
function indexRowMerge(i) {
  var rows = S.indexRows;
  if (!rows || !rows[i]) return;
  var target = window.prompt('Move \u201c' + rows[i].heading + (rows[i].subheading ? ': ' + rows[i].subheading : '') +
    '\u201d under which heading?\n\nType an existing heading. Its pages join that heading (or keep the subentry, if it has one).', '');
  if (target == null) return;
  target = target.trim();
  if (!target) return;
  var from = rows[i].heading;
  rows[i].heading = target;
  // If the old heading has no other lines left, leave a see-reference behind
  // unless the wording is trivially the same.
  var left = rows.some(function (r) { return r.heading.trim().toLowerCase() === from.trim().toLowerCase(); });
  if (!left && sortKey(from) !== sortKey(target)) {
    rows.push({ heading: from, subheading: '', see: target, seeAlso: '', anchors: [] });
  }
  S.indexDirty = true;
  renderIndex();
}
function indexRowDelete(i) {
  if (!S.indexRows || !S.indexRows[i]) return;
  S.indexRows.splice(i, 1);
  S.indexDirty = true;
  renderIndex();
}
function indexRowAdd() {
  if (!S.indexRows) return;
  var heading = window.prompt('New heading:', '');
  if (heading == null || !heading.trim()) return;
  var anchor = window.prompt('A short phrase (5\u201312 words) copied exactly from the manuscript where this topic is discussed. Leave empty for a see-reference.', '');
  if (anchor == null) return;
  var row = { heading: heading.trim(), subheading: '', see: '', seeAlso: '', anchors: [] };
  if (anchor.trim()) row.anchors.push({ chapter: 0, text: anchor.trim() });
  else {
    var see = window.prompt('See which heading?', '');
    if (see == null || !see.trim()) return;
    row.see = see.trim();
  }
  S.indexRows.push(row);
  S.indexDirty = true;
  renderIndex();
}

async function doIndexSave() {
  if (!S.current || !S.indexRows || S.indexSaving) return;
  var entries = S.indexRows.map(entryFromRow).filter(function (e) { return e.heading; });
  var status = $('fx-index-status');
  S.indexSaving = true;
  renderIndex();
  try {
    S.index = await api('/api/books/' + S.current.id + '/index', { method: 'PUT', body: JSON.stringify({ entries: entries }) });
    S.indexRows = (S.index.index.entries || []).map(rowFromEntry);
    S.indexDirty = false;
    S.indexSaving = false;
    S.includeIndex = true;
    renderAll();
    var notes = (S.index.index.notes || []);
    if (notes.length && status) {
      status.className = 'fx-status ok';
      status.innerHTML = 'Saved. The factory tidied ' + notes.length + ' ' + plural(notes.length, 'thing', 'things') + ':' +
        '<span class="fx-status-more">' + notes.slice(0, 6).map(esc).join('<br>') + (notes.length > 6 ? '<br>\u2026' : '') + '</span>';
    }
  } catch (e) {
    S.indexSaving = false;
    renderIndex();
    if (e.status === 401) { showAuth(function () { return refresh(); }); return; }
    if (status) {
      status.className = 'fx-status err';
      status.textContent = 'Not saved: ' + shortErr(e.message);
    }
  }
}

function wireIndex() {
  var sec = $('index');
  if (!sec) return;
  sec.addEventListener('click', function (e) {
    var t = e.target;
    if (!t.closest) return;
    if (t.closest('#fx-index-btn')) { doIndexDraft(); return; }
    if (t.closest('#fx-index-save')) { doIndexSave(); return; }
    var act = t.closest('[data-act]');
    if (!act) return;
    var tr = act.closest('tr[data-row]');
    var i = tr ? parseInt(tr.getAttribute('data-row'), 10) : -1;
    var a = act.getAttribute('data-act');
    if (a === 'add') indexRowAdd();
    else if (a === 'merge' && i >= 0) indexRowMerge(i);
    else if (a === 'delete' && i >= 0) indexRowDelete(i);
  });
  sec.addEventListener('input', function (e) {
    var t = e.target;
    var f = t.getAttribute && t.getAttribute('data-f');
    if (f) {
      var tr = t.closest('tr[data-row]');
      if (tr) indexRowEdit(parseInt(tr.getAttribute('data-row'), 10), f, t.value);
      // Only the action buttons and the status line change while typing —
      // re-rendering the table would steal the caret.
      var save = $('fx-index-save');
      if (save) { save.disabled = false; save.textContent = 'Save the review'; }
      var st = $('fx-index-status');
      if (st) { st.className = 'fx-status warn'; st.textContent = 'Unsaved edits \u2014 save the review before you build.'; }
    }
  });
  sec.addEventListener('change', function (e) {
    var t = e.target;
    if (t.id === 'fx-index-include') { S.includeIndex = !!t.checked; renderBuild(); }
    var f = t.getAttribute && t.getAttribute('data-f');
    if (f === 'heading' || f === 'see') renderIndex(); // regroup / recheck targets on commit
  });
  window.addEventListener('beforeunload', stopIndexPolling);
}

// ─── render: download ──────────────────────────────────────────────────────
function renderDownload() {
  var latest = $('fx-latest');
  var earlier = $('fx-earlier');
  var meta = $('fx-download-meta');

  if (meta) {
    meta.className = 'fx-sec-meta';
    setText(meta, S.outputs.length ? S.outputs.length + ' ' + plural(S.outputs.length, 'file', 'files') + ' kept' : '');
  }

  if (!latest) return;

  if (!S.current || (!isBuilt(S.current) && !S.outputs.length)) {
    latest.innerHTML = '<p class="fx-body">Your files appear here after the first build.</p>';
    if (earlier) earlier.innerHTML = '';
    return;
  }

  // Newest of each format within each kind (0.28): finals first, then the
  // latest proof. Rows before migration 049 carry kind 'final'.
  var base = '/api/books/' + S.current.id + '/download/';
  var newest = { final: {}, proof: {} };
  S.outputs.forEach(function (o) {
    var k = o.kind === 'proof' ? 'proof' : 'final';
    if (!newest[k][o.output_format]) newest[k][o.output_format] = o;
  });
  function when(o) {
    return o && o.created_at ? ' <span class="fx-dl-when">' + esc(fmtWhen(o.created_at)) + '</span>' : '';
  }
  function group(kind, label, note) {
    var pdf = newest[kind].pdf, epub = newest[kind].epub;
    if (!pdf && !epub) return '';
    var q = '?kind=' + kind;
    var h = '<div class="fx-list-head">' + label + '</div><div class="fx-dl">';
    h += pdf ? '<span><a ' + (kind === 'final' || !newest.final.pdf ? 'id="fx-dl-pdf" ' : '') + 'href="' + base + 'pdf' + q + '">Download ' +
               (kind === 'proof' ? 'proof PDF' : 'print PDF') + '</a>' + when(pdf) + '</span>'
             : '<span class="fx-dl-missing">Print PDF \u2014 not built yet</span>';
    h += epub ? '<span><a href="' + base + 'epub' + q + '">Download EPUB</a>' + when(epub) + '</span>'
              : '<span class="fx-dl-missing">EPUB \u2014 not built yet</span>';
    h += '</div>';
    if (note) h += '<p class="fx-fine">' + note + '</p>';
    return h;
  }
  var title = esc(S.current.title || 'your book');
  var html = group('final', 'Final files',
    'Clean print PDF for \u201c' + title + '\u201d \u2014 the one to send a printer. Save these somewhere of your own.');
  html += group('proof', 'Latest proof',
    'Read the EPUB first \u2014 it\u2019s the quickest way to see how the machine understood your file \u2014 then the PDF. ' +
    'The proof PDF carries a PROOF line on every page; export a final when it\u2019s right.');
  latest.innerHTML = html;

  if (!earlier) return;
  // The newest of each kind+format is shown above; the rest are history.
  var seen = {};
  var older = S.outputs.filter(function (o) {
    var key = (o.kind === 'proof' ? 'proof' : 'final') + '/' + o.output_format;
    if (!seen[key]) { seen[key] = true; return false; }
    return true;
  });
  if (!older.length) { earlier.innerHTML = ''; return; }
  var h2 = '<div class="fx-list-head">Earlier builds</div>';
  older.forEach(function (o) {
    h2 += '<div class="fx-row">' +
      '<div><div class="fx-row-main">' + (o.kind === 'proof' ? 'Proof ' : 'Final ') + esc(String(o.output_format || '').toUpperCase()) + '</div>' +
      '<div class="fx-row-sub">' + esc(fmtWhen(o.created_at)) +
      (o.size_bytes ? ' \u00b7 ' + esc(fmtSize(o.size_bytes)) : '') + '</div></div>' +
      '<div class="fx-row-right"><a class="link-action accent" href="/api/books/' + S.current.id +
      '/outputs/' + o.id + '/download">Download</a></div></div>';
  });
  earlier.innerHTML = h2;
}

// Design system rule 5: ONE filled button per page. This page is a wizard, so
// the filled button is the current step's action — which also makes "you are
// here" impossible to miss. Every other step's action stays an underlined link.
function emphasize() {
  var cur = currentStep();
  [[2, 'fx-upload-btn'], [3, 'fx-inspect-btn'], [4, 'fx-index-btn'], [5, 'fx-proof-btn']].forEach(function (pair) {
    var btn = $(pair[1]);
    if (!btn) return;
    var fill = pair[0] === cur;
    // The proof and index actions stay buttons (outlined) when not filled;
    // the rest fall back to text links. "Export final" is always outlined (0.28).
    btn.className = fill ? 'btn-fill' : (pair[0] >= 4 ? 'btn-line' : 'link-action accent');
  });
  // Step 6 has no button — its action is a download link — so when that's the
  // current step the PDF link wears the fill instead.
  var pdf = $('fx-dl-pdf');
  if (pdf) pdf.className = cur === 6 ? 'btn-fill' : '';
}

function renderCover() {
  var wrap = $('fx-cover');
  if (!wrap) return;
  var st = passState();
  var thumb = $('fx-cover-thumb');
  var img = $('fx-cover-img');
  var btn = $('fx-cover-btn');
  var rm = $('fx-cover-remove');
  show(thumb, !!S.hasCover);
  if (S.hasCover && img && S.coverURL && img.getAttribute('src') !== S.coverURL) img.setAttribute('src', S.coverURL);
  if (btn) {
    btn.disabled = !st.ok || S.coverBusy;
    btn.textContent = S.coverBusy ? 'Uploading…' : (S.hasCover ? 'Replace cover image' : 'Add a cover image');
  }
  if (rm) { show(rm, !!S.hasCover); rm.disabled = !st.ok || S.coverBusy; }
}

async function doCoverUpload(file) {
  if (!file || S.coverBusy || !passState().ok) return;
  var status = $('fx-cover-status');
  if (!/^image\/(jpeg|png)$/.test(file.type)) {
    status.className = 'fx-status err';
    status.textContent = 'That isn\u2019t a JPEG or PNG. Export the cover as one of those and try again.';
    return;
  }
  if (file.size > 10 * 1024 * 1024) {
    status.className = 'fx-status err';
    status.textContent = 'That file is over 10 MB. A JPEG at 1600 \u00d7 2400 is usually well under 2 MB.';
    return;
  }
  S.coverBusy = true;
  renderCover();
  status.className = 'fx-status busy';
  status.textContent = 'Uploading ' + file.name + '\u2026';
  var fd = new FormData();
  fd.append('cover', file);
  try {
    await api('/api/projects/' + S.projectId + '/book-spec/cover', { method: 'POST', body: fd, raw: true });
    await loadCover();
    status.className = 'fx-status ok';
    status.textContent = 'Cover saved. It goes into the EPUB on your next build.';
  } catch (e) {
    status.className = 'fx-status err';
    status.textContent = 'Couldn\u2019t save the cover: ' + (e.message || 'unknown error');
  }
  S.coverBusy = false;
  renderCover();
}

async function doCoverRemove() {
  if (S.coverBusy || !passState().ok) return;
  var status = $('fx-cover-status');
  S.coverBusy = true;
  renderCover();
  try {
    await api('/api/projects/' + S.projectId + '/book-spec/cover', { method: 'DELETE' });
    S.hasCover = false;
    if (S.coverURL) { URL.revokeObjectURL(S.coverURL); S.coverURL = null; }
    status.className = 'fx-status ok';
    status.textContent = 'Cover removed. The next EPUB builds without one.';
  } catch (e) {
    status.className = 'fx-status err';
    status.textContent = 'Couldn\u2019t remove the cover: ' + (e.message || 'unknown error');
  }
  S.coverBusy = false;
  renderCover();
}

function renderAll() {
  renderHeader();
  renderSteps();
  renderUpload();
  renderCover();
  renderInspect();
  renderIndex();
  renderBuild();
  renderDownload();
  emphasize();
}

// ─── loaders ───────────────────────────────────────────────────────────────
// Each loader is independently failure-tolerant: one dead endpoint must not
// take the page down.
async function loadPass() {
  try {
    S.pass = await api('/api/projects/' + S.projectId + '/pass');
  } catch (e) {
    if (e.status === 401) throw e;
    if (e.status === 403 || e.status === 404) { S.pass = { exists: false }; return; }
    S.pass = { exists: false, _error: e.message };
    banner('Couldn\u2019t read your pass (' + esc(e.message) + '). The page is showing read-only until that works \u2014 reload to retry.');
  }
}

async function loadBooks() {
  var rows = await api('/api/projects/' + S.projectId + '/books');
  var list = (Array.isArray(rows) ? rows : []).map(normBook).filter(function (b) { return b && b.id; });
  list.sort(function (a, b) {
    var ta = new Date(a.createdAt || 0).getTime() || 0;
    var tb = new Date(b.createdAt || 0).getTime() || 0;
    if (tb !== ta) return tb - ta;
    return b.id - a.id;
  });
  S.books = list;
  S.current = list[0] || null;
}

async function loadPreflight() {
  if (!S.current) { S.preflight = null; return; }
  try {
    // The contract doesn't spell out the query string; the existing handler
    // requires ?book_id=.
    S.preflight = await api('/api/projects/' + S.projectId + '/preflight?book_id=' + S.current.id);
  } catch (e) {
    if (e.status === 401) throw e;
    S.preflight = { exists: false };
  }
}

async function loadOutputs() {
  if (!S.current) { S.outputs = []; return; }
  try {
    var rows = await api('/api/books/' + S.current.id + '/outputs');
    // Newest first; the handler already orders that way, but don't rely on it.
    S.outputs = (Array.isArray(rows) ? rows : []).slice().sort(function (a, b) {
      var d = new Date(b.created_at || 0) - new Date(a.created_at || 0);
      return d || (b.id - a.id);
    });
  } catch (e) {
    if (e.status === 401) throw e;
    S.outputs = [];
  }
}

// The spec JSON is admin-only, so the cover route itself is the "is there
// one?" probe. GET rather than HEAD (proxies in the path don't all forward
// HEAD), and the bytes become the thumbnail so it downloads once.
async function loadCover() {
  try {
    var r = await fetch('/api/projects/' + S.projectId + '/book-spec/cover', { cache: 'no-store', credentials: 'same-origin' });
    if (!r.ok) { S.hasCover = false; S.coverURL = null; return; }
    var blob = await r.blob();
    if (S.coverURL) URL.revokeObjectURL(S.coverURL);
    S.coverURL = URL.createObjectURL(blob);
    S.hasCover = true;
  } catch (e) { S.hasCover = false; S.coverURL = null; }
}

// The transmittal form is section 1 of this page (0.17 C): transmittal.js
// (loaded before us) exposes JdbbTransmittal.mount; it renders into
// #fx-transmittal and reports its status via the tx:status event so the step
// strip stays in step. If the script failed to load we fall back to reading
// the status ourselves and say so in the section.
function mountTransmittal(info) {
  var el = $('fx-transmittal');
  if (!el) return;
  if (window.JdbbTransmittal && typeof JdbbTransmittal.mount === 'function') {
    JdbbTransmittal.mount(el, S.clientSlug, S.projectSlug, info);
    return;
  }
  el.innerHTML = '<p class="fx-status err">The transmittal form didn\u2019t load. Reload the page; if it keeps happening, email <a href="mailto:' +
    esc(S.contactEmail) + '">' + esc(S.contactEmail) + '</a>.</p>';
  loadTransmittal().then(renderSteps);
}
document.addEventListener('tx:status', function (e) {
  S.transmittalStatus = e.detail && e.detail.status ? e.detail.status : null;
  renderSteps();
});
document.addEventListener('tx:unauthorized', function () {
  showAuth(function () { return boot(); });
});

async function loadTransmittal() {
  try {
    var tx = await api('/api/projects/' + S.projectId + '/transmittal');
    S.transmittalStatus = tx && tx.status ? tx.status : null;
  } catch (e) { S.transmittalStatus = null; }
}

async function refresh() {
  await loadPass();
  await loadBooks();
  await Promise.all([loadPreflight(), loadOutputs(), loadCover(), loadIndex()]);
  renderAll();
  if (isBuilding(S.current)) startPolling();
  if (indexStatus() === 'drafting') startIndexPolling();
}

// ─── sign-in gate ──────────────────────────────────────────────────────────
// Primary path (5.7): the customer types the email their Factory Pass went
// to and we mail a one-shot sign-in link (POST /api/public/login-link →
// GET /auth/link). The password from the fulfilment email still works as a
// secondary path: try the client password first, fall back to a project token.
S.authMode = 'email'; // 'email' | 'password' | 'sent'

function setAuthMode(mode) {
  S.authMode = mode;
  show($('fx-auth-email'), mode === 'email');
  show($('fx-auth-sent'), mode === 'sent');
  show($('fx-auth-password'), mode === 'password');
  show($('fx-auth-err'), false);
  var toggle = $('fx-auth-toggle');
  if (toggle) {
    toggle.textContent = mode === 'password'
      ? 'Forgot the password? Get a sign-in link by email instead'
      : (mode === 'sent' ? 'Try a different address' : 'Have a password? Use it instead');
  }
  var focus = mode === 'password' ? $('fx-auth-pw') : $('fx-auth-em');
  if (focus && mode !== 'sent') focus.focus();
}

function showAuth(retryFn) {
  S.retry = retryFn || null;
  show($('fx-shell'), false);
  show($('fx-auth'), true);
  var sub = $('fx-auth-pw-sub');
  if (sub) {
    sub.textContent = 'Enter the password from your Factory Pass email' +
      (S.clientSlug ? ' (account: ' + S.clientSlug + ').' : '.');
  }
  var pw = $('fx-auth-pw');
  if (pw) pw.value = '';
  setAuthMode('email');
}
function hideAuth() {
  show($('fx-auth'), false);
  show($('fx-shell'), true);
}

// doSendLink asks the server to mail a sign-in link. The server always says
// ok (no account enumeration), so the copy on the "sent" panel is hedged.
async function doSendLink() {
  var em = $('fx-auth-em');
  var err = $('fx-auth-err');
  var btn = $('fx-auth-link-btn');
  var email = em ? em.value.trim() : '';
  if (!email || email.indexOf('@') < 0) { setText(err, 'Type the email address your Factory Pass was sent to.'); show(err, true); return; }
  if (!S.clientSlug) { setText(err, 'This page has no client account to sign in to; use the password instead.'); show(err, true); return; }
  show(err, false);
  if (btn) { btn.disabled = true; btn.textContent = 'Sending\u2026'; }
  try {
    await api('/api/public/login-link', { method: 'POST', body: JSON.stringify({ client: S.clientSlug, email: email }) });
    setAuthMode('sent');
  } catch (e) {
    setText(err, 'Could not send just now \u2014 try again in a moment.');
    show(err, true);
  }
  if (btn) { btn.disabled = false; btn.textContent = 'Email me a sign-in link'; }
}

async function doUnlock() {
  var pw = $('fx-auth-pw');
  var err = $('fx-auth-err');
  var btn = $('fx-auth-btn');
  var value = pw ? pw.value : '';
  if (!value) { setText(err, 'Type your password.'); show(err, true); return; }
  show(err, false);
  if (btn) { btn.disabled = true; btn.textContent = 'Checking\u2026'; }

  var body = JSON.stringify({ password: value });
  var ok = false;
  if (S.clientSlug) {
    try { await api('/api/clients/' + encodeURIComponent(S.clientSlug) + '/verify', { method: 'POST', body: body }); ok = true; }
    catch (e) { /* fall through to the project token */ }
  }
  if (!ok && S.projectId) {
    try { await api('/api/projects/' + S.projectId + '/verify', { method: 'POST', body: body }); ok = true; }
    catch (e) { /* handled below */ }
  }
  if (btn) { btn.disabled = false; btn.textContent = 'Unlock'; }
  if (!ok) {
    setText(err, 'That password didn\u2019t work. Check the Factory Pass email \u2014 copy and paste it if you can.');
    show(err, true);
    return;
  }
  hideAuth();
  banner('');
  var again = S.retry;
  S.retry = null;
  try { await (again ? again() : boot()); }
  catch (e) { handleFatal(e); }
}

// The server answers 403 for "no pass" / "expired" (requirePassAccess in
// srv/passes.go). passState() normally disables those buttons first, so a 403
// means our copy of the pass is stale — re-read it and say something human
// instead of echoing the server's phrasing.
async function handleForbidden(statusEl) {
  await loadPass();
  renderAll();
  if (!statusEl) return;
  statusEl.className = 'fx-status err';
  statusEl.textContent = passState().reason ||
    'That isn\u2019t available on this project any more. Reload the page to see where things stand.';
}

// Any 401 anywhere puts the gate up and re-runs what failed once unlocked.
function handleFatal(e) {
  if (e && e.status === 401) { showAuth(function () { return refresh(); }); return; }
  bail('Something went wrong: ' + esc((e && e.message) || 'unknown error') +
    '. Try reloading. If it keeps happening, email <a href="mailto:' + esc(S.contactEmail) + '">' +
    esc(S.contactEmail) + '</a>.');
}

// ─── upload ────────────────────────────────────────────────────────────────
function titleFromProject() {
  return (S.project && S.project.Name) || '';
}
function titleFromFile(name) {
  return String(name || '').replace(/\.docx$/i, '').replace(/[-_]+/g, ' ').trim();
}

function pickFile(file) {
  var status = $('fx-upload-status');
  if (!passState().ok) return;
  if (!file) return;
  if (!/\.docx$/i.test(file.name)) {
    status.className = 'fx-status err';
    status.textContent = 'That\u2019s not a .docx file. Export one from your editor: Word \u2192 Save As \u2192 Word Document; Google Docs \u2192 Download \u2192 Microsoft Word; Pages \u2192 Export To \u2192 Word; LibreOffice \u2192 Save As \u2192 Word 2007\u2013365.';
    return;
  }
  S.pendingFile = file;
  status.className = 'fx-status';
  status.textContent = '';

  var main = $('fx-drop-main');
  if (main) {
    main.innerHTML = '\u2713 ' + esc(file.name) + ' <span class="fx-drop-link">\u00b7 choose a different file</span>';
    main.dataset.picked = 'true';
  }

  // Prefill: project name is the authoritative title for a pass (the project
  // was created from the title they gave at redemption); fall back to the
  // filename. Author comes from the previous upload if there is one.
  var t = $('fx-book-title');
  var a = $('fx-book-author');
  if (t && !t.value.trim()) t.value = titleFromProject() || titleFromFile(file.name);
  if (a && !a.value.trim()) {
    a.value = (S.current && S.current.author) || (S.pass && S.pass.customer_name) || '';
  }
  show($('fx-upload-form'), true);
  renderUpload();
  if (a && !a.value.trim()) a.focus();
}

function resetUpload() {
  S.pendingFile = null;
  var f = $('fx-file');
  if (f) f.value = '';
  var main = $('fx-drop-main');
  if (main) {
    main.innerHTML = 'Drop your .docx here, or <span class="fx-drop-link">choose a file</span>';
    main.dataset.picked = 'false';
  }
  show($('fx-upload-form'), false);
  var t = $('fx-book-title'); if (t) t.value = '';
  var a = $('fx-book-author'); if (a) a.value = '';
  renderUpload();
}

async function doUpload() {
  if (!S.pendingFile || S.uploading) return;
  var status = $('fx-upload-status');
  S.uploading = true;
  renderUpload();
  status.className = 'fx-status busy';
  status.textContent = 'Uploading ' + S.pendingFile.name + '\u2026';

  var fd = new FormData();
  fd.append('file', S.pendingFile);
  fd.append('title', ($('fx-book-title').value || '').trim());
  fd.append('author', ($('fx-book-author').value || '').trim());
  fd.append('series', '');
  fd.append('project_id', String(S.projectId));

  try {
    await api('/api/books/upload', { method: 'POST', body: fd, raw: true });
    status.className = 'fx-status ok';
    status.textContent = 'Uploaded. This is now your current manuscript \u2014 inspect it next (step 3).';
    resetUpload();
    S.uploading = false;
    await refresh();
    var sec = $('inspect');
    if (sec && sec.scrollIntoView) sec.scrollIntoView({ behavior: 'smooth', block: 'start' });
  } catch (e) {
    S.uploading = false;
    if (e.status === 401) { showAuth(function () { return refresh(); }); return; }
    if (e.status === 403) { await handleForbidden(status); return; }
    status.className = 'fx-status err';
    status.textContent = e.status === 413
      ? 'That file is too large to upload here. Email ' + S.contactEmail + ' and we\u2019ll take it another way.'
      : 'Upload didn\u2019t work: ' + e.message;
    renderUpload();
  }
}

// ─── inspect ───────────────────────────────────────────────────────────────
async function doInspect() {
  if (!S.current || S.inspecting) return;
  var status = $('fx-inspect-status');
  S.inspecting = true;
  renderInspect();
  status.className = 'fx-status busy';
  status.textContent = 'Reading your manuscript\u2026';
  try {
    var pf = await api('/api/projects/' + S.projectId + '/preflight', {
      method: 'POST',
      body: JSON.stringify({ book_id: S.current.id }),
    });
    // POST returns the same shape as GET; re-GET anyway so history/report_url
    // are whatever the server considers current.
    S.preflight = pf && pf.exists !== undefined ? pf : null;
    S.inspecting = false;
    await loadPreflight();
    status.className = 'fx-status ok';
    status.textContent = 'Done. Nothing was charged \u2014 inspecting is unlimited.';
    renderAll();
  } catch (e) {
    S.inspecting = false;
    if (e.status === 401) { showAuth(function () { return refresh(); }); return; }
    if (e.status === 403) { await handleForbidden(status); return; }
    status.className = 'fx-status err';
    status.textContent = 'Inspection didn\u2019t run: ' + shortErr(e.message);
    renderInspect();
  }
}

// ─── build ─────────────────────────────────────────────────────────────────
function buildingText(kind) {
  return (kind === 'proof' ? 'Building your proof \u2014 EPUB and print PDF\u2026' : 'Exporting your final \u2014 EPUB and clean print PDF\u2026') +
    ' usually a minute or two, longer if other books are building at the same time. You can leave this page open.';
}

// One build = the EPUB and the print PDF together. kind is 'proof' (free,
// PROOF line on the PDF) or 'final' (one credit, clean PDF) — 0.28.
async function doBuild(kind) {
  if (!S.current || S.building) return;
  kind = kind === 'proof' ? 'proof' : 'final';
  var left = creditsLeft();
  var total = creditsTotal();
  if (kind === 'final') {
    var okToGo = window.confirm(
      'Export a final of \u201c' + (S.current.title || 'your book') + '\u201d now?\n\n' +
      'This makes the EPUB and the clean print PDF and uses 1 of your ' + total + ' ' + plural(total, 'final', 'finals') +
      '. You\u2019ll have ' + Math.max(0, left - 1) + ' left afterwards.\n\n' +
      'A final that fails is not counted. Proofs are always free.'
    );
    if (!okToGo) return;
  }

  var status = $('fx-build-status');
  S.building = true;
  S.buildingKind = kind;
  renderBuild();
  status.className = 'fx-status busy';
  status.textContent = kind === 'proof' ? 'Starting the proof\u2026' : 'Starting the final\u2026';

  try {
    var withIndex = indexIncludable() && S.includeIndex;
    await api('/api/books/' + S.current.id + '/convert', { method: 'POST', body: JSON.stringify({ format: 'both', kind: kind, index: withIndex }) });
    status.className = 'fx-status busy';
    status.textContent = buildingText(kind) + (withIndex ? ' The index goes in at the back.' : '');
    await loadPass();
    renderHeader();
    startPolling();
  } catch (e) {
    S.building = false;
    if (e.status === 401) { showAuth(function () { return refresh(); }); return; }
    if (e.status === 403) { await handleForbidden(status); return; }
    if (e.status === 402 && e.data && e.data.addon === 'index') {
      status.className = 'fx-status err';
      status.innerHTML = 'The index is a $100 add-on this pass doesn\u2019t have yet \u2014 buy it in step 4, or untick \u201cInclude the index\u201d and build without it.';
      await loadIndex();
      renderAll();
      return;
    }
    if (e.status === 402) {
      status.className = 'fx-status err';
      status.innerHTML = 'No finals left on this pass \u2014 proofs still work.' +
        '<span class="fx-status-more">Need more finals? Add +3 above, or email <a href="mailto:' + esc(S.contactEmail) + '">' +
        esc(S.contactEmail) + '</a>. Everything you\u2019ve already built stays downloadable below.</span>';
      await loadPass();
      renderAll();
      return;
    }
    if (e.status === 429) {
      status.className = 'fx-status err';
      status.textContent = 'That\u2019s a lot of proofs for one day \u2014 the limit is 30 per project per 24 hours. Try again later.';
      renderBuild();
      return;
    }
    if (e.status === 409) {
      status.className = 'fx-status busy';
      status.textContent = 'A build is already running for this project. Sit tight \u2014 this page will update when it finishes.';
      startPolling();
      return;
    }
    status.className = 'fx-status err';
    status.innerHTML = 'The build didn\u2019t start: ' + esc(shortErr(e.message)) +
      '<span class="fx-status-more">Nothing was charged. Try again, or email ' + esc(S.contactEmail) + '.</span>';
    renderBuild();
  }
}

// Poll the project's books every 3s while the current one is building.
// Capped so a stuck job doesn't poll forever.
var POLL_MS = 3000;
var POLL_MAX = 200; // 10 minutes
function startPolling() {
  if (S.polling) return;
  S.pollTicks = 0;
  S.polling = setInterval(pollTick, POLL_MS);
}
function stopPolling() {
  if (S.polling) clearInterval(S.polling);
  S.polling = null;
}
async function pollTick() {
  S.pollTicks++;
  if (S.pollTicks > POLL_MAX) {
    stopPolling();
    S.building = false;
    var st = $('fx-build-status');
    if (st) {
      st.className = 'fx-status err';
      st.innerHTML = 'This build is taking much longer than usual.' +
        '<span class="fx-status-more">Reload the page to check on it, or email ' + esc(S.contactEmail) + '.</span>';
    }
    return;
  }
  var wasId = S.current ? S.current.id : null;
  try {
    await loadBooks();
  } catch (e) {
    stopPolling();
    S.building = false;
    if (e.status === 401) { showAuth(function () { return refresh(); }); return; }
    banner('Lost contact with the server while building (' + esc(e.message) + '). Reload to check on it.');
    return;
  }
  var b = S.books.filter(function (x) { return x.id === wasId; })[0] || S.current;
  if (isBuilding(b)) { renderUpload(); return; }

  stopPolling();
  S.building = false;
  var kind = (b && b.buildKind) || S.buildingKind || 'final';
  S.buildingKind = null;

  if (isFailed(b)) {
    await loadPass();           // the refund lands here
    renderAll();
    var el = $('fx-build-status');
    if (el) {
      el.className = 'fx-status err';
      el.innerHTML = 'That ' + (kind === 'proof' ? 'proof' : 'final') + ' failed: ' + esc(shortErr(b.errorMsg || 'unknown error')) +
        errDetailHTML(b.errorMsg) +
        '<span class="fx-status-more">' + (kind === 'proof' ? 'Proofs are free, so nothing was used.' : 'Failed finals are not counted \u2014 your credit came back.') +
        ' Stuck? Email ' + esc(S.contactEmail) + ' with the message above.</span>';
    }
    return;
  }

  await loadPass();
  await loadOutputs();
  renderAll();
  var done = $('fx-build-status');
  if (done) {
    if (b && b.errorMsg) {
      done.className = 'fx-status warn';
      done.innerHTML = 'Build warning: ' + esc(shortErr(b.errorMsg)) +
        '<span class="fx-status-more">Your print PDF is ready in step 6 below.</span>';
    } else {
      done.className = 'fx-status ok';
      done.textContent = (kind === 'proof' ? 'Proof finished.' : 'Final exported.') + ' Your files are in step 6 below.';
    }
  }
  var sec = $('download');
  if (sec && sec.scrollIntoView) sec.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

// ─── wiring ────────────────────────────────────────────────────────────────
function wire() {
  var drop = $('fx-drop');
  var file = $('fx-file');
  if (drop && file) {
    drop.addEventListener('click', function () { if (passState().ok) file.click(); });
    drop.addEventListener('keydown', function (e) {
      if ((e.key === 'Enter' || e.key === ' ') && passState().ok) { e.preventDefault(); file.click(); }
    });
    drop.addEventListener('dragover', function (e) { e.preventDefault(); if (passState().ok) drop.classList.add('dragover'); });
    drop.addEventListener('dragleave', function () { drop.classList.remove('dragover'); });
    drop.addEventListener('drop', function (e) {
      e.preventDefault();
      drop.classList.remove('dragover');
      if (!passState().ok) return;
      var files = e.dataTransfer && e.dataTransfer.files;
      if (!files || !files.length) return;
      if (files.length > 1) {
        var st = $('fx-upload-status');
        st.className = 'fx-status err';
        st.textContent = 'One file at a time, please \u2014 the whole book in a single .docx.';
        return;
      }
      pickFile(files[0]);
    });
    file.addEventListener('change', function () { if (file.files.length) pickFile(file.files[0]); });
  }

  ['fx-book-title', 'fx-book-author'].forEach(function (id) {
    var el = $(id);
    if (el) el.addEventListener('input', renderUpload);
  });

  var up = $('fx-upload-btn');
  if (up) up.addEventListener('click', doUpload);
  var cancel = $('fx-upload-cancel');
  if (cancel) cancel.addEventListener('click', function (e) { e.preventDefault(); resetUpload(); });

  var coverBtn = $('fx-cover-btn');
  var coverFile = $('fx-cover-file');
  if (coverBtn && coverFile) {
    coverBtn.addEventListener('click', function (e) { e.preventDefault(); if (passState().ok) coverFile.click(); });
    coverFile.addEventListener('change', function () {
      if (coverFile.files.length) doCoverUpload(coverFile.files[0]);
      coverFile.value = '';
    });
  }
  var coverRm = $('fx-cover-remove');
  if (coverRm) coverRm.addEventListener('click', function (e) { e.preventDefault(); doCoverRemove(); });

  var ins = $('fx-inspect-btn');
  if (ins) ins.addEventListener('click', doInspect);

  var bld = $('fx-build-btn');
  if (bld) bld.addEventListener('click', function () { doBuild('final'); });
  var prf = $('fx-proof-btn');
  if (prf) prf.addEventListener('click', function () { doBuild('proof'); });
  wireIndex();

  var authBtn = $('fx-auth-btn');
  if (authBtn) authBtn.addEventListener('click', doUnlock);
  var authPw = $('fx-auth-pw');
  if (authPw) authPw.addEventListener('keydown', function (e) { if (e.key === 'Enter') doUnlock(); });
  var linkBtn = $('fx-auth-link-btn');
  if (linkBtn) linkBtn.addEventListener('click', doSendLink);
  var authEm = $('fx-auth-em');
  if (authEm) authEm.addEventListener('keydown', function (e) { if (e.key === 'Enter') doSendLink(); });
  var toggle = $('fx-auth-toggle');
  if (toggle) toggle.addEventListener('click', function (e) {
    e.preventDefault();
    setAuthMode(S.authMode === 'password' ? 'email' : (S.authMode === 'sent' ? 'email' : 'password'));
  });

  // Step links scroll (0.17 C: the transmittal is section 1 of this page).
  [1, 2, 3, 4, 5, 6].forEach(function (n) {
    var el = $('fx-step-' + n);
    if (!el) return;
    el.addEventListener('click', function (e) {
      var target = $(el.getAttribute('href').slice(1));
      if (target && target.scrollIntoView) { e.preventDefault(); target.scrollIntoView({ behavior: 'smooth', block: 'start' }); }
    });
  });

  // Stop polling when the tab is hidden for a long time? No — keep it simple
  // and cheap; the cap above bounds it. Just clean up on unload.
  window.addEventListener('beforeunload', stopPolling);
}

// ─── boot ──────────────────────────────────────────────────────────────────
async function boot() {
  // /{client}/{project}/factory/  → parts = [client, project, 'factory']
  var parts = window.location.pathname.replace(/\/+$/, '').split('/').filter(Boolean);
  var idx = parts.indexOf('factory');
  if (idx < 2) {
    bail('This page needs to be opened from your Factory link \u2014 it looks like ' +
      '<code>/your-account/your-book/factory/</code>. Check the link in your Factory Pass email.');
    return;
  }
  S.clientSlug = parts[idx - 2];
  S.projectSlug = parts[idx - 1];

  var info;
  try {
    info = await api('/api/project-by-path/' + encodeURIComponent(S.clientSlug) + '/' + encodeURIComponent(S.projectSlug));
  } catch (e) {
    if (e.status === 401) { showAuth(function () { return boot(); }); return; }
    if (e.status === 404) {
      bail('We don\u2019t have a book at this address. A small typo in the link is the usual cause \u2014 ' +
        'check the one in your Factory Pass email, or email <a href="mailto:' + esc(S.contactEmail) + '">' +
        esc(S.contactEmail) + '</a>.');
      return;
    }
    bail('Couldn\u2019t load this page: ' + esc(e.message) + '. Try reloading.');
    return;
  }

  S.project = info && info.project;
  S.projectId = S.project && (S.project.ID != null ? S.project.ID : S.project.id);
  renderHeader();

  if (!S.projectId) {
    bail('Couldn\u2019t work out which book this page is for. Email <a href="mailto:' +
      esc(S.contactEmail) + '">' + esc(S.contactEmail) + '</a> with the link you used.');
    return;
  }

  if (info.has_auth && !info.authenticated) { showAuth(function () { return boot(); }); return; }

  mountTransmittal(info);
  try {
    await refresh();
  } catch (e) {
    handleFatal(e);
  }
}

// Theme bar (shared JdbbTheme from theme.js) + first paint.
function mountTheme() {
  var bar = $('theme-bar');
  if (bar && window.JdbbTheme && bar.dataset.mounted !== 'true') {
    JdbbTheme.mount(bar);
    bar.dataset.mounted = 'true';
  }
}

fetch('/api/public/config')
  .then(function (r) { return r.ok ? r.json() : null; })
  .then(function (c) {
    if (c && c.contact_email) { S.contactEmail = c.contact_email; renderHeader(); }
  })
  .catch(function () {});

fetch('/api/public/store/config')
  .then(function (r) { return r.ok ? r.json() : null; })
  .then(function (c) { if (c && c.enabled) { S.store = c; renderHeader(); } })
  .catch(function () {});

document.addEventListener('click', function (e) {
  var b = e.target.closest && e.target.closest('[data-addon]');
  if (b) buyAddon(b.getAttribute('data-addon'));
  // In-page links from the transmittal section ("Continue to 2 · Upload →",
  // "upload it below") scroll like the step strip does.
  var a = e.target.closest && e.target.closest('#fx-transmittal a[href^="#"]');
  if (a) {
    var t = $(a.getAttribute('href').slice(1));
    if (t && t.scrollIntoView) { e.preventDefault(); t.scrollIntoView({ behavior: 'smooth', block: 'start' }); }
  }
});

// Scroll-spy for the sticky step strip (0.26): the step whose section is under
// the strip gets .here, so the reader can see where they are as well as where
// the work is (.current, set by state).
(function () {
  var ids = ['transmittal', 'upload', 'inspect', 'index', 'build', 'download'];
  var ticking = false;
  function update() {
    ticking = false;
    var strip = $('fx-steps');
    var line = (strip ? strip.getBoundingClientRect().bottom : 0) + 24;
    var here = null;
    ids.forEach(function (id, i) {
      var sec = $(id);
      if (sec && sec.getBoundingClientRect().top <= line) here = i + 1;
    });
    ids.forEach(function (_, i) {
      var el = $('fx-step-' + (i + 1));
      if (el) el.classList.toggle('here', here === i + 1);
    });
  }
  window.addEventListener('scroll', function () {
    if (!ticking) { ticking = true; requestAnimationFrame(update); }
  }, { passive: true });
  window.addEventListener('load', update);
})();

mountTheme();
// Workshop copy switchover: after end of day Tue Sep 22 (23:59:59 HKT =
// 15:59:59 UTC), drop the Sep 21–22 workshop sentences from the support
// footer. Mirrors workshopEndsAt in srv/passes.go — keep the two in step.
if (Date.now() > Date.parse('2026-09-22T15:59:59Z')) {
  document.querySelectorAll('[data-workshop]').forEach((el) => el.remove());
  document.querySelectorAll('[data-post-workshop]').forEach((el) => { el.hidden = false; });
}
wire();
renderAll();
boot().then(confirmOrderFromURL).catch(handleFatal);

})();
