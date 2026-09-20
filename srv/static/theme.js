// theme.js — canonical site-font + dark-mode state for jdbb studio.
//
// Storage key `prodcal-theme-v1` is kept from the previous design so existing
// visitors keep their dark preference. The selector switches the full site
// voice so headings, labels, controls, and prose move together.
//
// Usage per page:
//   <script src="/static/theme.js"></script>
//   JdbbTheme.mount(document.getElementById('theme-bar'));
// The mount point receives the standard .theme-bar structure (see theme.css).
// Pages that render their own bar markup can call JdbbTheme.bind(el) instead.

(function () {
  'use strict';

  var FONTS = {
    jetbrains: 'JetBrains',
    martian: 'Martian',
    plex: 'Plex',
    geist: 'Geist',
    literata: 'Literata',
    'ibm-serif': 'Plex Serif',
    'source-serif': 'Source Serif',
    newsreader: 'Newsreader',
  };
  var KEYS = Object.keys(FONTS);
  var GROUPS = [
    { label: 'Mono/Sans', keys: ['jetbrains', 'martian', 'plex', 'geist'] },
    { label: 'Serif', keys: ['literata', 'ibm-serif', 'source-serif', 'newsreader'] },
  ];
  var LEGACY = { menlo: 'jetbrains', 'ibm-sans': 'geist' };
  var STORAGE = 'prodcal-theme-v1';        // localStorage: an explicit choice (sticky forever)
  var SESSION = 'prodcal-theme-session';   // sessionStorage: first-visit random pick (sticky per tab session)
  // First visits open in a random sans/mono face; the serifs are there for the
  // visitor to try, never auto-picked (decision 2026-09-11).
  var RANDOM_POOL = GROUPS[0].keys;

  // Appearance: 'system' follows the OS (default), 'light' / 'dark' are
  // explicit overrides. Older saves stored `dark: true|false`; those are read
  // as an explicit choice.
  var MODES = ['system', 'light', 'dark'];
  var MODE_GLYPH = { system: '◐', light: '☀', dark: '☾' };
  var MODE_LABEL = { system: 'Appearance: system (click for light)', light: 'Appearance: light (click for dark)', dark: 'Appearance: dark (click for system)' };
  var mq = window.matchMedia ? window.matchMedia('(prefers-color-scheme: dark)') : null;
  var state = { font: null, mode: 'system', chosen: false };
  try {
    var saved = JSON.parse(localStorage.getItem(STORAGE));
    if (saved) {
      if (MODES.indexOf(saved.mode) >= 0) state.mode = saved.mode;
      else if (typeof saved.dark === 'boolean') state.mode = saved.dark ? 'dark' : 'light';
      if (FONTS[saved.font]) { state.font = saved.font; state.chosen = true; }
      else if (LEGACY[saved.font]) { state.font = LEGACY[saved.font]; state.chosen = true; }
    }
  } catch (e) { /* first visit */ }
  if (!state.font) {
    try {
      var sess = sessionStorage.getItem(SESSION);
      if (FONTS[sess]) state.font = sess;
    } catch (e) {}
  }
  if (!state.font) {
    state.font = RANDOM_POOL[Math.floor(Math.random() * RANDOM_POOL.length)];
    try { sessionStorage.setItem(SESSION, state.font); } catch (e) {}
  }

  function isDark() {
    if (state.mode === 'system') return !!(mq && mq.matches);
    return state.mode === 'dark';
  }

  // save persists only what the visitor has actually chosen: the appearance
  // mode always, the font only once they've picked one from the selector.
  function save() {
    try {
      var out = { mode: state.mode, dark: isDark() };
      if (state.chosen) out.font = state.font;
      localStorage.setItem(STORAGE, JSON.stringify(out));
    } catch (e) {}
  }

  function apply(bar) {
    document.documentElement.setAttribute('data-font', state.font);
    document.documentElement.classList.toggle('dark', isDark());
    if (!bar) return;
    var nameEl = bar.querySelector('.font-name');
    if (nameEl) nameEl.textContent = FONTS[state.font];
    bar.querySelectorAll('.theme-opt[data-font]').forEach(function (b) {
      var active = b.dataset.font === state.font;
      b.classList.toggle('active', active);
      b.setAttribute('aria-pressed', active ? 'true' : 'false');
      var marker = b.querySelector('.theme-current');
      if (marker) marker.hidden = !active;
    });
    var darkBtn = bar.querySelector('.dark-btn');
    if (darkBtn) {
      darkBtn.textContent = MODE_GLYPH[state.mode];
      darkBtn.title = MODE_LABEL[state.mode];
      darkBtn.setAttribute('aria-label', MODE_LABEL[state.mode]);
      darkBtn.dataset.mode = state.mode;
    }
  }

  function bind(bar) {
    var nameEl = bar.querySelector('.font-name');
    var expanded = false;
    function setExpanded(next) {
      expanded = next;
      bar.classList.toggle('expanded', expanded);
      if (nameEl) nameEl.setAttribute('aria-expanded', expanded ? 'true' : 'false');
    }
    if (nameEl) {
      nameEl.addEventListener('click', function () {
        setExpanded(!expanded);
      });
      document.addEventListener('click', function (e) {
        if (expanded && !bar.contains(e.target)) {
          setExpanded(false);
        }
      });
      document.addEventListener('keydown', function (e) {
        if (expanded && e.key === 'Escape') setExpanded(false);
      });
    }
    bar.querySelectorAll('.theme-opt[data-font]').forEach(function (btn) {
      btn.addEventListener('click', function () {
        state.font = this.dataset.font;
        state.chosen = true;
        setExpanded(false);
        apply(bar); save();
      });
    });
    var darkBtn = bar.querySelector('.dark-btn');
    if (darkBtn) {
      darkBtn.addEventListener('click', function () {
        state.mode = MODES[(MODES.indexOf(state.mode) + 1) % MODES.length];
        apply(bar); save();
      });
    }
    if (mq) {
      var onChange = function () { if (state.mode === 'system') apply(bar); };
      if (mq.addEventListener) mq.addEventListener('change', onChange);
      else if (mq.addListener) mq.addListener(onChange);
    }
    apply(bar);
  }

  function mount(el) {
    if (!el) return;
    el.classList.add('theme-bar');
    el.innerHTML =
      '<button type="button" class="font-name" aria-haspopup="true" aria-expanded="false"></button>' +
      '<div class="font-options" role="menu" aria-label="Site font">' +
      GROUPS.map(function (group) {
        return '<div class="theme-opt-group">' +
          '<div class="theme-opt-label">' + group.label + '</div>' +
          group.keys.map(function (k) {
            return '<button type="button" class="theme-opt" data-font="' + k + '" aria-pressed="false">' +
              '<span class="theme-opt-main">' +
                '<span class="theme-opt-name">' + FONTS[k] + '</span>' +
                '<span class="theme-current" hidden>[current]</span>' +
              '</span>' +
              '<span class="theme-sample">A quiet line of proof text</span>' +
            '</button>';
          }).join('') +
        '</div>';
      }).join('') +
      '</div>' +
      '<span class="theme-sep"></span>' +
      '<button type="button" class="dark-btn" title="Appearance"></button>';
    bind(el);
  }

  // Apply immediately (pre-mount) to avoid a flash of default type/theme.
  apply(null);

  // Auto-mount: any page with <div id="theme-bar"> gets the standard bar
  // without page-local glue (pages may still call mount/bind explicitly).
  function autoMount() {
    var el = document.getElementById('theme-bar');
    if (el && !el.classList.contains('theme-bar')) mount(el);
  }
  // Admin nav: one list, every admin page. A page opts in with
  // <nav data-admin-nav> inside .jdbb-masthead; any <a> children it ships as a
  // no-JS fallback are replaced, other children (alerts, theme bar) are kept.
  // Add a page here and it appears everywhere at once.
  var ADMIN_NAV = [
    ['/admin/', 'Admin', 'Projects, typesetting, files, archived, pages, mail'],
    ['/admin/registrations', 'Cohorts', 'Workshop registrations and list email'],
    ['/admin/store/', 'Store', 'Factory Pass: new pass, sales, grant, revoke'],
    ['/admin/factory/', 'Floor', 'Live activity: who is uploading, inspecting, building right now'],
    ['/admin/docs/', 'Docs', 'Edit the public pages and email sign-off'],
    ['/admin/email-preview/', 'Emails', 'Every outbound email template with fixture data'],
    ['/admin/content-review/', 'Review', 'Content review'],
    ['/admin/runs/', 'Runs', 'Archive of punch lists and run logs, read-only'],
    ['/admin/#pages', 'Pages', 'Every route we have spun up, and what we mean to do with it'],
    ['/2026-pi-symposium', 'Roster', 'Cohort roster (what attendees see)'],
    ['/2026-pi-symposium/map', 'Map', 'Factory map: who acts at each stage, and where']
  ];
  function adminNav() {
    var nav = document.querySelector('.jdbb-masthead nav[data-admin-nav]');
    if (!nav || nav.getAttribute('data-admin-nav') === 'done') return;
    Array.prototype.slice.call(nav.querySelectorAll(':scope > a')).forEach(function (a) { nav.removeChild(a); });
    var here = location.pathname.replace(/\/+$/, '') || '/';
    var first = nav.firstChild;
    ADMIN_NAV.forEach(function (item) {
      var a = document.createElement('a');
      a.href = item[0]; a.textContent = item[1]; a.title = item[2];
      var path = item[0].split('#')[0].replace(/\/+$/, '') || '/';
      if (path === here && item[0].indexOf('#') < 0) a.setAttribute('aria-current', 'page');
      nav.insertBefore(a, first);
    });
    nav.setAttribute('data-admin-nav', 'done');
  }
  // Client nav: the customer pages under /{client}/[{project}/factory/]
  // share one strip — Your books · Factory — derived from the URL, current
  // page marked. Opt in with <nav data-client-nav>. (The transmittal is
  // section 1 of the factory page since 0.17 C; no separate entry.)
  function clientNav() {
    var nav = document.querySelector('.jdbb-masthead nav[data-client-nav]');
    if (!nav || nav.getAttribute('data-client-nav') === 'done') return;
    var seg = location.pathname.split('/').filter(Boolean);
    if (!seg.length) return;
    var c = seg[0], p = seg[1] || '', page = seg[2] || '';
    var items = [['/' + c + '/', 'Your books', 'All your projects']];
    if (p) {
      var base = '/' + c + '/' + p + '/';
      items.push([base + 'factory/', 'Factory', 'Transmittal, upload, inspect, build, download']);
      items.push([base + 'stylesheet/', 'Style sheet', 'Your book\u2019s editorial style sheet: accept, edit or add rules']);
      // Calendar hidden from customers (Jenna, 2026-09-18: not going to use it);
      // the page itself still answers at /{client}/{project}/ for the admin.
      // items.push([base, 'Calendar', 'Schedule, tasks, budget']);
    }
    var here = location.pathname.replace(/\/+$/, '') + '/';
    var first = nav.firstChild;
    items.forEach(function (item) {
      var a = document.createElement('a');
      a.href = item[0]; a.textContent = item[1]; a.title = item[2];
      if (item[0] === here) a.setAttribute('aria-current', 'page');
      nav.insertBefore(a, first);
    });
    nav.setAttribute('data-client-nav', 'done');
  }
  // Public nav: every public-tier page (the docs in jdbbs-public, the cohort
  // roster, store thanks, house style) shares one strip. Opt in with
  // <nav data-public-nav>; hand-written <a> children are replaced, the theme
  // bar is kept. Vocabulary per PAGE-DESIGN-HOSTING-VISIBILITY §4.
  var PUBLIC_NAV = [
    ['/workshop', 'Workshop', 'Protocolize Your Book — the workshop'],
    ['/field-notes', 'Field notes', 'Notes from the studio'],
    ['/factory', 'Factory', 'Factory Pass: manuscript (.docx) → EPUB + print PDF'],
    ['/portal', 'Client portal', 'Sign in to your books']
  ];
  function fillNav(nav, items, attr) {
    Array.prototype.slice.call(nav.querySelectorAll(':scope > a')).forEach(function (a) { nav.removeChild(a); });
    var here = location.pathname.replace(/\/+$/, '') || '/';
    var first = nav.firstChild;
    items.forEach(function (item) {
      var a = document.createElement('a');
      a.href = item[0]; a.textContent = item[1]; a.title = item[2];
      var path = item[0].split('#')[0].replace(/\/+$/, '') || '/';
      if (path === here && item[0].indexOf('#') < 0) a.setAttribute('aria-current', 'page');
      nav.insertBefore(a, first);
    });
    nav.setAttribute(attr, 'done');
  }
  function publicNav() {
    var nav = document.querySelector('.jdbb-masthead nav[data-public-nav]');
    if (!nav || nav.getAttribute('data-public-nav') === 'done') return;
    fillNav(nav, PUBLIC_NAV, 'data-public-nav');
  }
  // Theme bar first (it is the last child of every nav; the nav fillers insert
  // links before it), then the three shared strips.
  function navs() { autoMount(); adminNav(); clientNav(); publicNav(); }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', navs);
  else navs();

  window.JdbbTheme = { mount: mount, bind: bind, state: state, isDark: isDark, apply: apply, save: save, adminNav: adminNav, clientNav: clientNav, publicNav: publicNav };
})();
