// theme.js — canonical font-pairing + dark-mode state for jdbb studio.
//
// Storage key `prodcal-theme-v1` is kept from the previous design so existing
// visitors keep their dark preference. Legacy font keys (literata / ibm-serif /
// menlo / ibm-sans) migrate to the nearest new pairing once, preserving the
// "every visitor gets a slightly different type voice" behavior.
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
  };
  var KEYS = Object.keys(FONTS);
  var LEGACY = { literata: 'martian', 'ibm-serif': 'plex', menlo: 'jetbrains', 'ibm-sans': 'geist' };
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
      b.classList.toggle('active', b.dataset.font === state.font);
    });
    var darkBtn = bar.querySelector('.dark-btn');
    if (darkBtn) darkBtn.textContent = state.dark ? '☀' : '☾';
  }

  function bind(bar) {
    var nameEl = bar.querySelector('.font-name');
    var expanded = false;
    if (nameEl) {
      nameEl.addEventListener('click', function () {
        expanded = !expanded;
        bar.classList.toggle('expanded', expanded);
      });
      document.addEventListener('click', function (e) {
        if (expanded && !bar.contains(e.target)) {
          expanded = false;
          bar.classList.remove('expanded');
        }
      });
    }
    bar.querySelectorAll('.theme-opt[data-font]').forEach(function (btn) {
      btn.addEventListener('click', function () {
        state.font = this.dataset.font;
        expanded = false;
        bar.classList.remove('expanded');
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
      '<span class="font-name"></span>' +
      '<span class="font-options">' +
      KEYS.map(function (k) {
        return '<button type="button" class="theme-opt" data-font="' + k + '">' + FONTS[k] + '</button>';
      }).join('') +
      '</span>' +
      '<span class="theme-sep"></span>' +
      '<button type="button" class="dark-btn" title="Toggle dark mode"></button>';
    bind(el);
  }

  // Apply immediately (pre-mount) to avoid a flash of default type/theme.
  apply(null);

  window.JdbbTheme = { mount: mount, bind: bind, state: state, apply: apply, save: save };
})();
