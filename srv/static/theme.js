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
  var STORAGE = 'prodcal-theme-v1';

  var state = { font: KEYS[Math.floor(Math.random() * KEYS.length)], dark: false };
  try {
    var saved = JSON.parse(localStorage.getItem(STORAGE));
    if (saved) {
      state.dark = !!saved.dark;
      if (FONTS[saved.font]) state.font = saved.font;
      else if (LEGACY[saved.font]) state.font = LEGACY[saved.font];
    }
  } catch (e) { /* first visit */ }

  function save() {
    try { localStorage.setItem(STORAGE, JSON.stringify({ font: state.font, dark: state.dark })); } catch (e) {}
  }

  function apply(bar) {
    document.documentElement.setAttribute('data-font', state.font);
    document.documentElement.classList.toggle('dark', state.dark);
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
    if (darkBtn) darkBtn.textContent = state.dark ? '☀' : '☾';
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
        setExpanded(false);
        apply(bar); save();
      });
    });
    var darkBtn = bar.querySelector('.dark-btn');
    if (darkBtn) {
      darkBtn.addEventListener('click', function () {
        state.dark = !state.dark;
        apply(bar); save();
      });
    }
    apply(bar);
    save(); // persist first-visit random pick so the voice follows the visitor
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
      '<button type="button" class="dark-btn" title="Toggle dark mode"></button>';
    bind(el);
  }

  // Apply immediately (pre-mount) to avoid a flash of default type/theme.
  apply(null);

  window.JdbbTheme = { mount: mount, bind: bind, state: state, apply: apply, save: save };
})();
