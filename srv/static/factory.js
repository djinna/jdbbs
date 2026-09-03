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
  inspecting: false,
  contactEmail: 'j@djinna.com',
  retry: null,         // re-run after the password gate clears
  passUnavailable: false,
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
    var builds = '<span class="' + cls + '">' + left + ' of ' + total + ' ' + plural(total, 'build', 'builds') + ' left</span>';
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
    contact.innerHTML = 'Need more builds, more time, or a pair of eyes on it? Email <a href="mailto:' +
      esc(S.contactEmail) + '">' + esc(S.contactEmail) + '</a>.';
  }
}

// ─── render: step strip ────────────────────────────────────────────────────
// Which step are you on? Cheap heuristic, deliberately generous: the point is
// to answer "what do I do next", not to police anything.
function currentStep() {
  if (!S.books.length) return 2;
  if (isBuilding(S.current)) return 4;
  if (hasDownloads()) return 5;
  if (!S.preflight || S.preflight.exists === false) return 3;
  return 4;
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
    4: hasDownloads(),
    5: false,
  };
  for (var i = 1; i <= 5; i++) {
    var el = $('fx-step-' + i);
    if (!el) continue;
    var cls = 'fx-step';
    if (i === cur) cls += ' current';
    else if (done[i]) cls += ' done';
    el.className = cls;
    el.setAttribute('aria-current', i === cur ? 'step' : 'false');
  }
  var txMeta = $('fx-tx-meta');
  if (txMeta) {
    txMeta.className = 'fx-sec-meta' + (S.transmittalStatus === 'final' ? ' ok' : '');
    setText(txMeta, S.transmittalStatus === 'final' ? '[FINAL]'
      : S.transmittalStatus ? '[' + String(S.transmittalStatus).toUpperCase() + ']' : '');
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
      main.innerHTML = 'Drop your Word file here, or <span class="fx-drop-link">choose a file</span>';
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
  return first.length > 220 ? first.slice(0, 217) + '\u2026' : first;
}

// ─── render: inspect ───────────────────────────────────────────────────────
var TYPE_LABELS = {
  undeclared_custom_style: 'Word styles not in your transmittal',
  declared_custom_style_used: 'Declared styles found in the file',
  observed_style: 'Styles seen in the file',
  image_inventory: 'Images',
  low_resolution_image: 'Images too low-resolution for print',
  manual_page_break: 'Manual page breaks',
  empty_paragraph: 'Empty paragraphs',
  tracked_change: 'Tracked changes still in the file',
  comment: 'Comments still in the file',
  footnote: 'Footnotes',
  table: 'Tables',
  hyperlink: 'Links',
};
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
      '</p><p class="fx-fine">Re-save it from Word as .docx and upload again. Nothing was charged.</p>';
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
    html += '<p class="fx-fine">You can build anyway \u2014 nothing here blocks it. But the report explains each item, and fixing the \u201cworth fixing\u201d ones in Word before you spend a build usually saves you one.</p>';
  } else {
    html += '<p class="fx-fine">Nothing serious. The report has the detail if you\u2019re curious.</p>';
  }

  if (pf.images && pf.images.length) {
    html += '<p class="fx-fine">' + pf.images.length + ' ' + plural(pf.images.length, 'image', 'images') +
      ' found \u2014 the report lists each one\u2019s size in print.</p>';
  }

  box.innerHTML = html;
}

// ─── render: build ─────────────────────────────────────────────────────────
function renderBuild() {
  var st = passState();
  var left = creditsLeft();
  var total = creditsTotal();
  var btn = $('fx-build-btn');
  var meta = $('fx-build-meta');

  if (meta) {
    meta.className = 'fx-sec-meta' + (left === 0 && S.pass && S.pass.exists !== false ? ' err' : '');
    if (!S.pass) setText(meta, '');
    else if (S.pass.exists === false) setText(meta, 'Read-only');
    else setText(meta, left + ' of ' + total + ' left');
  }

  if (btn) {
    if (!S.pass || S.pass.exists === false) {
      btn.textContent = 'Build PDF + EPUB';
      btn.disabled = true;
    } else if (S.building || isBuilding(S.current)) {
      btn.textContent = 'Building\u2026';
      btn.disabled = true;
    } else if (left <= 0) {
      btn.textContent = 'No builds left';
      btn.disabled = true;
    } else {
      btn.textContent = 'Build PDF + EPUB (uses 1 of ' + total + ' builds)';
      btn.disabled = !st.ok || !S.current;
    }
  }

  var status = $('fx-build-status');
  if (!status) return;

  // Only own this line when we're not mid-flight with our own message.
  if (S.building || isBuilding(S.current)) {
    status.className = 'fx-status busy';
    status.textContent = 'Building your PDF and EPUB\u2026 this usually takes a minute or two. You can leave this page open.';
  } else if (isFailed(S.current)) {
    status.className = 'fx-status err';
    status.innerHTML = 'That build failed: ' + esc(shortErr(S.current.errorMsg || 'unknown error')) +
      '<span class="fx-status-more">Failed builds are not counted \u2014 you still have ' + left + ' ' +
      plural(left, 'build', 'builds') + '. Try inspecting the file, or email ' + esc(S.contactEmail) + ' with the message above.</span>';
  } else if (left <= 0 && S.pass && S.pass.exists !== false) {
    status.className = 'fx-status';
    status.innerHTML = 'You\u2019ve used all ' + total + ' ' + plural(total, 'build', 'builds') + ' on this pass.' +
      '<span class="fx-status-more">Need more builds? Email <a href="mailto:' + esc(S.contactEmail) + '">' +
      esc(S.contactEmail) + '</a>. Everything you\u2019ve already built stays downloadable below.</span>';
  } else if (!S.current && S.pass) {
    status.className = 'fx-status';
    status.textContent = 'Upload your manuscript first (step 2).';
  } else {
    status.className = 'fx-status';
    status.textContent = '';
  }
}

