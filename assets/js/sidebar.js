// ABOUTME: Sidebar category expand/collapse persistence
// ABOUTME: Uses localStorage to persist collapse state across sessions

(function() {
  var STORAGE_KEY = 'sidebar-category-expand-state';

  // Toggle category expansion (called from template hx-on:click)
  function toggleCategory(header) {
    var group = header.parentElement;
    var isCollapsed = group.classList.contains('collapsed');

    if (isCollapsed) {
      group.classList.remove('collapsed');
      header.setAttribute('aria-expanded', 'true');
    } else {
      group.classList.add('collapsed');
      header.setAttribute('aria-expanded', 'false');
    }

    // Persist to localStorage
    saveExpandState();
  }

  // Save expand state to localStorage
  function saveExpandState() {
    var state = {};
    document.querySelectorAll('.category-group').forEach(function(group) {
      var categoryName = group.dataset.category;
      state[categoryName] = !group.classList.contains('collapsed');
    });
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  }

  // Load expand state from localStorage
  // Per D-05: 默认全部展开，仅在有 saved data 时应用折叠
  function loadExpandState() {
    var saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      try {
        var state = JSON.parse(saved);
        document.querySelectorAll('.category-group').forEach(function(group) {
          var categoryName = group.dataset.category;
          if (state[categoryName] === false) {
            group.classList.add('collapsed');
            var header = group.querySelector('.category-header');
            if (header) {
              header.setAttribute('aria-expanded', 'false');
            }
          }
        });
      } catch (e) {
        // Ignore parse errors
      }
    }
  }

  // Initialize on page load
  document.addEventListener('DOMContentLoaded', loadExpandState);

  // Re-apply after HTMX swaps (sidebar refreshes)
  document.body.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail && evt.detail.target && evt.detail.target.id === 'blog-list') {
      loadExpandState();
    }
  });

  // Expose toggleCategory globally for hx-on:click handlers
  window.toggleCategory = toggleCategory;
})();