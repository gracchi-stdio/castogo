// Admin sidebar + public navbar active-link state, and public nav scroll-collapse.
// Pure DOM helpers — no Swup dependency. Called from app.ts on initial load
// and after each SPA navigation.

export function updateActiveNavLinks(pathname: string): void {
  const adminLinks = document.querySelectorAll("[data-nav-link]");
  for (const link of adminLinks) {
    const isActive = link.getAttribute("href") === pathname;
    link.classList.toggle("sidebar-link-active", isActive);
    link.classList.toggle("sidebar-link-inactive", !isActive);
  }
  const publicLinks = document.querySelectorAll("[data-public-nav-link]");
  for (const link of publicLinks) {
    const isActive = link.getAttribute("href") === pathname;
    link.classList.toggle("text-primary", isActive);
    link.classList.toggle("text-muted-foreground", !isActive);
    // Reflect active state for assistive tech + the CSS marker hook
    if (isActive) {
      link.setAttribute("aria-current", "page");
      link.setAttribute("data-active", "");
    } else {
      link.removeAttribute("aria-current");
      link.removeAttribute("data-active");
    }
  }
}

// Brutalist nav scroll collapse — toggles data-scrolled on the public nav
// when the user scrolls past 80px.
export function initPublicNavScroll(): void {
  const nav = document.querySelector(".public-nav");
  if (!nav) return;
  window.addEventListener(
    "scroll",
    () => {
      if (window.scrollY > 80) {
        nav.setAttribute("data-scrolled", "");
      } else {
        nav.removeAttribute("data-scrolled");
      }
    },
    { passive: true },
  );
}
