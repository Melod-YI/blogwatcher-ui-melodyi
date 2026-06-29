// ABOUTME: Article card ID copy-to-clipboard handler
// ABOUTME: Exposes global copyArticleId for inline onclick in article cards

(function() {
  // Copy the article ID to clipboard and give brief visual feedback.
  // Bound via inline onclick="copyArticleId(this, event)" so it works for
  // cards loaded by htmx infinite scroll without re-binding.
  function copyArticleId(el, event) {
    if (event) {
      event.stopPropagation();
      event.preventDefault();
    }
    var id = el.dataset.id;
    if (!id) {
      return;
    }

    var feedbackEl = el;
    var originalText = el.textContent;
    var copied = false;

    function showFeedback() {
      feedbackEl.textContent = '已复制 ✓';
      feedbackEl.classList.add('copied');
      setTimeout(function() {
        feedbackEl.textContent = originalText;
        feedbackEl.classList.remove('copied');
      }, 1200);
    }

    // Fallback using a hidden textarea + execCommand for non-secure
    // contexts (e.g. plain HTTP) where the async Clipboard API is unavailable.
    function fallbackCopy() {
      var ta = document.createElement('textarea');
      ta.value = id;
      ta.setAttribute('readonly', '');
      ta.style.position = 'absolute';
      ta.style.left = '-9999px';
      document.body.appendChild(ta);
      ta.select();
      var ok = false;
      try {
        ok = document.execCommand('copy');
      } catch (e) {
        ok = false;
      }
      document.body.removeChild(ta);
      return ok;
    }

    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(id).then(function() {
        copied = true;
        showFeedback();
      }).catch(function() {
        if (fallbackCopy()) {
          copied = true;
          showFeedback();
        }
      });
    } else {
      if (fallbackCopy()) {
        copied = true;
        showFeedback();
      }
    }
  }

  // Expose globally for inline onclick handlers in templates.
  window.copyArticleId = copyArticleId;
})();
