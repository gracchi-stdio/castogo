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
const TOAST_STACK_ID = "toast-stack";
const DEFAULT_TOAST_TIMEOUT = 4000;
const ERROR_TOAST_TIMEOUT = 6000;
const TOAST_EXIT_DURATION = 160;

let adminNavigationController;
let publicNavigationController;

function getToastTimeout(variant, timeoutMs) {
  if (typeof timeoutMs === "number" && timeoutMs >= 0) {
    return timeoutMs;
  }

  if (variant === "error") {
    return ERROR_TOAST_TIMEOUT;
  }

  return DEFAULT_TOAST_TIMEOUT;
}

function getToastVariantClasses(variant) {
  if (variant === "error") {
    return "border-destructive bg-destructive text-destructive-foreground";
  }

  if (variant === "success") {
    return "border-emerald-300 bg-emerald-100 text-emerald-900 dark:border-emerald-700 dark:bg-emerald-900 dark:text-emerald-100";
  }

  return "border-border bg-card text-foreground";
}

function createCloseIcon() {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("viewBox", "0 0 24 24");
  svg.setAttribute("fill", "none");
  svg.setAttribute("stroke", "currentColor");
  svg.setAttribute("stroke-width", "2");
  svg.setAttribute("stroke-linecap", "round");
  svg.setAttribute("stroke-linejoin", "round");
  svg.setAttribute("class", "size-4");

  const line1 = document.createElementNS("http://www.w3.org/2000/svg", "line");
  line1.setAttribute("x1", "18");
  line1.setAttribute("y1", "6");
  line1.setAttribute("x2", "6");
  line1.setAttribute("y2", "18");

  const line2 = document.createElementNS("http://www.w3.org/2000/svg", "line");
  line2.setAttribute("x1", "6");
  line2.setAttribute("y1", "6");
  line2.setAttribute("x2", "18");
  line2.setAttribute("y2", "18");

  svg.appendChild(line1);
  svg.appendChild(line2);
  return svg;
}

function removeToast(toast) {
  if (!toast) {
    return;
  }

  if (toast.classList.contains("toast-exit")) {
    return;
  }

  toast.classList.remove("toast-enter");
  toast.classList.add("toast-exit");
  window.setTimeout(() => toast.remove(), TOAST_EXIT_DURATION);
}

function pushToast({ message, variant = "info", timeoutMs } = {}) {
  if (!message) {
    return;
  }

  const stack = document.getElementById(TOAST_STACK_ID);
  if (!stack) {
    return;
  }

  const toast = document.createElement("div");
  toast.className = `toast-enter rounded-lg border px-4 py-3 shadow-lg flex items-start gap-3 ${getToastVariantClasses(variant)}`;

  const text = document.createElement("div");
  text.className = "text-sm font-medium";
  text.textContent = message;

  const closeButton = document.createElement("button");
  closeButton.type = "button";
  closeButton.className = "opacity-70 hover:opacity-100 ml-auto";
  closeButton.setAttribute("aria-label", "Dismiss toast");
  closeButton.appendChild(createCloseIcon());
  closeButton.addEventListener("click", () => removeToast(toast));

  toast.appendChild(text);
  toast.appendChild(closeButton);
  if (typeof stack.prepend === "function") {
    stack.prepend(toast);
  } else if (stack.firstChild) {
    stack.insertBefore(toast, stack.firstChild);
  } else {
    stack.appendChild(toast);
  }

  const timeout = getToastTimeout(variant, timeoutMs);
  if (timeout > 0) {
    window.setTimeout(() => removeToast(toast), timeout);
  }
}

window.pushToast = pushToast;

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
