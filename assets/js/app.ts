// Podlog — app entry point
import "./audio-metadata";
import { initBlockSorter } from "./blocks";
import { initEpisodePlayers, destroyActivePlayer } from "./episode-player";
import { toastManager } from "./toast";
import { updateActiveNavLinks, initPublicNavScroll } from "./nav";

// Import CSS (Vite resolves @import and bundles)
import "../css/app.css";
import "@fontsource-variable/inter";
import "@fontsource-variable/playfair-display";
import "@fontsource/oswald";

import Swup from "swup";
import SwupPreloadPlugin from "@swup/preload-plugin";

// Paths that always full-reload. They change the app shell (login ↔ admin ↔ public)
// or are not HTML routes (feed, healthcheck, assets). Crossing the admin boundary
// also requires a full reload because the swap targets are different elements.
const RESERVED_PATHS = [
  "/login",
  "/logout",
  "/register",
  "/feed",
  "/healthcheck",
  "/assets",
] as const;

const SKIPPED_TRANSITION_MESSAGE = "Transition was skipped";

// Backend (helpers.go:toastScript) calls window.pushToast({message, variant}).
// Bridge to the class-based toastManager.
window.pushToast = (opts) => toastManager.push(opts);

function pathnameOf(url: string): string {
  return new URL(url, window.location.href).pathname;
}

function isReserved(pathname: string): boolean {
  return RESERVED_PATHS.some((p) => pathname === p || pathname.startsWith(p + "/"));
}

// Decide whether Swup should handle a given URL. Returning true = bypass Swup,
// falling back to a full browser navigation.
function ignoreVisit(url: string): boolean {
  const targetPath = pathnameOf(url);
  const currentPath = window.location.pathname;
  const isTargetAdmin = targetPath.startsWith("/admin");
  const isCurrentAdmin = currentPath.startsWith("/admin");

  // Cross-shell transitions (admin → public or vice versa) need a full reload
  // because the swap target element differs ([data-admin-content] vs [data-public-content]).
  if (isTargetAdmin !== isCurrentAdmin) {
    return true;
  }

  // Reserved paths always full-reload.
  if (isReserved(targetPath)) {
    return true;
  }

  return false;
}

const swup = new Swup({
  // Default containers — narrowed per-visit in visit:start so Swup never sees a
  // missing container. All listed so Swup's cache key is consistent.
  containers: [
    "[data-swap-title]",
    "[data-swap-actions]",
    "[data-swap-subnav]",
    "[data-admin-content]",
    "[data-public-content]",
  ],
  plugins: [new SwupPreloadPlugin()],
  ignoreVisit,
  // Use the browser's native View Transitions API for animations. Swup will wrap
  // its renderPage call in document.startViewTransition() automatically — the
  // existing ::view-transition-* CSS in app.css drives the animation.
  native: true,
  // We don't use Swup's CSS-class-based animation system (is-leaving, is-rendering,
  // transition-* classes). Setting this to false silences the "no CSS animation
  // duration defined" warning that would otherwise log on every navigation.
  animationSelector: false,
});

// Per-visit container selection: admin target swaps title + actions + content;
// public target swaps just [data-public-content]. Required because Swup 4 wants
// all containers to exist in both old and new documents. Cross-shell navigation
// is already filtered out by ignoreVisit.
swup.hooks.on("visit:start", (visit) => {
  // Tear down public-side players before the swap leaves the DOM.
  if (document.querySelector("[data-public-content]")) {
    destroyActivePlayer();
  }
  const targetPath = pathnameOf(visit.to.url);
  visit.containers = targetPath.startsWith("/admin")
    ? ["[data-swap-title]", "[data-swap-actions]", "[data-swap-subnav]", "[data-admin-content]"]
    : ["[data-public-content]"];
});

// content:replace fires inside Swup's startViewTransition callback (because
// native:true). Any DOM changes here are captured in the new snapshot — so this is
// the right place to update sidebar/nav active states. Updating them later (e.g. in
// page:view) would mutate the live DOM after the animation ends, causing a visible
// blink since the sidebar is a named view-transition element with animation: none.
swup.hooks.on("content:replace", (visit) => {
  updateActiveNavLinks(pathnameOf(visit.to.url));
  if (document.querySelector("[data-admin-content]")) {
    initBlockSorter();
  }
  if (document.querySelector("[data-public-content]")) {
    initEpisodePlayers();
    initPublicNavScroll();
  }
});

// Backward compat: handlers (admin_episodes.go, admin_pages.go) call this via
// SSE ExecuteScript. Delegates to Swup.
window.navigateAdmin = (url: string) => swup.navigate(url);

// Bust cached /edit/blocks snapshots after block mutations (add/delete/reorder).
// The Blocks tab mutates via SSE patches, not navigation, so Swup's cached page would
// otherwise go stale and resurface on the next Settings ↔ Blocks tab switch.
window.bustBlocksCache = () => {
  swup.cache.prune((url) => url.includes("/edit/blocks"));
};

// Suppress unhandled "Transition was skipped" rejections from the browser's
// internal view transition promises. The wrapSwapWithViewTransition handler
// catches most, but some race conditions still leak through.
window.addEventListener("unhandledrejection", (event) => {
  const reason = event.reason;
  if (
    reason?.name === "AbortError" &&
    typeof reason?.message === "string" &&
    reason.message.includes(SKIPPED_TRANSITION_MESSAGE)
  ) {
    event.preventDefault();
  }
});

// Initial page-load inits (Swup only fires hooks on subsequent navigations).
initBlockSorter();
initEpisodePlayers();
initPublicNavScroll();
