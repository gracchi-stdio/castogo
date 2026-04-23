// Podlog — app entry point

// Import CSS (Vite resolves @import and bundles)
import '../css/app.css';

// Shoelace — import components from /dist/, Vite resolves bare imports (lit, etc.)
import '@shoelace-style/shoelace/dist/components/alert/alert.js';
import '@shoelace-style/shoelace/dist/components/avatar/avatar.js';
import '@shoelace-style/shoelace/dist/components/button/button.js';
import '@shoelace-style/shoelace/dist/components/card/card.js';
import '@shoelace-style/shoelace/dist/components/details/details.js';
import '@shoelace-style/shoelace/dist/components/divider/divider.js';
import '@shoelace-style/shoelace/dist/components/drawer/drawer.js';
import '@shoelace-style/shoelace/dist/components/dropdown/dropdown.js';
import '@shoelace-style/shoelace/dist/components/icon/icon.js';
import '@shoelace-style/shoelace/dist/components/icon-button/icon-button.js';
import '@shoelace-style/shoelace/dist/components/input/input.js';
import '@shoelace-style/shoelace/dist/components/menu/menu.js';
import '@shoelace-style/shoelace/dist/components/menu-item/menu-item.js';
import '@shoelace-style/shoelace/dist/components/menu-label/menu-label.js';

// Set base path so Shoelace finds icons at /shoelace/dist/assets/icons/
import { setBasePath } from '@shoelace-style/shoelace/dist/utilities/base-path.js';
setBasePath('/shoelace/dist');

// View Transitions: save/restore state across navigations
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