// ─── render: download ──────────────────────────────────────────────────────
function renderDownload() {
  var latest = $('fx-latest');
  var earlier = $('fx-earlier');
  var meta = $('fx-download-meta');

  var haveFormats = {};
  S.outputs.forEach(function (o) { haveFormats[o.output_format] = true; });

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

  var base = '/api/books/' + S.current.id + '/download/';
  var pdfOK = haveFormats.pdf || isBuilt(S.current);
  var epubOK = haveFormats.epub;
  var html = '<div class="fx-dl">';
  html += pdfOK ? '<a id="fx-dl-pdf" href="' + base + 'pdf">Download print PDF</a>'
                : '<span class="fx-dl-missing">Print PDF \u2014 not built yet</span>';
  html += epubOK ? '<a href="' + base + 'epub">Download EPUB</a>'
                 : '<span class="fx-dl-missing">EPUB \u2014 not built yet</span>';
  html += '</div>';
  html += '<p class="fx-fine">Latest build of \u201c' + esc(S.current.title || 'your book') + '\u201d' +
    (S.current.updatedAt ? ' \u00b7 ' + esc(fmtWhen(S.current.updatedAt)) : '') +
    '. Save these somewhere of your own \u2014 the PDF is the one to send a printer.</p>';
  // A build is meant to produce both formats. If only one showed up, say so
  // plainly rather than leaving a dead-looking label.
  if (pdfOK && !epubOK) {
    html += '<p class="fx-fine">The EPUB from this build isn\u2019t here yet. Reload in a moment; if it still doesn\u2019t appear, email ' +
      esc(S.contactEmail) + ' \u2014 you shouldn\u2019t need to spend another build for it.</p>';
  }
  latest.innerHTML = html;

  if (!earlier) return;
  // Newest of each format is the "latest" above; the rest are history.
  var seen = {};
  var older = S.outputs.filter(function (o) {
    if (!seen[o.output_format]) { seen[o.output_format] = true; return false; }
    return true;
  });
  if (!older.length) { earlier.innerHTML = ''; return; }
  var h2 = '<div class="fx-list-head">Earlier builds</div>';
  older.forEach(function (o) {
    h2 += '<div class="fx-row">' +
      '<div><div class="fx-row-main">' + esc(String(o.output_format || '').toUpperCase()) + '</div>' +
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
  [[2, 'fx-upload-btn'], [3, 'fx-inspect-btn'], [4, 'fx-build-btn']].forEach(function (pair) {
    var btn = $(pair[1]);
    if (!btn) return;
    btn.className = pair[0] === cur ? 'btn-fill' : 'link-action accent';
  });
  // Step 5 has no button — its action is a download link — so when that's the
  // current step the PDF link wears the fill instead.
  var pdf = $('fx-dl-pdf');
  if (pdf) pdf.className = cur === 5 ? 'btn-fill' : '';
}

function renderAll() {
  renderHeader();
  renderSteps();
  renderUpload();
  renderInspect();
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

async function loadTransmittal() {
  try {
    var tx = await api('/api/projects/' + S.projectId + '/transmittal');
    S.transmittalStatus = tx && tx.status ? tx.status : null;
  } catch (e) { S.transmittalStatus = null; }
}

async function refresh() {
  await loadPass();
  await loadBooks();
  await Promise.all([loadPreflight(), loadOutputs()]);
  renderAll();
  if (isBuilding(S.current)) startPolling();
}

// ─── password gate ─────────────────────────────────────────────────────────
// Workshop attendees get a client slug + password from their fulfillment
// email, so try the client password first and fall back to a project token.
function showAuth(retryFn) {
  S.retry = retryFn || null;
  show($('fx-shell'), false);
  show($('fx-auth'), true);
  var sub = $('fx-auth-sub');
  if (sub) {
    sub.textContent = 'Enter the password from your Factory Pass email' +
      (S.clientSlug ? ' (account: ' + S.clientSlug + ').' : '.');
  }
  var pw = $('fx-auth-pw');
  if (pw) { pw.value = ''; pw.focus(); }
}
function hideAuth() {
  show($('fx-auth'), false);
  show($('fx-shell'), true);
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
    status.textContent = 'That\u2019s not a .docx file. In Word: File \u2192 Save As \u2192 Word Document (.docx).';
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
    main.innerHTML = 'Drop your Word file here, or <span class="fx-drop-link">choose a file</span>';
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
async function doBuild() {
  if (!S.current || S.building) return;
  var left = creditsLeft();
  var total = creditsTotal();
  var okToGo = window.confirm(
    'Build \u201c' + (S.current.title || 'your book') + '\u201d now?\n\n' +
    'This uses 1 of your ' + total + ' ' + plural(total, 'build', 'builds') +
    '. You\u2019ll have ' + Math.max(0, left - 1) + ' left afterwards.\n\n' +
    'A build that fails is not counted.'
  );
  if (!okToGo) return;

  var status = $('fx-build-status');
  S.building = true;
  renderBuild();
  status.className = 'fx-status busy';
  status.textContent = 'Starting the build\u2026';

  try {
    await api('/api/books/' + S.current.id + '/convert', { method: 'POST' });
    status.className = 'fx-status busy';
    status.textContent = 'Building your PDF and EPUB\u2026 this usually takes a minute or two. You can leave this page open.';
    await loadPass();
    renderHeader();
    startPolling();
  } catch (e) {
    S.building = false;
    if (e.status === 401) { showAuth(function () { return refresh(); }); return; }
    if (e.status === 403) { await handleForbidden(status); return; }
    if (e.status === 402) {
      status.className = 'fx-status err';
      status.innerHTML = 'No builds left on this pass.' +
        '<span class="fx-status-more">Need more builds? Email <a href="mailto:' + esc(S.contactEmail) + '">' +
        esc(S.contactEmail) + '</a>. Everything you\u2019ve already built stays downloadable below.</span>';
      await loadPass();
      renderAll();
      return;
    }
    if (e.status === 409) {
      status.className = 'fx-status busy';
      status.textContent = 'A build is already running for this book. Sit tight \u2014 this page will update when it finishes.';
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

  if (isFailed(b)) {
    await loadPass();           // the refund lands here
    renderAll();
    var el = $('fx-build-status');
    if (el) {
      el.className = 'fx-status err';
      el.innerHTML = 'That build failed: ' + esc(shortErr(b.errorMsg || 'unknown error')) +
        '<span class="fx-status-more">Failed builds are not counted \u2014 your credit came back. Inspect the file for clues, or email ' +
        esc(S.contactEmail) + ' with the message above.</span>';
    }
    return;
  }

  await loadPass();
  await loadOutputs();
  renderAll();
  var done = $('fx-build-status');
  if (done) {
    done.className = 'fx-status ok';
    done.textContent = 'Build finished. Your files are in step 5 below.';
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

  var ins = $('fx-inspect-btn');
  if (ins) ins.addEventListener('click', doInspect);

  var bld = $('fx-build-btn');
  if (bld) bld.addEventListener('click', doBuild);

  var authBtn = $('fx-auth-btn');
  if (authBtn) authBtn.addEventListener('click', doUnlock);
  var authPw = $('fx-auth-pw');
  if (authPw) authPw.addEventListener('keydown', function (e) { if (e.key === 'Enter') doUnlock(); });
  var forgot = $('fx-auth-forgot');
  if (forgot) forgot.addEventListener('click', function (e) {
    e.preventDefault();
    var subject = encodeURIComponent('Factory Pass access — ' + (S.clientSlug || ''));
    var body = encodeURIComponent('Hi,\n\nI can\u2019t get into my Factory page and need the password again.\n\n' +
      'Page: ' + window.location.href + '\n\nThanks.\n');
    window.location.href = 'mailto:' + S.contactEmail + '?subject=' + subject + '&body=' + body;
  });

  // Step links scroll; step 1 is a real link to the transmittal.
  [2, 3, 4, 5].forEach(function (n) {
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

  loadTransmittal().then(renderSteps);
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

mountTheme();
wire();
renderAll();
boot().catch(handleFatal);

})();
