// Podlog — app entry point

// Import CSS (Vite resolves @import and bundles)
import '../css/app.css';

// View Transitions: save/restore scroll position across navigations
window.addEventListener("pageswap", (event) => {
  if (event.viewTransition) {
    sessionStorage.setItem("scrollPosition", window.scrollY);
  }
});

window.addEventListener("pagereveal", (event) => {
  if (event.viewTransition) {
    const scrollY = sessionStorage.getItem("scrollPosition");
    if (scrollY) {
      window.scrollTo(0, parseInt(scrollY));
    }
  }
});
