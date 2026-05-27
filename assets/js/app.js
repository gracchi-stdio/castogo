// Podlog — app entry point
import "./audio-metadata.js";
import { initBlockSorter } from "./blocks.js";
import { initEpisodePlayers, destroyActivePlayer } from "./episode-player.js";

// Import CSS (Vite resolves @import and bundles)
import "../css/app.css";
import "@fontsource-variable/inter";
import "@fontsource-variable/playfair-display";
import "@fontsource/oswald";

const SKIPPED_TRANSITION_MESSAGE = "Transition was skipped";
const ADMIN_CONTENT_SELECTOR = "[data-admin-content]";
const PUBLIC_CONTENT_SELECTOR = "[data-public-content]";
const RESERVED_PATHS = ["/admin", "/login", "/logout", "/register", "/feed", "/healthcheck", "/assets"];

let adminNavigationController;
let publicNavigationController;

function hasAdminContent(doc = document) {
  return !!doc.querySelector(ADMIN_CONTENT_SELECTOR);
}

function hasPublicContent(doc = document) {
  return !!doc.querySelector(PUBLIC_CONTENT_SELECTOR);
}

function isSkippedTransitionAbortError(error) {
  return (
    error instanceof DOMException &&
    error.name === "AbortError" &&
    typeof error.message === "string" &&
    error.message.includes(SKIPPED_TRANSITION_MESSAGE)
  );
}

function shouldHandleAdminNavigation(event, anchor) {
  if (!anchor || event.defaultPrevented) {
    return false;
  }

  if (anchor.target && anchor.target !== "_self") {
    return false;
  }

  if (
    anchor.hasAttribute("download") ||
    anchor.getAttribute("rel") === "external"
  ) {
    return false;
  }

  if (
    event.metaKey ||
    event.ctrlKey ||
    event.shiftKey ||
    event.altKey ||
    event.button !== 0
  ) {
    return false;
  }

  const url = new URL(anchor.href, window.location.href);
  if (url.origin !== window.location.origin) {
    return false;
  }

  if (
    url.pathname === window.location.pathname &&
    url.search === window.location.search
  ) {
    return false;
  }

  return url.pathname.startsWith("/admin") && hasAdminContent();
}

function shouldHandlePublicNavigation(event, anchor) {
  if (!anchor || event.defaultPrevented) {
    return false;
  }

  if (anchor.target && anchor.target !== "_self") {
    return false;
  }

  if (
    anchor.hasAttribute("download") ||
    anchor.getAttribute("rel") === "external"
  ) {
    return false;
  }

  if (
    event.metaKey ||
    event.ctrlKey ||
    event.shiftKey ||
    event.altKey ||
    event.button !== 0
  ) {
    return false;
  }

  const url = new URL(anchor.href, window.location.href);
  if (url.origin !== window.location.origin) {
    return false;
  }

  if (
    url.pathname === window.location.pathname &&
    url.search === window.location.search
  ) {
    return false;
  }

  const isReserved = RESERVED_PATHS.some(
    (p) => url.pathname === p || url.pathname.startsWith(p + "/")
  );
  if (isReserved) {
    return false;
  }

  return hasPublicContent();
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
  const { historyMode = "push", scrollToTop = true } = options;

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
  const nextDocument = parser.parseFromString(html, "text/html");
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

  updateActiveNavLink(new URL(url, window.location.href).pathname);

  if (historyMode === "push") {
    window.history.pushState({ adminNav: true }, "", url);
  } else if (historyMode === "replace") {
    window.history.replaceState({ adminNav: true }, "", url);
  }

  if (scrollToTop) {
    window.scrollTo(0, 0);
  }

  initBlockSorter();
}

async function navigatePublic(url, options = {}) {
  const { historyMode = "push", scrollToTop = true } = options;

  if (publicNavigationController) {
    publicNavigationController.abort();
  }

  destroyActivePlayer();

  publicNavigationController = new AbortController();
  const { signal } = publicNavigationController;

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
  const nextDocument = parser.parseFromString(html, "text/html");
  const nextContent = nextDocument.querySelector(PUBLIC_CONTENT_SELECTOR);
  const currentContent = document.querySelector(PUBLIC_CONTENT_SELECTOR);

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

  updateActivePublicNavLink(new URL(url, window.location.href).pathname);

  if (historyMode === "push") {
    window.history.pushState({ publicNav: true }, "", url);
  } else if (historyMode === "replace") {
    window.history.replaceState({ publicNav: true }, "", url);
  }

  if (scrollToTop) {
    window.scrollTo(0, 0);
  }

  initEpisodePlayers();
  initPublicNavScroll();
}

