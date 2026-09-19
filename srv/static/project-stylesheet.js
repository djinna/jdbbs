// Per-project editorial style sheet (punch list 6.1 B).
//
// One page per book at /{client}/{project}/stylesheet/. The server seeds the
// sheet from the house sheet on first open (srv/project_stylesheet.go); this
// file renders it and saves each decision as it is made. Everything it talks
// to is behind the project auth gate, so a 401 anywhere puts the sign-in
// panel up — the same two paths as the factory page (emailed link first,
// password as a fallback).
(function () {
  'use strict';

  var S = {
    clientSlug: '', projectSlug: '', projectId: null,
    sheet: null,
    editing: null,   // item id being edited inline
    adding: null,    // section ord whose "add a rule" form is open
    authMode: 'email',
    retry: null,
    contactEmail: 'j@djinna.com'
  };

  // ── small helpers ────────────────────────────────────────────────────
  function $(id) { return document.getElementById(id); }
  function show(el, on) { if (el) el.hidden = !on; }
  function setText(el, t) { if (el) el.textContent = t; }
  function esc(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }
  // The house sheet stores light Markdown in its cells; render the same
  // subset here so a rule reads identically in both places.
  function inlineMd(s) {
    var t = esc(s), codes = [];
    t = t.replace(/`([^`]+)`/g, function (_, c) { codes.push(c); return '\u0000' + (codes.length - 1) + '\u0000'; });
    t = t.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    t = t.replace(/(^|[^*])\*([^*\n]+)\*/g, '$1<em>$2</em>');
    t = t.replace(/\u0000(\d+)\u0000/g, function (_, i) { return '<code>' + codes[+i] + '</code>'; });
    return t;
  }
  function blockMd(src) {
    var lines = String(src || '').replace(/\r\n?/g, '\n').split('\n'), out = [], i = 0;
    function list(tag, items) {
      out.push('<' + tag + '>' + items.map(function (x) { return '<li>' + inlineMd(x) + '</li>'; }).join('') + '</' + tag + '>');
    }
    while (i < lines.length) {
      var l = lines[i];
      if (/^\s*$/.test(l)) { i++; continue; }
      if (/^\s*[-*+]\s+/.test(l)) {
        var ul = [];
        while (i < lines.length && /^\s*[-*+]\s+/.test(lines[i])) { ul.push(lines[i].replace(/^\s*[-*+]\s+/, '')); i++; }
        list('ul', ul); continue;
      }
      if (/^\s*\d+[.)]\s+/.test(l)) {
        var ol = [];
        while (i < lines.length && /^\s*\d+[.)]\s+/.test(lines[i])) { ol.push(lines[i].replace(/^\s*\d+[.)]\s+/, '')); i++; }
        list('ol', ol); continue;
      }
      var p = [];
      while (i < lines.length && !/^\s*$/.test(lines[i]) && !/^\s*[-*+]\s+/.test(lines[i]) && !/^\s*\d+[.)]\s+/.test(lines[i])) {
        p.push(lines[i]); i++;
      }
      out.push('<p>' + inlineMd(p.join(' ')) + '</p>');
    }
    return out.join('');
  }

  function ApiError(message, status) { this.message = message; this.status = status; }
  ApiError.prototype = Object.create(Error.prototype);

  async function api(url, opts) {
    opts = opts || {};
    var r;
    try {
      r = await fetch(url, {
        method: opts.method || 'GET',
        headers: opts.body ? { 'Content-Type': 'application/json' } : undefined,
        body: opts.body
      });
    } catch (e) {
      throw new ApiError("Couldn't reach the server. Check your connection and try again.", 0);
    }
    var text = await r.text().catch(function () { return ''; });
    var data = null;
    if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
    if (!r.ok) throw new ApiError((data && data.error) || (r.status + ' ' + r.statusText), r.status);
    return data;
  }

  // ── save indicator (same idea as the transmittal's) ──────────────────
  var saveTimer = null;
  function mark(state, msg) {
    var el = $('ss-save');
    if (!el) return;
    el.className = 'ss-save ' + (state || '');
    el.textContent = msg || '';
    if (saveTimer) clearTimeout(saveTimer);
    if (state === 'saved') {
      saveTimer = setTimeout(function () { if (el.className.indexOf('saved') >= 0) { el.textContent = ''; el.className = 'ss-save'; } }, 2200);
    }
  }
  function banner(msg) {
    var el = $('ss-banner');
    if (!el) return;
    el.innerHTML = msg || '';
    show(el, !!msg);
  }

  // Any 401 puts the gate up and re-runs what failed once unlocked.
  function handle(e, retry) {
    if (e && e.status === 401) { showAuth(retry || boot); return; }
    mark('err', 'Not saved');
    banner('Something went wrong: ' + esc((e && e.message) || 'unknown error') +
      '. Nothing was lost &mdash; try again, or reload the page.');
  }

  function base() { return '/api/projects/' + S.projectId + '/stylesheet'; }

  // ── rendering ────────────────────────────────────────────────────────
  var MARKS = {
    accepted: 'Our default, accepted',
    edited: 'Edited by you',
    rejected: 'Not for this book',
    added: 'Your rule'
  };

  function render() {
    var sheet = S.sheet;
    if (!sheet) return;
    setText($('ss-title'), sheet.project_name || S.projectSlug);
    setText($('ss-byline'), sheet.client_slug + ' / ' + sheet.project_slug +
      ' · ' + (sheet.book_kind === 'both' ? 'general rules' : sheet.book_kind));
    show($('ss-kind'), !sheet.kind_chosen);

    var c = sheet.counts || {};
    $('ss-counts').innerHTML =
      '<b>' + c.house + '</b> studio rules · <b>' + c.accepted + '</b> accepted · <b>' +
      c.edited + '</b> edited · <b>' + c.rejected + '</b> rejected · <b>' + c.added + '</b> added by you';

    var body = $('ss-body');
    if (!sheet.sections || !sheet.sections.length) {
      body.innerHTML = '<p class="ss-empty">This sheet is empty.</p>';
      return;
    }
    body.innerHTML = sheet.sections.map(function (sec, i) {
      return '<section class="ss-section" id="s' + sec.ord + '">' +
        '<h2><span class="n">' + (i + 1) + '</span>' + esc(sec.title) + '</h2>' +
        sec.items.map(renderItem).join('') +
        renderAdd(sec) +
        '</section>';
    }).join('');
  }

  function renderMeta(it) {
    var acts = [];
    if (it.status === 'rejected') {
      acts.push(btn('accept', it.id, 'Put back'));
    } else {
      acts.push(btn('reject', it.id, 'Reject'));
      acts.push(btn('edit', it.id, 'Edit'));
    }
    if (it.house_id && (it.status === 'edited' || it.house)) acts.push(btn('restore', it.id, 'Restore ours'));
    if (!it.house_id) acts.push(btn('delete', it.id, 'Delete'));
    acts.push(btn('note', it.id, it.note ? 'Edit note' : 'Add note'));
    return '<div class="ss-meta">' +
      '<span class="ss-mark ' + esc(it.status) + '">' + esc(MARKS[it.status] || it.status) + '</span>' +
      '<span class="ss-acts">' + acts.join('') + '</span></div>';
  }

  function btn(act, id, label) {
    return '<button type="button" data-act="' + act + '" data-id="' + id + '">' + esc(label) + '</button>';
  }

  function renderNote(it) {
    if (S.editing === 'note:' + it.id) {
      return '<div class="ss-note"><textarea class="ss-note-input" data-note="' + it.id + '" rows="2" ' +
        'placeholder="A note for your copyeditor">' + esc(it.note) + '</textarea>' +
        '<span class="ss-edit-actions">' + btn('note-save', it.id, 'Save note') + btn('cancel', it.id, 'Cancel') + '</span></div>';
    }
    return it.note ? '<div class="ss-note">Note: ' + esc(it.note) + '</div>' : '';
  }

  function renderItem(it) {
    var cls = 'ss-item is-' + it.status + (it.house_id ? '' : ' is-added');
    if (S.editing === it.id) return '<div class="' + cls + '">' + renderEditor(it) + '</div>';

    if (it.kind === 'rule') {
      return '<div class="' + cls + '"><div class="ss-grid">' +
        '<div class="ss-c1">' + inlineMd(it.col1) + '</div>' +
        '<div class="ss-c2">' + inlineMd(it.col2) + '</div>' +
        '<div class="ss-c3">' + inlineMd(it.col3) + '</div>' +
        renderMeta(it) + renderNote(it) +
        '</div></div>';
    }
    return '<div class="' + cls + '"><div class="ss-proserow">' +
      '<div class="ss-prose">' + blockMd(it.body) + '</div>' +
      renderMeta(it) + renderNote(it) +
      '</div></div>';
  }

  // Inline edit: a plain textarea swap, three cells for a rule, one for prose.
  function renderEditor(it) {
    var fields;
    if (it.kind === 'rule') {
      fields = '<div class="ss-edit">' +
        '<textarea data-f="col1" rows="2" placeholder="Item">' + esc(it.col1) + '</textarea>' +
        '<textarea data-f="col2" rows="2" placeholder="Rule">' + esc(it.col2) + '</textarea>' +
        '<textarea data-f="col3" rows="2" placeholder="Example / note">' + esc(it.col3) + '</textarea>';
    } else {
      fields = '<div class="ss-edit one">' +
        '<textarea data-f="body" rows="5" placeholder="Your wording">' + esc(it.body) + '</textarea>';
    }
    return fields + '<div class="ss-edit-actions">' +
      btn('save', it.id, 'Save') + btn('cancel', it.id, 'Cancel') +
      (it.house_id ? btn('restore', it.id, 'Restore ours') : '') +
      '</div></div>';
  }

  function renderAdd(sec) {
    if (S.adding !== sec.ord) {
      return '<div class="ss-add">' +
        '<button type="button" data-act="add-open" data-ord="' + sec.ord + '">+ Add a rule to ' + esc(sec.title) + '</button></div>';
    }
    return '<div class="ss-add"><div class="ss-edit">' +
      '<textarea data-nf="col1" rows="2" placeholder="Item (e.g. Serial comma)"></textarea>' +
      '<textarea data-nf="col2" rows="2" placeholder="Rule (what we should do)"></textarea>' +
      '<textarea data-nf="col3" rows="2" placeholder="Example / note (optional)"></textarea>' +
      '<div class="ss-edit-actions">' +
      '<button type="button" data-act="add-save" data-ord="' + sec.ord + '">Add rule</button>' +
      '<button type="button" data-act="add-cancel" data-ord="' + sec.ord + '">Cancel</button>' +
      '</div></div></div>';
  }

  // ── actions ──────────────────────────────────────────────────────────
  function findItem(id) {
    var found = null;
    (S.sheet.sections || []).forEach(function (sec) {
      sec.items.forEach(function (it) { if (it.id === id) found = it; });
    });
    return found;
  }

  async function reload() {
    S.sheet = await api(base());
    render();
  }

  async function patchItem(id, body) {
    mark('saving', 'Saving…');
    var updated = await api(base() + '/items/' + id, { method: 'PATCH', body: JSON.stringify(body) });
    // Counts live on the sheet, so re-read it: one small request, always right.
    await reload();
    mark('saved', 'Saved ✓');
    return updated;
  }

  async function onAction(act, id, el) {
    var it = id ? findItem(id) : null;
    switch (act) {
      case 'reject':
        await patchItem(id, { status: 'rejected' });
        break;
      case 'accept':
        await patchItem(id, { status: 'accepted' });
        break;
      case 'edit':
        S.editing = id; render();
        var first = document.querySelector('.ss-edit textarea');
        if (first) first.focus();
        break;
      case 'note':
        S.editing = 'note:' + id; render();
        var na = document.querySelector('.ss-note-input');
        if (na) na.focus();
        break;
      case 'cancel':
        S.editing = null; render();
        break;
      case 'save': {
        var box = el.closest('.ss-item');
        var body = {};
        box.querySelectorAll('[data-f]').forEach(function (t) { body[t.dataset.f] = t.value; });
        S.editing = null;
        await patchItem(id, body);
        break;
      }
      case 'note-save': {
        var ta = el.closest('.ss-item').querySelector('.ss-note-input');
        S.editing = null;
        await patchItem(id, { note: ta ? ta.value : '' });
        break;
      }
      case 'restore':
        mark('saving', 'Saving…');
        await api(base() + '/items/' + id + '/restore', { method: 'POST' });
        S.editing = null;
        await reload();
        mark('saved', 'Saved ✓');
        break;
      case 'delete':
        if (!window.confirm('Delete your rule “' + ((it && (it.col1 || it.col2 || it.body)) || '') + '”?')) return;
        mark('saving', 'Saving…');
        await api(base() + '/items/' + id, { method: 'DELETE' });
        await reload();
        mark('saved', 'Saved ✓');
        break;
    }
  }

  async function onSectionAction(act, ord, el) {
    if (act === 'add-open') { S.adding = ord; render();
      var t = document.querySelector('[data-nf="col1"]'); if (t) t.focus(); return; }
    if (act === 'add-cancel') { S.adding = null; render(); return; }
    if (act === 'add-save') {
      var wrap = el.closest('.ss-add');
      var body = { section_ord: ord };
      wrap.querySelectorAll('[data-nf]').forEach(function (t) { body[t.dataset.nf] = t.value; });
      if (!String(body.col1 || '').trim() && !String(body.col2 || '').trim()) {
        banner('A new rule needs at least an item or a rule.');
        return;
      }
      banner('');
      mark('saving', 'Saving…');
      await api(base() + '/items', { method: 'POST', body: JSON.stringify(body) });
      S.adding = null;
      await reload();
      mark('saved', 'Saved ✓');
    }
  }

  document.addEventListener('click', async function (e) {
    var el = e.target.closest('[data-act]');
    if (!el) return;
    e.preventDefault();
    var act = el.dataset.act;
    try {
      if (el.dataset.ord !== undefined) await onSectionAction(act, parseInt(el.dataset.ord, 10), el);
      else await onAction(act, parseInt(el.dataset.id, 10), el);
    } catch (err) {
      handle(err, function () { return reload(); });
    }
  });

  // Accept all: puts every rejected rule back; edits and your own rules stay.
  function wireToolbar() {
    var acc = $('ss-accept-all');
    if (acc) acc.addEventListener('click', async function () {
      if (!window.confirm('Accept the studio defaults for every rule? Your edits and your own rules are kept; rules you rejected come back.')) return;
      try {
        mark('saving', 'Saving…');
        S.sheet = await api(base() + '/accept-all', { method: 'POST' });
        render();
        mark('saved', 'Saved ✓');
      } catch (err) { handle(err); }
    });
    var kind = $('ss-kind');
    if (kind) kind.addEventListener('click', async function (e) {
      var b = e.target.closest('[data-kind]');
      if (b) {
        try {
          mark('saving', 'Saving…');
          S.sheet = await api(base(), { method: 'PATCH', body: JSON.stringify({ book_kind: b.dataset.kind }) });
          render();
          mark('saved', 'Saved ✓');
        } catch (err) { handle(err); }
        return;
      }
      if (e.target.id === 'ss-kind-later') show($('ss-kind'), false);
    });
  }

  // ── sign-in gate (mirrors factory.js; the server gate is unchanged) ───
  function setAuthMode(mode) {
    S.authMode = mode;
    show($('ss-auth-email'), mode === 'email');
    show($('ss-auth-sent'), mode === 'sent');
    show($('ss-auth-password'), mode === 'password');
    show($('ss-auth-err'), false);
    var toggle = $('ss-auth-toggle');
    if (toggle) {
      toggle.textContent = mode === 'password'
        ? 'Forgot the password? Get a sign-in link by email instead'
        : (mode === 'sent' ? 'Try a different address' : 'Have a password? Use it instead');
    }
    var focus = mode === 'password' ? $('ss-auth-pw') : $('ss-auth-em');
    if (focus && mode !== 'sent') focus.focus();
  }

  function showAuth(retry) {
    S.retry = retry || null;
    show($('ss-shell'), false);
    show($('ss-auth'), true);
    var sub = $('ss-auth-pw-sub');
    if (sub) sub.textContent = 'Enter the password from your Factory Pass email' + (S.clientSlug ? ' (account: ' + S.clientSlug + ').' : '.');
    setAuthMode('email');
  }
  function hideAuth() { show($('ss-auth'), false); show($('ss-shell'), true); }

  async function sendLink() {
    var err = $('ss-auth-err'), em = $('ss-auth-em');
    var email = em ? em.value.trim() : '';
    if (!email || email.indexOf('@') < 0) { setText(err, 'Type the email address your Factory Pass was sent to.'); show(err, true); return; }
    if (!S.clientSlug) { setText(err, 'This page has no account to sign in to; use the password instead.'); show(err, true); return; }
    show(err, false);
    try {
      await api('/api/public/login-link', { method: 'POST', body: JSON.stringify({ client: S.clientSlug, email: email }) });
      setAuthMode('sent');
    } catch (e) {
      setText(err, 'Could not send just now — try again in a moment.');
      show(err, true);
    }
  }

  async function unlock() {
    var err = $('ss-auth-err'), pw = $('ss-auth-pw');
    var value = pw ? pw.value : '';
    if (!value) { setText(err, 'Type your password.'); show(err, true); return; }
    show(err, false);
    var body = JSON.stringify({ password: value }), ok = false;
    if (S.clientSlug) {
      try { await api('/api/clients/' + encodeURIComponent(S.clientSlug) + '/verify', { method: 'POST', body: body }); ok = true; } catch (e) { /* try the project token */ }
    }
    if (!ok && S.projectId) {
      try { await api('/api/projects/' + S.projectId + '/verify', { method: 'POST', body: body }); ok = true; } catch (e) { /* reported below */ }
    }
    if (!ok) {
      setText(err, 'That password didn’t work. Check the Factory Pass email — copy and paste it if you can.');
      show(err, true);
      return;
    }
    hideAuth();
    var again = S.retry; S.retry = null;
    try { await (again ? again() : boot()); } catch (e) { handle(e); }
  }

  function wireAuth() {
    var link = $('ss-auth-link-btn'); if (link) link.addEventListener('click', sendLink);
    var btn = $('ss-auth-btn'); if (btn) btn.addEventListener('click', unlock);
    var toggle = $('ss-auth-toggle');
    if (toggle) toggle.addEventListener('click', function (e) {
      e.preventDefault();
      setAuthMode(S.authMode === 'password' ? 'email' : 'password');
    });
    var em = $('ss-auth-em'); if (em) em.addEventListener('keydown', function (e) { if (e.key === 'Enter') sendLink(); });
    var pw = $('ss-auth-pw'); if (pw) pw.addEventListener('keydown', function (e) { if (e.key === 'Enter') unlock(); });
  }

  // ── boot ─────────────────────────────────────────────────────────────
  async function boot() {
    var parts = window.location.pathname.replace(/\/+$/, '').split('/').filter(Boolean);
    var idx = parts.indexOf('stylesheet');
    if (idx < 2) {
      banner('This page needs to be opened from your book: <code>/your-account/your-book/stylesheet/</code>.');
      return;
    }
    S.clientSlug = parts[idx - 2];
    S.projectSlug = parts[idx - 1];
    $('ss-home').href = '/' + S.clientSlug + '/';

    var info;
    try {
      info = await api('/api/project-by-path/' + encodeURIComponent(S.clientSlug) + '/' + encodeURIComponent(S.projectSlug));
    } catch (e) {
      if (e.status === 401) { showAuth(boot); return; }
      if (e.status === 404) {
        banner('We don’t have a book at this address. Check the link in your Factory Pass email, or email <a href="mailto:' +
          esc(S.contactEmail) + '">' + esc(S.contactEmail) + '</a>.');
        return;
      }
      banner('Couldn’t load this page: ' + esc(e.message) + '. Try reloading.');
      return;
    }

    var p = info && info.project;
    S.projectId = p && (p.ID != null ? p.ID : p.id);
    if (!S.projectId) { banner('Couldn’t work out which book this page is for.'); return; }
    if (info.has_auth && !info.authenticated) { showAuth(boot); return; }

    try {
      S.sheet = await api(base());
      banner('');
      render();
    } catch (e) { handle(e); }
  }

  fetch('/api/public/config')
    .then(function (r) { return r.ok ? r.json() : null; })
    .then(function (c) { if (c && c.contact_email) S.contactEmail = c.contact_email; })
    .catch(function () {});

  wireAuth();
  wireToolbar();
  boot();
})();
