/* PI Editorial Stylesheet — working review tool (vanilla JS, no deps).
   Shared by the editor SPA (index.html) and the authors' view (authors.html).
   The page sets window.SS_MODE = 'editor' | 'authors' before loading this. */
(function () {
  'use strict';

  var MODE = window.SS_MODE || 'editor';
  var DEMO = /[?&]demo=1\b/.test(location.search);

  // ── tiny DOM helpers (mirrors transmittal.js conventions) ───────────
  var $ = function (s, p) { return (p || document).querySelector(s); };
  function h(tag, attrs) {
    var el = document.createElement(tag);
    if (attrs) Object.keys(attrs).forEach(function (k) {
      var v = attrs[k];
      if (v == null || v === false) return;
      if (k.indexOf('on') === 0) el.addEventListener(k.slice(2).toLowerCase(), v);
      else if (k === 'className') el.className = v;
      else if (k === 'html') el.innerHTML = v;
      else el.setAttribute(k, v);
    });
    for (var i = 2; i < arguments.length; i++) append(el, arguments[i]);
    return el;
  }
  function append(el, c) {
    if (c == null || c === false) return;
    if (Array.isArray(c)) { c.forEach(function (x) { append(el, x); }); return; }
    el.appendChild(c instanceof Node ? c : document.createTextNode(String(c)));
  }

  // ── API wrapper ───────────────────────────────────────────
  function api(url, opts) {
    opts = opts || {};
    return fetch(url, {
      headers: Object.assign({ 'Content-Type': 'application/json' }, opts.headers || {}),
      method: opts.method || 'GET',
      body: opts.body ? JSON.stringify(opts.body) : undefined,
    }).then(function (r) {
      return r.text().then(function (txt) {
        var data = null;
        try { data = txt ? JSON.parse(txt) : null; } catch (e) { data = null; }
        if (!r.ok) {
          var err = new Error((data && data.error) || r.statusText || ('HTTP ' + r.status));
          err.status = r.status;
          throw err;
        }
        return data;
      });
    });
  }

  // ── State ──────────────────────────────────────────────
  var state = {
    items: [],
    me: { email: '', name: '', can_write: false },
    filter: 'all',        // all | proposed | accepted | rejected | author
    scopeFilter: 'all',   // all | universal | zoothesia
    editingId: null,      // item id currently being inline-edited
    loadError: null,
  };

  // ──────────────────────────────────────────────────
  // Inline markdown → HTML (minimal, safe). Handles **bold**, *italic*,
  // `code`. Everything else (curly quotes, em dashes) passes through verbatim.
  // ──────────────────────────────────────────────────
  function esc(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }
  function inlineMd(s) {
    // operate on escaped text so raw HTML can't be injected
    var t = esc(s);
    // `code` first so its contents aren't touched by other rules
    var codes = [];
    t = t.replace(/`([^`]+)`/g, function (_, c) {
      codes.push(c); return '\u0000' + (codes.length - 1) + '\u0000';
    });
    t = t.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    t = t.replace(/(^|[^*])\*([^*\n]+)\*/g, '$1<em>$2</em>');
    t = t.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, function (_, txt, url) {
      return '<a href="' + esc(url) + '" target="_blank" rel="noopener">' + txt + '</a>';
    });
    t = t.replace(/\u0000(\d+)\u0000/g, function (_, i) {
      return '<code>' + codes[+i] + '</code>';
    });
    return t;
  }

  // Lightweight block markdown → HTML for prose bodies.
  // Supports: #/##/### headings, - and * bullets, 1. ordered lists,
  // > blockquotes, blank-line paragraphs, plus the inline rules above.
  function blockMd(src) {
    var lines = String(src == null ? '' : src).replace(/\r\n?/g, '\n').split('\n');
    var out = [], i = 0;
    function flushList(tag, items) {
      out.push('<' + tag + '>' + items.map(function (x) {
        return '<li>' + inlineMd(x) + '</li>';
      }).join('') + '</' + tag + '>');
    }
    while (i < lines.length) {
      var line = lines[i];
      if (/^\s*$/.test(line)) { i++; continue; }
      var hd = /^(#{1,4})\s+(.*)$/.exec(line);
      if (hd) {
        var lvl = Math.min(hd[1].length + 1, 4); // never emit h1 inside a card
        out.push('<h' + lvl + '>' + inlineMd(hd[2]) + '</h' + lvl + '>');
        i++; continue;
      }
      if (/^\s*>/.test(line)) {
        var q = [];
        while (i < lines.length && /^\s*>/.test(lines[i])) {
          q.push(lines[i].replace(/^\s*>\s?/, '')); i++;
        }
        out.push('<blockquote>' + inlineMd(q.join(' ')) + '</blockquote>');
        continue;
      }
      if (/^\s*[-*+]\s+/.test(line)) {
        var ul = [];
        while (i < lines.length && /^\s*[-*+]\s+/.test(lines[i])) {
          ul.push(lines[i].replace(/^\s*[-*+]\s+/, '')); i++;
        }
        flushList('ul', ul); continue;
      }
      if (/^\s*\d+[.)]\s+/.test(line)) {
        var ol = [];
        while (i < lines.length && /^\s*\d+[.)]\s+/.test(lines[i])) {
          ol.push(lines[i].replace(/^\s*\d+[.)]\s+/, '')); i++;
        }
        flushList('ol', ol); continue;
      }
      // paragraph: gather until blank line
      var para = [];
      while (i < lines.length && !/^\s*$/.test(lines[i]) &&
             !/^(#{1,4})\s+/.test(lines[i]) && !/^\s*>/.test(lines[i]) &&
             !/^\s*[-*+]\s+/.test(lines[i]) && !/^\s*\d+[.)]\s+/.test(lines[i])) {
        para.push(lines[i]); i++;
      }
      out.push('<p>' + inlineMd(para.join(' ')) + '</p>');
    }
    return out.join('\n');
  }

  // ──────────────────────────────────────────────────
  // Word-level diff via LCS. Returns HTML with <del.diff>/<ins.diff> spans.
  // Tokenise into words + whitespace so spacing is preserved and only
  // changed *words* are highlighted.
  // ──────────────────────────────────────────────────
  function tokenize(s) {
    // split keeping whitespace runs as their own tokens
    return String(s == null ? '' : s).split(/(\s+)/).filter(function (t) { return t !== ''; });
  }
  function lcsDiff(aStr, bStr) {
    var a = tokenize(aStr), b = tokenize(bStr);
    var n = a.length, m = b.length;
    // DP table of LCS lengths (word equality; whitespace compared literally)
    var dp = [];
    for (var i = 0; i <= n; i++) { dp.push(new Array(m + 1).fill(0)); }
    for (i = n - 1; i >= 0; i--) {
      for (var j = m - 1; j >= 0; j--) {
        dp[i][j] = a[i] === b[j] ? dp[i + 1][j + 1] + 1
          : Math.max(dp[i + 1][j], dp[i][j + 1]);
      }
    }
    var ops = []; i = 0; var k = 0;
    while (i < n && k < m) {
      if (a[i] === b[k]) { ops.push(['=', a[i]]); i++; k++; }
      else if (dp[i + 1][k] >= dp[i][k + 1]) { ops.push(['-', a[i]]); i++; }
      else { ops.push(['+', b[k]]); k++; }
    }
    while (i < n) { ops.push(['-', a[i]]); i++; }
    while (k < m) { ops.push(['+', b[k]]); k++; }
    return ops;
  }
  // Merge adjacent same-type ops, then render. Whitespace-only tokens that
  // are added/removed are emitted without a highlight box (cleaner look).
  function diffHtml(oldStr, newStr, side) {
    var ops = lcsDiff(oldStr, newStr);
    var want = side === 'old' ? '-' : '+';
    var buf = '', run = '', runType = '=';
    function flush() {
      if (run === '') return;
      if (runType === '=') buf += inlineMd(run);
      else if (/^\s+$/.test(run)) buf += run; // don't box pure whitespace
      else if (runType === '-') buf += '<del class="diff">' + inlineMd(run) + '</del>';
      else buf += '<ins class="diff">' + inlineMd(run) + '</ins>';
      run = '';
    }
    ops.forEach(function (op) {
      var type = op[0], tok = op[1];
      // show '=' on both sides; show only the side's changes
      var visible = type === '=' || type === want;
      if (!visible) return;
      if (type !== runType) { flush(); runType = type; }
      run += tok;
    });
    flush();
    return buf;
  }

  // ── Rendering ──────────────────────────────────────────
  function groupBySection(items) {
    var map = {}, order = [];
    items.forEach(function (it) {
      var key = it.section_ord + '\u0000' + it.section;
      if (!map[key]) { map[key] = { ord: it.section_ord, name: it.section, items: [] }; order.push(key); }
      map[key].items.push(it);
    });
    var groups = order.map(function (k) { return map[k]; });
    groups.sort(function (a, b) { return a.ord - b.ord; });
    groups.forEach(function (g) {
      g.items.sort(function (a, b) { return a.item_ord - b.item_ord; });
    });
    return groups;
  }

  function passesFilter(it) {
    if (state.scopeFilter !== 'all' && it.scope !== state.scopeFilter) return false;
    switch (state.filter) {
      case 'proposed': return it.status === 'proposed';
      case 'accepted': return it.status === 'accepted';
      case 'rejected': return it.status === 'rejected';
      case 'author': return it.author_facing;
      default: return true;
    }
  }

  function fmtWhen(s) {
    if (!s) return '';
    var d = new Date(s);
    if (isNaN(d)) return '';
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
  }

  // one field row inside a pending-edit diff block
  function diffRow(label, oldVal, newVal) {
    if ((oldVal || '') === (newVal || '')) return null;
    return h('div', { className: 'diff-row' },
      h('div', { className: 'dr-label' }, label),
      h('div', { className: 'diff-col' },
        h('span', { className: 'col-tag' }, 'current'),
        h('span', { html: diffHtml(oldVal || '', newVal || '', 'old') })),
      h('div', { className: 'diff-col' },
        h('span', { className: 'col-tag' }, 'proposed'),
        h('span', { html: diffHtml(oldVal || '', newVal || '', 'new') })));
  }

  function renderPending(it) {
    var e = it.pending_edit;
    var rows = [];
    if (it.kind === 'rule') {
      rows.push(diffRow('Item', it.col1, e.col1));
      rows.push(diffRow('Rule', it.col2, e.col2));
      rows.push(diffRow('Example', it.col3, e.col3));
    } else {
      rows.push(diffRow('Body', it.body, e.body));
    }
    rows = rows.filter(Boolean);
    if (!rows.length) rows.push(h('div', { className: 'empty-section' }, '(no textual change)'));

    var actions = state.me.can_write ? h('div', { className: 'pending-actions' },
      h('button', { className: 'link-action accent', onclick: function () { acceptEdit(e.id); } }, 'Accept edit'),
      h('button', { className: 'link-action danger', onclick: function () { rejectEdit(e.id); } }, 'Reject edit')
    ) : null;

    return h('div', { className: 'pending' },
      h('div', { className: 'pending-head' },
        h('span', {}, 'Pending edit'),
        h('span', { className: 'status-meta' },
          'proposed by ' + (e.proposed_by || 'someone') +
          (e.proposed_at ? ' · ' + fmtWhen(e.proposed_at) : ''))),
      e.note ? h('div', { className: 'pending-note' }, '“' + e.note + '”') : null,
      rows, actions);
  }

  function renderRuleBody(it) {
    return h('div', { className: 'rule-grid' },
      h('div', { className: 'rg-label' }, 'Item'),
      h('div', { className: 'rg-item', html: inlineMd(it.col1) }),
      h('div', { className: 'rg-label' }, 'Rule'),
      h('div', { className: 'rg-rule', html: inlineMd(it.col2) }),
      it.col3 ? h('div', { className: 'rg-label' }, 'Example') : null,
      it.col3 ? h('div', { className: 'rg-ex', html: inlineMd(it.col3) }) : null);
  }

  function renderEditor(it) {
    var isRule = it.kind === 'rule';
    var f = {};
    var fields = isRule
      ? [['col1', 'Item', false, it.col1], ['col2', 'Rule', true, it.col2], ['col3', 'Example', true, it.col3]]
      : [['body', 'Body (markdown)', true, it.body]];
    var inputs = fields.map(function (fd) {
      var el = fd[2]
        ? h('textarea', {}, fd[3] || '')
        : h('input', { type: 'text', value: fd[3] || '' });
      f[fd[0]] = el;
      return h('div', {}, h('label', {}, fd[1]), el);
    });
    var noteEl = h('input', { type: 'text', placeholder: 'why this change? (optional)' });
    inputs.push(h('div', {}, h('label', {}, 'Note'), noteEl));

    return h('div', { className: 'editor' }, inputs,
      h('div', { className: 'editor-actions' },
        h('button', {
          className: 'btn-fill', onclick: function () {
            var body = {
              col1: isRule ? f.col1.value : it.col1,
              col2: isRule ? f.col2.value : it.col2,
              col3: isRule ? f.col3.value : it.col3,
              body: isRule ? it.body : f.body.value,
              note: noteEl.value,
            };
            saveEdit(it.id, body);
          }
        }, 'Save proposed edit'),
        h('button', { className: 'link-action', onclick: function () { state.editingId = null; render(); } }, 'Cancel')));
  }

  // scope control: two-way ledger toggle (universal | zoothesia)
  function renderScope(it) {
    var canWrite = state.me.can_write;
    if (!canWrite) {
      return h('span', { className: 'scope-readonly', title: 'Scope' },
        it.scope === 'zoothesia' ? 'zoothesia' : 'universal');
    }
    function opt(val, label) {
      return h('button', {
        className: 'scope-btn' + (it.scope === val ? ' active' : ''),
        title: 'Scope: ' + label,
        'aria-pressed': it.scope === val ? 'true' : 'false',
        onclick: it.scope === val ? null : function () { setScope(it, val); },
      }, label);
    }
    return h('span', { className: 'scope-ctl', title: 'Scope' },
      opt('universal', 'universal'),
      h('span', { className: 'scope-sep' }, '·'),
      opt('zoothesia', 'zoothesia'));
  }

  function renderCard(it) {
    var canWrite = state.me.can_write;
    var head = h('div', { className: 'item-head' },
      h('span', { className: 'tag kind brk' }, it.kind),
      h('span', { className: 'tag status-' + it.status + ' brk' }, it.status),
      renderScope(it),
      it.status_by ? h('span', { className: 'status-meta' },
        it.status + ' by ' + it.status_by + (it.status_at ? ' · ' + fmtWhen(it.status_at) : '')) : null,
      h('span', { className: 'grow' }),
      h('button', {
        className: 'star-btn' + (it.author_facing ? ' on' : '') + (canWrite ? '' : ' readonly'),
        title: it.author_facing ? 'Author-facing (click to remove)' : 'Mark author-facing',
        disabled: !canWrite,
        'aria-pressed': it.author_facing ? 'true' : 'false',
        onclick: canWrite ? function () { toggleAuthor(it); } : null,
      }, it.author_facing ? '★' : '☆'));

    var bodyEl = it.kind === 'rule' ? renderRuleBody(it)
      : h('div', { className: 'prose', html: blockMd(it.body) });

    var kids = [head, bodyEl];

    if (it.pending_edit) kids.push(renderPending(it));

    if (state.editingId === it.id) {
      kids.push(renderEditor(it));
    } else if (canWrite) {
      kids.push(h('div', { className: 'actions' },
        h('button', { className: 'link-action accent', disabled: it.status === 'accepted',
          onclick: it.status === 'accepted' ? null : function () { setStatus(it, 'accepted'); } }, 'Accept'),
        h('button', { className: 'link-action danger', disabled: it.status === 'rejected',
          onclick: it.status === 'rejected' ? null : function () { setStatus(it, 'rejected'); } }, 'Reject'),
        it.status !== 'proposed' ? h('button', { className: 'link-action',
          onclick: function () { setStatus(it, 'proposed'); } }, 'Reset') : null,
        h('button', { className: 'link-action', onclick: function () { state.editingId = it.id; render(); } },
          'Edit')));
    }

    return h('div', {
      className: 'item status-' + it.status + (it.pending_edit ? ' has-pending' : ''),
    }, kids);
  }

  function renderList(container) {
    var visible = state.items.filter(passesFilter);
    if (!visible.length) {
      append(container, h('div', { className: 'notice' },
        h('h2', {}, 'Nothing here'),
        h('p', {}, 'No items match the “' + state.filter + '” filter.')));
      return;
    }
    groupBySection(visible).forEach(function (g) {
      var sec = h('section', { className: 'section' }, h('h2', {}, g.name));
      g.items.forEach(function (it) { append(sec, renderCard(it)); });
      append(container, sec);
    });
  }

  // ── Authors' view render (read-only, filtered) ───────────────────
  function renderAuthors(container) {
    var visible = state.items.filter(function (it) {
      return it.status === 'accepted' && it.author_facing;
    });
    if (!visible.length) {
      append(container, h('div', { className: 'notice' },
        h('h2', {}, 'Sheet is empty'),
        h('p', {}, 'No accepted, author-facing items yet.')));
      return;
    }
    groupBySection(visible).forEach(function (g) {
      var sec = h('section', { className: 'section' }, h('h2', {}, g.name));
      g.items.forEach(function (it) {
        var body = it.kind === 'rule' ? renderRuleBody(it)
          : h('div', { className: 'prose', html: blockMd(it.body) });
        append(sec, h('div', { className: 'item' }, body));
      });
      append(container, sec);
    });
  }

  // ── Top-level render ─────────────────────────────────────
  function render() {
    var main = $('#content');
    if (!main) return;
    main.innerHTML = '';

    if (state.loadError) {
      append(main, h('div', { className: 'notice error' },
        h('h2', {}, 'Backend not ready yet'),
        h('p', {}, state.loadError),
        h('p', { className: 'status-meta' },
          'Tip: append “?demo=1” to the URL to preview with sample data.')));
      return;
    }

    if (MODE === 'authors') { renderAuthors(main); return; }

    // editor: read-only note + list
    if (!state.me.can_write) {
      append(main, h('div', { className: 'readonly-note' },
        'Read-only. Sign in as an editor to make changes.'));
    }
    renderList(main);
  }

  // update just the sub-header identity + filter active states
  function renderChrome() {
    var who = $('#who');
    if (who) {
      if (state.me.name) {
        who.textContent = state.me.name + (state.me.can_write ? ' · editor' : ' · read-only');
      } else {
        who.textContent = 'not signed in';
      }
    }
    document.querySelectorAll('.filter-btn').forEach(function (b) {
      b.classList.toggle('active', b.dataset.filter === state.filter);
    });
    document.querySelectorAll('.scope-filter-btn').forEach(function (b) {
      b.classList.toggle('active', b.dataset.scope === state.scopeFilter);
    });
  }

  // ── Writes ─────────────────────────────────────────────
  function replaceItem(updated) {
    if (!updated || updated.id == null) return;
    for (var i = 0; i < state.items.length; i++) {
      if (state.items[i].id === updated.id) { state.items[i] = updated; break; }
    }
  }

  function handleWriteErr(e) {
    if (e.status === 401) toast('You are not signed in (401).', true);
    else if (e.status === 403) toast('You are not allowed to make changes (403).', true);
    else toast('Save failed: ' + e.message, true);
  }

  function patchItem(it, body, okMsg) {
    if (DEMO) { Object.assign(it, body); render(); toast(okMsg + ' (demo)'); return Promise.resolve(); }
    return api('/api/stylesheet/items/' + it.id, { method: 'PATCH', body: body })
      .then(function (updated) { replaceItem(updated); render(); toast(okMsg); })
      .catch(handleWriteErr);
  }

  function setStatus(it, status) { patchItem(it, { status: status }, 'Marked ' + status + '.'); }
  function setScope(it, scope) { patchItem(it, { scope: scope }, 'Scope → ' + scope + '.'); }
  function toggleAuthor(it) {
    patchItem(it, { author_facing: !it.author_facing },
      it.author_facing ? 'Removed from authors\u2019 sheet.' : 'Added to authors\u2019 sheet.');
  }

  function saveEdit(id, body) {
    if (DEMO) {
      var it = state.items.filter(function (x) { return x.id === id; })[0];
      if (it) it.pending_edit = {
        id: 9000 + id, item_id: id, col1: body.col1, col2: body.col2, col3: body.col3,
        body: body.body, note: body.note, proposed_by: state.me.name || 'You',
        proposed_at: new Date().toISOString(), state: 'pending',
      };
      state.editingId = null; render(); toast('Edit proposed (demo).'); return;
    }
    api('/api/stylesheet/items/' + id + '/edits', { method: 'POST', body: body })
      .then(function (updated) { replaceItem(updated); state.editingId = null; render(); toast('Edit proposed.'); })
      .catch(handleWriteErr);
  }

  function acceptEdit(editId) { resolveEdit(editId, 'accept'); }
  function rejectEdit(editId) { resolveEdit(editId, 'reject'); }
  function resolveEdit(editId, action) {
    if (DEMO) {
      state.items.forEach(function (it) {
        if (it.pending_edit && it.pending_edit.id === editId) {
          if (action === 'accept') {
            it.col1 = it.pending_edit.col1; it.col2 = it.pending_edit.col2;
            it.col3 = it.pending_edit.col3; it.body = it.pending_edit.body;
          }
          it.pending_edit = null;
        }
      });
      render(); toast('Edit ' + action + 'ed (demo).'); return;
    }
    api('/api/stylesheet/edits/' + editId + '/' + action, { method: 'POST' })
      .then(function (updated) { replaceItem(updated); render(); toast('Edit ' + action + 'ed.'); })
      .catch(handleWriteErr);
  }

  // ── Toast ─────────────────────────────────────────────
  var toastTimer = null;
  function toast(msg, isErr) {
    var t = $('#toast');
    if (!t) return;
    t.textContent = msg;
    t.className = 'toast show' + (isErr ? ' err' : '');
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { t.className = 'toast'; }, 2600);
  }

  // ── Theme (canonical theme.js owns font + dark state) ───────────
  function mountTheme() {
    var bar = $('#theme-bar');
    if (bar && window.JdbbTheme) JdbbTheme.mount(bar);
  }


  // ── Filters wiring (editor only) ─────────────────────────────
  function initFilters() {
    document.querySelectorAll('.filter-btn').forEach(function (b) {
      b.addEventListener('click', function () {
        state.filter = b.dataset.filter;
        renderChrome(); render();
      });
    });
    document.querySelectorAll('.scope-filter-btn').forEach(function (b) {
      b.addEventListener('click', function () {
        state.scopeFilter = b.dataset.scope;
        renderChrome(); render();
      });
    });
  }

  // ── Load ──────────────────────────────────────────────
  function boot() {
    mountTheme();
    if (MODE === 'editor') initFilters();

    if (DEMO) {
      state.items = SAMPLE_ITEMS.map(function (x) { return JSON.parse(JSON.stringify(x)); });
      state.me = { email: 'demo@local', name: 'Demo Editor', can_write: true };
      renderChrome(); render();
      toast('Demo mode — sample data, writes are local only.');
      return;
    }

    api('/api/stylesheet/items').then(function (data) {
      state.items = (data && data.items) || [];
      state.me = (data && data.me) || state.me;
      state.loadError = null;
      renderChrome(); render();
    }).catch(function (e) {
      if (e.status === 401) {
        state.loadError = 'You need to be signed in to view the stylesheet (401).';
      } else {
        state.loadError = 'Could not reach the API (' + (e.message || 'network error') +
          '). The backend may still be starting up.';
      }
      renderChrome(); render();
    });
  }

  // ── Sample data for ?demo=1 visual testing ──────────────────────
  var SAMPLE_ITEMS = [
    { id: 1, section_ord: 1, section: '1. House Style Basics', item_ord: 1, kind: 'rule',
      col1: 'Serial comma', col2: 'Use the **Oxford comma** in lists of three or more.',
      col3: 'red, white, and blue', status: 'accepted', scope: 'universal', author_facing: true,
      status_by: 'James Langdon (PI Editor)', status_at: '2026-07-10T12:00:00Z',
      updated_at: '2026-07-10T12:00:00Z', pending_edit: null },
    { id: 2, section_ord: 1, section: '1. House Style Basics', item_ord: 2, kind: 'rule',
      col1: 'Em dashes', col2: 'Use unspaced em dashes — like this — for breaks in thought.',
      col3: 'She paused—then ran.', status: 'proposed', scope: 'zoothesia', author_facing: false,
      status_by: '', status_at: '', updated_at: '2026-07-11T09:00:00Z',
      pending_edit: { id: 91, item_id: 2, col1: 'Em dashes',
        col2: 'Use unspaced em dashes—like this—for a sharp break in thought.',
        col3: 'She paused—then ran.', body: '', note: 'Prefer unspaced per Chicago.',
        proposed_by: 'JD (Publisher)', proposed_at: '2026-07-12T15:30:00Z', state: 'pending' } },
    { id: 3, section_ord: 1, section: '1. House Style Basics', item_ord: 3, kind: 'rule',
      col1: 'Numbers', col2: 'Spell out `zero` through `nine`; use numerals for 10+.',
      col3: 'three cats, 42 dogs', status: 'rejected', scope: 'universal', author_facing: false,
      status_by: 'JD (Publisher)', status_at: '2026-07-09T08:00:00Z',
      updated_at: '2026-07-09T08:00:00Z', pending_edit: null },
    { id: 4, section_ord: 2, section: '2. Voice & Tone', item_ord: 1, kind: 'prose',
      col1: '', col2: '', col3: '',
      body: '## Preserve the author\u2019s voice\n\nThis is a *light* copy edit regime — fix mechanics, don\u2019t rewrite. When in doubt, **leave the author\u2019s choice**.\n\n- Fix clear errors of grammar and spelling.\n- Query, don\u2019t change, stylistic risks.\n- Keep curly quotes and em dashes as-is.\n\n> When in doubt, leave it alone.',
      status: 'accepted', scope: 'universal', author_facing: true,
      status_by: 'James Langdon (PI Editor)', status_at: '2026-07-08T10:00:00Z',
      updated_at: '2026-07-08T10:00:00Z', pending_edit: null },
    { id: 5, section_ord: 2, section: '2. Voice & Tone', item_ord: 2, kind: 'prose',
      col1: '', col2: '', col3: '',
      body: 'Dialogue tags stay simple: `said` and `asked` do most of the work.',
      status: 'proposed', scope: 'zoothesia', author_facing: false, status_by: '', status_at: '',
      updated_at: '2026-07-13T11:00:00Z',
      pending_edit: { id: 92, item_id: 5, col1: '', col2: '', col3: '',
        body: 'Dialogue tags stay simple: `said` and `asked` carry most scenes. Reserve fancy tags for real emphasis.',
        note: 'Expanded slightly for clarity.', proposed_by: 'James Langdon (PI Editor)',
        proposed_at: '2026-07-14T09:15:00Z', state: 'pending' } },
  ];

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
  } else { boot(); }

  // expose for debugging
  window.SS = { state: state, diffHtml: diffHtml, blockMd: blockMd, inlineMd: inlineMd };
})();