function updateActiveNavLink(pathname) {
  const links = document.querySelectorAll("[data-nav-link]");
  for (const link of links) {
    const isActive = link.getAttribute("href") === pathname;
    if (isActive) {
      link.classList.remove("sidebar-link-inactive");
      link.classList.add("sidebar-link-active");
    } else {
      link.classList.remove("sidebar-link-active");
      link.classList.add("sidebar-link-inactive");
    }
  }
}

function updateActivePublicNavLink(pathname) {
  const links = document.querySelectorAll("[data-public-nav-link]");
  for (const link of links) {
    const isActive = link.getAttribute("href") === pathname;
    if (isActive) {
      link.classList.remove("text-muted-foreground");
      link.classList.add("text-primary");
    } else {
      link.classList.remove("text-primary");
      link.classList.add("text-muted-foreground");
    }
  }
}

document.addEventListener("click", (event) => {
  const target = event.target;
  if (!(target instanceof Element)) {
    return;
  }

  const anchor = target.closest("a[href]");
  if (!(anchor instanceof HTMLAnchorElement)) {
    return;
  }

  if (shouldHandleAdminNavigation(event, anchor)) {
    event.preventDefault();
    void navigateAdmin(anchor.href).catch(() => {
      window.location.assign(anchor.href);
    });
    return;
  }

  if (shouldHandlePublicNavigation(event, anchor)) {
    event.preventDefault();
    void navigatePublic(anchor.href).catch(() => {
      window.location.assign(anchor.href);
    });
    return;
  }
});

// Suppress unhandled "Transition was skipped" rejections from the browser's
// internal view transition promises — harmless, happens on rapid navigation
window.addEventListener("unhandledrejection", (event) => {
  if (
    event.reason instanceof DOMException &&
    event.reason.name === "AbortError" &&
    typeof event.reason.message === "string" &&
    event.reason.message.includes("Transition was skipped")
  ) {
    event.preventDefault();
  }
});

// Expose for server-side SSE redirect (ExecuteScript)
window.navigateAdmin = navigateAdmin;
window.navigatePublic = navigatePublic;

// Init block sorter on initial page load
initBlockSorter();
initEpisodePlayers();
initPublicNavScroll();

window.addEventListener("popstate", () => {
  if (window.location.pathname.startsWith("/admin") && hasAdminContent()) {
    void navigateAdmin(window.location.href, {
      historyMode: "none",
      scrollToTop: false,
    }).catch(() => {
      window.location.reload();
    });
    return;
  }

  if (hasPublicContent()) {
    const isReserved = RESERVED_PATHS.some(
      (p) =>
        window.location.pathname === p ||
        window.location.pathname.startsWith(p + "/")
    );
    if (!isReserved) {
      void navigatePublic(window.location.href, {
        historyMode: "none",
        scrollToTop: false,
      }).catch(() => {
        window.location.reload();
      });
    }
  }
});

// Brutalist nav scroll collapse
function initPublicNavScroll() {
  const nav = document.querySelector(".public-nav");
  if (!nav) return;
  window.addEventListener("scroll", () => {
    if (window.scrollY > 80) {
      nav.setAttribute("data-scrolled", "");
    } else {
      nav.removeAttribute("data-scrolled");
    }
  }, { passive: true });
}

// Search form SPA interception
document.addEventListener("submit", (event) => {
  const form = event.target;
  if (!(form instanceof HTMLFormElement)) return;
  if (!form.matches("[data-public-search]")) return;
  if (!hasPublicContent()) return;

  event.preventDefault();
  const url = new URL(form.action, window.location.href);
  const formData = new FormData(form);
  for (const [key, value] of formData.entries()) {
    url.searchParams.set(key, value);
  }
  void navigatePublic(url.toString()).catch(() => {
    window.location.assign(url.toString());
  });
});
