// Podlog — app entry point

// Import CSS (Vite resolves @import and bundles)
import '../css/app.css';

const SKIPPED_TRANSITION_MESSAGE = 'Transition was skipped';
const ADMIN_CONTENT_SELECTOR = '[data-admin-content]';

let adminNavigationController;

function hasAdminContent(doc = document) {
  return !!doc.querySelector(ADMIN_CONTENT_SELECTOR);
}

function isSkippedTransitionAbortError(error) {
  return error instanceof DOMException &&
    error.name === 'AbortError' &&
    typeof error.message === 'string' &&
    error.message.includes(SKIPPED_TRANSITION_MESSAGE);
}

function shouldHandleAdminNavigation(event, anchor) {
  if (!anchor || event.defaultPrevented) {
    return false;
  }

  if (anchor.target && anchor.target !== '_self') {
    return false;
  }

  if (anchor.hasAttribute('download') || anchor.getAttribute('rel') === 'external') {
    return false;
  }

  if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) {
    return false;
  }

  const url = new URL(anchor.href, window.location.href);
  if (url.origin !== window.location.origin) {
    return false;
  }

  if (url.pathname === window.location.pathname && url.search === window.location.search) {
    return false;
  }

  return url.pathname.startsWith('/admin') && hasAdminContent();
}

async function runWithOptionalViewTransition(callback) {
  if (!document.startViewTransition) {
    callback();
    return;
  }

  const transition = document.startViewTransition(() => {
    callback();
  });

  try {
    await transition.finished;
  } catch (error) {
    if (!isSkippedTransitionAbortError(error)) {
      throw error;
    }
  }
}

async function navigateAdmin(url, options = {}) {
  const { historyMode = 'push', scrollToTop = true } = options;

  if (adminNavigationController) {
    adminNavigationController.abort();
  }

  adminNavigationController = new AbortController();
  const { signal } = adminNavigationController;

  let response;
  try {
    response = await fetch(url, { signal });
  } catch (error) {
    if (signal.aborted || isSkippedTransitionAbortError(error)) {
      return;
    }
    throw error;
  }

  if (!response.ok) {
    window.location.assign(url);
    return;
  }

  const html = await response.text();
  if (signal.aborted) {
    return;
  }

  const parser = new DOMParser();
  const nextDocument = parser.parseFromString(html, 'text/html');
  const nextContent = nextDocument.querySelector(ADMIN_CONTENT_SELECTOR);
  const currentContent = document.querySelector(ADMIN_CONTENT_SELECTOR);

  if (!nextContent || !currentContent) {
    window.location.assign(url);
    return;
  }

  await runWithOptionalViewTransition(() => {
    currentContent.replaceWith(nextContent);
    if (nextDocument.title) {
      document.title = nextDocument.title;
    }
  });

  if (historyMode === 'push') {
    window.history.pushState({ adminNav: true }, '', url);
  } else if (historyMode === 'replace') {
    window.history.replaceState({ adminNav: true }, '', url);
  }

  if (scrollToTop) {
    window.scrollTo(0, 0);
  }
}

document.addEventListener('click', (event) => {
  const target = event.target;
  if (!(target instanceof Element)) {
    return;
  }

  const anchor = target.closest('a[href]');
  if (!(anchor instanceof HTMLAnchorElement)) {
    return;
  }

  if (!shouldHandleAdminNavigation(event, anchor)) {
    return;
  }

  event.preventDefault();
  void navigateAdmin(anchor.href).catch(() => {
    window.location.assign(anchor.href);
  });
});

// Expose for server-side SSE redirect (ExecuteScript)
window.navigateAdmin = navigateAdmin;

window.addEventListener('popstate', () => {
  if (!window.location.pathname.startsWith('/admin') || !hasAdminContent()) {
    return;
  }

  void navigateAdmin(window.location.href, { historyMode: 'none', scrollToTop: false }).catch(() => {
    window.location.reload();
  });
});
