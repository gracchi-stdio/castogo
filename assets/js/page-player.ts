const MOUNTED = "data-player-mounted";
const SPEEDS = [1, 1.25, 1.5, 2];

const PLAY_ICON = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="6 3 20 12 6 21 6 3"/></svg>';
const PAUSE_ICON = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>';

function formatTime(seconds: number): string {
  if (!Number.isFinite(seconds)) return "0:00";
  const total = Math.max(0, Math.floor(seconds));
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, "0")}`;
}

export function initPagePlayer(): void {
  const dock = document.querySelector<HTMLElement>("[data-episode-dock]");
  if (!dock || dock.hasAttribute(MOUNTED)) return;
  dock.setAttribute(MOUNTED, "");

  const audios = Array.from(document.querySelectorAll<HTMLAudioElement>("[data-page-audio]"));
  const play = dock.querySelector<HTMLButtonElement>("[data-player-play]");
  const progress = dock.querySelector<HTMLInputElement>("[data-player-progress]");
  const title = dock.querySelector<HTMLElement>("[data-player-title]");
  const current = dock.querySelector<HTMLElement>("[data-player-current]");
  const duration = dock.querySelector<HTMLElement>("[data-player-duration]");
  const speed = dock.querySelector<HTMLButtonElement>("[data-player-speed]");
  const slideSlot = dock.querySelector<HTMLElement>("[data-player-slides]");
  if (!play || !progress || !title || !current || !duration || !speed || !slideSlot) return;
  const playerDock = dock;
  const playButton = play;
  const progressInput = progress;
  const titleElement = title;
  const currentElement = current;
  const durationElement = duration;
  const speedButton = speed;
  const slideRoots = Array.from(document.querySelectorAll<HTMLElement>("[data-slideshow]"));
  slideRoots.forEach((slideRoot) => {
    slideRoot.hidden = true;
    slideSlot.append(slideRoot);
  });

  let active: HTMLAudioElement | null = null;
  let speedIndex = 0;

  function setState(state: "closed" | "open" | "expanded" | "hidden"): void {
    playerDock.dataset.playerState = state;
  }

  function update(): void {
    if (!active) return;
    const ratio = active.duration ? (active.currentTime / active.duration) * 100 : 0;
    progressInput.value = String(ratio);
    currentElement.textContent = formatTime(active.currentTime);
    durationElement.textContent = formatTime(active.duration);
    progressInput.style.setProperty("--player-progress", `${ratio}%`);
    const playing = !active.paused && !active.ended;
    playButton.innerHTML = playing ? PAUSE_ICON : PLAY_ICON;
    playButton.setAttribute("aria-label", playing ? "Pause episode" : "Play episode");
  }

  function attach(audio: HTMLAudioElement): void {
    if (active && active !== audio) active.pause();
    active = audio;
    slideRoots.forEach((slideRoot) => {
      const isActive = slideRoot.dataset.slideshowAudio === audio.id;
      slideRoot.hidden = !isActive;
    });
    titleElement.textContent = audio.dataset.episodeTitle || "Episode";
    setState(playerDock.dataset.playerState === "hidden" ? "hidden" : "open");
    update();
  }

  audios.forEach((audio) => {
    ["play", "pause", "timeupdate", "loadedmetadata", "durationchange", "ended"].forEach((event) => {
      audio.addEventListener(event, () => {
        if (event === "play") attach(audio);
        if (active === audio) update();
      });
    });
    audio.addEventListener("click", () => attach(audio));
  });

  playButton.addEventListener("click", () => {
    if (!active) {
      const first = audios[0];
      if (!first) return;
      attach(first);
      void first.play();
      return;
    }
    if (active.paused) void active.play();
    else active.pause();
  });

  progressInput.addEventListener("input", () => {
    if (!active || !active.duration) return;
    active.currentTime = (Number(progressInput.value) / 100) * active.duration;
    update();
  });

  speedButton.addEventListener("click", () => {
    speedIndex = (speedIndex + 1) % SPEEDS.length;
    const rate = SPEEDS[speedIndex];
    audios.forEach((audio) => { audio.playbackRate = rate; });
    speedButton.textContent = `${rate}x`;
  });

  playerDock.querySelector("[data-player-expand]")?.addEventListener("click", () => {
    setState(playerDock.dataset.playerState === "expanded" ? "open" : "expanded");
  });
  playerDock.querySelector("[data-player-hide]")?.addEventListener("click", () => setState("hidden"));
  playerDock.querySelector("[data-player-show]")?.addEventListener("click", () => setState("open"));

  if (audios.length > 0) {
    titleElement.textContent = audios[0].dataset.episodeTitle || "Choose an episode";
    setState("open");
  }
}