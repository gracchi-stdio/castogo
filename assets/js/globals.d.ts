// Type declarations for the window.* bridges consumed by backend Datastar
// ExecuteScript calls (helpers.go:toastScript, admin_pages.go, admin_episodes.go)
// and data-on:* attributes in templates (episode_new_page.templ).
//
// These globals are assigned at runtime in app.ts / audio-metadata.ts.

import type { ToastOptions } from "./types";

declare global {
  interface Window {
    // app.ts — bridge from backend SSE `ExecuteScript("window.navigateAdmin('/admin/...')")`.
    navigateAdmin: (url: string) => void;

    // app.ts — bridge from backend helpers.go:toastScript to toastManager.push.
    pushToast: (opts: ToastOptions) => void;

    // audio-metadata.ts — called from episode_new_page.templ data-on:change.
    extractAudioMetadata: (el: HTMLInputElement) => void;

    // app.ts — bust Swup's cached editor page after block mutations (add/delete/reorder).
    // The Blocks tab mutates via SSE patches, not navigation, so a cached /edit/blocks
    // snapshot would otherwise resurface stale on the next Settings ↔ Blocks switch.
    bustBlocksCache: () => void;

    // Safari-prefixed AudioContext. Optional — only present on older Safari.
    webkitAudioContext?: typeof AudioContext;
  }
}

export {};
