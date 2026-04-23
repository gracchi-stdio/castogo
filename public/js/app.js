// Podlog — Datastar signals and custom behaviors

// View Transitions: save state on navigation away, restore on arrival
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
